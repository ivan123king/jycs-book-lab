// Package library 定义书库、导入预览和单本小说项目的本地数据服务边界。
package library

// AppSettings 保存应用级偏好，只记录与具体书籍无关的全局配置。
type AppSettings struct {
	Version            int    `json:"version"`
	LibraryPath        string `json:"libraryPath"`
	LastOpenedBookPath string `json:"lastOpenedBookPath"`
}

// LibraryBookEntry 是书库 index.json 中用于快速展示书籍列表的轻量入口。
type LibraryBookEntry struct {
	BookName               string `json:"bookName"`
	BookPath               string `json:"bookPath"`
	BookSize               int64  `json:"bookSize"`
	ReadingProgressPercent int    `json:"readingProgressPercent"`
}

// LibraryIndex 对应书库根目录的 index.json，不保存章节、拆解或素材详情。
type LibraryIndex struct {
	Version int                `json:"version"`
	Books   []LibraryBookEntry `json:"books"`
}

// LibraryState 返回前端展示书库页所需的当前目录和轻量索引。
type LibraryState struct {
	LibraryPath string       `json:"libraryPath"`
	Index       LibraryIndex `json:"index"`
}

// BookMeta 保存单本小说 book.json 中的书籍基础信息。
type BookMeta struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	SourceFileName string `json:"sourceFileName"`
	SourceTextPath string `json:"sourceTextPath"`
	BookPath       string `json:"bookPath"`
	BookSize       int64  `json:"bookSize"`
	CreatedAt      string `json:"createdAt"`
	UpdatedAt      string `json:"updatedAt"`
}

// BookChapter 保存章节标题、正文 TXT 相对路径和原文位置，后续用于来源追溯。
type BookChapter struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Order       int    `json:"order"`
	ContentPath string `json:"contentPath"`
	StartOffset int    `json:"startOffset"`
	EndOffset   int    `json:"endOffset"`
	WordCount   int    `json:"wordCount"`
}

// BookProject 对应单本小说目录中的 book.json。
type BookProject struct {
	Version  int           `json:"version"`
	Book     BookMeta      `json:"book"`
	Chapters []BookChapter `json:"chapters"`
}

// ImportChapterPreview 是导入预览阶段展示给用户快速检查的章节标题行。
type ImportChapterPreview struct {
	ID          string `json:"id"`
	Order       int    `json:"order"`
	Title       string `json:"title"`
	StartOffset int    `json:"startOffset"`
	EndOffset   int    `json:"endOffset"`
	WordCount   int    `json:"wordCount"`
}

// ImportTimingStage 记录导入链路中某个阶段的耗时，用于定位本地文件读取、章节识别或落盘卡顿。
type ImportTimingStage struct {
	Name       string `json:"name"`
	Label      string `json:"label"`
	DurationMs int64  `json:"durationMs"`
}

// ImportDiagnostics 是导入预览和确认导入的轻量诊断信息，只通过 API 返回，不写入正式 book.json。
type ImportDiagnostics struct {
	Action       string              `json:"action"`
	TotalMs      int64               `json:"totalMs"`
	BookSize     int64               `json:"bookSize"`
	ChapterCount int                 `json:"chapterCount"`
	Stages       []ImportTimingStage `json:"stages"`
}

// ImportDraft 是 TXT 小说被确认入库前的预览结果，正式数据尚未写入书库。
type ImportDraft struct {
	BookID       string                 `json:"bookId"`
	BookName     string                 `json:"bookName"`
	Description  string                 `json:"description"`
	SourcePath   string                 `json:"sourcePath"`
	LibraryPath  string                 `json:"libraryPath"`
	TargetPath   string                 `json:"targetPath"`
	BookSize     int64                  `json:"bookSize"`
	ChapterCount int                    `json:"chapterCount"`
	Chapters     []ImportChapterPreview `json:"chapters"`
	Diagnostics  ImportDiagnostics      `json:"diagnostics"`
}

// ConfirmImportRequest 承载用户在导入面板确认后的小说名、简介和预览来源。
type ConfirmImportRequest struct {
	BookID      string `json:"bookId"`
	SourcePath  string `json:"sourcePath"`
	BookName    string `json:"bookName"`
	Description string `json:"description"`
}

