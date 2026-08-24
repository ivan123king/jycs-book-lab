// reading_service.go 负责阅读工作台的章节正文、阅读位置、选文来源和拆解卡片本地持久化。
package library

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	readingStateRelativePath        = "reading/reading-state.json"
	sourceTextsRelativePath         = "source-texts/source-texts.json"
	deconstructionCardsRelativePath = "deconstructions/deconstruction-cards.json"
	deconstructionPlanRelativePath  = "deconstruction-plan.json"
	deconstructionTemplatesDir      = "deconstruction-templates"
	systemTemplateFileName          = "系统内置模板.json"
	systemTemplateName              = "系统内置模板"
	defaultReaderFontSize           = 18
	minReaderFontSize               = 14
	maxReaderFontSize               = 28
)

// OpenReadingSession 读取当前书阅读页所需的正式项目数据，并为旧项目补齐缺失文件。
func OpenReadingSession(bookPath string) (ReadingSession, error) {
	project, err := OpenBookProject(bookPath)
	if err != nil {
		return ReadingSession{}, err
	}
	bookPath = projectRootPath(bookPath)
	if err := ensureReadingSupportFiles(bookPath, project); err != nil {
		return ReadingSession{}, err
	}

	state, err := loadReadingState(bookPath, project)
	if err != nil {
		return ReadingSession{}, err
	}
	currentChapter := chooseChapterForPosition(project, state.LastPosition.ChapterID)
	content, err := readChapterText(bookPath, currentChapter)
	if err != nil {
		return ReadingSession{}, err
	}
	plan, err := loadDeconstructionPlan(bookPath, project.Book.ID)
	if err != nil {
		return ReadingSession{}, err
	}
	sourceTexts, err := loadSourceTexts(bookPath)
	if err != nil {
		return ReadingSession{}, err
	}
	cards, err := loadDeconstructionCards(bookPath)
	if err != nil {
		return ReadingSession{}, err
	}

	state.LastPosition.ChapterID = currentChapter.ID
	state.LastPosition.ChapterOrder = currentChapter.Order
	return ReadingSession{
		Project:             project,
		State:               state,
		CurrentChapter:      currentChapter,
		ChapterContent:      content,
		Plan:                plan,
		SourceTexts:         sourceTexts,
		DeconstructionCards: cards,
	}, nil
}

// GetChapterContent 读取用户修正后的章节 TXT，供翻页和目录跳章使用。
func GetChapterContent(request ChapterContentRequest) (ChapterContentResult, error) {
	project, err := OpenBookProject(request.BookPath)
	if err != nil {
		return ChapterContentResult{}, err
	}
	chapter, ok := findChapterByID(project, request.ChapterID)
	if !ok {
		return ChapterContentResult{}, fmt.Errorf("未找到章节: %s", request.ChapterID)
	}
	content, err := readChapterText(projectRootPath(request.BookPath), chapter)
	if err != nil {
		return ChapterContentResult{}, err
	}
	return ChapterContentResult{Chapter: chapter, Content: content}, nil
}

// SaveReadingPosition 将阅读位置保存到单本小说项目目录，作为恢复阅读的正式来源。
func SaveReadingPosition(request SaveReadingPositionRequest) (ReadingState, error) {
	project, err := OpenBookProject(request.BookPath)
	if err != nil {
		return ReadingState{}, err
	}
	bookPath := projectRootPath(request.BookPath)
	chapter := chooseChapterForPosition(project, request.ChapterID)
	mode := strings.TrimSpace(request.LastMode)
	if mode != "edit" {
		mode = "read"
	}
	now := time.Now().Format(time.RFC3339)
	position := ReadingPosition{
		ChapterID:          chapter.ID,
		ChapterOrder:       chapter.Order,
		CharOffset:         maxInt(0, request.CharOffset),
		ParagraphIndex:     maxInt(0, request.ParagraphIndex),
		RenderedPageNumber: maxInt(1, request.RenderedPageNumber),
		UpdatedAt:          now,
	}
	progressPercent := clampPercent(request.ReadingProgressPercent)
	if progressPercent == 0 && (position.ChapterOrder > 1 || position.CharOffset > 0) {
		progressPercent = estimateReadingProgressPercent(project, position)
	}
	existingState, err := loadReadingState(bookPath, project)
	if err != nil {
		return ReadingState{}, err
	}
	state := ReadingState{
		Version:                1,
		BookID:                 project.Book.ID,
		LastPosition:           position,
		LastMode:               mode,
		ReadingProgressPercent: progressPercent,
		Preferences:            normalizeReadingPreferences(existingState.Preferences),
	}
	if err := writeJSONFileAtomic(filepath.Join(bookPath, filepath.FromSlash(readingStateRelativePath)), state); err != nil {
		return ReadingState{}, fmt.Errorf("保存阅读位置失败: %w", err)
	}
	return state, nil
}

// SaveReadingPreferences 将单本小说自己的阅读界面偏好保存到项目目录。
func SaveReadingPreferences(request SaveReadingPreferencesRequest) (ReadingState, error) {
	project, err := OpenBookProject(request.BookPath)
	if err != nil {
		return ReadingState{}, err
	}
	bookPath := projectRootPath(request.BookPath)
	if err := ensureReadingSupportFiles(bookPath, project); err != nil {
		return ReadingState{}, err
	}
	state, err := loadReadingState(bookPath, project)
	if err != nil {
		return ReadingState{}, err
	}
	state.Preferences = normalizeReadingPreferences(ReadingPreferences{FontSize: request.FontSize})
	if err := writeJSONFileAtomic(filepath.Join(bookPath, filepath.FromSlash(readingStateRelativePath)), state); err != nil {
		return ReadingState{}, fmt.Errorf("保存阅读偏好失败: %w", err)
	}
	return state, nil
}

