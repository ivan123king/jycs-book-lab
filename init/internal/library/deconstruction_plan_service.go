// deconstruction_plan_service.go 负责单本小说拆解方案引用和模板目录管理。
package library

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// OpenDeconstructionPlanWorkspace 读取当前书的当前模板引用和可编辑模板列表。
func OpenDeconstructionPlanWorkspace(bookPath string) (DeconstructionPlanWorkspace, error) {
	project, err := OpenBookProject(bookPath)
	if err != nil {
		return DeconstructionPlanWorkspace{}, err
	}
	bookPath = projectRootPath(bookPath)
	if err := ensureDeconstructionPlanFiles(bookPath, project.Book.ID); err != nil {
		return DeconstructionPlanWorkspace{}, err
	}
	state, err := loadDeconstructionPlanState(bookPath, project.Book.ID)
	if err != nil {
		return DeconstructionPlanWorkspace{}, err
	}
	templates, err := listDeconstructionTemplates(bookPath, project.Book.ID)
	if err != nil {
		return DeconstructionPlanWorkspace{}, err
	}
	currentPlans := map[string]DeconstructionPlan{}
	currentTemplateFileNames := map[string]string{}
	for _, analysisKind := range supportedAnalysisKinds() {
		plan, err := loadDeconstructionPlanByKind(bookPath, project.Book.ID, analysisKind)
		if err != nil {
			return DeconstructionPlanWorkspace{}, err
		}
		currentPlans[analysisKind] = plan
		currentTemplateFileNames[analysisKind] = state.CurrentTemplates[analysisKind]
	}
	return DeconstructionPlanWorkspace{
		CurrentPlan:              currentPlans[analysisKindWritingTechnique],
		CurrentTemplateFileName:  currentTemplateFileNames[analysisKindWritingTechnique],
		CurrentPlanState:         state,
		CurrentPlans:             currentPlans,
		CurrentTemplateFileNames: currentTemplateFileNames,
		Templates:                templates,
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

// SaveDeconstructionPlan 兼容旧调用：将方案保存到该分析类型当前引用的模板文件。
func SaveDeconstructionPlan(request SaveDeconstructionPlanRequest) (DeconstructionPlanWorkspace, error) {
	project, err := OpenBookProject(request.BookPath)
	if err != nil {
		return DeconstructionPlanWorkspace{}, err
	}
	bookPath := projectRootPath(request.BookPath)
	if err := ensureDeconstructionPlanFiles(bookPath, project.Book.ID); err != nil {
		return DeconstructionPlanWorkspace{}, err
	}
	state, err := loadDeconstructionPlanState(bookPath, project.Book.ID)
	if err != nil {
		return DeconstructionPlanWorkspace{}, err
	}
	analysisKind := normalizeAnalysisKind(request.Plan.AnalysisKind)
	plan := normalizeDeconstructionPlan(request.Plan, project.Book.ID, analysisKind)
	if err := validateDeconstructionPlan(plan); err != nil {
		return DeconstructionPlanWorkspace{}, err
	}
	currentFileName := state.CurrentTemplates[analysisKind]
	if currentFileName == "" {
		currentFileName = defaultTemplateFileNameForKind(analysisKind)
		state.CurrentTemplates[analysisKind] = currentFileName
	}
	templatePath, err := safeTemplatePath(bookPath, currentFileName)
	if err != nil {
		return DeconstructionPlanWorkspace{}, err
	}
	if err := writeJSONFileAtomic(templatePath, plan); err != nil {
		return DeconstructionPlanWorkspace{}, fmt.Errorf("保存当前拆解模板失败: %w", err)
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
	analysisKind := normalizeAnalysisKind(firstNonEmpty(request.AnalysisKind, request.Plan.AnalysisKind))
	plan := normalizeDeconstructionPlan(request.Plan, project.Book.ID, analysisKind)
	plan.TemplateName = templateName
	if err := validateDeconstructionPlan(plan); err != nil {
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
	return OpenDeconstructionPlanWorkspace(bookPath)
}

// SetCurrentDeconstructionTemplate 只更新当前模板引用，不把模板内容复制进 deconstruction-plan.json。
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
	analysisKind := normalizeAnalysisKind(firstNonEmpty(request.AnalysisKind, plan.AnalysisKind))
	if plan.AnalysisKind != analysisKind {
		return DeconstructionPlanWorkspace{}, fmt.Errorf("模板类型不匹配，不能将“%s”模板设为“%s”当前模板", plan.AnalysisKind, analysisKind)
	}
	state, err := loadDeconstructionPlanState(bookPath, project.Book.ID)
	if err != nil {
		return DeconstructionPlanWorkspace{}, err
	}
	state.CurrentTemplates[analysisKind] = sanitizeTemplateFileName(request.FileName)
	state.UpdatedAt = time.Now().Format(time.RFC3339)
	if err := writeDeconstructionPlanState(bookPath, state); err != nil {
		return DeconstructionPlanWorkspace{}, fmt.Errorf("设置当前拆解模板失败: %w", err)
	}
	return OpenDeconstructionPlanWorkspace(bookPath)
}

// defaultDeconstructionPlanState 创建当前书默认的各类拆解路线模板引用。
func defaultDeconstructionPlanState(bookID string) DeconstructionPlanState {
	return DeconstructionPlanState{
		Version: 1,
		BookID:  bookID,
		CurrentTemplates: map[string]string{
			analysisKindWritingTechnique: systemWritingTemplateFileName,
			analysisKindStoryPlot:        systemStoryTemplateFileName,
		},
		UpdatedAt: time.Now().Format(time.RFC3339),
	}
}

// defaultWritingTechniqueTemplate 创建阅读选文拆解使用的系统内置写作手法模板。
func defaultWritingTechniqueTemplate(bookID string) DeconstructionPlan {
	now := time.Now().Format(time.RFC3339)
	return DeconstructionPlan{
		Version:      1,
		BookID:       bookID,
		AnalysisKind: analysisKindWritingTechnique,
		TemplateName: systemWritingTemplateName,
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

// defaultStoryPlotTemplate 创建暂供方案页维护的系统内置剧情分析模板。
func defaultStoryPlotTemplate(bookID string) DeconstructionPlan {
	now := time.Now().Format(time.RFC3339)
	return DeconstructionPlan{
		Version:      1,
		BookID:       bookID,
		AnalysisKind: analysisKindStoryPlot,
		TemplateName: systemStoryTemplateName,
		UpdatedAt:    now,
		Types: []DeconstructionType{
			{
				ID:    "plot",
				Name:  "剧情分析",
				Order: 1,
				Dimensions: []DeconstructionDimension{
					{ID: "core-event", Name: "核心事件", Order: 1},
					{ID: "plot-progress", Name: "剧情推进", Order: 2},
					{ID: "conflict-change", Name: "冲突变化", Order: 3},
					{ID: "character-goal", Name: "人物目标", Order: 4},
					{ID: "relationship-change", Name: "人物关系变化", Order: 5},
					{ID: "information-reveal", Name: "信息揭示", Order: 6},
					{ID: "suspense", Name: "悬念设置", Order: 7},
					{ID: "emotion-curve", Name: "情绪曲线", Order: 8},
					{ID: "foreshadowing-payoff", Name: "伏笔与回收", Order: 9},
					{ID: "plot-function", Name: "本段剧情作用", Order: 10},
				},
			},
		},
	}
}

// ensureDeconstructionPlanFiles 补齐模板目录、两类系统模板，并迁移旧版当前方案文件。
func ensureDeconstructionPlanFiles(bookPath string, bookID string) error {
	if err := os.MkdirAll(filepath.Join(bookPath, deconstructionTemplatesDir), 0o755); err != nil {
		return fmt.Errorf("创建拆解模板目录失败: %w", err)
	}
	currentPlanPath := filepath.Join(bookPath, filepath.FromSlash(deconstructionPlanRelativePath))
	state, err := readDeconstructionPlanStateFile(currentPlanPath, bookID)
	if errors.Is(err, os.ErrNotExist) {
		state = defaultDeconstructionPlanState(bookID)
		if err := writeDeconstructionPlanState(bookPath, state); err != nil {
			return err
		}
	} else if err != nil {
		// 旧项目的 deconstruction-plan.json 保存的是模板内容；迁移时先保留为写作手法模板，再改成引用状态。
		legacyPlan, legacyErr := readDeconstructionPlanFileWithKind(currentPlanPath, bookID, analysisKindWritingTechnique)
		if legacyErr != nil {
			return fmt.Errorf("读取当前拆解方案失败: %w", err)
		}
		legacyPlan.AnalysisKind = analysisKindWritingTechnique
		if strings.TrimSpace(legacyPlan.TemplateName) == "" {
			legacyPlan.TemplateName = systemWritingTemplateName
		}
		if err := writeTemplateIfNeeded(bookPath, systemWritingTemplateFileName, legacyPlan, true); err != nil {
			return err
		}
		state = defaultDeconstructionPlanState(bookID)
		if err := writeDeconstructionPlanState(bookPath, state); err != nil {
			return err
		}
	}
	state = normalizeDeconstructionPlanState(state, bookID)
	if err := writeTemplateIfNeeded(bookPath, systemWritingTemplateFileName, defaultWritingTechniqueTemplate(bookID), false); err != nil {
		return err
	}
	if err := writeTemplateIfNeeded(bookPath, systemStoryTemplateFileName, defaultStoryPlotTemplate(bookID), false); err != nil {
		return err
	}
	if err := writeDeconstructionPlanState(bookPath, state); err != nil {
		return err
	}
	return nil
}

// loadDeconstructionPlan 兼容阅读工作台旧调用，始终返回当前写作手法模板内容。
func loadDeconstructionPlan(bookPath string, bookID string) (DeconstructionPlan, error) {
	return loadDeconstructionPlanByKind(bookPath, bookID, analysisKindWritingTechnique)
}

// loadDeconstructionPlanByKind 按当前方案引用读取指定分析类型的模板内容。
func loadDeconstructionPlanByKind(bookPath string, bookID string, analysisKind string) (DeconstructionPlan, error) {
	state, err := loadDeconstructionPlanState(bookPath, bookID)
	if err != nil {
		return DeconstructionPlan{}, err
	}
	analysisKind = normalizeAnalysisKind(analysisKind)
	fileName := state.CurrentTemplates[analysisKind]
	if fileName == "" {
		fileName = defaultTemplateFileNameForKind(analysisKind)
	}
	templatePath, err := safeTemplatePath(bookPath, fileName)
	if err != nil {
		return DeconstructionPlan{}, err
	}
	plan, err := readDeconstructionPlanFileWithKind(templatePath, bookID, analysisKind)
	if err == nil {
		return plan, nil
	}
	defaultFileName := defaultTemplateFileNameForKind(analysisKind)
	if errors.Is(err, os.ErrNotExist) && !strings.EqualFold(fileName, defaultFileName) {
		defaultTemplatePath, pathErr := safeTemplatePath(bookPath, defaultFileName)
		if pathErr != nil {
			return DeconstructionPlan{}, pathErr
		}
		return readDeconstructionPlanFileWithKind(defaultTemplatePath, bookID, analysisKind)
	}
	return DeconstructionPlan{}, err
}

// readDeconstructionPlanFile 读取模板文件，并按模板自身类型归一化。
func readDeconstructionPlanFile(path string, bookID string) (DeconstructionPlan, error) {
	return readDeconstructionPlanFileWithKind(path, bookID, "")
}

// readDeconstructionPlanFileWithKind 读取模板文件，旧模板缺少类型时使用传入类型补齐。
func readDeconstructionPlanFileWithKind(path string, bookID string, fallbackAnalysisKind string) (DeconstructionPlan, error) {
	var plan DeconstructionPlan
	if err := readJSONFile(path, &plan); err != nil {
		return DeconstructionPlan{}, fmt.Errorf("读取拆解方案失败: %w", err)
	}
	return normalizeDeconstructionPlan(plan, bookID, fallbackAnalysisKind), nil
}

// loadDeconstructionPlanState 读取 deconstruction-plan.json 中的当前模板引用。
func loadDeconstructionPlanState(bookPath string, bookID string) (DeconstructionPlanState, error) {
	path := filepath.Join(bookPath, filepath.FromSlash(deconstructionPlanRelativePath))
	state, err := readDeconstructionPlanStateFile(path, bookID)
	if err != nil {
		return DeconstructionPlanState{}, fmt.Errorf("读取拆解方案引用失败: %w", err)
	}
	return normalizeDeconstructionPlanState(state, bookID), nil
}

// readDeconstructionPlanStateFile 读取方案引用文件，并识别旧版完整模板结构。
func readDeconstructionPlanStateFile(path string, bookID string) (DeconstructionPlanState, error) {
	var state DeconstructionPlanState
	if err := readJSONFile(path, &state); err != nil {
		return DeconstructionPlanState{}, err
	}
	if len(state.CurrentTemplates) == 0 {
		return DeconstructionPlanState{}, errors.New("旧版拆解方案不是模板引用结构")
	}
	return normalizeDeconstructionPlanState(state, bookID), nil
}

// writeDeconstructionPlanState 将当前模板引用写回单本小说项目目录。
func writeDeconstructionPlanState(bookPath string, state DeconstructionPlanState) error {
	state = normalizeDeconstructionPlanState(state, state.BookID)
	if err := writeJSONFileAtomic(filepath.Join(bookPath, filepath.FromSlash(deconstructionPlanRelativePath)), state); err != nil {
		return fmt.Errorf("保存拆解方案引用失败: %w", err)
	}
	return nil
}

// normalizeDeconstructionPlanState 补齐当前书缺失的分析类型引用和基础元数据。
func normalizeDeconstructionPlanState(state DeconstructionPlanState, bookID string) DeconstructionPlanState {
	if state.Version == 0 {
		state.Version = 1
	}
	state.BookID = bookID
	if state.CurrentTemplates == nil {
		state.CurrentTemplates = map[string]string{}
	}
	for _, analysisKind := range supportedAnalysisKinds() {
		if strings.TrimSpace(state.CurrentTemplates[analysisKind]) == "" {
			state.CurrentTemplates[analysisKind] = defaultTemplateFileNameForKind(analysisKind)
			continue
		}
		state.CurrentTemplates[analysisKind] = sanitizeTemplateFileName(state.CurrentTemplates[analysisKind])
	}
	if strings.TrimSpace(state.UpdatedAt) == "" {
		state.UpdatedAt = time.Now().Format(time.RFC3339)
	}
	return state
}

// normalizeDeconstructionPlan 统一模板版本、书籍 ID、分析类型、类型维度和旧字段迁移。
func normalizeDeconstructionPlan(plan DeconstructionPlan, bookID string, fallbackAnalysisKind string) DeconstructionPlan {
	if plan.Version == 0 {
		plan.Version = 1
	}
	plan.BookID = bookID
	plan.AnalysisKind = normalizeAnalysisKind(firstNonEmpty(plan.AnalysisKind, fallbackAnalysisKind))
	if strings.TrimSpace(plan.TemplateName) == "" {
		plan.TemplateName = defaultTemplateNameForKind(plan.AnalysisKind)
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

// validateDeconstructionPlan 校验模板至少保留可用于创建拆解的类型和维度。
func validateDeconstructionPlan(plan DeconstructionPlan) error {
	if strings.TrimSpace(plan.TemplateName) == "" {
		return errors.New("模板名称不能为空")
	}
	if normalizeAnalysisKind(plan.AnalysisKind) != plan.AnalysisKind {
		return fmt.Errorf("不支持的拆解模板类型: %s", plan.AnalysisKind)
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

// supportedAnalysisKinds 返回第一阶段支持维护的拆解路线类型。
func supportedAnalysisKinds() []string {
	return []string{analysisKindWritingTechnique, analysisKindStoryPlot}
}

// normalizeAnalysisKind 将空值或未知值归一到写作手法路线，保持旧数据可打开。
func normalizeAnalysisKind(value string) string {
	value = strings.TrimSpace(value)
	for _, analysisKind := range supportedAnalysisKinds() {
		if value == analysisKind {
			return analysisKind
		}
	}
	return analysisKindWritingTechnique
}

// defaultTemplateFileNameForKind 返回指定分析类型的系统内置模板文件名。
func defaultTemplateFileNameForKind(analysisKind string) string {
	switch normalizeAnalysisKind(analysisKind) {
	case analysisKindStoryPlot:
		return systemStoryTemplateFileName
	default:
		return systemWritingTemplateFileName
	}
}

// defaultTemplateNameForKind 返回指定分析类型的系统内置模板展示名。
func defaultTemplateNameForKind(analysisKind string) string {
	switch normalizeAnalysisKind(analysisKind) {
	case analysisKindStoryPlot:
		return systemStoryTemplateName
	default:
		return systemWritingTemplateName
	}
}

// analysisKindOrder 保证模板列表按写作模板、剧情模板的顺序稳定展示。
func analysisKindOrder(analysisKind string) int {
	switch normalizeAnalysisKind(analysisKind) {
	case analysisKindWritingTechnique:
		return 1
	case analysisKindStoryPlot:
		return 2
	default:
		return 99
	}
}

// writeTemplateIfNeeded 在模板缺失或迁移覆盖时写入指定模板文件。
func writeTemplateIfNeeded(bookPath string, fileName string, plan DeconstructionPlan, overwrite bool) error {
	templatePath, err := safeTemplatePath(bookPath, fileName)
	if err != nil {
		return err
	}
	if _, err := os.Stat(templatePath); err == nil && !overwrite {
		return nil
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("检查拆解模板失败: %w", err)
	}
	if err := writeJSONFileAtomic(templatePath, plan); err != nil {
		return fmt.Errorf("保存拆解模板失败: %w", err)
	}
	return nil
}

// listDeconstructionTemplates 读取当前书模板目录并生成方案页列表。
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
			Name:         plan.TemplateName,
			FileName:     entry.Name(),
			AnalysisKind: plan.AnalysisKind,
			UpdatedAt:    plan.UpdatedAt,
		})
	}
	sort.SliceStable(templates, func(i, j int) bool {
		if analysisKindOrder(templates[i].AnalysisKind) != analysisKindOrder(templates[j].AnalysisKind) {
			return analysisKindOrder(templates[i].AnalysisKind) < analysisKindOrder(templates[j].AnalysisKind)
		}
		if templates[i].FileName == defaultTemplateFileNameForKind(templates[i].AnalysisKind) {
			return true
		}
		if templates[j].FileName == defaultTemplateFileNameForKind(templates[j].AnalysisKind) {
			return false
		}
		return strings.ToLower(templates[i].Name) < strings.ToLower(templates[j].Name)
	})
	return templates, nil
}

// safeTemplatePath 将模板文件名限制在当前书模板目录内。
func safeTemplatePath(bookPath string, fileName string) (string, error) {
	cleanName := sanitizeTemplateFileName(fileName)
	if cleanName == "" {
		return "", errors.New("模板文件名不能为空")
	}
	return safeProjectPath(bookPath, filepath.ToSlash(filepath.Join(deconstructionTemplatesDir, cleanName)))
}

// sanitizeTemplateFileName 清理模板文件名，避免非法字符或路径穿越。
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

// normalizeStableID 生成稳定且不重复的类型或维度 ID。
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

// stableIDFromName 为缺失 ID 的旧数据生成顺序兜底 ID。
func stableIDFromName(name string, prefix string, fallbackOrder int) string {
	return fmt.Sprintf("%s-%02d", prefix, fallbackOrder)
}

// sanitizeID 清理结构标识，只保留前端和 JSON 稳定使用的字符。
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

// firstNonEmpty 返回第一个非空字符串，用于请求参数和旧数据字段的兜底选择。
func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
