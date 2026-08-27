// import_service.go 负责 TXT 小说导入预览、确认入库和初始小说项目目录生成。
package library

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode"
)

type importAnalysis struct {
	draft    ImportDraft
	text     string
	chapters []BookChapter
}

type importTimer struct {
	action string
	start  time.Time
	last   time.Time
	stages []ImportTimingStage
}

// newImportTimer 创建导入链路计时器，所有耗时只返回给前端诊断，不写入正式小说数据。
func newImportTimer(action string) *importTimer {
	now := time.Now()
	return &importTimer{action: action, start: now, last: now, stages: make([]ImportTimingStage, 0, 8)}
}

// mark 记录从上一个阶段到当前阶段的耗时，方便定位卡在读文件、章节识别还是落盘。
func (timer *importTimer) mark(name string, label string) {
	if timer == nil {
		return
	}
	now := time.Now()
	timer.stages = append(timer.stages, ImportTimingStage{
		Name:       name,
		Label:      label,
		DurationMs: now.Sub(timer.last).Milliseconds(),
	})
	timer.last = now
}

// finish 汇总导入诊断信息；章节数和文件大小用于把慢操作与具体导入规模对应起来。
func (timer *importTimer) finish(bookSize int64, chapterCount int) ImportDiagnostics {
	if timer == nil {
		return ImportDiagnostics{}
	}
	return ImportDiagnostics{
		Action:       timer.action,
		TotalMs:      time.Since(timer.start).Milliseconds(),
		BookSize:     bookSize,
		ChapterCount: chapterCount,
		Stages:       timer.stages,
	}
}

// PreviewNovelImport 读取 TXT 文件并生成导入预览，不写入任何正式书籍数据。
func PreviewNovelImport(libraryPath string, sourcePath string) (ImportDraft, error) {
	timer := newImportTimer("previewImport")
	analysis, err := analyzeNovelImport(libraryPath, sourcePath, timer)
	if err != nil {
		return ImportDraft{}, err
	}
	analysis.draft.Diagnostics = timer.finish(analysis.draft.BookSize, analysis.draft.ChapterCount)
	return analysis.draft, nil
}

// analyzeNovelImport 完成导入预览所需的读取、编码处理和章节识别，确认导入会复用该结果避免重复解析。
func analyzeNovelImport(libraryPath string, sourcePath string, timer *importTimer) (importAnalysis, error) {
	libraryPath = filepath.Clean(strings.TrimSpace(libraryPath))
	sourcePath = filepath.Clean(strings.TrimSpace(sourcePath))
	if sourcePath == "." || sourcePath == "" {
		return importAnalysis{}, errors.New("导入文件不能为空")
	}
	info, err := os.Stat(sourcePath)
	timer.mark("statSource", "读取文件信息")
	if err != nil {
		return importAnalysis{}, fmt.Errorf("读取导入文件信息失败: %w", err)
	}
	if info.IsDir() {
		return importAnalysis{}, errors.New("请选择 TXT 小说文件，不能选择目录")
	}
	if !strings.EqualFold(filepath.Ext(sourcePath), ".txt") {
		return importAnalysis{}, errors.New("第一阶段只支持导入 TXT 小说文件")
	}

	text, err := readNovelText(sourcePath)
	timer.mark("readText", "读取并转换文本")
	if err != nil {
		return importAnalysis{}, err
	}
	chapters := DetectChapters(text)
	timer.mark("detectChapters", "识别章节目录")

	bookID := nextBookID(libraryPath)
	previewChapters := make([]ImportChapterPreview, 0, len(chapters))
	for _, chapter := range chapters {
		previewChapters = append(previewChapters, ImportChapterPreview{
			ID:          chapter.ID,
			Order:       chapter.Order,
			Title:       chapter.Title,
			StartOffset: chapter.StartOffset,
			EndOffset:   chapter.EndOffset,
			WordCount:   chapter.WordCount,
		})
	}
	timer.mark("buildDraft", "生成导入预览")

	draft := ImportDraft{
		BookID:       bookID,
		BookName:     strings.TrimSuffix(filepath.Base(sourcePath), filepath.Ext(sourcePath)),
		SourcePath:   sourcePath,
		LibraryPath:  libraryPath,
		TargetPath:   filepath.Join(libraryPath, "books", bookID),
		BookSize:     info.Size(),
		ChapterCount: len(previewChapters),
		Chapters:     previewChapters,
	}

	return importAnalysis{draft: draft, text: text, chapters: chapters}, nil
}

