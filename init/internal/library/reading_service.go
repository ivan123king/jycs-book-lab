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
	if !planHasType(plan, request.PlanTypeID) {
		return CreateDeconstructionResult{}, fmt.Errorf("拆解类型不存在: %s", request.PlanTypeID)
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
		ID:           nextSequentialID("deconstruction", len(cards.Items)+1),
		BookID:       project.Book.ID,
		SourceTextID: sourceText.ID,
		PlanTypeID:   request.PlanTypeID,
		FieldValues:  normalizeFieldValues(request.FieldValues),
		TagIDs:       normalizeTags(request.TagIDs),
		CreatedAt:    now,
		UpdatedAt:    now,
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
	if _, err := os.Stat(filepath.Join(bookPath, filepath.FromSlash(deconstructionPlanRelativePath))); errors.Is(err, os.ErrNotExist) {
		if err := writeJSONFileAtomic(filepath.Join(bookPath, filepath.FromSlash(deconstructionPlanRelativePath)), defaultDeconstructionPlan(project.Book.ID)); err != nil {
			return err
		}
	}
	return nil
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
		Version:   1,
		BookID:    bookID,
		UpdatedAt: now,
		Types: []DeconstructionType{
			{
				ID:        "scene",
				Name:      "场景拆解",
				Dimension: "场景",
				Order:     1,
				Fields: []DeconstructionField{
					{ID: "scenePurpose", Label: "场景目的", Type: "textarea", Order: 1, Required: true, Visible: true},
					{ID: "emotionChange", Label: "情绪变化", Type: "textarea", Order: 2, Required: false, Visible: true},
				},
			},
			{
				ID:        "dialogue",
				Name:      "对白拆解",
				Dimension: "人物",
				Order:     2,
				Fields: []DeconstructionField{
					{ID: "subtext", Label: "潜台词", Type: "textarea", Order: 1, Required: true, Visible: true},
					{ID: "relationship", Label: "关系变化", Type: "text", Order: 2, Required: false, Visible: true},
				},
			},
			{
				ID:        "character",
				Name:      "人物弧光",
				Dimension: "人物",
				Order:     3,
				Fields: []DeconstructionField{
					{ID: "motivation", Label: "人物动机", Type: "textarea", Order: 1, Required: true, Visible: true},
					{ID: "change", Label: "转变线索", Type: "textarea", Order: 2, Required: false, Visible: true},
				},
			},
		},
	}
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
	var plan DeconstructionPlan
	if err := readJSONFile(filepath.Join(bookPath, filepath.FromSlash(deconstructionPlanRelativePath)), &plan); err != nil {
		return DeconstructionPlan{}, fmt.Errorf("读取拆解方案失败: %w", err)
	}
	if plan.Version == 0 {
		plan.Version = 1
	}
	if plan.BookID == "" {
		plan.BookID = bookID
	}
	sort.SliceStable(plan.Types, func(i, j int) bool {
		return plan.Types[i].Order < plan.Types[j].Order
	})
	for index := range plan.Types {
		sort.SliceStable(plan.Types[index].Fields, func(i, j int) bool {
			return plan.Types[index].Fields[i].Order < plan.Types[index].Fields[j].Order
		})
	}
	return plan, nil
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

func planHasType(plan DeconstructionPlan, planTypeID string) bool {
	for _, planType := range plan.Types {
		if planType.ID == planTypeID {
			return true
		}
	}
	return false
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
