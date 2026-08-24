// chapter_memo_service.go 负责章节级自由小记的读写和排序。
package library

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// SaveChapterMemo 保存或清空章节目录中的自由小记，正式数据写入单本小说项目目录。
func SaveChapterMemo(request SaveChapterMemoRequest) (SaveChapterMemoResult, error) {
	project, err := OpenBookProject(request.BookPath)
	if err != nil {
		return SaveChapterMemoResult{}, err
	}
	bookPath := projectRootPath(request.BookPath)
	if err := ensureReadingSupportFiles(bookPath, project); err != nil {
		return SaveChapterMemoResult{}, err
	}
	chapter, ok := findChapterByID(project, request.ChapterID)
	if !ok {
		return SaveChapterMemoResult{}, fmt.Errorf("未找到章节: %s", request.ChapterID)
	}
	collection, err := loadChapterMemos(bookPath)
	if err != nil {
		return SaveChapterMemoResult{}, err
	}

	now := time.Now().Format(time.RFC3339)
	memoIndex := -1
	for index := range collection.Items {
		if collection.Items[index].BookID == project.Book.ID && collection.Items[index].ChapterID == chapter.ID {
			memoIndex = index
			break
		}
	}
	memo := ChapterMemo{
		BookID:       project.Book.ID,
		ChapterID:    chapter.ID,
		ChapterOrder: chapter.Order,
		Content:      request.Content,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if memoIndex >= 0 {
		memo.CreatedAt = collection.Items[memoIndex].CreatedAt
		if strings.TrimSpace(request.Content) == "" {
			collection.Items = append(collection.Items[:memoIndex], collection.Items[memoIndex+1:]...)
		} else {
			collection.Items[memoIndex] = memo
		}
	} else if strings.TrimSpace(request.Content) != "" {
		collection.Items = append(collection.Items, memo)
	}
	sort.SliceStable(collection.Items, func(i, j int) bool {
		if collection.Items[i].ChapterOrder != collection.Items[j].ChapterOrder {
			return collection.Items[i].ChapterOrder < collection.Items[j].ChapterOrder
		}
		return collection.Items[i].UpdatedAt < collection.Items[j].UpdatedAt
	})
	if err := writeJSONFileAtomic(filepath.Join(bookPath, filepath.FromSlash(chapterMemosRelativePath)), collection); err != nil {
		return SaveChapterMemoResult{}, fmt.Errorf("保存章节小记失败: %w", err)
	}
	return SaveChapterMemoResult{Memo: memo, ChapterMemos: collection}, nil
}

// loadChapterMemos 读取章节级小记集合，缺失文件由阅读支持文件初始化负责补齐。
func loadChapterMemos(bookPath string) (ChapterMemoCollection, error) {
	collection := ChapterMemoCollection{Version: 1, Items: []ChapterMemo{}}
	if err := readJSONFile(filepath.Join(bookPath, filepath.FromSlash(chapterMemosRelativePath)), &collection); err != nil {
		return collection, fmt.Errorf("读取章节小记失败: %w", err)
	}
	if collection.Version == 0 {
		collection.Version = 1
	}
	if collection.Items == nil {
		collection.Items = []ChapterMemo{}
	}
	return collection, nil
}
