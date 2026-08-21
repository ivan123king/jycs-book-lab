// service.go 负责书库目录、小说导入和本地 JSON/TXT 持久化。
package library

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/encoding/simplifiedchinese"
)

const (
	appDataFolderName = "jiayi-book-lab"
	settingsFileName  = "settings.json"
	libraryIndexName  = "index.json"
	bookJSONName      = "book.json"
	originalTextName  = "original.txt"
)

// loadSettings 从应用系统数据目录读取全局设置，不读取或写入单本小说业务数据。
func loadSettings() (AppSettings, error) {
	settings := AppSettings{Version: 1}
	path, err := settingsPath()
	if err != nil {
		return settings, err
	}

	content, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return settings, nil
	}
	if err != nil {
		return settings, fmt.Errorf("读取应用设置失败: %w", err)
	}
	if err := json.Unmarshal(content, &settings); err != nil {
		return settings, fmt.Errorf("解析应用设置失败: %w", err)
	}
	if settings.Version == 0 {
		settings.Version = 1
	}
	return settings, nil
}

// saveSettings 将书库目录等全局偏好保存到应用系统数据目录。
func saveSettings(settings AppSettings) error {
	if settings.Version == 0 {
		settings.Version = 1
	}
	path, err := settingsPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("创建应用设置目录失败: %w", err)
	}
	return writeJSONFileAtomic(path, settings)
}

func settingsPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("获取应用设置目录失败: %w", err)
	}
	return filepath.Join(configDir, appDataFolderName, settingsFileName), nil
}

// LoadLibraryState 读取当前书库目录和 index.json；没有设置书库时返回空状态。
func LoadLibraryState() (LibraryState, error) {
	settings, err := loadSettings()
	if err != nil {
		return LibraryState{}, err
	}
	if strings.TrimSpace(settings.LibraryPath) == "" {
		return LibraryState{Index: emptyLibraryIndex()}, nil
	}
	index, err := ensureLibraryIndex(settings.LibraryPath)
	if err != nil {
		return LibraryState{LibraryPath: settings.LibraryPath, Index: emptyLibraryIndex()}, err
	}
	return LibraryState{LibraryPath: settings.LibraryPath, Index: index}, nil
}

// SetLibraryPath 初始化用户选择的书库目录，并保存为应用级设置。
func SetLibraryPath(libraryPath string) (LibraryState, error) {
	cleanPath := filepath.Clean(strings.TrimSpace(libraryPath))
	if cleanPath == "." || cleanPath == "" {
		return LibraryState{}, errors.New("书库目录不能为空")
	}
	if err := os.MkdirAll(filepath.Join(cleanPath, "books"), 0o755); err != nil {
		return LibraryState{}, fmt.Errorf("创建书库目录失败: %w", err)
	}
	index, err := ensureLibraryIndex(cleanPath)
	if err != nil {
		return LibraryState{}, err
	}
	if err := saveSettings(AppSettings{Version: 1, LibraryPath: cleanPath}); err != nil {
		return LibraryState{}, err
	}
	return LibraryState{LibraryPath: cleanPath, Index: index}, nil
}

func emptyLibraryIndex() LibraryIndex {
	return LibraryIndex{Version: 1, Books: []LibraryBookEntry{}}
}

func ensureLibraryIndex(libraryPath string) (LibraryIndex, error) {
	if err := os.MkdirAll(filepath.Join(libraryPath, "books"), 0o755); err != nil {
		return LibraryIndex{}, fmt.Errorf("创建书库 books 目录失败: %w", err)
	}

	indexPath := filepath.Join(libraryPath, libraryIndexName)
	index := emptyLibraryIndex()
	content, err := os.ReadFile(indexPath)
	if errors.Is(err, os.ErrNotExist) {
		if err := writeJSONFileAtomic(indexPath, index); err != nil {
			return LibraryIndex{}, err
		}
		return index, nil
	}
	if err != nil {
		return LibraryIndex{}, fmt.Errorf("读取书库索引失败: %w", err)
	}
	if err := json.Unmarshal(content, &index); err != nil {
		return LibraryIndex{}, fmt.Errorf("解析书库索引失败: %w", err)
	}
	if index.Version == 0 {
		index.Version = 1
	}
	if index.Books == nil {
		index.Books = []LibraryBookEntry{}
	}
	return index, nil
}

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
		Library: LibraryState{LibraryPath: libraryPath, Index: index},
		Project: project,
	}, nil
}

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
	if err := os.Rename(tempPath, targetPath); err != nil {
		_ = os.RemoveAll(tempPath)
		return BookProject{}, fmt.Errorf("提交小说目录失败: %w", err)
	}
	return project, nil
}

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