// ConfirmImportResult 返回确认导入后的书库索引和新建小说项目。
type ConfirmImportResult struct {
	Library     LibraryState      `json:"library"`
	Project     BookProject       `json:"project"`
	Diagnostics ImportDiagnostics `json:"diagnostics"`
}

// ReadingPosition 保存当前书最近阅读到的章节和正文偏移。
type ReadingPosition struct {
	ChapterID          string `json:"chapterId"`
	ChapterOrder       int    `json:"chapterOrder"`
	CharOffset         int    `json:"charOffset"`
	ParagraphIndex     int    `json:"paragraphIndex"`
	RenderedPageNumber int    `json:"renderedPageNumber"`
	UpdatedAt          string `json:"updatedAt"`
}

// ReadingPreferences 保存单本小说自己的阅读界面偏好，随小说项目目录迁移。
type ReadingPreferences struct {
	FontSize int `json:"fontSize"`
}

// ReadingState 是单本小说 reading/reading-state.json 的正式阅读位置数据。
type ReadingState struct {
	Version                int                `json:"version"`
	BookID                 string             `json:"bookId"`
	LastPosition           ReadingPosition    `json:"lastPosition"`
	LastMode               string             `json:"lastMode"`
	ReadingProgressPercent int                `json:"readingProgressPercent"`
	Preferences            ReadingPreferences `json:"preferences"`
}

// DeconstructionField 是旧版拆解方案字段结构，仅用于兼容旧项目迁移。
type DeconstructionField struct {
	ID       string   `json:"id"`
	Label    string   `json:"label"`
	Type     string   `json:"type"`
	Order    int      `json:"order"`
	Required bool     `json:"required"`
	Visible  bool     `json:"visible"`
	Options  []string `json:"options"`
}

// DeconstructionDimension 定义某个拆解类型下可选择的分析维度。
type DeconstructionDimension struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Order int    `json:"order"`
}

// DeconstructionType 定义当前书拆解方案中的一种拆解类型。
type DeconstructionType struct {
	ID         string                    `json:"id"`
	Name       string                    `json:"name"`
	Order      int                       `json:"order"`
	Dimensions []DeconstructionDimension `json:"dimensions"`
	Dimension  string                    `json:"dimension,omitempty"`
	Fields     []DeconstructionField     `json:"fields,omitempty"`
}

// DeconstructionPlanState 保存当前书各类拆解路线正在使用的模板引用。
type DeconstructionPlanState struct {
	Version          int               `json:"version"`
	BookID           string            `json:"bookId"`
	CurrentTemplates map[string]string `json:"currentTemplates"`
	UpdatedAt        string            `json:"updatedAt"`
}

// DeconstructionPlan 保存每本书独立的拆解模板配置。
type DeconstructionPlan struct {
	Version      int                  `json:"version"`
	BookID       string               `json:"bookId"`
	AnalysisKind string               `json:"analysisKind"`
	TemplateName string               `json:"templateName"`
	Types        []DeconstructionType `json:"types"`
	UpdatedAt    string               `json:"updatedAt"`
}

// DeconstructionTemplateSummary 是方案页模板列表中的轻量展示项。
type DeconstructionTemplateSummary struct {
	Name         string `json:"name"`
	FileName     string `json:"fileName"`
	AnalysisKind string `json:"analysisKind"`
	UpdatedAt    string `json:"updatedAt"`
}

// DeconstructionPlanWorkspace 返回方案页编辑所需的当前方案和模板列表。
type DeconstructionPlanWorkspace struct {
	CurrentPlan              DeconstructionPlan              `json:"currentPlan"`
	CurrentTemplateFileName  string                          `json:"currentTemplateFileName"`
	CurrentPlanState         DeconstructionPlanState         `json:"currentPlanState"`
	CurrentPlans             map[string]DeconstructionPlan   `json:"currentPlans"`
	CurrentTemplateFileNames map[string]string               `json:"currentTemplateFileNames"`
	Templates                []DeconstructionTemplateSummary `json:"templates"`
}

// SaveDeconstructionPlanRequest 保存当前书正在使用的拆解方案。
type SaveDeconstructionPlanRequest struct {
	BookPath string             `json:"bookPath"`
	Plan     DeconstructionPlan `json:"plan"`
}

