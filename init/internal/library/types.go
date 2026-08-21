// Package library 定义书库、导入预览和单本小说项目的本地数据服务边界。
package library

// AppSettings 保存应用级偏好，只记录与具体书籍无关的全局配置。
type AppSettings struct {
	Version     int    `json:"version"`
	LibraryPath string `json:"libraryPath"`
}

// LibraryBookEntry 是书库 index.json 中用于快速展示书籍列表的轻量入口。
type LibraryBookEntry struct {
	BookName string `json:"bookName"`
	BookPath string `json:"bookPath"`
	BookSize int64  `json:"bookSize"`
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
