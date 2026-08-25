// story_plot_service.go 负责剧情拆解卡片的章节范围解析、临时类型维度快照和本地 JSON 持久化。
package library

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	storyPlotStatusDraft     = "draft"
	storyPlotStatusCompleted = "completed"
)

var storyChapterRangeSegmentPattern = regexp.MustCompile(`^(\d+)(?:\s*-\s*(\d+))?$`)

// OpenStoryPlotWorkspace 读取剧情页需要的当前书、当前剧情模板和已有剧情卡片。
func OpenStoryPlotWorkspace(bookPath string) (StoryPlotWorkspace, error) {
	project, err := OpenBookProject(bookPath)
	if err != nil {
		return StoryPlotWorkspace{}, err
	}
	bookPath = projectRootPath(bookPath)
	if err := ensureReadingSupportFiles(bookPath, project); err != nil {
		return StoryPlotWorkspace{}, err
	}
	state, err := loadDeconstructionPlanState(bookPath, project.Book.ID)
	if err != nil {
		return StoryPlotWorkspace{}, err
	}
	plan, err := loadDeconstructionPlanByKind(bookPath, project.Book.ID, analysisKindStoryPlot)
	if err != nil {
		return StoryPlotWorkspace{}, err
	}
	cards, err := loadStoryPlotCards(bookPath)
	if err != nil {
		return StoryPlotWorkspace{}, err
	}
	return StoryPlotWorkspace{
		Project:                 project,
		Plan:                    plan,
		CurrentTemplateFileName: state.CurrentTemplates[analysisKindStoryPlot],
		StoryPlotCards:          cards,
	}, nil
}

// CreateStoryPlotCard 保存一张新的剧情卡片；模板只作为初始快照，不被回写。
func CreateStoryPlotCard(request SaveStoryPlotCardRequest) (SaveStoryPlotCardResult, error) {
	project, err := OpenBookProject(request.BookPath)
	if err != nil {
		return SaveStoryPlotCardResult{}, err
	}
	bookPath := projectRootPath(request.BookPath)
	if err := ensureReadingSupportFiles(bookPath, project); err != nil {
		return SaveStoryPlotCardResult{}, err
	}
	cards, err := loadStoryPlotCards(bookPath)
	if err != nil {
		return SaveStoryPlotCardResult{}, err
	}
	now := time.Now().Format(time.RFC3339)
	card, err := buildStoryPlotCardFromRequest(project, request, now, now, true)
	if err != nil {
		return SaveStoryPlotCardResult{}, err
	}
	card.ID = nextStoryPlotCardID(cards)
	cards.Items = append(cards.Items, card)
	if err := writeStoryPlotCards(bookPath, cards); err != nil {
		return SaveStoryPlotCardResult{}, err
	}
	return SaveStoryPlotCardResult{Card: card, StoryPlotCards: cards}, nil
}

// UpdateStoryPlotCard 更新已有剧情卡片，不改变创建时间和卡片 ID。
func UpdateStoryPlotCard(request SaveStoryPlotCardRequest) (SaveStoryPlotCardResult, error) {
	project, err := OpenBookProject(request.BookPath)
	if err != nil {
		return SaveStoryPlotCardResult{}, err
	}
	bookPath := projectRootPath(request.BookPath)
	if err := ensureReadingSupportFiles(bookPath, project); err != nil {
		return SaveStoryPlotCardResult{}, err
	}
	cards, err := loadStoryPlotCards(bookPath)
	if err != nil {
		return SaveStoryPlotCardResult{}, err
	}
	for index := range cards.Items {
		if cards.Items[index].ID != request.CardID {
			continue
		}
		card, err := buildStoryPlotCardFromRequest(project, request, cards.Items[index].CreatedAt, time.Now().Format(time.RFC3339), false)
		if err != nil {
			return SaveStoryPlotCardResult{}, err
		}
		card.ID = cards.Items[index].ID
		card.MaterialIDs = append([]string{}, cards.Items[index].MaterialIDs...)
		cards.Items[index] = card
		if err := writeStoryPlotCards(bookPath, cards); err != nil {
			return SaveStoryPlotCardResult{}, err
		}
		return SaveStoryPlotCardResult{Card: card, StoryPlotCards: cards}, nil
	}
	return SaveStoryPlotCardResult{}, fmt.Errorf("剧情卡片不存在: %s", request.CardID)
}