// SaveDeconstructionTemplateRequest 保存或另存当前书模板目录中的模板文件。
type SaveDeconstructionTemplateRequest struct {
	BookPath     string             `json:"bookPath"`
	AnalysisKind string             `json:"analysisKind"`
	TemplateName string             `json:"templateName"`
	FileName     string             `json:"fileName"`
	Plan         DeconstructionPlan `json:"plan"`
	Overwrite    bool               `json:"overwrite"`
}

// SetCurrentDeconstructionTemplateRequest 将模板目录中的某个模板设为某类分析的当前模板。
type SetCurrentDeconstructionTemplateRequest struct {
	BookPath     string `json:"bookPath"`
	AnalysisKind string `json:"analysisKind"`
	FileName     string `json:"fileName"`
}

// SourceText 是从章节正文中选出的原文片段快照，可被多张拆解卡片引用。
type SourceText struct {
	ID           string `json:"id"`
	BookID       string `json:"bookId"`
	ChapterID    string `json:"chapterId"`
	ChapterOrder int    `json:"chapterOrder"`
	StartOffset  int    `json:"startOffset"`
	EndOffset    int    `json:"endOffset"`
	SelectedText string `json:"selectedText"`
	CreatedAt    string `json:"createdAt"`
}

// SourceTextCollection 对应 source-texts/source-texts.json。
type SourceTextCollection struct {
	Version int          `json:"version"`
	Items   []SourceText `json:"items"`
}

// DeconstructionCard 保存用户从原文片段创建的结构化拆解卡片。
type DeconstructionCard struct {
	ID            string            `json:"id"`
	BookID        string            `json:"bookId"`
	SourceTextID  string            `json:"sourceTextId"`
	AnalysisKind  string            `json:"analysisKind"`
	PlanTypeID    string            `json:"planTypeId"`
	PlanTypeName  string            `json:"planTypeName"`
	DimensionID   string            `json:"dimensionId"`
	DimensionName string            `json:"dimensionName"`
	Content       string            `json:"content"`
	FieldValues   map[string]string `json:"fieldValues"`
	TagIDs        []string          `json:"tagIds"`
	MaterialIDs   []string          `json:"materialIds"`
	CreatedAt     string            `json:"createdAt"`
	UpdatedAt     string            `json:"updatedAt"`
}

// DeconstructionCardCollection 对应 deconstructions/writing-technique-cards.json。
type DeconstructionCardCollection struct {
	Version int                  `json:"version"`
	Items   []DeconstructionCard `json:"items"`
}

// DeconstructionCardsWorkspace 返回写作手法拆解管理页所需的项目、来源和卡片数据。
type DeconstructionCardsWorkspace struct {
	Project             BookProject                  `json:"project"`
	Plan                DeconstructionPlan           `json:"plan"`
	SourceTexts         SourceTextCollection         `json:"sourceTexts"`
	DeconstructionCards DeconstructionCardCollection `json:"deconstructionCards"`
}

// StoryPlotChapterRef 保存剧情拆解引用章节的可跳转 ID 和标题快照。
type StoryPlotChapterRef struct {
	ChapterID            string `json:"chapterId"`
	ChapterOrder         int    `json:"chapterOrder"`
	ChapterTitleSnapshot string `json:"chapterTitleSnapshot"`
}

// StoryPlotDimension 保存剧情卡片内某个临时维度的名称、顺序和用户分析正文。
type StoryPlotDimension struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Order   int    `json:"order"`
	Content string `json:"content"`
}

// StoryPlotTypeSnapshot 保存某张剧情卡片自己的临时类型和维度列表，不回写模板。
type StoryPlotTypeSnapshot struct {
	ID         string               `json:"id"`
	Name       string               `json:"name"`
	Order      int                  `json:"order"`
	Dimensions []StoryPlotDimension `json:"dimensions"`
}

