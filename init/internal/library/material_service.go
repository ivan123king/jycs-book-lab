// material_service.go 负责单本小说素材库的正式数据读写、来源追溯和全局索引轻量刷新。
package library

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var errMaterialSourceCardNotFound = errors.New("素材来源拆解卡片不存在")

// OpenMaterialWorkspace 读取当前书素材、素材正文和来源上下文，所有正式数据都来自小说项目目录。
func OpenMaterialWorkspace(bookPath string) (MaterialWorkspace, error) {
	project, err := OpenBookProject(bookPath)
	if err != nil {
		return MaterialWorkspace{}, err
	}
	bookPath = projectRootPath(bookPath)
	if err := ensureReadingSupportFiles(bookPath, project); err != nil {
		return MaterialWorkspace{}, err
	}
	materials, err := loadMaterials(bookPath)
	if err != nil {
		return MaterialWorkspace{}, err
	}
	contents, err := loadMaterialContents(bookPath, materials)
	if err != nil {
		return MaterialWorkspace{}, err
	}
	sourceTexts, err := loadSourceTexts(bookPath)
	if err != nil {
		return MaterialWorkspace{}, err
	}
	cards, err := loadDeconstructionCards(bookPath)
	if err != nil {
		return MaterialWorkspace{}, err
	}
	return MaterialWorkspace{
		Project:             project,
		Materials:           materials,
		MaterialContents:    contents,
		SourceTexts:         sourceTexts,
		DeconstructionCards: cards,
	}, nil
}

// BuildMaterialDraftFromCard 根据写作手法或剧情拆解卡片生成可编辑素材草稿，不直接保存为素材。
func BuildMaterialDraftFromCard(request BuildMaterialDraftFromCardRequest) (MaterialDraft, error) {
	project, err := OpenBookProject(request.BookPath)
	if err != nil {
		return MaterialDraft{}, err
	}
	bookPath := projectRootPath(request.BookPath)
	if err := ensureReadingSupportFiles(bookPath, project); err != nil {
		return MaterialDraft{}, err
	}
	card, sourceText, chapter, err := loadWritingTechniqueSourceContext(bookPath, project, request.CardID)
	if err == nil {
		sourceRef := buildWritingTechniqueMaterialSourceRef(project, card, sourceText, chapter)
		title := strings.TrimSpace(card.DimensionName)
		if title == "" {
			title = strings.TrimSpace(card.PlanTypeName)
		}
		if title == "" {
			title = "写作手法素材"
		}
		return MaterialDraft{
			MaterialType: analysisKindWritingTechnique,
			Title:        title,
			Summary:      sourceText.SelectedText,
			Content:      card.Content,
			CategoryIDs:  []string{},
			TagIDs:       append([]string{}, card.TagIDs...),
			SourceRefs:   []MaterialSourceRef{sourceRef},
		}, nil
	}
	if !errors.Is(err, errMaterialSourceCardNotFound) {
		return MaterialDraft{}, err
	}
	storyCard, err := loadStoryPlotSourceContext(bookPath, request.CardID)
	if err != nil {
		return MaterialDraft{}, err
	}
	title := strings.TrimSpace(storyCard.Title)
	if title == "" {
		title = strings.TrimSpace(storyCard.PlanTypeName)
	}
	if title == "" {
		title = "剧情结构素材"
	}
	return MaterialDraft{
		MaterialType: analysisKindStoryPlot,
		Title:        title,
		Summary:      storyCard.Summary,
		Content:      buildStoryPlotMaterialContent(storyCard),
		CategoryIDs:  []string{},
		TagIDs:       append([]string{}, storyCard.TagIDs...),
		SourceRefs:   []MaterialSourceRef{buildStoryPlotMaterialSourceRef(project, storyCard)},
	}, nil
}