func appendBookToIndex(libraryPath string, entry LibraryBookEntry) (LibraryIndex, error) {
	index, err := ensureLibraryIndex(libraryPath)
	if err != nil {
		return LibraryIndex{}, err
	}

	replaced := false
	for i := range index.Books {
		if filepath.Clean(index.Books[i].BookPath) == filepath.Clean(entry.BookPath) {
			index.Books[i] = entry
			replaced = true
			break
		}
	}
	if !replaced {
		index.Books = append(index.Books, entry)
	}
	sort.SliceStable(index.Books, func(i, j int) bool {
		return strings.ToLower(index.Books[i].BookName) < strings.ToLower(index.Books[j].BookName)
	})

	if err := writeJSONFileAtomic(filepath.Join(libraryPath, libraryIndexName), index); err != nil {
		return LibraryIndex{}, err
	}
	return index, nil
}

// OpenBookProject 以单本小说目录中的 book.json 作为入口读取正式书籍数据。
func OpenBookProject(bookPath string) (BookProject, error) {
	cleanBookPath := filepath.Clean(bookPath)
	bookJSONPath := filepath.Join(cleanBookPath, bookJSONName)
	content, err := os.ReadFile(bookJSONPath)
	if err != nil {
		return BookProject{}, fmt.Errorf("读取小说数据失败: %w", err)
	}
	var project BookProject
	if err := json.Unmarshal(content, &project); err != nil {
		return BookProject{}, fmt.Errorf("解析小说数据失败: %w", err)
	}
	if project.Version == 0 {
		project.Version = 1
	}
	if err := recoverLeadingUnknownChapter(cleanBookPath, &project); err != nil {
		return BookProject{}, err
	}
	return project, nil
}

// recoverLeadingUnknownChapter 修复旧导入数据：第一个可识别标题前的正文应成为“未知标题”章节。
func recoverLeadingUnknownChapter(bookPath string, project *BookProject) error {
	if len(project.Chapters) == 0 || project.Chapters[0].StartOffset <= 0 {
		return nil
	}

	sourceTextPath := project.Book.SourceTextPath
	if strings.TrimSpace(sourceTextPath) == "" {
		sourceTextPath = originalTextName
	}
	originalPath := filepath.Join(bookPath, filepath.FromSlash(sourceTextPath))
	content, err := os.ReadFile(originalPath)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("读取原始导入文本用于修复章节失败: %w", err)
	}
	text := string(trimUTF8BOM(content))
	firstChapter := project.Chapters[0]
	if firstChapter.StartOffset > len(text) {
		return nil
	}
	prefixText := text[:firstChapter.StartOffset]
	if !hasReadableText(prefixText) {
		return nil
	}

	prefixID := "chapter-000000"
	prefixChapter := BookChapter{
		ID:          prefixID,
		Title:       unknownChapterTitle,
		Order:       1,
		ContentPath: "chapters/chapter-000000.txt",
		StartOffset: 0,
		EndOffset:   firstChapter.StartOffset,
		WordCount:   countTextRunes(prefixText),
	}
	for index := range project.Chapters {
		project.Chapters[index].Order++
	}
	project.Chapters = append([]BookChapter{prefixChapter}, project.Chapters...)
	project.Book.UpdatedAt = time.Now().Format(time.RFC3339)

	prefixPath := filepath.Join(bookPath, filepath.FromSlash(prefixChapter.ContentPath))
	if err := os.MkdirAll(filepath.Dir(prefixPath), 0o755); err != nil {
		return fmt.Errorf("创建未知标题章节目录失败: %w", err)
	}
	if err := os.WriteFile(prefixPath, []byte(strings.TrimLeft(prefixText, "\r\n")), 0o644); err != nil {
		return fmt.Errorf("写入未知标题章节正文失败: %w", err)
	}
	if err := writeJSONFileAtomic(filepath.Join(bookPath, bookJSONName), *project); err != nil {
		return fmt.Errorf("保存修复后的小说数据失败: %w", err)
	}
	if err := shiftPersistedChapterOrdersAfterLeadingRecovery(bookPath, prefixID); err != nil {
		return err
	}
	return nil
}