// StoryPlotCard 保存用户围绕章节范围整理出的剧情拆解卡片。
type StoryPlotCard struct {
	ID                string                  `json:"id"`
	BookID            string                  `json:"bookId"`
	AnalysisKind      string                  `json:"analysisKind"`
	Status            string                  `json:"status"`
	Title             string                  `json:"title"`
	Summary           string                  `json:"summary"`
	ChapterRangeInput string                  `json:"chapterRangeInput"`
	ChapterRefs       []StoryPlotChapterRef   `json:"chapterRefs"`
	ParseErrors       []string                `json:"parseErrors"`
	TemplateFileName  string                  `json:"templateFileName"`
	TemplateName      string                  `json:"templateName"`
	PlanTypeID        string                  `json:"planTypeId"`
	PlanTypeName      string                  `json:"planTypeName"`
	Dimensions        []StoryPlotDimension    `json:"dimensions"`
	TypeSnapshots     []StoryPlotTypeSnapshot `json:"typeSnapshots"`
	TagIDs            []string                `json:"tagIds"`
	MaterialIDs       []string                `json:"materialIds"`
	CreatedAt         string                  `json:"createdAt"`
	UpdatedAt         string                  `json:"updatedAt"`
}

// StoryPlotCardCollection 对应 deconstructions/story-plot-cards.json。
type StoryPlotCardCollection struct {
	Version int             `json:"version"`
	Items   []StoryPlotCard `json:"items"`
}

// StoryPlotWorkspace 返回剧情页需要的当前书、当前剧情模板和已有剧情卡片。
type StoryPlotWorkspace struct {
	Project                 BookProject             `json:"project"`
	Plan                    DeconstructionPlan      `json:"plan"`
	CurrentTemplateFileName string                  `json:"currentTemplateFileName"`
	StoryPlotCards          StoryPlotCardCollection `json:"storyPlotCards"`
}

// SaveStoryPlotCardRequest 承载剧情弹框保存草稿或正式完成的数据。
type SaveStoryPlotCardRequest struct {
	BookPath          string                  `json:"bookPath"`
	CardID            string                  `json:"cardId"`
	Status            string                  `json:"status"`
	Title             string                  `json:"title"`
	Summary           string                  `json:"summary"`
	ChapterRangeInput string                  `json:"chapterRangeInput"`
	TemplateFileName  string                  `json:"templateFileName"`
	TemplateName      string                  `json:"templateName"`
	PlanTypeID        string                  `json:"planTypeId"`
	PlanTypeName      string                  `json:"planTypeName"`
	Dimensions        []StoryPlotDimension    `json:"dimensions"`
	TypeSnapshots     []StoryPlotTypeSnapshot `json:"typeSnapshots"`
	TagIDs            []string                `json:"tagIds"`
}

// SaveStoryPlotCardResult 返回保存后的剧情卡片和完整集合，便于前端刷新列表。
type SaveStoryPlotCardResult struct {
	Card           StoryPlotCard           `json:"card"`
	StoryPlotCards StoryPlotCardCollection `json:"storyPlotCards"`
}

// DeleteStoryPlotCardRequest 指定要从当前书删除的剧情卡片。
type DeleteStoryPlotCardRequest struct {
	BookPath string `json:"bookPath"`
	CardID   string `json:"cardId"`
}

// DeleteStoryPlotCardResult 返回删除后的剧情卡片集合。
type DeleteStoryPlotCardResult struct {
	StoryPlotCards StoryPlotCardCollection `json:"storyPlotCards"`
}

// MaterialChapterRef 保存素材来源章节的可跳转 ID 和标题快照。
type MaterialChapterRef struct {
	ChapterID            string `json:"chapterId"`
	ChapterOrder         int    `json:"chapterOrder"`
	ChapterTitleSnapshot string `json:"chapterTitleSnapshot"`
}

// MaterialSourceSnapshot 保存来源失效时仍可阅读的基础上下文。
type MaterialSourceSnapshot struct {
	BookName            string `json:"bookName"`
	ChapterTitle        string `json:"chapterTitle"`
	DeconstructionTitle string `json:"deconstructionTitle"`
	PlanTypeName        string `json:"planTypeName"`
	DimensionName       string `json:"dimensionName"`
	SourceSummary       string `json:"sourceSummary"`
	SelectedText        string `json:"selectedText"`
	CreatedAt           string `json:"createdAt"`
}

// MaterialSourceRef 记录素材由哪类拆解沉淀而来，保留跳转 ID 和快照。
type MaterialSourceRef struct {
	SourceKind       string                 `json:"sourceKind"`
	DeconstructionID string                 `json:"deconstructionId"`
	SourceTextID     string                 `json:"sourceTextId"`
	ChapterRefs      []MaterialChapterRef   `json:"chapterRefs"`
	Snapshot         MaterialSourceSnapshot `json:"snapshot"`
}