// SaveChapterContent 保存编辑 TAB 中的章节标题和正文，并同步更新 book.json 章节元数据。
func SaveChapterContent(request SaveChapterContentRequest) (ChapterContentResult, error) {
	project, err := OpenBookProject(request.BookPath)
	if err != nil {
		return ChapterContentResult{}, err
	}
	bookPath := projectRootPath(request.BookPath)
	chapterIndex := -1
	for index := range project.Chapters {
		if project.Chapters[index].ID == request.ChapterID {
			chapterIndex = index
			break
		}
	}
	if chapterIndex < 0 {
		return ChapterContentResult{}, fmt.Errorf("未找到章节: %s", request.ChapterID)
	}
	title := strings.TrimSpace(request.Title)
	if title == "" {
		return ChapterContentResult{}, errors.New("章节标题不能为空")
	}

	chapter := project.Chapters[chapterIndex]
	chapter.Title = title
	chapter.WordCount = countTextRunes(request.Content)
	project.Chapters[chapterIndex] = chapter
	project.Book.UpdatedAt = time.Now().Format(time.RFC3339)

	contentPath, err := safeProjectPath(bookPath, chapter.ContentPath)
	if err != nil {
		return ChapterContentResult{}, err
	}
	if err := writeTextFileAtomic(contentPath, request.Content); err != nil {
		return ChapterContentResult{}, fmt.Errorf("保存章节正文失败: %w", err)
	}
	if err := writeJSONFileAtomic(filepath.Join(bookPath, bookJSONName), project); err != nil {
		return ChapterContentResult{}, fmt.Errorf("保存章节元数据失败: %w", err)
	}
	return ChapterContentResult{Chapter: chapter, Content: request.Content}, nil
}

// SplitChapterFromStart 从当前章节开头拆出一个新章节，当前章节保留剩余正文。
func SplitChapterFromStart(request SplitChapterRequest) (SplitChapterResult, error) {
	project, err := OpenBookProject(request.BookPath)
	if err != nil {
		return SplitChapterResult{}, err
	}
	bookPath := projectRootPath(request.BookPath)
	if err := ensureReadingSupportFiles(bookPath, project); err != nil {
		return SplitChapterResult{}, err
	}

	chapterIndex := -1
	for index := range project.Chapters {
		if project.Chapters[index].ID == request.ChapterID {
			chapterIndex = index
			break
		}
	}
	if chapterIndex < 0 {
		return SplitChapterResult{}, fmt.Errorf("未找到章节: %s", request.ChapterID)
	}
	newTitle := strings.TrimSpace(request.NewTitle)
	if newTitle == "" {
		return SplitChapterResult{}, errors.New("新章节标题不能为空")
	}
	if strings.TrimSpace(request.NewContent) == "" {
		return SplitChapterResult{}, errors.New("新章节正文不能为空")
	}
	if strings.TrimSpace(request.RemainingContent) == "" {
		return SplitChapterResult{}, errors.New("当前章节剩余正文不能为空")
	}
	if err := ensureChapterHasNoSourceTexts(bookPath, request.ChapterID); err != nil {
		return SplitChapterResult{}, err
	}

	currentChapter := project.Chapters[chapterIndex]
	newChapterID := nextChapterID(project)
	newChapter := BookChapter{
		ID:          newChapterID,
		Title:       newTitle,
		Order:       currentChapter.Order,
		ContentPath: fmt.Sprintf("chapters/%s.txt", newChapterID),
		StartOffset: currentChapter.StartOffset,
		EndOffset:   currentChapter.StartOffset + len(request.NewContent),
		WordCount:   countTextRunes(request.NewContent),
	}
	remainingChapter := currentChapter
	remainingChapter.Order = currentChapter.Order + 1
	remainingChapter.StartOffset = newChapter.EndOffset
	remainingChapter.EndOffset = newChapter.EndOffset + len(request.RemainingContent)
	remainingChapter.WordCount = countTextRunes(request.RemainingContent)

	nextChapters := make([]BookChapter, 0, len(project.Chapters)+1)
	for index, chapter := range project.Chapters {
		if index == chapterIndex {
			nextChapters = append(nextChapters, newChapter, remainingChapter)
			continue
		}
		if chapter.Order > currentChapter.Order {
			chapter.Order++
		}
		nextChapters = append(nextChapters, chapter)
	}
	project.Chapters = nextChapters
	project.Book.UpdatedAt = time.Now().Format(time.RFC3339)

	newContentPath, err := safeProjectPath(bookPath, newChapter.ContentPath)
	if err != nil {
		return SplitChapterResult{}, err
	}
	remainingContentPath, err := safeProjectPath(bookPath, remainingChapter.ContentPath)
	if err != nil {
		return SplitChapterResult{}, err
	}
	if err := writeTextFileAtomic(newContentPath, request.NewContent); err != nil {
		return SplitChapterResult{}, fmt.Errorf("保存新章节正文失败: %w", err)
	}
	if err := writeTextFileAtomic(remainingContentPath, request.RemainingContent); err != nil {
		return SplitChapterResult{}, fmt.Errorf("保存当前章节剩余正文失败: %w", err)
	}
	if err := writeJSONFileAtomic(filepath.Join(bookPath, bookJSONName), project); err != nil {
		return SplitChapterResult{}, fmt.Errorf("保存拆分后的章节目录失败: %w", err)
	}
	if err := syncPersistedChapterOrdersAfterSplit(bookPath, currentChapter.Order, currentChapter.ID, remainingChapter); err != nil {
		return SplitChapterResult{}, err
	}

	return SplitChapterResult{
		Project:          project,
		NewChapter:       newChapter,
		RemainingChapter: remainingChapter,
		NewContent:       request.NewContent,
		RemainingContent: request.RemainingContent,
	}, nil
}