// shiftPersistedChapterOrdersAfterLeadingRecovery 同步被前置章节后移影响的轻量位置序号。
func shiftPersistedChapterOrdersAfterLeadingRecovery(bookPath string, prefixID string) error {
	readingStatePath := filepath.Join(bookPath, "reading", "reading-state.json")
	var state ReadingState
	if err := readOptionalJSONFile(readingStatePath, &state); err == nil && state.LastPosition.ChapterID != "" {
		if state.LastPosition.ChapterID != prefixID && state.LastPosition.ChapterOrder > 0 {
			state.LastPosition.ChapterOrder++
			if err := writeJSONFileAtomic(readingStatePath, state); err != nil {
				return fmt.Errorf("同步阅读位置章节序号失败: %w", err)
			}
		}
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("读取阅读位置用于迁移失败: %w", err)
	}

	sourceTextsPath := filepath.Join(bookPath, "source-texts", "source-texts.json")
	collection := SourceTextCollection{Version: 1}
	if err := readOptionalJSONFile(sourceTextsPath, &collection); err == nil && collection.Items != nil {
		changed := false
		for index := range collection.Items {
			if collection.Items[index].ChapterID != prefixID && collection.Items[index].ChapterOrder > 0 {
				collection.Items[index].ChapterOrder++
				changed = true
			}
		}
		if changed {
			if err := writeJSONFileAtomic(sourceTextsPath, collection); err != nil {
				return fmt.Errorf("同步选文来源章节序号失败: %w", err)
			}
		}
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("读取选文来源用于迁移失败: %w", err)
	}
	return nil
}

func readOptionalJSONFile(path string, target any) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(content, target); err != nil {
		return err
	}
	return nil
}

func readNovelText(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("读取小说文本失败: %w", err)
	}
	content = trimUTF8BOM(content)
	if utf8.Valid(content) {
		return string(content), nil
	}
	decoded, err := simplifiedchinese.GB18030.NewDecoder().Bytes(content)
	if err != nil {
		return "", fmt.Errorf("小说文本编码识别失败，请先转换为 UTF-8 或 GB18030 后重试: %w", err)
	}
	return string(decoded), nil
}

func trimUTF8BOM(content []byte) []byte {
	if len(content) >= 3 && content[0] == 0xEF && content[1] == 0xBB && content[2] == 0xBF {
		return content[3:]
	}
	return content
}

func copyFile(source string, target string) error {
	input, err := os.Open(source)
	if err != nil {
		return fmt.Errorf("打开导入文件失败: %w", err)
	}
	defer input.Close()

	output, err := os.Create(target)
	if err != nil {
		return fmt.Errorf("创建原文副本失败: %w", err)
	}
	defer output.Close()

	if _, err := io.Copy(output, input); err != nil {
		return fmt.Errorf("复制原文副本失败: %w", err)
	}
	return output.Sync()
}

func writeJSONFileAtomic(path string, value any) error {
	content, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化 JSON 失败: %w", err)
	}
	content = append(content, '\n')

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("创建 JSON 目录失败: %w", err)
	}
	tempFile, err := os.CreateTemp(filepath.Dir(path), filepath.Base(path)+".*.tmp")
	if err != nil {
		return fmt.Errorf("创建临时 JSON 文件失败: %w", err)
	}
	tempPath := tempFile.Name()
	if _, err := tempFile.Write(content); err != nil {
		tempFile.Close()
		_ = os.Remove(tempPath)
		return fmt.Errorf("写入临时 JSON 文件失败: %w", err)
	}
	if err := tempFile.Close(); err != nil {
		_ = os.Remove(tempPath)
		return fmt.Errorf("关闭临时 JSON 文件失败: %w", err)
	}
	if err := os.Rename(tempPath, path); err != nil {
		if _, statErr := os.Stat(path); statErr != nil {
			_ = os.Remove(tempPath)
			return fmt.Errorf("替换 JSON 文件失败: %w", err)
		}
		if removeErr := os.Remove(path); removeErr != nil {
			_ = os.Remove(tempPath)
			return fmt.Errorf("删除旧 JSON 文件失败: %w", removeErr)
		}
		if renameErr := os.Rename(tempPath, path); renameErr != nil {
			_ = os.Remove(tempPath)
			return fmt.Errorf("替换 JSON 文件失败: %w", renameErr)
		}
	}
	return nil
}

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