// Material 是单本小说项目目录中的正式素材元数据，正文保存在 contentPath 指向的 TXT。
type Material struct {
	ID           string              `json:"id"`
	BookID       string              `json:"bookId"`
	MaterialType string              `json:"materialType"`
	Title        string              `json:"title"`
	Summary      string              `json:"summary"`
	ContentPath  string              `json:"contentPath"`
	CategoryIDs  []string            `json:"categoryIds"`
	TagIDs       []string            `json:"tagIds"`
	SourceRefs   []MaterialSourceRef `json:"sourceRefs"`
	CreatedAt    string              `json:"createdAt"`
	UpdatedAt    string              `json:"updatedAt"`
}

// MaterialCollection 对应当前小说项目目录中的 materials/materials.json。
type MaterialCollection struct {
	Version int        `json:"version"`
	Items   []Material `json:"items"`
}

// MaterialWorkspace 返回小说内素材页需要的正式素材、正文和来源上下文。
type MaterialWorkspace struct {
	Project             BookProject                  `json:"project"`
	Materials           MaterialCollection           `json:"materials"`
	MaterialContents    map[string]string            `json:"materialContents"`
	SourceTexts         SourceTextCollection         `json:"sourceTexts"`
	DeconstructionCards DeconstructionCardCollection `json:"deconstructionCards"`
}

// MaterialDraft 是从拆解卡片生成的可编辑素材草稿，不直接落盘。
type MaterialDraft struct {
	MaterialType string              `json:"materialType"`
	Title        string              `json:"title"`
	Summary      string              `json:"summary"`
	Content      string              `json:"content"`
	CategoryIDs  []string            `json:"categoryIds"`
	TagIDs       []string            `json:"tagIds"`
	SourceRefs   []MaterialSourceRef `json:"sourceRefs"`
}

// BuildMaterialDraftFromCardRequest 指定要从哪张写作手法拆解卡片生成素材草稿。
type BuildMaterialDraftFromCardRequest struct {
	BookPath string `json:"bookPath"`
	CardID   string `json:"cardId"`
}

// CreateMaterialFromCardRequest 保存用户基于拆解卡片二次整理后的素材。
type CreateMaterialFromCardRequest struct {
	BookPath    string   `json:"bookPath"`
	CardID      string   `json:"cardId"`
	Title       string   `json:"title"`
	Summary     string   `json:"summary"`
	Content     string   `json:"content"`
	CategoryIDs []string `json:"categoryIds"`
	TagIDs      []string `json:"tagIds"`
}

// CreateMaterialResult 返回保存后的素材、完整集合和正文映射。
type CreateMaterialResult struct {
	Material            Material                     `json:"material"`
	Materials           MaterialCollection           `json:"materials"`
	MaterialContents    map[string]string            `json:"materialContents"`
	DeconstructionCards DeconstructionCardCollection `json:"deconstructionCards"`
}

// UpdateMaterialRequest 承载素材详情页可编辑的元数据和正文。
type UpdateMaterialRequest struct {
	BookPath    string   `json:"bookPath"`
	MaterialID  string   `json:"materialId"`
	Title       string   `json:"title"`
	Summary     string   `json:"summary"`
	Content     string   `json:"content"`
	CategoryIDs []string `json:"categoryIds"`
	TagIDs      []string `json:"tagIds"`
}

// UpdateMaterialResult 返回编辑后的素材集合和正文映射，便于前端无二次读取刷新。
type UpdateMaterialResult struct {
	Material         Material           `json:"material"`
	Materials        MaterialCollection `json:"materials"`
	MaterialContents map[string]string  `json:"materialContents"`
}

// DeleteMaterialRequest 指定要从当前书素材库删除的素材。
type DeleteMaterialRequest struct {
	BookPath   string `json:"bookPath"`
	MaterialID string `json:"materialId"`
}

// DeleteMaterialResult 返回删除后的素材、正文和卡片引用集合。
type DeleteMaterialResult struct {
	Materials           MaterialCollection           `json:"materials"`
	MaterialContents    map[string]string            `json:"materialContents"`
	DeconstructionCards DeconstructionCardCollection `json:"deconstructionCards"`
}

