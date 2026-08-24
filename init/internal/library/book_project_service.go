// book_project_service.go 负责单本小说项目 book.json 读取和旧项目数据恢复。
package library

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// readBookProjectSummary 读取单本小说 book.json 的轻量项目数据。
func readBookProjectSummary(bookPath string) (BookProject, error) {
	cleanBookPath := filepath.Clean(bookPath)
	content, err := os.ReadFile(filepath.Join(cleanBookPath, bookJSONName))
	if err != nil {
		return BookProject{}, err
	}
	var project BookProject
	if err := json.Unmarshal(content, &project); err != nil {
		return BookProject{}, err
	}
	if project.Version == 0 {
		project.Version = 1
	}
	return project, nil
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

	chapterMemosPath := filepath.Join(bookPath, "reading", "chapter-memos.json")
	memos := ChapterMemoCollection{Version: 1}
	if err := readOptionalJSONFile(chapterMemosPath, &memos); err == nil && memos.Items != nil {
		changed := false
		for index := range memos.Items {
			if memos.Items[index].ChapterID != prefixID && memos.Items[index].ChapterOrder > 0 {
				memos.Items[index].ChapterOrder++
				changed = true
			}
		}
		if changed {
			if err := writeJSONFileAtomic(chapterMemosPath, memos); err != nil {
				return fmt.Errorf("同步章节小记序号失败: %w", err)
			}
		}
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("读取章节小记用于迁移失败: %w", err)
	}
	return nil
}
