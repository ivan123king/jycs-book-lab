// app.go 暴露 Wails 前端可调用的应用入口，并把书库业务转发给 internal/library 模块。
package main

import (
	"context"
	"fmt"

	booklibrary "jiayi-book-lab/internal/library"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type AppSettings = booklibrary.AppSettings
type LibraryBookEntry = booklibrary.LibraryBookEntry
type LibraryIndex = booklibrary.LibraryIndex
type LibraryState = booklibrary.LibraryState
type BookMeta = booklibrary.BookMeta
type BookChapter = booklibrary.BookChapter
type BookProject = booklibrary.BookProject
type ImportChapterPreview = booklibrary.ImportChapterPreview
type ImportDraft = booklibrary.ImportDraft
type ConfirmImportRequest = booklibrary.ConfirmImportRequest
type ConfirmImportResult = booklibrary.ConfirmImportResult
type ReadingPosition = booklibrary.ReadingPosition
type ReadingState = booklibrary.ReadingState
type DeconstructionField = booklibrary.DeconstructionField
type DeconstructionType = booklibrary.DeconstructionType
type DeconstructionPlan = booklibrary.DeconstructionPlan
type SourceText = booklibrary.SourceText
type SourceTextCollection = booklibrary.SourceTextCollection
type DeconstructionCard = booklibrary.DeconstructionCard
type DeconstructionCardCollection = booklibrary.DeconstructionCardCollection
type ReadingSession = booklibrary.ReadingSession
type ChapterContentRequest = booklibrary.ChapterContentRequest
type ChapterContentResult = booklibrary.ChapterContentResult
type SaveReadingPositionRequest = booklibrary.SaveReadingPositionRequest
type SaveChapterContentRequest = booklibrary.SaveChapterContentRequest
type SourceSelection = booklibrary.SourceSelection
type CreateDeconstructionRequest = booklibrary.CreateDeconstructionRequest
type CreateDeconstructionResult = booklibrary.CreateDeconstructionResult

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

// GetLibraryState 读取当前书库目录和轻量索引，用于启动时恢复书库页面。
func (a *App) GetLibraryState() (LibraryState, error) {
	return booklibrary.LoadLibraryState()
}

// ChooseLibraryDirectory 让用户选择书库根目录，并在目录中创建或读取 index.json。
func (a *App) ChooseLibraryDirectory() (LibraryState, error) {
	path, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "选择佳依拆书书库目录",
	})
	if err != nil {
		return LibraryState{}, fmt.Errorf("选择书库目录失败: %w", err)
	}
	if path == "" {
		return booklibrary.LoadLibraryState()
	}
	return booklibrary.SetLibraryPath(path)
}

// PreviewNovelImport 选择 TXT 小说文件并生成导入预览，不写入正式数据。
func (a *App) PreviewNovelImport() (*ImportDraft, error) {
	state, err := booklibrary.LoadLibraryState()
	if err != nil {
		return nil, err
	}
	if state.LibraryPath == "" {
		return nil, fmt.Errorf("请先设置书库目录")
	}

	sourcePath, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "选择 TXT 小说文件",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "TXT 小说文件 (*.txt)",
				Pattern:     "*.txt",
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("选择小说文件失败: %w", err)
	}
	if sourcePath == "" {
		return nil, nil
	}

	draft, err := booklibrary.PreviewNovelImport(state.LibraryPath, sourcePath)
	if err != nil {
		return nil, err
	}
	return &draft, nil
}

// ReDetectNovelImport 对同一个 TXT 文件重新执行章节识别，供导入失败或重试时使用。
func (a *App) ReDetectNovelImport(sourcePath string) (*ImportDraft, error) {
	state, err := booklibrary.LoadLibraryState()
	if err != nil {
		return nil, err
	}
	if state.LibraryPath == "" {
		return nil, fmt.Errorf("请先设置书库目录")
	}
	draft, err := booklibrary.PreviewNovelImport(state.LibraryPath, sourcePath)
	if err != nil {
		return nil, err
	}
	return &draft, nil
}

// ConfirmNovelImport 将预览中的 TXT 小说复制入库，并生成 book.json、章节 TXT 和 index.json。
func (a *App) ConfirmNovelImport(request ConfirmImportRequest) (ConfirmImportResult, error) {
	state, err := booklibrary.LoadLibraryState()
	if err != nil {
		return ConfirmImportResult{}, err
	}
	if state.LibraryPath == "" {
		return ConfirmImportResult{}, fmt.Errorf("请先设置书库目录")
	}
	return booklibrary.ConfirmNovelImport(state.LibraryPath, request)
}

// OpenBookProject 读取单本小说目录中的 book.json，作为进入当前书工作区的入口。
func (a *App) OpenBookProject(bookPath string) (BookProject, error) {
	return booklibrary.OpenBookProject(bookPath)
}

// OpenReadingSession 打开阅读工作台，并从单本小说目录恢复阅读位置和当前章节正文。
func (a *App) OpenReadingSession(bookPath string) (ReadingSession, error) {
	return booklibrary.OpenReadingSession(bookPath)
}

// GetChapterContent 按章节 ID 读取用户修正后的章节正文 TXT。
func (a *App) GetChapterContent(request ChapterContentRequest) (ChapterContentResult, error) {
	return booklibrary.GetChapterContent(request)
}

// SaveReadingPosition 把翻页和跳章产生的阅读位置写入当前小说项目目录。
func (a *App) SaveReadingPosition(request SaveReadingPositionRequest) (ReadingState, error) {
	return booklibrary.SaveReadingPosition(request)
}

// SaveChapterContent 保存编辑 TAB 中的章节标题和正文。
func (a *App) SaveChapterContent(request SaveChapterContentRequest) (ChapterContentResult, error) {
	return booklibrary.SaveChapterContent(request)
}

// CreateDeconstructionCard 从当前选文创建拆解卡片，并保留来源追溯。
func (a *App) CreateDeconstructionCard(request CreateDeconstructionRequest) (CreateDeconstructionResult, error) {
	return booklibrary.CreateDeconstructionCard(request)
}