// CreateMaterialFromCard 把用户二次编辑后的素材草稿保存到当前小说项目目录，并保留来源拆解引用。
func CreateMaterialFromCard(request CreateMaterialFromCardRequest) (CreateMaterialResult, error) {
	project, err := OpenBookProject(request.BookPath)
	if err != nil {
		return CreateMaterialResult{}, err
	}
	bookPath := projectRootPath(request.BookPath)
	if err := ensureReadingSupportFiles(bookPath, project); err != nil {
		return CreateMaterialResult{}, err
	}
	title, summary, content, err := normalizeMaterialText(request.Title, request.Summary, request.Content)
	if err != nil {
		return CreateMaterialResult{}, err
	}
	card, sourceText, chapter, err := loadWritingTechniqueSourceContext(bookPath, project, request.CardID)
	isWritingTechnique := err == nil
	storyCard := StoryPlotCard{}
	if err != nil {
		if !errors.Is(err, errMaterialSourceCardNotFound) {
			return CreateMaterialResult{}, err
		}
		storyCard, err = loadStoryPlotSourceContext(bookPath, request.CardID)
		if err != nil {
			return CreateMaterialResult{}, err
		}
	}
	materials, err := loadMaterials(bookPath)
	if err != nil {
		return CreateMaterialResult{}, err
	}
	cards, err := loadDeconstructionCards(bookPath)
	if err != nil {
		return CreateMaterialResult{}, err
	}

	now := time.Now().Format(time.RFC3339)
	materialID := nextMaterialID(materials)
	materialType := analysisKindWritingTechnique
	sourceRefs := []MaterialSourceRef{}
	if !isWritingTechnique {
		materialType = analysisKindStoryPlot
		sourceRefs = []MaterialSourceRef{buildStoryPlotMaterialSourceRef(project, storyCard)}
	} else {
		sourceRefs = []MaterialSourceRef{buildWritingTechniqueMaterialSourceRef(project, card, sourceText, chapter)}
	}
	material := Material{
		ID:           materialID,
		BookID:       project.Book.ID,
		MaterialType: materialType,
		Title:        title,
		Summary:      summary,
		ContentPath:  filepath.ToSlash(filepath.Join(materialTextsDir, materialID+".txt")),
		CategoryIDs:  normalizeTags(request.CategoryIDs),
		TagIDs:       normalizeTags(request.TagIDs),
		SourceRefs:   sourceRefs,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := writeMaterialContent(bookPath, material, content); err != nil {
		return CreateMaterialResult{}, err
	}
	materials.Items = append(materials.Items, material)
	if err := writeJSONFileAtomic(filepath.Join(bookPath, filepath.FromSlash(materialsRelativePath)), materials); err != nil {
		return CreateMaterialResult{}, fmt.Errorf("保存素材索引失败: %w", err)
	}
	if isWritingTechnique {
		cards = addMaterialIDToCard(cards, card.ID, material.ID)
		if err := writeJSONFileAtomic(filepath.Join(bookPath, filepath.FromSlash(deconstructionCardsRelativePath)), cards); err != nil {
			return CreateMaterialResult{}, fmt.Errorf("同步写作手法拆解卡片素材引用失败: %w", err)
		}
	} else {
		storyCards, err := loadStoryPlotCards(bookPath)
		if err != nil {
			return CreateMaterialResult{}, err
		}
		storyCards = addMaterialIDToStoryPlotCard(storyCards, storyCard.ID, material.ID)
		if err := writeStoryPlotCards(bookPath, storyCards); err != nil {
			return CreateMaterialResult{}, fmt.Errorf("同步剧情拆解卡片素材引用失败: %w", err)
		}
	}
	contents, err := loadMaterialContents(bookPath, materials)
	if err != nil {
		return CreateMaterialResult{}, err
	}
	// 全局索引只是跨书快捷入口，刷新失败不能影响单书素材正式保存。
	_ = refreshGlobalMaterialIndexForBook(bookPath, project, materials)
	return CreateMaterialResult{
		Material:            material,
		Materials:           materials,
		MaterialContents:    contents,
		DeconstructionCards: cards,
	}, nil
}

// UpdateMaterial 更新素材元数据和 TXT 正文，不改变已有来源追溯。
func UpdateMaterial(request UpdateMaterialRequest) (UpdateMaterialResult, error) {
	project, err := OpenBookProject(request.BookPath)
	if err != nil {
		return UpdateMaterialResult{}, err
	}
	bookPath := projectRootPath(request.BookPath)
	if err := ensureReadingSupportFiles(bookPath, project); err != nil {
		return UpdateMaterialResult{}, err
	}
	title, summary, content, err := normalizeMaterialText(request.Title, request.Summary, request.Content)
	if err != nil {
		return UpdateMaterialResult{}, err
	}
	materials, err := loadMaterials(bookPath)
	if err != nil {
		return UpdateMaterialResult{}, err
	}
	for index := range materials.Items {
		if materials.Items[index].ID != request.MaterialID {
			continue
		}
		materials.Items[index].Title = title
		materials.Items[index].Summary = summary
		materials.Items[index].CategoryIDs = normalizeTags(request.CategoryIDs)
		materials.Items[index].TagIDs = normalizeTags(request.TagIDs)
		materials.Items[index].UpdatedAt = time.Now().Format(time.RFC3339)
		if err := writeMaterialContent(bookPath, materials.Items[index], content); err != nil {
			return UpdateMaterialResult{}, err
		}
		if err := writeJSONFileAtomic(filepath.Join(bookPath, filepath.FromSlash(materialsRelativePath)), materials); err != nil {
			return UpdateMaterialResult{}, fmt.Errorf("保存素材索引失败: %w", err)
		}
		contents, err := loadMaterialContents(bookPath, materials)
		if err != nil {
			return UpdateMaterialResult{}, err
		}
		_ = refreshGlobalMaterialIndexForBook(bookPath, project, materials)
		return UpdateMaterialResult{Material: materials.Items[index], Materials: materials, MaterialContents: contents}, nil
	}
	return UpdateMaterialResult{}, fmt.Errorf("素材不存在: %s", request.MaterialID)
}

// DeleteMaterial 删除单本书素材和对应 TXT，并从写作手法和剧情拆解卡片中移除反向引用。
func DeleteMaterial(request DeleteMaterialRequest) (DeleteMaterialResult, error) {
	project, err := OpenBookProject(request.BookPath)
	if err != nil {
		return DeleteMaterialResult{}, err
	}
	bookPath := projectRootPath(request.BookPath)
	if err := ensureReadingSupportFiles(bookPath, project); err != nil {
		return DeleteMaterialResult{}, err
	}
	materials, err := loadMaterials(bookPath)
	if err != nil {
		return DeleteMaterialResult{}, err
	}
	removed := Material{}
	kept := make([]Material, 0, len(materials.Items))
	for _, material := range materials.Items {
		if material.ID == request.MaterialID {
			removed = material
			continue
		}
		kept = append(kept, material)
	}
	if removed.ID == "" {
		return DeleteMaterialResult{}, fmt.Errorf("素材不存在: %s", request.MaterialID)
	}
	materials.Items = kept
	if err := writeJSONFileAtomic(filepath.Join(bookPath, filepath.FromSlash(materialsRelativePath)), materials); err != nil {
		return DeleteMaterialResult{}, fmt.Errorf("保存素材索引失败: %w", err)
	}
	if removed.ContentPath != "" {
		if contentPath, pathErr := safeProjectPath(bookPath, removed.ContentPath); pathErr == nil {
			if err := os.Remove(contentPath); err != nil && !errors.Is(err, os.ErrNotExist) {
				return DeleteMaterialResult{}, fmt.Errorf("删除素材正文失败: %w", err)
			}
		}
	}
	cards, err := loadDeconstructionCards(bookPath)
	if err != nil {
		return DeleteMaterialResult{}, err
	}
	cards = removeMaterialIDFromCards(cards, removed.ID)
	if err := writeJSONFileAtomic(filepath.Join(bookPath, filepath.FromSlash(deconstructionCardsRelativePath)), cards); err != nil {
		return DeleteMaterialResult{}, fmt.Errorf("同步写作手法拆解卡片素材引用失败: %w", err)
	}
	storyCards, err := loadStoryPlotCards(bookPath)
	if err != nil {
		return DeleteMaterialResult{}, err
	}
	storyCards = removeMaterialIDFromStoryPlotCards(storyCards, removed.ID)
	if err := writeStoryPlotCards(bookPath, storyCards); err != nil {
		return DeleteMaterialResult{}, fmt.Errorf("同步剧情拆解卡片素材引用失败: %w", err)
	}
	contents, err := loadMaterialContents(bookPath, materials)
	if err != nil {
		return DeleteMaterialResult{}, err
	}
	_ = refreshGlobalMaterialIndexForBook(bookPath, project, materials)
	return DeleteMaterialResult{Materials: materials, MaterialContents: contents, DeconstructionCards: cards}, nil
}

// loadMaterials 读取小说内素材集合，并补齐旧数据中的空切片字段。
func loadMaterials(bookPath string) (MaterialCollection, error) {
	collection := MaterialCollection{Version: 1, Items: []Material{}}
	if err := readJSONFile(filepath.Join(bookPath, filepath.FromSlash(materialsRelativePath)), &collection); err != nil {
		return collection, fmt.Errorf("读取素材库失败: %w", err)
	}
	if collection.Version == 0 {
		collection.Version = 1
	}
	if collection.Items == nil {
		collection.Items = []Material{}
		return collection, nil
	}
	for index := range collection.Items {
		if collection.Items[index].CategoryIDs == nil {
			collection.Items[index].CategoryIDs = []string{}
		}
		if collection.Items[index].TagIDs == nil {
			collection.Items[index].TagIDs = []string{}
		}
		if collection.Items[index].SourceRefs == nil {
			collection.Items[index].SourceRefs = []MaterialSourceRef{}
		}
	}
	return collection, nil
}

// loadMaterialContents 按 contentPath 读取素材正文，列表搜索和详情展示都以 TXT 为准。
func loadMaterialContents(bookPath string, materials MaterialCollection) (map[string]string, error) {
	contents := map[string]string{}
	for _, material := range materials.Items {
		content, err := readMaterialContent(bookPath, material)
		if err != nil {
			return contents, err
		}
		contents[material.ID] = content
	}
	return contents, nil
}

func readMaterialContent(bookPath string, material Material) (string, error) {
	if strings.TrimSpace(material.ContentPath) == "" {
		return "", nil
	}
	path, err := safeProjectPath(bookPath, material.ContentPath)
	if err != nil {
		return "", err
	}
	content, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("读取素材正文失败: %w", err)
	}
	return string(trimUTF8BOM(content)), nil
}

