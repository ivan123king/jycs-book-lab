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
	Library LibraryState `json:"library"`
	Project BookProject  `json:"project"`
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
	BookPath    string            `json:"bookPath"`
	Selection   SourceSelection   `json:"selection"`
	PlanTypeID  string            `json:"planTypeId"`
	DimensionID string            `json:"dimensionId"`
	Content     string            `json:"content"`
	FieldValues map[string]string `json:"fieldValues"`
	TagIDs      []string          `json:"tagIds"`
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
	BookPath    string   `json:"bookPath"`
	CardID      string   `json:"cardId"`
	PlanTypeID  string   `json:"planTypeId"`
	DimensionID string   `json:"dimensionId"`
	Content     string   `json:"content"`
	TagIDs      []string `json:"tagIds"`
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