// LibraryMaterial 是书库目录自己的正式素材元数据，正文保存在书库 materials/texts 下。
type LibraryMaterial struct {
	ID               string              `json:"id"`
	SourceMode       string              `json:"sourceMode"`
	BookID           string              `json:"bookId"`
	BookName         string              `json:"bookName"`
	OriginalBookPath string              `json:"originalBookPath"`
	BookMaterialID   string              `json:"bookMaterialId"`
	MaterialType     string              `json:"materialType"`
	PlanTypeName     string              `json:"planTypeName"`
	DimensionName    string              `json:"dimensionName"`
	Title            string              `json:"title"`
	Summary          string              `json:"summary"`
	ContentPath      string              `json:"contentPath"`
	TagIDs           []string            `json:"tagIds"`
	SourceRefs       []MaterialSourceRef `json:"sourceRefs"`
	CreatedAt        string              `json:"createdAt"`
	UpdatedAt        string              `json:"updatedAt"`
	MigratedAt       string              `json:"migratedAt"`
}

// LibraryMaterialCollection 对应书库根目录中的 materials/library-materials.json。
type LibraryMaterialCollection struct {
	Version int               `json:"version"`
	Items   []LibraryMaterial `json:"items"`
}

// LibraryMaterialListItem 是书库素材页的分页展示条目，合并书库自有素材和本书素材引用。
type LibraryMaterialListItem struct {
	ID                string              `json:"id"`
	SourceMode        string              `json:"sourceMode"`
	LibraryMaterialID string              `json:"libraryMaterialId"`
	BookID            string              `json:"bookId"`
	BookName          string              `json:"bookName"`
	BookPath          string              `json:"bookPath"`
	BookMaterialID    string              `json:"bookMaterialId"`
	MaterialType      string              `json:"materialType"`
	PlanTypeName      string              `json:"planTypeName"`
	DimensionName     string              `json:"dimensionName"`
	Title             string              `json:"title"`
	Summary           string              `json:"summary"`
	Content           string              `json:"content"`
	TagIDs            []string            `json:"tagIds"`
	SourceRefs        []MaterialSourceRef `json:"sourceRefs"`
	CreatedAt         string              `json:"createdAt"`
	UpdatedAt         string              `json:"updatedAt"`
	SourceAvailable   bool                `json:"sourceAvailable"`
}

// LibraryMaterialPageRequest 承载书库素材页的筛选和分页参数。
type LibraryMaterialPageRequest struct {
	LibraryPath  string `json:"libraryPath"`
	Keyword      string `json:"keyword"`
	BookID       string `json:"bookId"`
	MaterialType string `json:"materialType"`
	TagID        string `json:"tagId"`
	Page         int    `json:"page"`
	PageSize     int    `json:"pageSize"`
}

// LibraryMaterialPage 返回书库素材页当前页数据和筛选选项。
type LibraryMaterialPage struct {
	LibraryPath string                      `json:"libraryPath"`
	Items       []LibraryMaterialListItem   `json:"items"`
	Total       int                         `json:"total"`
	Page        int                         `json:"page"`
	PageSize    int                         `json:"pageSize"`
	Books       []LibraryMaterialBookOption `json:"books"`
	Types       []string                    `json:"types"`
	Tags        []string                    `json:"tags"`
}

// LibraryMaterialBookOption 是书库素材页按书筛选的轻量选项。
type LibraryMaterialBookOption struct {
	BookID   string `json:"bookId"`
	BookName string `json:"bookName"`
}

// CreateLibraryMaterialRequest 保存用户直接在书库素材页新建的素材。
type CreateLibraryMaterialRequest struct {
	LibraryPath   string   `json:"libraryPath"`
	Title         string   `json:"title"`
	Summary       string   `json:"summary"`
	Content       string   `json:"content"`
	MaterialType  string   `json:"materialType"`
	PlanTypeName  string   `json:"planTypeName"`
	DimensionName string   `json:"dimensionName"`
	TagIDs        []string `json:"tagIds"`
}