func writeMaterialContent(bookPath string, material Material, content string) error {
	path, err := safeProjectPath(bookPath, material.ContentPath)
	if err != nil {
		return err
	}
	if err := writeTextFileAtomic(path, content); err != nil {
		return fmt.Errorf("保存素材正文失败: %w", err)
	}
	return nil
}

func loadWritingTechniqueSourceContext(bookPath string, project BookProject, cardID string) (DeconstructionCard, SourceText, BookChapter, error) {
	sourceTexts, err := loadSourceTexts(bookPath)
	if err != nil {
		return DeconstructionCard{}, SourceText{}, BookChapter{}, err
	}
	cards, err := loadDeconstructionCards(bookPath)
	if err != nil {
		return DeconstructionCard{}, SourceText{}, BookChapter{}, err
	}
	for _, card := range cards.Items {
		if card.ID != cardID {
			continue
		}
		sourceText, ok := findSourceTextByID(sourceTexts, card.SourceTextID)
		if !ok {
			return DeconstructionCard{}, SourceText{}, BookChapter{}, fmt.Errorf("拆解来源选文不存在: %s", card.SourceTextID)
		}
		chapter, ok := findChapterByID(project, sourceText.ChapterID)
		if !ok {
			return DeconstructionCard{}, SourceText{}, BookChapter{}, fmt.Errorf("拆解来源章节不存在: %s", sourceText.ChapterID)
		}
		return card, sourceText, chapter, nil
	}
	return DeconstructionCard{}, SourceText{}, BookChapter{}, fmt.Errorf("%w: %s", errMaterialSourceCardNotFound, cardID)
}