// CreateDeconstructionCard 从当前选文创建拆解卡片；同一来源片段允许被多张卡片引用。
func CreateDeconstructionCard(request CreateDeconstructionRequest) (CreateDeconstructionResult, error) {
	project, err := OpenBookProject(request.BookPath)
	if err != nil {
		return CreateDeconstructionResult{}, err
	}
	bookPath := projectRootPath(request.BookPath)
	if err := ensureReadingSupportFiles(bookPath, project); err != nil {
		return CreateDeconstructionResult{}, err
	}
	chapter, ok := findChapterByID(project, request.Selection.ChapterID)
	if !ok {
		return CreateDeconstructionResult{}, fmt.Errorf("未找到章节: %s", request.Selection.ChapterID)
	}
	if strings.TrimSpace(request.Selection.SelectedText) == "" {
		return CreateDeconstructionResult{}, errors.New("选文内容不能为空")
	}
	plan, err := loadDeconstructionPlan(bookPath, project.Book.ID)
	if err != nil {
		return CreateDeconstructionResult{}, err
	}
	planType, ok := findPlanType(plan, request.PlanTypeID)
	if !ok {
		return CreateDeconstructionResult{}, fmt.Errorf("拆解类型不存在: %s", request.PlanTypeID)
	}
	dimension, ok := findPlanDimension(planType, request.DimensionID)
	if !ok {
		return CreateDeconstructionResult{}, fmt.Errorf("拆解维度不存在: %s", request.DimensionID)
	}
	content := strings.TrimSpace(request.Content)
	if content == "" {
		return CreateDeconstructionResult{}, errors.New("拆解看法不能为空")
	}

	sourceTexts, err := loadSourceTexts(bookPath)
	if err != nil {
		return CreateDeconstructionResult{}, err
	}
	cards, err := loadDeconstructionCards(bookPath)
	if err != nil {
		return CreateDeconstructionResult{}, err
	}

	now := time.Now().Format(time.RFC3339)
	sourceText := findReusableSourceText(sourceTexts, project.Book.ID, request.Selection)
	if sourceText.ID == "" {
		sourceText = SourceText{
			ID:           nextSequentialID("source", len(sourceTexts.Items)+1),
			BookID:       project.Book.ID,
			ChapterID:    chapter.ID,
			ChapterOrder: chapter.Order,
			StartOffset:  maxInt(0, request.Selection.StartOffset),
			EndOffset:    maxInt(0, request.Selection.EndOffset),
			SelectedText: request.Selection.SelectedText,
			CreatedAt:    now,
		}
		sourceTexts.Items = append(sourceTexts.Items, sourceText)
		if err := writeJSONFileAtomic(filepath.Join(bookPath, filepath.FromSlash(sourceTextsRelativePath)), sourceTexts); err != nil {
			return CreateDeconstructionResult{}, fmt.Errorf("保存选文来源失败: %w", err)
		}
	}

	card := DeconstructionCard{
		ID:            nextSequentialID("deconstruction", len(cards.Items)+1),
		BookID:        project.Book.ID,
		SourceTextID:  sourceText.ID,
		PlanTypeID:    planType.ID,
		PlanTypeName:  planType.Name,
		DimensionID:   dimension.ID,
		DimensionName: dimension.Name,
		Content:       content,
		FieldValues:   normalizeFieldValues(request.FieldValues),
		TagIDs:        normalizeTags(request.TagIDs),
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	cards.Items = append(cards.Items, card)
	if err := writeJSONFileAtomic(filepath.Join(bookPath, filepath.FromSlash(deconstructionCardsRelativePath)), cards); err != nil {
		return CreateDeconstructionResult{}, fmt.Errorf("保存拆解卡片失败: %w", err)
	}

	return CreateDeconstructionResult{
		SourceText:          sourceText,
		Card:                card,
		SourceTexts:         sourceTexts,
		DeconstructionCards: cards,
	}, nil
}

// OpenDeconstructionPlanWorkspace 读取当前书的当前方案和可编辑模板列表。
func OpenDeconstructionPlanWorkspace(bookPath string) (DeconstructionPlanWorkspace, error) {
	project, err := OpenBookProject(bookPath)
	if err != nil {
		return DeconstructionPlanWorkspace{}, err
	}
	bookPath = projectRootPath(bookPath)
	if err := ensureDeconstructionPlanFiles(bookPath, project.Book.ID); err != nil {
		return DeconstructionPlanWorkspace{}, err
	}
	plan, err := loadDeconstructionPlan(bookPath, project.Book.ID)
	if err != nil {
		return DeconstructionPlanWorkspace{}, err
	}
	templates, err := listDeconstructionTemplates(bookPath, project.Book.ID)
	if err != nil {
		return DeconstructionPlanWorkspace{}, err
	}
	return DeconstructionPlanWorkspace{
		CurrentPlan:             plan,
		CurrentTemplateFileName: findCurrentTemplateFileName(bookPath, project.Book.ID, plan, templates),
		Templates:               templates,
	}, nil
}

// OpenDeconstructionTemplate 读取模板目录中的一个模板，供方案页选择后编辑。
func OpenDeconstructionTemplate(bookPath string, fileName string) (DeconstructionPlan, error) {
	project, err := OpenBookProject(bookPath)
	if err != nil {
		return DeconstructionPlan{}, err
	}
	bookPath = projectRootPath(bookPath)
	if err := ensureDeconstructionPlanFiles(bookPath, project.Book.ID); err != nil {
		return DeconstructionPlan{}, err
	}
	templatePath, err := safeTemplatePath(bookPath, fileName)
	if err != nil {
		return DeconstructionPlan{}, err
	}
	return readDeconstructionPlanFile(templatePath, project.Book.ID)
}

// SaveDeconstructionPlan 保存当前书正在使用的方案，阅读工作台会按它创建后续拆解。
func SaveDeconstructionPlan(request SaveDeconstructionPlanRequest) (DeconstructionPlanWorkspace, error) {
	project, err := OpenBookProject(request.BookPath)
	if err != nil {
		return DeconstructionPlanWorkspace{}, err
	}
	bookPath := projectRootPath(request.BookPath)
	if err := ensureDeconstructionPlanFiles(bookPath, project.Book.ID); err != nil {
		return DeconstructionPlanWorkspace{}, err
	}
	plan := normalizeDeconstructionPlan(request.Plan, project.Book.ID)
	if err := validateDeconstructionPlan(plan); err != nil {
		return DeconstructionPlanWorkspace{}, err
	}
	if err := writeJSONFileAtomic(filepath.Join(bookPath, filepath.FromSlash(deconstructionPlanRelativePath)), plan); err != nil {
		return DeconstructionPlanWorkspace{}, fmt.Errorf("保存当前拆解方案失败: %w", err)
	}
	return OpenDeconstructionPlanWorkspace(bookPath)
}

// SaveDeconstructionTemplate 保存当前书模板目录中的一个模板，另存为时会创建新的 JSON 文件。
func SaveDeconstructionTemplate(request SaveDeconstructionTemplateRequest) (DeconstructionPlanWorkspace, error) {
	project, err := OpenBookProject(request.BookPath)
	if err != nil {
		return DeconstructionPlanWorkspace{}, err
	}
	bookPath := projectRootPath(request.BookPath)
	if err := ensureDeconstructionPlanFiles(bookPath, project.Book.ID); err != nil {
		return DeconstructionPlanWorkspace{}, err
	}
	templateName := strings.TrimSpace(request.TemplateName)
	if templateName == "" {
		return DeconstructionPlanWorkspace{}, errors.New("模板名称不能为空")
	}
	plan := normalizeDeconstructionPlan(request.Plan, project.Book.ID)
	plan.TemplateName = templateName
	if err := validateDeconstructionPlan(plan); err != nil {
		return DeconstructionPlanWorkspace{}, err
	}
	currentWorkspace, err := OpenDeconstructionPlanWorkspace(bookPath)
	if err != nil {
		return DeconstructionPlanWorkspace{}, err
	}
	targetFileName := templateName + ".json"
	if request.Overwrite && strings.TrimSpace(request.FileName) != "" {
		targetFileName = request.FileName
	}
	templatePath, err := safeTemplatePath(bookPath, targetFileName)
	if err != nil {
		return DeconstructionPlanWorkspace{}, err
	}
	if _, err := os.Stat(templatePath); err == nil && !request.Overwrite {
		return DeconstructionPlanWorkspace{}, fmt.Errorf("模板已存在: %s", templateName)
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return DeconstructionPlanWorkspace{}, fmt.Errorf("检查模板文件失败: %w", err)
	}
	if err := writeJSONFileAtomic(templatePath, plan); err != nil {
		return DeconstructionPlanWorkspace{}, fmt.Errorf("保存拆解模板失败: %w", err)
	}
	if request.Overwrite && sameTemplateFile(currentWorkspace.CurrentTemplateFileName, targetFileName) {
		if err := writeJSONFileAtomic(filepath.Join(bookPath, filepath.FromSlash(deconstructionPlanRelativePath)), plan); err != nil {
			return DeconstructionPlanWorkspace{}, fmt.Errorf("同步当前拆解方案失败: %w", err)
		}
	}
	return OpenDeconstructionPlanWorkspace(bookPath)
}

// SetCurrentDeconstructionTemplate 将选中的模板复制为当前方案，不改写历史拆解卡片。
func SetCurrentDeconstructionTemplate(request SetCurrentDeconstructionTemplateRequest) (DeconstructionPlanWorkspace, error) {
	project, err := OpenBookProject(request.BookPath)
	if err != nil {
		return DeconstructionPlanWorkspace{}, err
	}
	bookPath := projectRootPath(request.BookPath)
	if err := ensureDeconstructionPlanFiles(bookPath, project.Book.ID); err != nil {
		return DeconstructionPlanWorkspace{}, err
	}
	templatePath, err := safeTemplatePath(bookPath, request.FileName)
	if err != nil {
		return DeconstructionPlanWorkspace{}, err
	}
	plan, err := readDeconstructionPlanFile(templatePath, project.Book.ID)
	if err != nil {
		return DeconstructionPlanWorkspace{}, err
	}
	if err := writeJSONFileAtomic(filepath.Join(bookPath, filepath.FromSlash(deconstructionPlanRelativePath)), plan); err != nil {
		return DeconstructionPlanWorkspace{}, fmt.Errorf("设置当前拆解模板失败: %w", err)
	}
	return OpenDeconstructionPlanWorkspace(bookPath)
}

func ensureReadingSupportFiles(bookPath string, project BookProject) error {
	for _, relativeDir := range []string{"reading", "source-texts", "deconstructions"} {
		if err := os.MkdirAll(filepath.Join(bookPath, relativeDir), 0o755); err != nil {
			return fmt.Errorf("创建阅读数据目录失败: %w", err)
		}
	}
	if _, err := os.Stat(filepath.Join(bookPath, filepath.FromSlash(readingStateRelativePath))); errors.Is(err, os.ErrNotExist) {
		chapter := chooseChapterForPosition(project, "")
		state := defaultReadingState(project.Book.ID, chapter)
		if err := writeJSONFileAtomic(filepath.Join(bookPath, filepath.FromSlash(readingStateRelativePath)), state); err != nil {
			return err
		}
	}
	if _, err := os.Stat(filepath.Join(bookPath, filepath.FromSlash(sourceTextsRelativePath))); errors.Is(err, os.ErrNotExist) {
		if err := writeJSONFileAtomic(filepath.Join(bookPath, filepath.FromSlash(sourceTextsRelativePath)), SourceTextCollection{Version: 1, Items: []SourceText{}}); err != nil {
			return err
		}
	}
	if _, err := os.Stat(filepath.Join(bookPath, filepath.FromSlash(deconstructionCardsRelativePath))); errors.Is(err, os.ErrNotExist) {
		if err := writeJSONFileAtomic(filepath.Join(bookPath, filepath.FromSlash(deconstructionCardsRelativePath)), DeconstructionCardCollection{Version: 1, Items: []DeconstructionCard{}}); err != nil {
			return err
		}
	}
	return ensureDeconstructionPlanFiles(bookPath, project.Book.ID)
}

func defaultReadingState(bookID string, chapter BookChapter) ReadingState {
	return ReadingState{
		Version:                1,
		BookID:                 bookID,
		ReadingProgressPercent: 0,
		Preferences:            ReadingPreferences{FontSize: defaultReaderFontSize},
		LastPosition: ReadingPosition{
			ChapterID:          chapter.ID,
			ChapterOrder:       chapter.Order,
			CharOffset:         0,
			ParagraphIndex:     0,
			RenderedPageNumber: 1,
			UpdatedAt:          time.Now().Format(time.RFC3339),
		},
		LastMode: "read",
	}
}

func defaultDeconstructionPlan(bookID string) DeconstructionPlan {
	now := time.Now().Format(time.RFC3339)
	return DeconstructionPlan{
		Version:      1,
		BookID:       bookID,
		TemplateName: systemTemplateName,
		UpdatedAt:    now,
		Types: []DeconstructionType{
			{
				ID:    "technique",
				Name:  "表现手法",
				Order: 1,
				Dimensions: []DeconstructionDimension{
					{ID: "environment", Name: "环境描写", Order: 1},
					{ID: "rhythm", Name: "节奏控制", Order: 2},
					{ID: "dialogue", Name: "对白技巧", Order: 3},
					{ID: "imagery", Name: "意象象征", Order: 4},
				},
			},
			{
				ID:    "emotion",
				Name:  "人物感情",
				Order: 2,
				Dimensions: []DeconstructionDimension{
					{ID: "emotion-change", Name: "情绪变化", Order: 1},
					{ID: "motivation", Name: "人物动机", Order: 2},
					{ID: "relationship", Name: "关系变化", Order: 3},
					{ID: "subtext", Name: "潜台词", Order: 4},
				},
			},
			{
				ID:    "structure",
				Name:  "结构设计",
				Order: 3,
				Dimensions: []DeconstructionDimension{
					{ID: "foreshadowing", Name: "伏笔铺垫", Order: 1},
					{ID: "conflict", Name: "冲突推进", Order: 2},
					{ID: "information", Name: "信息释放", Order: 3},
					{ID: "scene-function", Name: "场景功能", Order: 4},
				},
			},
		},
	}
}

func ensureDeconstructionPlanFiles(bookPath string, bookID string) error {
	if err := os.MkdirAll(filepath.Join(bookPath, deconstructionTemplatesDir), 0o755); err != nil {
		return fmt.Errorf("创建拆解模板目录失败: %w", err)
	}
	currentPlanPath := filepath.Join(bookPath, filepath.FromSlash(deconstructionPlanRelativePath))
	if _, err := os.Stat(currentPlanPath); errors.Is(err, os.ErrNotExist) {
		if err := writeJSONFileAtomic(currentPlanPath, defaultDeconstructionPlan(bookID)); err != nil {
			return err
		}
	} else if err != nil {
		return fmt.Errorf("检查当前拆解方案失败: %w", err)
	}
	systemTemplatePath := filepath.Join(bookPath, deconstructionTemplatesDir, systemTemplateFileName)
	if _, err := os.Stat(systemTemplatePath); errors.Is(err, os.ErrNotExist) {
		plan, err := readDeconstructionPlanFile(currentPlanPath, bookID)
		if err != nil {
			plan = defaultDeconstructionPlan(bookID)
		}
		plan.TemplateName = systemTemplateName
		if err := writeJSONFileAtomic(systemTemplatePath, plan); err != nil {
			return fmt.Errorf("保存系统内置拆解模板失败: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("检查系统内置拆解模板失败: %w", err)
	}
	return nil
}

func loadReadingState(bookPath string, project BookProject) (ReadingState, error) {
	var state ReadingState
	if err := readJSONFile(filepath.Join(bookPath, filepath.FromSlash(readingStateRelativePath)), &state); err != nil {
		return ReadingState{}, fmt.Errorf("读取阅读位置失败: %w", err)
	}
	if state.Version == 0 {
		state.Version = 1
	}
	if state.BookID == "" {
		state.BookID = project.Book.ID
	}
	if state.LastMode == "" {
		state.LastMode = "read"
	}
	state.Preferences = normalizeReadingPreferences(state.Preferences)
	if state.LastPosition.ChapterID == "" {
		defaultState := defaultReadingState(project.Book.ID, chooseChapterForPosition(project, ""))
		defaultState.Preferences = state.Preferences
		return defaultState, nil
	}
	if state.ReadingProgressPercent == 0 && (state.LastPosition.ChapterOrder > 1 || state.LastPosition.CharOffset > 0) {
		state.ReadingProgressPercent = estimateReadingProgressPercent(project, state.LastPosition)
	} else {
		state.ReadingProgressPercent = clampPercent(state.ReadingProgressPercent)
	}
	return state, nil
}

func normalizeReadingPreferences(preferences ReadingPreferences) ReadingPreferences {
	if preferences.FontSize == 0 {
		preferences.FontSize = defaultReaderFontSize
	}
	preferences.FontSize = clampInt(preferences.FontSize, minReaderFontSize, maxReaderFontSize)
	return preferences
}

func estimateReadingProgressPercent(project BookProject, position ReadingPosition) int {
	totalWordCount := 0
	completedWordCount := 0
	for _, chapter := range project.Chapters {
		chapterWordCount := maxInt(0, chapter.WordCount)
		totalWordCount += chapterWordCount
		if chapter.ID == position.ChapterID {
			completedWordCount += minInt(chapterWordCount, maxInt(0, position.CharOffset))
			continue
		}
		if chapter.Order < position.ChapterOrder {
			completedWordCount += chapterWordCount
		}
	}
	if totalWordCount <= 0 {
		return 0
	}
	progress := int(float64(completedWordCount) / float64(totalWordCount) * 100)
	return clampPercent(progress)
}

func clampPercent(value int) int {
	if value < 0 {
		return 0
	}
	if value > 100 {
		return 100
	}
	return value
}

func clampInt(value int, minValue int, maxValue int) int {
	if value < minValue {
		return minValue
	}
	if value > maxValue {
		return maxValue
	}
	return value
}

func loadDeconstructionPlan(bookPath string, bookID string) (DeconstructionPlan, error) {
	return readDeconstructionPlanFile(filepath.Join(bookPath, filepath.FromSlash(deconstructionPlanRelativePath)), bookID)
}

func readDeconstructionPlanFile(path string, bookID string) (DeconstructionPlan, error) {
	var plan DeconstructionPlan
	if err := readJSONFile(path, &plan); err != nil {
		return DeconstructionPlan{}, fmt.Errorf("读取拆解方案失败: %w", err)
	}
	return normalizeDeconstructionPlan(plan, bookID), nil
}

func normalizeDeconstructionPlan(plan DeconstructionPlan, bookID string) DeconstructionPlan {
	if plan.Version == 0 {
		plan.Version = 1
	}
	plan.BookID = bookID
	if strings.TrimSpace(plan.TemplateName) == "" {
		plan.TemplateName = systemTemplateName
	}
	if strings.TrimSpace(plan.UpdatedAt) == "" {
		plan.UpdatedAt = time.Now().Format(time.RFC3339)
	}
	usedTypeIDs := map[string]int{}
	for index := range plan.Types {
		plan.Types[index].Name = strings.TrimSpace(plan.Types[index].Name)
		if plan.Types[index].Name == "" {
			plan.Types[index].Name = fmt.Sprintf("拆解类型 %d", index+1)
		}
		plan.Types[index].ID = normalizeStableID(plan.Types[index].ID, "type", plan.Types[index].Name, index+1, usedTypeIDs)
		if plan.Types[index].Order <= 0 {
			plan.Types[index].Order = index + 1
		}
		if len(plan.Types[index].Dimensions) == 0 && strings.TrimSpace(plan.Types[index].Dimension) != "" {
			plan.Types[index].Dimensions = []DeconstructionDimension{
				{ID: stableIDFromName(plan.Types[index].Dimension, "dimension", 1), Name: strings.TrimSpace(plan.Types[index].Dimension), Order: 1},
			}
		}
		if len(plan.Types[index].Dimensions) == 0 {
			plan.Types[index].Dimensions = []DeconstructionDimension{
				{ID: fmt.Sprintf("dimension-%02d", index+1), Name: "通用分析", Order: 1},
			}
		}
		usedDimensionIDs := map[string]int{}
		for dimensionIndex := range plan.Types[index].Dimensions {
			dimension := &plan.Types[index].Dimensions[dimensionIndex]
			dimension.Name = strings.TrimSpace(dimension.Name)
			if dimension.Name == "" {
				dimension.Name = fmt.Sprintf("拆解维度 %d", dimensionIndex+1)
			}
			dimension.ID = normalizeStableID(dimension.ID, "dimension", dimension.Name, dimensionIndex+1, usedDimensionIDs)
			if dimension.Order <= 0 {
				dimension.Order = dimensionIndex + 1
			}
		}
		plan.Types[index].Dimension = ""
		plan.Types[index].Fields = nil
	}
	sort.SliceStable(plan.Types, func(i, j int) bool {
		return plan.Types[i].Order < plan.Types[j].Order
	})
	for index := range plan.Types {
		sort.SliceStable(plan.Types[index].Dimensions, func(i, j int) bool {
			return plan.Types[index].Dimensions[i].Order < plan.Types[index].Dimensions[j].Order
		})
	}
	return plan
}

func validateDeconstructionPlan(plan DeconstructionPlan) error {
	if strings.TrimSpace(plan.TemplateName) == "" {
		return errors.New("模板名称不能为空")
	}
	if len(plan.Types) == 0 {
		return errors.New("至少需要保留一个拆解类型")
	}
	for _, planType := range plan.Types {
		if strings.TrimSpace(planType.Name) == "" {
			return errors.New("拆解类型名称不能为空")
		}
		if len(planType.Dimensions) == 0 {
			return fmt.Errorf("拆解类型“%s”至少需要一个维度", planType.Name)
		}
		for _, dimension := range planType.Dimensions {
			if strings.TrimSpace(dimension.Name) == "" {
				return fmt.Errorf("拆解类型“%s”存在空维度名称", planType.Name)
			}
		}
	}
	return nil
}

func listDeconstructionTemplates(bookPath string, bookID string) ([]DeconstructionTemplateSummary, error) {
	templatesDir := filepath.Join(bookPath, deconstructionTemplatesDir)
	entries, err := os.ReadDir(templatesDir)
	if err != nil {
		return nil, fmt.Errorf("读取拆解模板目录失败: %w", err)
	}
	templates := []DeconstructionTemplateSummary{}
	for _, entry := range entries {
		if entry.IsDir() || !strings.EqualFold(filepath.Ext(entry.Name()), ".json") {
			continue
		}
		plan, err := readDeconstructionPlanFile(filepath.Join(templatesDir, entry.Name()), bookID)
		if err != nil {
			continue
		}
		templates = append(templates, DeconstructionTemplateSummary{
			Name:      plan.TemplateName,
			FileName:  entry.Name(),
			UpdatedAt: plan.UpdatedAt,
		})
	}
	sort.SliceStable(templates, func(i, j int) bool {
		if templates[i].FileName == systemTemplateFileName {
			return true
		}
		if templates[j].FileName == systemTemplateFileName {
			return false
		}
		return strings.ToLower(templates[i].Name) < strings.ToLower(templates[j].Name)
	})
	return templates, nil
}

func findCurrentTemplateFileName(bookPath string, bookID string, currentPlan DeconstructionPlan, templates []DeconstructionTemplateSummary) string {
	templatesDir := filepath.Join(bookPath, deconstructionTemplatesDir)
	for _, template := range templates {
		plan, err := readDeconstructionPlanFile(filepath.Join(templatesDir, template.FileName), bookID)
		if err == nil && deconstructionPlansEquivalent(currentPlan, plan) {
			return template.FileName
		}
	}
	expectedFileName := sanitizeTemplateFileName(currentPlan.TemplateName + ".json")
	for _, template := range templates {
		if sameTemplateFile(template.FileName, expectedFileName) {
			return template.FileName
		}
	}
	return ""
}

func deconstructionPlansEquivalent(left DeconstructionPlan, right DeconstructionPlan) bool {
	left.UpdatedAt = ""
	right.UpdatedAt = ""
	left.BookID = ""
	right.BookID = ""
	leftContent, leftErr := json.Marshal(left)
	rightContent, rightErr := json.Marshal(right)
	return leftErr == nil && rightErr == nil && string(leftContent) == string(rightContent)
}

func sameTemplateFile(left string, right string) bool {
	return strings.EqualFold(sanitizeTemplateFileName(left), sanitizeTemplateFileName(right))
}

func safeTemplatePath(bookPath string, fileName string) (string, error) {
	cleanName := sanitizeTemplateFileName(fileName)
	if cleanName == "" {
		return "", errors.New("模板文件名不能为空")
	}
	return safeProjectPath(bookPath, filepath.ToSlash(filepath.Join(deconstructionTemplatesDir, cleanName)))
}

func sanitizeTemplateFileName(fileName string) string {
	fileName = strings.TrimSpace(fileName)
	fileName = strings.TrimSuffix(fileName, filepath.Ext(fileName))
	var builder strings.Builder
	for _, value := range fileName {
		switch value {
		case '\\', '/', ':', '*', '?', '"', '<', '>', '|':
			builder.WriteByte('-')
		default:
			builder.WriteRune(value)
		}
	}
	cleaned := strings.Trim(builder.String(), " .-_")
	if cleaned == "" {
		return ""
	}
	return cleaned + ".json"
}

func normalizeStableID(currentID string, prefix string, name string, fallbackOrder int, used map[string]int) string {
	base := strings.TrimSpace(currentID)
	if base == "" {
		base = stableIDFromName(name, prefix, fallbackOrder)
	}
	base = sanitizeID(base, prefix, fallbackOrder)
	used[base]++
	if used[base] == 1 {
		return base
	}
	return fmt.Sprintf("%s-%02d", base, used[base])
}

func stableIDFromName(name string, prefix string, fallbackOrder int) string {
	return fmt.Sprintf("%s-%02d", prefix, fallbackOrder)
}

func sanitizeID(value string, prefix string, fallbackOrder int) string {
	value = strings.TrimSpace(value)
	var builder strings.Builder
	for _, item := range strings.ToLower(value) {
		switch {
		case item >= 'a' && item <= 'z', item >= '0' && item <= '9':
			builder.WriteRune(item)
		case item == '-' || item == '_':
			builder.WriteRune(item)
		}
	}
	cleaned := strings.Trim(builder.String(), "-_")
	if cleaned == "" {
		return fmt.Sprintf("%s-%02d", prefix, fallbackOrder)
	}
	return cleaned
}

func loadSourceTexts(bookPath string) (SourceTextCollection, error) {
	collection := SourceTextCollection{Version: 1, Items: []SourceText{}}
	if err := readJSONFile(filepath.Join(bookPath, filepath.FromSlash(sourceTextsRelativePath)), &collection); err != nil {
		return collection, fmt.Errorf("读取选文来源失败: %w", err)
	}
	if collection.Version == 0 {
		collection.Version = 1
	}
	if collection.Items == nil {
		collection.Items = []SourceText{}
	}
	return collection, nil
}

func loadDeconstructionCards(bookPath string) (DeconstructionCardCollection, error) {
	collection := DeconstructionCardCollection{Version: 1, Items: []DeconstructionCard{}}
	if err := readJSONFile(filepath.Join(bookPath, filepath.FromSlash(deconstructionCardsRelativePath)), &collection); err != nil {
		return collection, fmt.Errorf("读取拆解卡片失败: %w", err)
	}
	if collection.Version == 0 {
		collection.Version = 1
	}
	if collection.Items == nil {
		collection.Items = []DeconstructionCard{}
	}
	return collection, nil
}

func ensureChapterHasNoSourceTexts(bookPath string, chapterID string) error {
	sourceTexts, err := loadSourceTexts(bookPath)
	if err != nil {
		return err
	}
	for _, sourceText := range sourceTexts.Items {
		if sourceText.ChapterID == chapterID {
			return errors.New("该章节已有选文来源，暂不能直接拆分；请先确认相关拆解或素材的来源迁移")
		}
	}
	return nil
}

func syncPersistedChapterOrdersAfterSplit(bookPath string, splitOrder int, splitChapterID string, remainingChapter BookChapter) error {
	readingStatePath := filepath.Join(bookPath, filepath.FromSlash(readingStateRelativePath))
	var state ReadingState
	if err := readOptionalJSONFile(readingStatePath, &state); err == nil && state.LastPosition.ChapterID != "" {
		changed := false
		if state.LastPosition.ChapterID == splitChapterID {
			state.LastPosition.ChapterOrder = remainingChapter.Order
			state.LastPosition.CharOffset = 0
			state.LastPosition.ParagraphIndex = 0
			state.LastPosition.RenderedPageNumber = 1
			state.ReadingProgressPercent = 0
			changed = true
		} else if state.LastPosition.ChapterOrder > splitOrder {
			state.LastPosition.ChapterOrder++
			changed = true
		}
		if changed {
			state.Preferences = normalizeReadingPreferences(state.Preferences)
			if err := writeJSONFileAtomic(readingStatePath, state); err != nil {
				return fmt.Errorf("同步阅读位置章节序号失败: %w", err)
			}
		}
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("读取阅读位置用于章节拆分失败: %w", err)
	}

	sourceTextsPath := filepath.Join(bookPath, filepath.FromSlash(sourceTextsRelativePath))
	collection := SourceTextCollection{Version: 1}
	if err := readOptionalJSONFile(sourceTextsPath, &collection); err == nil && collection.Items != nil {
		changed := false
		for index := range collection.Items {
			if collection.Items[index].ChapterOrder > splitOrder {
				collection.Items[index].ChapterOrder++
				changed = true
			}
		}
		if changed {
			if err := writeJSONFileAtomic(sourceTextsPath, collection); err != nil {
				return fmt.Errorf("同步选文来源章节序号失败: %w", err)
			}
		}
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("读取选文来源用于章节拆分失败: %w", err)
	}
	return nil
}

func readChapterText(bookPath string, chapter BookChapter) (string, error) {
	contentPath, err := safeProjectPath(bookPath, chapter.ContentPath)
	if err != nil {
		return "", err
	}
	content, err := os.ReadFile(contentPath)
	if err != nil {
		return "", fmt.Errorf("读取章节正文失败: %w", err)
	}
	return string(trimUTF8BOM(content)), nil
}

func readJSONFile(path string, target any) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(content, target); err != nil {
		return err
	}
	return nil
}

func writeTextFileAtomic(path string, value string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("创建文本目录失败: %w", err)
	}
	tempFile, err := os.CreateTemp(filepath.Dir(path), filepath.Base(path)+".*.tmp")
	if err != nil {
		return fmt.Errorf("创建临时文本文件失败: %w", err)
	}
	tempPath := tempFile.Name()
	if _, err := tempFile.WriteString(value); err != nil {
		tempFile.Close()
		_ = os.Remove(tempPath)
		return fmt.Errorf("写入临时文本文件失败: %w", err)
	}
	if err := tempFile.Close(); err != nil {
		_ = os.Remove(tempPath)
		return fmt.Errorf("关闭临时文本文件失败: %w", err)
	}
	if err := os.Rename(tempPath, path); err != nil {
		if removeErr := os.Remove(path); removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) {
			_ = os.Remove(tempPath)
			return fmt.Errorf("删除旧文本文件失败: %w", removeErr)
		}
		if renameErr := os.Rename(tempPath, path); renameErr != nil {
			_ = os.Remove(tempPath)
			return fmt.Errorf("替换文本文件失败: %w", renameErr)
		}
	}
	return nil
}