// DeleteStoryPlotCard 删除当前书中的剧情卡片；已沉淀素材后续接入时应保留来源快照。
func DeleteStoryPlotCard(request DeleteStoryPlotCardRequest) (DeleteStoryPlotCardResult, error) {
	project, err := OpenBookProject(request.BookPath)
	if err != nil {
		return DeleteStoryPlotCardResult{}, err
	}
	bookPath := projectRootPath(request.BookPath)
	if err := ensureReadingSupportFiles(bookPath, project); err != nil {
		return DeleteStoryPlotCardResult{}, err
	}
	cards, err := loadStoryPlotCards(bookPath)
	if err != nil {
		return DeleteStoryPlotCardResult{}, err
	}
	kept := make([]StoryPlotCard, 0, len(cards.Items))
	removed := false
	for _, card := range cards.Items {
		if card.ID == request.CardID {
			removed = true
			continue
		}
		kept = append(kept, card)
	}
	if !removed {
		return DeleteStoryPlotCardResult{}, fmt.Errorf("剧情卡片不存在: %s", request.CardID)
	}
	cards.Items = kept
	if err := writeStoryPlotCards(bookPath, cards); err != nil {
		return DeleteStoryPlotCardResult{}, err
	}
	return DeleteStoryPlotCardResult{StoryPlotCards: cards}, nil
}

// loadStoryPlotCards 读取剧情卡片集合；缺失文件视为空集合，便于旧项目打开。
func loadStoryPlotCards(bookPath string) (StoryPlotCardCollection, error) {
	collection := StoryPlotCardCollection{Version: 1, Items: []StoryPlotCard{}}
	err := readJSONFile(filepath.Join(bookPath, filepath.FromSlash(storyPlotCardsRelativePath)), &collection)
	if errors.Is(err, os.ErrNotExist) {
		return collection, nil
	}
	if err != nil {
		return collection, fmt.Errorf("读取剧情卡片失败: %w", err)
	}
	return normalizeStoryPlotCardCollection(collection), nil
}

func writeStoryPlotCards(bookPath string, cards StoryPlotCardCollection) error {
	if err := writeJSONFileAtomic(filepath.Join(bookPath, filepath.FromSlash(storyPlotCardsRelativePath)), normalizeStoryPlotCardCollection(cards)); err != nil {
		return fmt.Errorf("保存剧情卡片失败: %w", err)
	}
	return nil
}