// ConfirmNovelImport 将导入草稿写成完整的单本小说项目目录，并更新书库根 index.json。
func ConfirmNovelImport(libraryPath string, request ConfirmImportRequest) (ConfirmImportResult, error) {
	timer := newImportTimer("confirmImport")
	libraryPath = filepath.Clean(strings.TrimSpace(libraryPath))
	if libraryPath == "." || libraryPath == "" {
		return ConfirmImportResult{}, errors.New("请先设置书库目录")
	}
	if _, err := ensureLibraryIndex(libraryPath); err != nil {
		return ConfirmImportResult{}, err
	}
	timer.mark("ensureIndex", "读取书库索引")

	analysis, err := analyzeNovelImport(libraryPath, request.SourcePath, timer)
	if err != nil {
		return ConfirmImportResult{}, err
	}
	draft := analysis.draft
	if strings.TrimSpace(request.BookID) != "" {
		draft.BookID = sanitizeBookID(request.BookID)
		draft.TargetPath = filepath.Join(libraryPath, "books", draft.BookID)
	}
	if strings.TrimSpace(request.BookName) != "" {
		draft.BookName = strings.TrimSpace(request.BookName)
	}
	draft.Description = strings.TrimSpace(request.Description)

	project, err := writeBookProjectFromDraft(draft, analysis.text, analysis.chapters, timer)
	if err != nil {
		return ConfirmImportResult{}, err
	}

	index, err := appendBookToIndex(libraryPath, LibraryBookEntry{
		BookName: project.Book.Name,
		BookPath: project.Book.BookPath,
		BookSize: project.Book.BookSize,
	})
	if err != nil {
		_ = os.RemoveAll(project.Book.BookPath)
		return ConfirmImportResult{}, err
	}
	timer.mark("appendIndex", "更新书库索引")

	diagnostics := timer.finish(project.Book.BookSize, len(project.Chapters))

	return ConfirmImportResult{
		Library:     LibraryState{LibraryPath: libraryPath, Index: enrichLibraryReadingProgress(index)},
		Project:     project,
		Diagnostics: diagnostics,
	}, nil
}