func chooseChapterForPosition(project BookProject, chapterID string) BookChapter {
	if chapter, ok := findChapterByID(project, chapterID); ok {
		return chapter
	}
	if len(project.Chapters) == 0 {
		return BookChapter{ID: "chapter-000001", Title: "正文", Order: 1, ContentPath: "chapters/chapter-000001.txt"}
	}
	return project.Chapters[0]
}

func findChapterByID(project BookProject, chapterID string) (BookChapter, bool) {
	for _, chapter := range project.Chapters {
		if chapter.ID == chapterID {
			return chapter, true
		}
	}
	return BookChapter{}, false
}

func safeProjectPath(bookPath string, relativePath string) (string, error) {
	root := projectRootPath(bookPath)
	cleanRelative := filepath.Clean(filepath.FromSlash(strings.TrimSpace(relativePath)))
	if cleanRelative == "." || filepath.IsAbs(cleanRelative) || strings.HasPrefix(cleanRelative, "..") {
		return "", fmt.Errorf("非法项目内路径: %s", relativePath)
	}
	fullPath := filepath.Join(root, cleanRelative)
	rel, err := filepath.Rel(root, fullPath)
	if err != nil || strings.HasPrefix(rel, "..") || filepath.IsAbs(rel) {
		return "", fmt.Errorf("路径越界: %s", relativePath)
	}
	return fullPath, nil
}

