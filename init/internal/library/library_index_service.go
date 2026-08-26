// library_index_service.go 负责书库根目录、index.json 轻量索引和书籍移除流程。
package library

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

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
	index = enrichLibraryReadingProgress(index)
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
	index = enrichLibraryReadingProgress(index)
	return LibraryState{LibraryPath: cleanPath, Index: index}, nil
}

// emptyLibraryIndex 创建空书库索引，保证 books 字段始终可序列化为数组。
func emptyLibraryIndex() LibraryIndex {
	return LibraryIndex{Version: 1, Books: []LibraryBookEntry{}}
}

// enrichLibraryReadingProgress 从单书正式阅读状态补齐书库列表进度缓存。
func enrichLibraryReadingProgress(index LibraryIndex) LibraryIndex {
	if index.Books == nil {
		index.Books = []LibraryBookEntry{}
	}
	for bookIndex := range index.Books {
		progress, err := readBookReadingProgress(index.Books[bookIndex].BookPath)
		if err != nil {
			index.Books[bookIndex].ReadingProgressPercent = 0
			continue
		}
		index.Books[bookIndex].ReadingProgressPercent = progress
	}
	return index
}

// readBookReadingProgress 从单本小说目录估算或读取阅读进度，index.json 不是唯一来源。
func readBookReadingProgress(bookPath string) (int, error) {
	project, err := readBookProjectSummary(bookPath)
	if err != nil {
		return 0, err
	}
	var state ReadingState
	if err := readOptionalJSONFile(filepath.Join(projectRootPath(bookPath), filepath.FromSlash(readingStateRelativePath)), &state); errors.Is(err, os.ErrNotExist) {
		return 0, nil
	} else if err != nil {
		return 0, err
	}
	if state.ReadingProgressPercent > 0 {
		return clampPercent(state.ReadingProgressPercent), nil
	}
	return estimateReadingProgressPercent(project, state.LastPosition), nil
}

// ensureLibraryIndex 确保书库根目录和轻量索引文件存在。
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

// appendBookToIndex 将导入后的小说入口写入书库轻量索引；最近导入的书排在最前。
func appendBookToIndex(libraryPath string, entry LibraryBookEntry) (LibraryIndex, error) {
	index, err := ensureLibraryIndex(libraryPath)
	if err != nil {
		return LibraryIndex{}, err
	}

	index.Books = moveLibraryEntryToFront(index.Books, entry)

	if err := writeJSONFileAtomic(filepath.Join(libraryPath, libraryIndexName), index); err != nil {
		return LibraryIndex{}, err
	}
	return index, nil
}

// PromoteBookInLibrary 将已打开的书移动到书库列表最前，方便几百本书时快速回到最近使用项。
func PromoteBookInLibrary(libraryPath string, bookPath string) (LibraryIndex, error) {
	libraryPath = filepath.Clean(strings.TrimSpace(libraryPath))
	if libraryPath == "." || libraryPath == "" {
		return LibraryIndex{}, errors.New("书库目录不能为空")
	}
	bookPath = filepath.Clean(strings.TrimSpace(bookPath))
	if bookPath == "." || bookPath == "" {
		return LibraryIndex{}, errors.New("小说目录不能为空")
	}

	index, err := ensureLibraryIndex(libraryPath)
	if err != nil {
		return LibraryIndex{}, err
	}
	for _, entry := range index.Books {
		if samePath(entry.BookPath, bookPath) {
			index.Books = moveLibraryEntryToFront(index.Books, entry)
			if err := writeJSONFileAtomic(filepath.Join(libraryPath, libraryIndexName), index); err != nil {
				return LibraryIndex{}, err
			}
			return index, nil
		}
	}
	return index, nil
}

// moveLibraryEntryToFront 在保持其他书籍相对顺序不变的前提下，把目标入口移到第一位。
func moveLibraryEntryToFront(books []LibraryBookEntry, entry LibraryBookEntry) []LibraryBookEntry {
	nextBooks := make([]LibraryBookEntry, 0, len(books)+1)
	nextBooks = append(nextBooks, entry)
	for _, book := range books {
		if samePath(book.BookPath, entry.BookPath) {
			continue
		}
		nextBooks = append(nextBooks, book)
	}
	return nextBooks
}

// RemoveBookFromLibrary 从书库 index.json 移除入口，并删除该书在 books 目录下的完整项目数据。
func RemoveBookFromLibrary(libraryPath string, bookPath string) (LibraryState, error) {
	libraryPath = filepath.Clean(strings.TrimSpace(libraryPath))
	if libraryPath == "." || libraryPath == "" {
		return LibraryState{}, errors.New("请先设置书库目录")
	}
	bookPath = filepath.Clean(strings.TrimSpace(bookPath))
	if bookPath == "." || bookPath == "" {
		return LibraryState{}, errors.New("要移除的小说目录不能为空")
	}

	index, err := ensureLibraryIndex(libraryPath)
	if err != nil {
		return LibraryState{}, err
	}

	booksRoot := filepath.Join(libraryPath, "books")
	nextBooks := make([]LibraryBookEntry, 0, len(index.Books))
	var removedEntry LibraryBookEntry
	found := false
	for _, entry := range index.Books {
		if samePath(entry.BookPath, bookPath) {
			removedEntry = entry
			found = true
			continue
		}
		nextBooks = append(nextBooks, entry)
	}
	if !found {
		return LibraryState{}, errors.New("书库中未找到要移除的小说")
	}

	// 重要数据删除只允许落在当前书库 books 子目录内，避免异常索引误删外部文件。
	if !isPathInside(booksRoot, removedEntry.BookPath) {
		return LibraryState{}, errors.New("小说目录不在当前书库 books 目录下，已拒绝移除")
	}
	// 移除书籍前先把该书已有素材复制到书库素材正式数据，避免书库素材引用随单书目录删除而失效。
	if _, err := MigrateBookMaterialsToLibrary(libraryPath, removedEntry.BookPath); err != nil {
		return LibraryState{}, fmt.Errorf("迁移小说素材到书库素材失败: %w", err)
	}
	if err := os.RemoveAll(removedEntry.BookPath); err != nil {
		return LibraryState{}, fmt.Errorf("删除小说项目目录失败: %w", err)
	}

	index.Books = nextBooks
	if index.Books == nil {
		index.Books = []LibraryBookEntry{}
	}
	if err := writeJSONFileAtomic(filepath.Join(libraryPath, libraryIndexName), index); err != nil {
		return LibraryState{}, fmt.Errorf("更新书库索引失败: %w", err)
	}
	if err := clearLastOpenedBookPathIfMatches(bookPath); err != nil {
		return LibraryState{}, err
	}
	return LibraryState{LibraryPath: libraryPath, Index: enrichLibraryReadingProgress(index)}, nil
}