// UpdateLibraryMaterialRequest 更新书库自有素材，不能用于实时引用的本书素材。
type UpdateLibraryMaterialRequest struct {
	LibraryPath       string   `json:"libraryPath"`
	LibraryMaterialID string   `json:"libraryMaterialId"`
	Title             string   `json:"title"`
	Summary           string   `json:"summary"`
	Content           string   `json:"content"`
	MaterialType      string   `json:"materialType"`
	PlanTypeName      string   `json:"planTypeName"`
	DimensionName     string   `json:"dimensionName"`
	TagIDs            []string `json:"tagIds"`
}

// DeleteLibraryMaterialRequest 删除书库自有素材；本书素材引用必须回到单书素材页处理。
type DeleteLibraryMaterialRequest struct {
	LibraryPath       string `json:"libraryPath"`
	LibraryMaterialID string `json:"libraryMaterialId"`
}

// GlobalMaterialIndex 是应用系统数据目录中的跨书素材引用索引，不是素材正式数据源。
type GlobalMaterialIndex struct {
	Version int                       `json:"version"`
	Items   []GlobalMaterialIndexItem `json:"items"`
}

// GlobalMaterialIndexItem 保存全局素材列表所需的引用和展示快照。
type GlobalMaterialIndexItem struct {
	BookID       string                      `json:"bookId"`
	BookPath     string                      `json:"bookPath"`
	MaterialID   string                      `json:"materialId"`
	MaterialType string                      `json:"materialType"`
	Snapshot     GlobalMaterialIndexSnapshot `json:"snapshot"`
	IndexedAt    string                      `json:"indexedAt"`
}

// GlobalMaterialIndexSnapshot 保存来源项目不可用时的必要展示信息。
type GlobalMaterialIndexSnapshot struct {
	Title    string   `json:"title"`
	Summary  string   `json:"summary"`
	BookName string   `json:"bookName"`
	TagNames []string `json:"tagNames"`
}

// ChapterMemo 保存用户针对单个章节写下的自由小记，随小说项目目录迁移。
type ChapterMemo struct {
	BookID       string `json:"bookId"`
	ChapterID    string `json:"chapterId"`
	ChapterOrder int    `json:"chapterOrder"`
	Content      string `json:"content"`
	CreatedAt    string `json:"createdAt"`
	UpdatedAt    string `json:"updatedAt"`
}

// ChapterMemoCollection 对应 reading/chapter-memos.json。
type ChapterMemoCollection struct {
	Version int           `json:"version"`
	Items   []ChapterMemo `json:"items"`
}

// ReadingSession 是前端打开阅读工作台所需的当前书完整上下文。
type ReadingSession struct {
	Project             BookProject                  `json:"project"`
	State               ReadingState                 `json:"state"`
	CurrentChapter      BookChapter                  `json:"currentChapter"`
	ChapterContent      string                       `json:"chapterContent"`
	Plan                DeconstructionPlan           `json:"plan"`
	SourceTexts         SourceTextCollection         `json:"sourceTexts"`
	DeconstructionCards DeconstructionCardCollection `json:"deconstructionCards"`
	ChapterMemos        ChapterMemoCollection        `json:"chapterMemos"`
}

// ChapterContentRequest 指定要读取的单本书章节正文。
type ChapterContentRequest struct {
	BookPath  string `json:"bookPath"`
	ChapterID string `json:"chapterId"`
}

// ChapterContentResult 返回章节元数据和用户修正后的 TXT 正文。
type ChapterContentResult struct {
	Chapter BookChapter `json:"chapter"`
	Content string      `json:"content"`
}

// SaveReadingPositionRequest 承载阅读页翻页、跳章或切换模式后的阅读位置。
type SaveReadingPositionRequest struct {
	BookPath               string `json:"bookPath"`
	ChapterID              string `json:"chapterId"`
	CharOffset             int    `json:"charOffset"`
	ParagraphIndex         int    `json:"paragraphIndex"`
	RenderedPageNumber     int    `json:"renderedPageNumber"`
	LastMode               string `json:"lastMode"`
	ReadingProgressPercent int    `json:"readingProgressPercent"`
}

// SaveReadingPreferencesRequest 保存当前书独立的阅读界面偏好。
type SaveReadingPreferencesRequest struct {
	BookPath string `json:"bookPath"`
	FontSize int    `json:"fontSize"`
}

