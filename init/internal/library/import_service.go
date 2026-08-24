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

// PreviewNovelImport 读取 TXT 文件并生成导入预览，不写入任何正式书籍数据。
func PreviewNovelImport(libraryPath string, sourcePath string) (ImportDraft, error) {
	sourcePath = filepath.Clean(strings.TrimSpace(sourcePath))
	if sourcePath == "." || sourcePath == "" {
		return ImportDraft{}, errors.New("导入文件不能为空")
	}
	info, err := os.Stat(sourcePath)
	if err != nil {
		return ImportDraft{}, fmt.Errorf("读取导入文件信息失败: %w", err)
	}
	if info.IsDir() {
		return ImportDraft{}, errors.New("请选择 TXT 小说文件，不能选择目录")
	}
	if !strings.EqualFold(filepath.Ext(sourcePath), ".txt") {
		return ImportDraft{}, errors.New("第一阶段只支持导入 TXT 小说文件")
	}

	text, err := readNovelText(sourcePath)
	if err != nil {
		return ImportDraft{}, err
	}
	chapters := DetectChapters(text)
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

	return ImportDraft{
		BookID:       bookID,
		BookName:     strings.TrimSuffix(filepath.Base(sourcePath), filepath.Ext(sourcePath)),
		SourcePath:   sourcePath,
		LibraryPath:  libraryPath,
		TargetPath:   filepath.Join(libraryPath, "books", bookID),
		BookSize:     info.Size(),
		ChapterCount: len(previewChapters),
		Chapters:     previewChapters,
	}, nil
}

// ConfirmNovelImport 将导入草稿写成完整的单本小说项目目录，并更新书库根 index.json。
func ConfirmNovelImport(libraryPath string, request ConfirmImportRequest) (ConfirmImportResult, error) {
	libraryPath = filepath.Clean(strings.TrimSpace(libraryPath))
	if libraryPath == "." || libraryPath == "" {
		return ConfirmImportResult{}, errors.New("请先设置书库目录")
	}
	if _, err := ensureLibraryIndex(libraryPath); err != nil {
		return ConfirmImportResult{}, err
	}

	draft, err := PreviewNovelImport(libraryPath, request.SourcePath)
	if err != nil {
		return ConfirmImportResult{}, err
	}
	if strings.TrimSpace(request.BookID) != "" {
		draft.BookID = sanitizeBookID(request.BookID)
		draft.TargetPath = filepath.Join(libraryPath, "books", draft.BookID)
	}
	if strings.TrimSpace(request.BookName) != "" {
		draft.BookName = strings.TrimSpace(request.BookName)
	}
	draft.Description = strings.TrimSpace(request.Description)

	project, err := writeBookProjectFromDraft(draft)
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

	return ConfirmImportResult{
		Library: LibraryState{LibraryPath: libraryPath, Index: enrichLibraryReadingProgress(index)},
		Project: project,
	}, nil
}

// writeBookProjectFromDraft 以临时目录完成导入落盘，成功后再提交为正式小说项目目录。
func writeBookProjectFromDraft(draft ImportDraft) (BookProject, error) {
	if strings.TrimSpace(draft.BookName) == "" {
		return BookProject{}, errors.New("小说名不能为空")
	}

	text, err := readNovelText(draft.SourcePath)
	if err != nil {
		return BookProject{}, err
	}
	chapters := DetectChapters(text)

	targetPath := filepath.Clean(draft.TargetPath)
	if _, err := os.Stat(targetPath); err == nil {
		return BookProject{}, fmt.Errorf("小说目录已存在: %s", targetPath)
	} else if !errors.Is(err, os.ErrNotExist) {
		return BookProject{}, fmt.Errorf("检查小说目录失败: %w", err)
	}

	tempPath := targetPath + ".importing"
	_ = os.RemoveAll(tempPath)
	if err := os.MkdirAll(filepath.Join(tempPath, "chapters"), 0o755); err != nil {
		return BookProject{}, fmt.Errorf("创建小说临时目录失败: %w", err)
	}

	if err := copyFile(draft.SourcePath, filepath.Join(tempPath, originalTextName)); err != nil {
		_ = os.RemoveAll(tempPath)
		return BookProject{}, err
	}

	for _, chapter := range chapters {
		bodyStart := chapter.StartOffset
		if isRecognizedChapterTitleAt(text, chapter.StartOffset, chapter.Title) {
			markerEnd := findChapterBodyStart(text, chapter.StartOffset)
			bodyStart = markerEnd
		}
		content := text[bodyStart:chapter.EndOffset]
		content = strings.TrimLeft(content, "\r\n")
		if err := os.WriteFile(filepath.Join(tempPath, filepath.FromSlash(chapter.ContentPath)), []byte(content), 0o644); err != nil {
			_ = os.RemoveAll(tempPath)
			return BookProject{}, fmt.Errorf("写入章节正文失败: %w", err)
		}
	}

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
	if err := ensureDeconstructionPlanFiles(tempPath, project.Book.ID); err != nil {
		_ = os.RemoveAll(tempPath)
		return BookProject{}, err
	}
	if err := os.Rename(tempPath, targetPath); err != nil {
		_ = os.RemoveAll(tempPath)
		return BookProject{}, fmt.Errorf("提交小说目录失败: %w", err)
	}
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

// isRecognizedChapterTitleAt 判断指定偏移处是否是章节识别器确认过的标题行。
func isRecognizedChapterTitleAt(text string, startOffset int, title string) bool {
	if startOffset < 0 || startOffset >= len(text) {
		return false
	}
	lineEnd := strings.IndexByte(text[startOffset:], '\n')
	if lineEnd < 0 {
		lineEnd = len(text) - startOffset
	}
	line := strings.TrimSpace(strings.TrimRight(text[startOffset:startOffset+lineEnd], "\r\n"))
	return line == strings.TrimSpace(title) && chapterTitlePattern.MatchString(line)
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