// writeBookProjectFromDraft 以临时目录完成导入落盘，成功后再提交为正式小说项目目录。
func writeBookProjectFromDraft(draft ImportDraft, text string, chapters []BookChapter, timer *importTimer) (BookProject, error) {
	if strings.TrimSpace(draft.BookName) == "" {
		return BookProject{}, errors.New("小说名不能为空")
	}

	targetPath := filepath.Clean(draft.TargetPath)
	if _, err := os.Stat(targetPath); err == nil {
		return BookProject{}, fmt.Errorf("小说目录已存在: %s", targetPath)
	} else if !errors.Is(err, os.ErrNotExist) {
		return BookProject{}, fmt.Errorf("检查小说目录失败: %w", err)
	}
	timer.mark("checkTarget", "检查目标目录")

	tempPath := targetPath + ".importing"
	_ = os.RemoveAll(tempPath)
	if err := os.MkdirAll(filepath.Join(tempPath, "chapters"), 0o755); err != nil {
		return BookProject{}, fmt.Errorf("创建小说临时目录失败: %w", err)
	}
	timer.mark("createTempDir", "创建临时目录")

	if err := copyFile(draft.SourcePath, filepath.Join(tempPath, originalTextName)); err != nil {
		_ = os.RemoveAll(tempPath)
		return BookProject{}, err
	}
	timer.mark("copyOriginal", "复制原始文本")

	for _, chapter := range chapters {
		bodyStart := chapter.StartOffset
		if chapter.Title != unknownChapterTitle {
			bodyStart = findChapterBodyStart(text, chapter.StartOffset)
		}
		content := text[bodyStart:chapter.EndOffset]
		content = strings.TrimLeft(content, "\r\n")
		if err := os.WriteFile(filepath.Join(tempPath, filepath.FromSlash(chapter.ContentPath)), []byte(content), 0o644); err != nil {
			_ = os.RemoveAll(tempPath)
			return BookProject{}, fmt.Errorf("写入章节正文失败: %w", err)
		}
	}
	timer.mark("writeChapters", "写入章节正文")

	now := time.Now().Format(time.RFC3339)
	project := BookProject{
		Version: 1,
		Book: BookMeta{
			ID:             draft.BookID,
			Name:           draft.BookName,
			Description:    draft.Description,
			SourceFileName: filepath.Base(draft.SourcePath),
			SourceTextPath: originalTextName,
			BookPath:       targetPath,
			BookSize:       draft.BookSize,
			CreatedAt:      now,
			UpdatedAt:      now,
		},
		Chapters: chapters,
	}

	if err := writeJSONFileAtomic(filepath.Join(tempPath, bookJSONName), project); err != nil {
		_ = os.RemoveAll(tempPath)
		return BookProject{}, err
	}
	timer.mark("writeBookJSON", "写入书籍元数据")
	preferences, err := defaultReadingPreferencesForNewBook()
	if err != nil {
		_ = os.RemoveAll(tempPath)
		return BookProject{}, fmt.Errorf("读取新书默认阅读偏好失败: %w", err)
	}
	initialReadingState := defaultReadingStateWithPreferences(project.Book.ID, chooseChapterForPosition(project, ""), preferences)
	if err := writeJSONFileAtomic(filepath.Join(tempPath, filepath.FromSlash(readingStateRelativePath)), initialReadingState); err != nil {
		_ = os.RemoveAll(tempPath)
		return BookProject{}, fmt.Errorf("写入初始阅读状态失败: %w", err)
	}
	timer.mark("writeReadingState", "写入初始阅读状态")
	if err := ensureDeconstructionPlanFiles(tempPath, project.Book.ID); err != nil {
		_ = os.RemoveAll(tempPath)
		return BookProject{}, err
	}
	timer.mark("ensurePlanFiles", "初始化拆解方案")
	if err := os.Rename(tempPath, targetPath); err != nil {
		_ = os.RemoveAll(tempPath)
		return BookProject{}, fmt.Errorf("提交小说目录失败: %w", err)
	}
	timer.mark("commitProject", "提交小说目录")
	return project, nil
}

// findChapterBodyStart 找到章节标题行后的正文起点，避免把标题重复写入章节 TXT。
func findChapterBodyStart(text string, startOffset int) int {
	if startOffset < 0 || startOffset >= len(text) {
		return startOffset
	}
	nextLine := strings.IndexByte(text[startOffset:], '\n')
	if nextLine < 0 {
		return len(text)
	}
	return startOffset + nextLine + 1
}

// nextBookID 根据当前时间生成不冲突的本地书籍目录 ID。
func nextBookID(libraryPath string) string {
	base := "book-" + time.Now().Format("20060102-150405")
	bookID := base
	for i := 1; ; i++ {
		if _, err := os.Stat(filepath.Join(libraryPath, "books", bookID)); errors.Is(err, os.ErrNotExist) {
			return bookID
		}
		bookID = fmt.Sprintf("%s-%02d", base, i)
	}
}

// sanitizeBookID 清理用户传入的书籍 ID，避免生成非法目录名。
func sanitizeBookID(bookID string) string {
	bookID = strings.TrimSpace(bookID)
	var builder strings.Builder
	for _, value := range bookID {
		switch {
		case unicode.IsLetter(value), unicode.IsDigit(value), value == '-', value == '_':
			builder.WriteRune(value)
		default:
			builder.WriteByte('-')
		}
	}
	cleaned := strings.Trim(builder.String(), "-_.")
	if cleaned == "" || cleaned == "." || cleaned == ".." {
		return "book-" + time.Now().Format("20060102-150405")
	}
	return cleaned
}