func buildStoryPlotCardFromRequest(project BookProject, request SaveStoryPlotCardRequest, createdAt string, updatedAt string, isCreate bool) (StoryPlotCard, error) {
	status := normalizeStoryPlotStatus(request.Status)
	refs, parseErrors := parseStoryChapterRange(project, request.ChapterRangeInput, status == storyPlotStatusCompleted)
	if status == storyPlotStatusCompleted {
		if len(parseErrors) > 0 {
			return StoryPlotCard{}, errors.New(parseErrors[0])
		}
		if strings.TrimSpace(request.Title) == "" {
			return StoryPlotCard{}, errors.New("剧情名称不能为空")
		}
	}
	if !isCreate && strings.TrimSpace(request.CardID) == "" {
		return StoryPlotCard{}, errors.New("剧情卡片 ID 不能为空")
	}
	typeSnapshots := normalizeStoryPlotTypeSnapshots(request.TypeSnapshots)
	dimensions := normalizeStoryPlotDimensions(request.Dimensions)
	planTypeID := strings.TrimSpace(request.PlanTypeID)
	planTypeName := strings.TrimSpace(request.PlanTypeName)
	if planTypeID == "" && len(typeSnapshots) > 0 {
		planTypeID = typeSnapshots[0].ID
	}
	if planTypeName == "" {
		planTypeName = storyPlotTypeName(typeSnapshots, planTypeID)
	}
	if len(dimensions) == 0 {
		dimensions = storyPlotDimensionsForType(typeSnapshots, planTypeID)
	}
	if planTypeName == "" {
		planTypeName = "剧情分析"
	}
	if len(typeSnapshots) == 0 {
		typeSnapshots = []StoryPlotTypeSnapshot{
			{
				ID:         firstNonEmpty(planTypeID, "plot"),
				Name:       planTypeName,
				Order:      1,
				Dimensions: dimensions,
			},
		}
	}
	return StoryPlotCard{
		BookID:            project.Book.ID,
		AnalysisKind:      analysisKindStoryPlot,
		Status:            status,
		Title:             strings.TrimSpace(request.Title),
		Summary:           strings.TrimRight(request.Summary, "\r\n"),
		ChapterRangeInput: strings.TrimSpace(request.ChapterRangeInput),
		ChapterRefs:       refs,
		ParseErrors:       parseErrors,
		TemplateFileName:  strings.TrimSpace(request.TemplateFileName),
		TemplateName:      strings.TrimSpace(request.TemplateName),
		PlanTypeID:        planTypeID,
		PlanTypeName:      planTypeName,
		Dimensions:        dimensions,
		TypeSnapshots:     typeSnapshots,
		TagIDs:            normalizeTags(request.TagIDs),
		MaterialIDs:       []string{},
		CreatedAt:         createdAt,
		UpdatedAt:         updatedAt,
	}, nil
}

// parseStoryChapterRange 将 1-3, 8, 12-15 解析为可追溯章节引用。
func parseStoryChapterRange(project BookProject, input string, required bool) ([]StoryPlotChapterRef, []string) {
	cleanInput := strings.TrimSpace(input)
	if cleanInput == "" {
		if required {
			return []StoryPlotChapterRef{}, []string{"请输入章节范围，例如：1-3, 8, 12-15"}
		}
		return []StoryPlotChapterRef{}, []string{}
	}
	chaptersByOrder := map[int]BookChapter{}
	for _, chapter := range project.Chapters {
		chaptersByOrder[chapter.Order] = chapter
	}
	orderSet := map[int]bool{}
	parseErrors := []string{}
	for _, rawSegment := range strings.FieldsFunc(cleanInput, func(value rune) bool {
		return value == ',' || value == '，'
	}) {
		segment := strings.TrimSpace(rawSegment)
		if segment == "" {
			continue
		}
		match := storyChapterRangeSegmentPattern.FindStringSubmatch(segment)
		if match == nil {
			parseErrors = append(parseErrors, fmt.Sprintf("章节范围 %q 格式不正确", segment))
			continue
		}
		startOrder, _ := strconv.Atoi(match[1])
		endOrder := startOrder
		if match[2] != "" {
			endOrder, _ = strconv.Atoi(match[2])
		}
		if startOrder > endOrder {
			parseErrors = append(parseErrors, fmt.Sprintf("章节范围 %q 的起始章节不能大于结束章节", segment))
			continue
		}
		for order := startOrder; order <= endOrder; order++ {
			if _, ok := chaptersByOrder[order]; !ok {
				parseErrors = append(parseErrors, fmt.Sprintf("第 %d 章不存在", order))
				continue
			}
			orderSet[order] = true
		}
	}
	orders := make([]int, 0, len(orderSet))
	for order := range orderSet {
		orders = append(orders, order)
	}
	sort.Ints(orders)
	refs := make([]StoryPlotChapterRef, 0, len(orders))
	for _, order := range orders {
		chapter := chaptersByOrder[order]
		refs = append(refs, StoryPlotChapterRef{
			ChapterID:            chapter.ID,
			ChapterOrder:         chapter.Order,
			ChapterTitleSnapshot: chapter.Title,
		})
	}
	if required && len(refs) == 0 && len(parseErrors) == 0 {
		parseErrors = append(parseErrors, "请输入有效的章节序号")
	}
	return refs, parseErrors
}

func normalizeStoryPlotStatus(status string) string {
	if strings.TrimSpace(status) == storyPlotStatusCompleted {
		return storyPlotStatusCompleted
	}
	return storyPlotStatusDraft
}