func buildWritingTechniqueMaterialSourceRef(project BookProject, card DeconstructionCard, sourceText SourceText, chapter BookChapter) MaterialSourceRef {
	return MaterialSourceRef{
		SourceKind:       analysisKindWritingTechnique,
		DeconstructionID: card.ID,
		SourceTextID:     sourceText.ID,
		ChapterRefs: []MaterialChapterRef{
			{
				ChapterID:            chapter.ID,
				ChapterOrder:         chapter.Order,
				ChapterTitleSnapshot: chapter.Title,
			},
		},
		Snapshot: MaterialSourceSnapshot{
			BookName:            project.Book.Name,
			ChapterTitle:        chapter.Title,
			DeconstructionTitle: strings.TrimSpace(card.PlanTypeName + " / " + card.DimensionName),
			PlanTypeName:        card.PlanTypeName,
			DimensionName:       card.DimensionName,
			SourceSummary:       excerptText(card.Content, 120),
			SelectedText:        sourceText.SelectedText,
			CreatedAt:           card.CreatedAt,
		},
	}
}

// loadStoryPlotSourceContext 读取剧情拆解卡片来源，供素材沉淀时保留剧情章节范围追溯。
func loadStoryPlotSourceContext(bookPath string, cardID string) (StoryPlotCard, error) {
	cards, err := loadStoryPlotCards(bookPath)
	if err != nil {
		return StoryPlotCard{}, err
	}
	for _, card := range cards.Items {
		if card.ID == cardID {
			return card, nil
		}
	}
	return StoryPlotCard{}, fmt.Errorf("%w: %s", errMaterialSourceCardNotFound, cardID)
}