func projectRootPath(bookPath string) string {
	return filepath.Clean(strings.TrimSpace(bookPath))
}

func findPlanType(plan DeconstructionPlan, planTypeID string) (DeconstructionType, bool) {
	for _, planType := range plan.Types {
		if planType.ID == planTypeID {
			return planType, true
		}
	}
	return DeconstructionType{}, false
}

func findPlanDimension(planType DeconstructionType, dimensionID string) (DeconstructionDimension, bool) {
	for _, dimension := range planType.Dimensions {
		if dimension.ID == dimensionID {
			return dimension, true
		}
	}
	return DeconstructionDimension{}, false
}

func findReusableSourceText(collection SourceTextCollection, bookID string, selection SourceSelection) SourceText {
	for _, item := range collection.Items {
		if item.BookID == bookID &&
			item.ChapterID == selection.ChapterID &&
			item.StartOffset == selection.StartOffset &&
			item.EndOffset == selection.EndOffset &&
			item.SelectedText == selection.SelectedText {
			return item
		}
	}
	return SourceText{}
}

func normalizeFieldValues(values map[string]string) map[string]string {
	normalized := map[string]string{}
	for key, value := range values {
		cleanKey := strings.TrimSpace(key)
		if cleanKey == "" {
			continue
		}
		normalized[cleanKey] = strings.TrimSpace(value)
	}
	return normalized
}

func normalizeTags(values []string) []string {
	seen := map[string]bool{}
	tags := make([]string, 0, len(values))
	for _, value := range values {
		tag := strings.TrimSpace(value)
		if tag == "" || seen[tag] {
			continue
		}
		seen[tag] = true
		tags = append(tags, tag)
	}
	return tags
}

func nextSequentialID(prefix string, next int) string {
	return fmt.Sprintf("%s-%s-%04d", prefix, time.Now().Format("20060102"), next)
}

func nextChapterID(project BookProject) string {
	used := map[string]bool{}
	for _, chapter := range project.Chapters {
		used[chapter.ID] = true
	}
	for next := len(project.Chapters) + 1; ; next++ {
		chapterID := fmt.Sprintf("chapter-%06d", next)
		if !used[chapterID] {
			return chapterID
		}
	}
}

func maxInt(left int, right int) int {
	if left > right {
		return left
	}
	return right
}

func minInt(left int, right int) int {
	if left < right {
		return left
	}
	return right
}