func normalizeStoryPlotCardCollection(collection StoryPlotCardCollection) StoryPlotCardCollection {
	if collection.Version == 0 {
		collection.Version = 1
	}
	if collection.Items == nil {
		collection.Items = []StoryPlotCard{}
		return collection
	}
	for index := range collection.Items {
		collection.Items[index] = normalizeStoryPlotCard(collection.Items[index])
	}
	return collection
}

func normalizeStoryPlotCard(card StoryPlotCard) StoryPlotCard {
	card.AnalysisKind = analysisKindStoryPlot
	card.Status = normalizeStoryPlotStatus(card.Status)
	if card.ChapterRefs == nil {
		card.ChapterRefs = []StoryPlotChapterRef{}
	}
	if card.ParseErrors == nil {
		card.ParseErrors = []string{}
	}
	card.Dimensions = normalizeStoryPlotDimensions(card.Dimensions)
	card.TypeSnapshots = normalizeStoryPlotTypeSnapshots(card.TypeSnapshots)
	if card.TagIDs == nil {
		card.TagIDs = []string{}
	}
	if card.MaterialIDs == nil {
		card.MaterialIDs = []string{}
	}
	return card
}

func normalizeStoryPlotTypeSnapshots(values []StoryPlotTypeSnapshot) []StoryPlotTypeSnapshot {
	if values == nil {
		return []StoryPlotTypeSnapshot{}
	}
	normalized := make([]StoryPlotTypeSnapshot, 0, len(values))
	usedIDs := map[string]int{}
	for index, item := range values {
		name := strings.TrimSpace(item.Name)
		if name == "" {
			name = fmt.Sprintf("剧情类型 %d", index+1)
		}
		id := stableStoryPlotID(item.ID, "story-type", index+1, usedIDs)
		normalized = append(normalized, StoryPlotTypeSnapshot{
			ID:         id,
			Name:       name,
			Order:      index + 1,
			Dimensions: normalizeStoryPlotDimensions(item.Dimensions),
		})
	}
	return normalized
}

func normalizeStoryPlotDimensions(values []StoryPlotDimension) []StoryPlotDimension {
	if values == nil {
		return []StoryPlotDimension{}
	}
	normalized := make([]StoryPlotDimension, 0, len(values))
	usedIDs := map[string]int{}
	for index, item := range values {
		name := strings.TrimSpace(item.Name)
		if name == "" {
			name = fmt.Sprintf("剧情维度 %d", index+1)
		}
		normalized = append(normalized, StoryPlotDimension{
			ID:      stableStoryPlotID(item.ID, "story-dimension", index+1, usedIDs),
			Name:    name,
			Order:   index + 1,
			Content: strings.TrimRight(item.Content, "\r\n"),
		})
	}
	return normalized
}

func stableStoryPlotID(value string, prefix string, order int, used map[string]int) string {
	cleanID := strings.TrimSpace(value)
	if cleanID == "" {
		cleanID = fmt.Sprintf("%s-%d", prefix, order)
	}
	if count := used[cleanID]; count > 0 {
		used[cleanID] = count + 1
		return fmt.Sprintf("%s-%d", cleanID, count+1)
	}
	used[cleanID] = 1
	return cleanID
}

func storyPlotTypeName(types []StoryPlotTypeSnapshot, typeID string) string {
	for _, item := range types {
		if item.ID == typeID {
			return item.Name
		}
	}
	return ""
}

func storyPlotDimensionsForType(types []StoryPlotTypeSnapshot, typeID string) []StoryPlotDimension {
	for _, item := range types {
		if item.ID == typeID {
			return normalizeStoryPlotDimensions(item.Dimensions)
		}
	}
	return []StoryPlotDimension{}
}

func nextStoryPlotCardID(cards StoryPlotCardCollection) string {
	existing := map[string]bool{}
	for _, card := range cards.Items {
		existing[card.ID] = true
	}
	for next := len(cards.Items) + 1; ; next++ {
		id := nextSequentialID("story-plot", next)
		if !existing[id] {
			return id
		}
	}
}