// buildStoryPlotMaterialSourceRef 把剧情卡片的章节范围和分析快照转换为素材来源引用。
func buildStoryPlotMaterialSourceRef(project BookProject, card StoryPlotCard) MaterialSourceRef {
	chapterRefs := make([]MaterialChapterRef, 0, len(card.ChapterRefs))
	chapterTitles := make([]string, 0, len(card.ChapterRefs))
	for _, chapter := range card.ChapterRefs {
		chapterRefs = append(chapterRefs, MaterialChapterRef{
			ChapterID:            chapter.ChapterID,
			ChapterOrder:         chapter.ChapterOrder,
			ChapterTitleSnapshot: chapter.ChapterTitleSnapshot,
		})
		chapterTitles = append(chapterTitles, chapter.ChapterTitleSnapshot)
	}
	return MaterialSourceRef{
		SourceKind:       analysisKindStoryPlot,
		DeconstructionID: card.ID,
		SourceTextID:     "",
		ChapterRefs:      chapterRefs,
		Snapshot: MaterialSourceSnapshot{
			BookName:            project.Book.Name,
			ChapterTitle:        strings.Join(chapterTitles, " / "),
			DeconstructionTitle: firstNonEmpty(card.Title, "未命名剧情拆解"),
			PlanTypeName:        card.PlanTypeName,
			DimensionName:       primaryStoryPlotDimensionName(card),
			SourceSummary:       excerptText(firstNonEmpty(card.Summary, buildStoryPlotMaterialContent(card)), 120),
			SelectedText:        card.Summary,
			CreatedAt:           card.CreatedAt,
		},
	}
}

// buildStoryPlotMaterialContent 生成剧情素材草稿正文，用户保存前仍可二次整理为可复用方法。
func buildStoryPlotMaterialContent(card StoryPlotCard) string {
	dimensionLines := []string{}
	for _, dimension := range card.Dimensions {
		name := strings.TrimSpace(dimension.Name)
		content := strings.TrimSpace(dimension.Content)
		// 剧情素材的可复用方法只沉淀用户真正填写过的维度，章节和简介留在来源信息里展示。
		if content == "" {
			continue
		}
		if name == "" {
			name = "剧情维度"
		}
		dimensionLines = append(dimensionLines, name+"：\n"+content)
	}
	if len(dimensionLines) > 0 {
		return strings.Join(dimensionLines, "\n\n")
	}
	return firstNonEmpty(card.Summary, card.Title, card.PlanTypeName, "剧情结构素材")
}

// primaryStoryPlotDimensionName 选择最能代表剧情素材的维度名称，用于素材列表展示和筛选。
func primaryStoryPlotDimensionName(card StoryPlotCard) string {
	for _, dimension := range card.Dimensions {
		if strings.TrimSpace(dimension.Content) != "" {
			return dimension.Name
		}
	}
	if len(card.Dimensions) > 0 {
		return card.Dimensions[0].Name
	}
	return "剧情结构"
}

func normalizeMaterialText(title string, summary string, content string) (string, string, string, error) {
	cleanTitle := strings.TrimSpace(title)
	cleanSummary := strings.TrimRight(summary, "\r\n")
	cleanContent := strings.TrimRight(content, "\r\n")
	if cleanTitle == "" {
		return "", "", "", errors.New("素材标题不能为空")
	}
	if strings.TrimSpace(cleanContent) == "" {
		return "", "", "", errors.New("素材正文不能为空")
	}
	if strings.TrimSpace(cleanSummary) == "" {
		cleanSummary = excerptText(cleanContent, 80)
	}
	return cleanTitle, cleanSummary, cleanContent, nil
}

func nextMaterialID(materials MaterialCollection) string {
	existing := map[string]bool{}
	for _, material := range materials.Items {
		existing[material.ID] = true
	}
	for next := len(materials.Items) + 1; ; next++ {
		id := nextSequentialID("material", next)
		if !existing[id] {
			return id
		}
	}
}

func findSourceTextByID(collection SourceTextCollection, sourceTextID string) (SourceText, bool) {
	for _, sourceText := range collection.Items {
		if sourceText.ID == sourceTextID {
			return sourceText, true
		}
	}
	return SourceText{}, false
}

