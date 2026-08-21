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

// ReadingState 是单本小说 reading/reading-state.json 的正式阅读位置数据。
type ReadingState struct {
	Version                int             `json:"version"`
	BookID                 string          `json:"bookId"`
	LastPosition           ReadingPosition `json:"lastPosition"`
	LastMode               string          `json:"lastMode"`
	ReadingProgressPercent int             `json:"readingProgressPercent"`
}

// DeconstructionField 定义当前书拆解类型下的一个动态表单字段。
type DeconstructionField struct {
	ID       string   `json:"id"`
	Label    string   `json:"label"`
	Type     string   `json:"type"`
	Order    int      `json:"order"`
	Required bool     `json:"required"`
	Visible  bool     `json:"visible"`
	Options  []string `json:"options"`
}

// DeconstructionType 定义当前书拆解方案中的一种拆解卡片类型。
type DeconstructionType struct {
	ID        string                `json:"id"`
	Name      string                `json:"name"`
	Dimension string                `json:"dimension"`
	Order     int                   `json:"order"`
	Fields    []DeconstructionField `json:"fields"`
}

// DeconstructionPlan 保存每本书独立的拆解方案配置。
type DeconstructionPlan struct {
	Version   int                  `json:"version"`
	BookID    string               `json:"bookId"`
	Types     []DeconstructionType `json:"types"`
	UpdatedAt string               `json:"updatedAt"`
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
	ID           string            `json:"id"`
	BookID       string            `json:"bookId"`
	SourceTextID string            `json:"sourceTextId"`
	PlanTypeID   string            `json:"planTypeId"`
	FieldValues  map[string]string `json:"fieldValues"`
	TagIDs       []string          `json:"tagIds"`
	CreatedAt    string            `json:"createdAt"`
	UpdatedAt    string            `json:"updatedAt"`
}

// DeconstructionCardCollection 对应 deconstructions/deconstruction-cards.json。
type DeconstructionCardCollection struct {
	Version int                  `json:"version"`
	Items   []DeconstructionCard `json:"items"`
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

// SaveChapterContentRequest 保存编辑 TAB 中修改后的章节标题和正文。
type SaveChapterContentRequest struct {
	BookPath  string `json:"bookPath"`
	ChapterID string `json:"chapterId"`
	Title     string `json:"title"`
	Content   string `json:"content"`
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
