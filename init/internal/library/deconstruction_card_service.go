// deconstruction_card_service.go 负责选文来源片段和拆解卡片的创建与读取。
package library

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

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
		AnalysisKind:  plan.AnalysisKind,
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

// loadSourceTexts 读取选文来源片段集合。
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

// loadDeconstructionCards 读取拆解卡片集合。
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

// ensureChapterHasNoSourceTexts 阻止拆分已有来源引用的章节，避免破坏追溯偏移。
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

// findPlanType 在当前拆解方案中查找拆解类型。
func findPlanType(plan DeconstructionPlan, planTypeID string) (DeconstructionType, bool) {
	for _, planType := range plan.Types {
		if planType.ID == planTypeID {
			return planType, true
		}
	}
	return DeconstructionType{}, false
}

// findPlanDimension 在拆解类型中查找拆解维度。
func findPlanDimension(planType DeconstructionType, dimensionID string) (DeconstructionDimension, bool) {
	for _, dimension := range planType.Dimensions {
		if dimension.ID == dimensionID {
			return dimension, true
		}
	}
	return DeconstructionDimension{}, false
}

// findReusableSourceText 复用完全一致的选文来源，保证同一片段可挂多张拆解卡片。
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

// normalizeFieldValues 清理旧版字段值，保留历史卡片兼容数据。
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

// normalizeTags 去重并清理用户输入的标签。
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

// nextSequentialID 生成带日期和序号的本地数据 ID。
func nextSequentialID(prefix string, next int) string {
	return fmt.Sprintf("%s-%s-%04d", prefix, time.Now().Format("20060102"), next)
}