func addMaterialIDToCard(cards DeconstructionCardCollection, cardID string, materialID string) DeconstructionCardCollection {
	for index := range cards.Items {
		if cards.Items[index].ID != cardID {
			continue
		}
		for _, existing := range cards.Items[index].MaterialIDs {
			if existing == materialID {
				return cards
			}
		}
		cards.Items[index].MaterialIDs = append(cards.Items[index].MaterialIDs, materialID)
		cards.Items[index].UpdatedAt = time.Now().Format(time.RFC3339)
		return cards
	}
	return cards
}

func removeMaterialIDFromCards(cards DeconstructionCardCollection, materialID string) DeconstructionCardCollection {
	for index := range cards.Items {
		kept := make([]string, 0, len(cards.Items[index].MaterialIDs))
		for _, existing := range cards.Items[index].MaterialIDs {
			if existing != materialID {
				kept = append(kept, existing)
			}
		}
		cards.Items[index].MaterialIDs = kept
	}
	return cards
}

// addMaterialIDToStoryPlotCard 在剧情卡片中记录已沉淀素材，避免素材和来源卡片关系丢失。
func addMaterialIDToStoryPlotCard(cards StoryPlotCardCollection, cardID string, materialID string) StoryPlotCardCollection {
	for index := range cards.Items {
		if cards.Items[index].ID != cardID {
			continue
		}
		for _, existing := range cards.Items[index].MaterialIDs {
			if existing == materialID {
				return cards
			}
		}
		cards.Items[index].MaterialIDs = append(cards.Items[index].MaterialIDs, materialID)
		cards.Items[index].UpdatedAt = time.Now().Format(time.RFC3339)
		return cards
	}
	return cards
}

// removeMaterialIDFromStoryPlotCards 删除素材时清理剧情卡片里的反向引用，素材来源快照仍保存在素材自身。
func removeMaterialIDFromStoryPlotCards(cards StoryPlotCardCollection, materialID string) StoryPlotCardCollection {
	for index := range cards.Items {
		kept := make([]string, 0, len(cards.Items[index].MaterialIDs))
		for _, existing := range cards.Items[index].MaterialIDs {
			if existing != materialID {
				kept = append(kept, existing)
			}
		}
		cards.Items[index].MaterialIDs = kept
	}
	return cards
}

func excerptText(value string, limit int) string {
	clean := strings.Join(strings.Fields(value), " ")
	runes := []rune(clean)
	if len(runes) <= limit {
		return clean
	}
	return string(runes[:limit])
}

func refreshGlobalMaterialIndexForBook(bookPath string, project BookProject, materials MaterialCollection) error {
	index, err := loadGlobalMaterialIndex()
	if err != nil {
		return err
	}
	cleanBookPath := projectRootPath(bookPath)
	kept := make([]GlobalMaterialIndexItem, 0, len(index.Items)+len(materials.Items))
	for _, item := range index.Items {
		if item.BookID == project.Book.ID || samePath(item.BookPath, cleanBookPath) {
			continue
		}
		kept = append(kept, item)
	}
	now := time.Now().Format(time.RFC3339)
	for _, material := range materials.Items {
		kept = append(kept, GlobalMaterialIndexItem{
			BookID:       project.Book.ID,
			BookPath:     cleanBookPath,
			MaterialID:   material.ID,
			MaterialType: material.MaterialType,
			Snapshot: GlobalMaterialIndexSnapshot{
				Title:    material.Title,
				Summary:  material.Summary,
				BookName: project.Book.Name,
				TagNames: append([]string{}, material.TagIDs...),
			},
			IndexedAt: now,
		})
	}
	index.Version = 1
	index.Items = kept
	return saveGlobalMaterialIndex(index)
}

func loadGlobalMaterialIndex() (GlobalMaterialIndex, error) {
	index := GlobalMaterialIndex{Version: 1, Items: []GlobalMaterialIndexItem{}}
	path, err := globalMaterialIndexPath()
	if err != nil {
		return index, err
	}
	if err := readJSONFile(path, &index); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return index, nil
		}
		return index, fmt.Errorf("读取全局素材索引失败: %w", err)
	}
	if index.Version == 0 {
		index.Version = 1
	}
	if index.Items == nil {
		index.Items = []GlobalMaterialIndexItem{}
	}
	return index, nil
}

func saveGlobalMaterialIndex(index GlobalMaterialIndex) error {
	path, err := globalMaterialIndexPath()
	if err != nil {
		return err
	}
	return writeJSONFileAtomic(path, index)
}

func globalMaterialIndexPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("获取应用设置目录失败: %w", err)
	}
	return filepath.Join(configDir, appDataFolderName, globalMaterialIndexFileName), nil
}