// SaveChapterContentRequest 保存编辑 TAB 中修改后的章节标题和正文。
type SaveChapterContentRequest struct {
	BookPath  string `json:"bookPath"`
	ChapterID string `json:"chapterId"`
	Title     string `json:"title"`
	Content   string `json:"content"`
}

// SplitChapterRequest 从当前章节开头拆出一个新章节，当前章节保留剩余正文。
type SplitChapterRequest struct {
	BookPath         string `json:"bookPath"`
	ChapterID        string `json:"chapterId"`
	NewTitle         string `json:"newTitle"`
	NewContent       string `json:"newContent"`
	RemainingContent string `json:"remainingContent"`
}

// SplitChapterResult 返回章节拆分后更新的项目数据和两段章节正文。
type SplitChapterResult struct {
	Project          BookProject `json:"project"`
	NewChapter       BookChapter `json:"newChapter"`
	RemainingChapter BookChapter `json:"remainingChapter"`
	NewContent       string      `json:"newContent"`
	RemainingContent string      `json:"remainingContent"`
}

// SourceSelection 是创建拆解时从阅读正文中捕获的来源范围。
type SourceSelection struct {
	ChapterID    string `json:"chapterId"`
	ChapterOrder int    `json:"chapterOrder"`
	StartOffset  int    `json:"startOffset"`
	EndOffset    int    `json:"endOffset"`
	SelectedText string `json:"selectedText"`
}

// CreateDeconstructionRequest 保存一张拆解卡片，并按需创建或复用来源片段。
type CreateDeconstructionRequest struct {
	BookPath      string            `json:"bookPath"`
	Selection     SourceSelection   `json:"selection"`
	PlanTypeID    string            `json:"planTypeId"`
	PlanTypeName  string            `json:"planTypeName"`
	DimensionID   string            `json:"dimensionId"`
	DimensionName string            `json:"dimensionName"`
	Content       string            `json:"content"`
	FieldValues   map[string]string `json:"fieldValues"`
	TagIDs        []string          `json:"tagIds"`
}

// CreateDeconstructionResult 返回本次保存后的来源片段和拆解卡片。
type CreateDeconstructionResult struct {
	SourceText          SourceText                   `json:"sourceText"`
	Card                DeconstructionCard           `json:"card"`
	SourceTexts         SourceTextCollection         `json:"sourceTexts"`
	DeconstructionCards DeconstructionCardCollection `json:"deconstructionCards"`
}

// UpdateDeconstructionCardRequest 承载写作手法拆解卡片正文和标签的编辑结果。
type UpdateDeconstructionCardRequest struct {
	BookPath      string   `json:"bookPath"`
	CardID        string   `json:"cardId"`
	PlanTypeID    string   `json:"planTypeId"`
	PlanTypeName  string   `json:"planTypeName"`
	DimensionID   string   `json:"dimensionId"`
	DimensionName string   `json:"dimensionName"`
	Content       string   `json:"content"`
	TagIDs        []string `json:"tagIds"`
}

// UpdateDeconstructionCardResult 返回编辑后完整卡片集合，便于前端刷新列表和详情。
type UpdateDeconstructionCardResult struct {
	Card                DeconstructionCard           `json:"card"`
	DeconstructionCards DeconstructionCardCollection `json:"deconstructionCards"`
}

// DeleteDeconstructionCardRequest 指定要从当前书删除的写作手法拆解卡片。
type DeleteDeconstructionCardRequest struct {
	BookPath string `json:"bookPath"`
	CardID   string `json:"cardId"`
}

// DeleteDeconstructionCardResult 返回删除后同步清理过的来源片段和卡片集合。
type DeleteDeconstructionCardResult struct {
	SourceTexts         SourceTextCollection         `json:"sourceTexts"`
	DeconstructionCards DeconstructionCardCollection `json:"deconstructionCards"`
}

// SaveChapterMemoRequest 保存目录中某章的小记内容。
type SaveChapterMemoRequest struct {
	BookPath  string `json:"bookPath"`
	ChapterID string `json:"chapterId"`
	Content   string `json:"content"`
}

// SaveChapterMemoResult 返回保存后的当前章节小记和完整小记集合。
type SaveChapterMemoResult struct {
	Memo         ChapterMemo           `json:"memo"`
	ChapterMemos ChapterMemoCollection `json:"chapterMemos"`
}
