// chapter_edit_service.go 负责章节正文编辑、标题修改和目录拆分落盘。
package library

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// SaveChapterContent 保存编辑 TAB 中的章节标题和正文，并同步更新 book.json 章节元数据。
func SaveChapterContent(request SaveChapterContentRequest) (ChapterContentResult, error) {
	project, err := OpenBookProject(request.BookPath)
	if err != nil {
		return ChapterContentResult{}, err
	}
	bookPath := projectRootPath(request.BookPath)
	chapterIndex := -1
	for index := range project.Chapters {
		if project.Chapters[index].ID == request.ChapterID {
			chapterIndex = index
			break
		}
	}
	if chapterIndex < 0 {
		return ChapterContentResult{}, fmt.Errorf("未找到章节: %s", request.ChapterID)
	}
	title := strings.TrimSpace(request.Title)
	if title == "" {
		return ChapterContentResult{}, errors.New("章节标题不能为空")
	}

	chapter := project.Chapters[chapterIndex]
	chapter.Title = title
	chapter.WordCount = countTextRunes(request.Content)
	project.Chapters[chapterIndex] = chapter
	project.Book.UpdatedAt = time.Now().Format(time.RFC3339)

	contentPath, err := safeProjectPath(bookPath, chapter.ContentPath)
	if err != nil {
		return ChapterContentResult{}, err
	}
	if err := writeTextFileAtomic(contentPath, request.Content); err != nil {
		return ChapterContentResult{}, fmt.Errorf("保存章节正文失败: %w", err)
	}
	if err := writeJSONFileAtomic(filepath.Join(bookPath, bookJSONName), project); err != nil {
		return ChapterContentResult{}, fmt.Errorf("保存章节元数据失败: %w", err)
	}
	return ChapterContentResult{Chapter: chapter, Content: request.Content}, nil
}

// SplitChapterFromStart 从当前章节开头拆出一个新章节，当前章节保留剩余正文。
func SplitChapterFromStart(request SplitChapterRequest) (SplitChapterResult, error) {
	project, err := OpenBookProject(request.BookPath)
	if err != nil {
		return SplitChapterResult{}, err
	}
	bookPath := projectRootPath(request.BookPath)
	if err := ensureReadingSupportFiles(bookPath, project); err != nil {
		return SplitChapterResult{}, err
	}

	chapterIndex := -1
	for index := range project.Chapters {
		if project.Chapters[index].ID == request.ChapterID {
			chapterIndex = index
			break
		}
	}
	if chapterIndex < 0 {
		return SplitChapterResult{}, fmt.Errorf("未找到章节: %s", request.ChapterID)
	}
	newTitle := strings.TrimSpace(request.NewTitle)
	if newTitle == "" {
		return SplitChapterResult{}, errors.New("新章节标题不能为空")
	}
	if strings.TrimSpace(request.NewContent) == "" {
		return SplitChapterResult{}, errors.New("新章节正文不能为空")
	}
	if strings.TrimSpace(request.RemainingContent) == "" {
		return SplitChapterResult{}, errors.New("当前章节剩余正文不能为空")
	}
	if err := ensureChapterHasNoSourceTexts(bookPath, request.ChapterID); err != nil {
		return SplitChapterResult{}, err
	}

	currentChapter := project.Chapters[chapterIndex]
	newChapterID := nextChapterID(project)
	newChapter := BookChapter{
		ID:          newChapterID,
		Title:       newTitle,
		Order:       currentChapter.Order,
		ContentPath: fmt.Sprintf("chapters/%s.txt", newChapterID),
		StartOffset: currentChapter.StartOffset,
		EndOffset:   currentChapter.StartOffset + len(request.NewContent),
		WordCount:   countTextRunes(request.NewContent),
	}
	remainingChapter := currentChapter
	remainingChapter.Order = currentChapter.Order + 1
	remainingChapter.StartOffset = newChapter.EndOffset
	remainingChapter.EndOffset = newChapter.EndOffset + len(request.RemainingContent)
	remainingChapter.WordCount = countTextRunes(request.RemainingContent)

	nextChapters := make([]BookChapter, 0, len(project.Chapters)+1)
	for index, chapter := range project.Chapters {
		if index == chapterIndex {
			nextChapters = append(nextChapters, newChapter, remainingChapter)
			continue
		}
		if chapter.Order > currentChapter.Order {
			chapter.Order++
		}
		nextChapters = append(nextChapters, chapter)
	}
	project.Chapters = nextChapters
	project.Book.UpdatedAt = time.Now().Format(time.RFC3339)

	newContentPath, err := safeProjectPath(bookPath, newChapter.ContentPath)
	if err != nil {
		return SplitChapterResult{}, err
	}
	remainingContentPath, err := safeProjectPath(bookPath, remainingChapter.ContentPath)
	if err != nil {
		return SplitChapterResult{}, err
	}
	if err := writeTextFileAtomic(newContentPath, request.NewContent); err != nil {
		return SplitChapterResult{}, fmt.Errorf("保存新章节正文失败: %w", err)
	}
	if err := writeTextFileAtomic(remainingContentPath, request.RemainingContent); err != nil {
		return SplitChapterResult{}, fmt.Errorf("保存当前章节剩余正文失败: %w", err)
	}
	if err := writeJSONFileAtomic(filepath.Join(bookPath, bookJSONName), project); err != nil {
		return SplitChapterResult{}, fmt.Errorf("保存拆分后的章节目录失败: %w", err)
	}
	if err := syncPersistedChapterOrdersAfterSplit(bookPath, currentChapter.Order, currentChapter.ID, remainingChapter); err != nil {
		return SplitChapterResult{}, err
	}

	return SplitChapterResult{
		Project:          project,
		NewChapter:       newChapter,
		RemainingChapter: remainingChapter,
		NewContent:       request.NewContent,
		RemainingContent: request.RemainingContent,
	}, nil
}

// syncPersistedChapterOrdersAfterSplit 在章节拆分后同步阅读位置、选文来源和章节小记的章节序号。
func syncPersistedChapterOrdersAfterSplit(bookPath string, splitOrder int, splitChapterID string, remainingChapter BookChapter) error {
	readingStatePath := filepath.Join(bookPath, filepath.FromSlash(readingStateRelativePath))
	var state ReadingState
	if err := readOptionalJSONFile(readingStatePath, &state); err == nil && state.LastPosition.ChapterID != "" {
		changed := false
		if state.LastPosition.ChapterID == splitChapterID {
			state.LastPosition.ChapterOrder = remainingChapter.Order
			state.LastPosition.CharOffset = 0
			state.LastPosition.ParagraphIndex = 0
			state.LastPosition.RenderedPageNumber = 1
			state.ReadingProgressPercent = 0
			changed = true
		} else if state.LastPosition.ChapterOrder > splitOrder {
			state.LastPosition.ChapterOrder++
			changed = true
		}
		if changed {
			state.Preferences = normalizeReadingPreferences(state.Preferences)
			if err := writeJSONFileAtomic(readingStatePath, state); err != nil {
				return fmt.Errorf("同步阅读位置章节序号失败: %w", err)
			}
		}
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("读取阅读位置用于章节拆分失败: %w", err)
	}

	sourceTextsPath := filepath.Join(bookPath, filepath.FromSlash(sourceTextsRelativePath))
	collection := SourceTextCollection{Version: 1}
	if err := readOptionalJSONFile(sourceTextsPath, &collection); err == nil && collection.Items != nil {
		changed := false
		for index := range collection.Items {
			if collection.Items[index].ChapterOrder > splitOrder {
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
		return fmt.Errorf("读取选文来源用于章节拆分失败: %w", err)
	}

	chapterMemosPath := filepath.Join(bookPath, filepath.FromSlash(chapterMemosRelativePath))
	memos := ChapterMemoCollection{Version: 1}
	if err := readOptionalJSONFile(chapterMemosPath, &memos); err == nil && memos.Items != nil {
		changed := false
		for index := range memos.Items {
			if memos.Items[index].ChapterID == splitChapterID {
				memos.Items[index].ChapterOrder = remainingChapter.Order
				changed = true
				continue
			}
			if memos.Items[index].ChapterOrder > splitOrder {
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
		return fmt.Errorf("读取章节小记用于章节拆分失败: %w", err)
	}
	return nil
}

// nextChapterID 为拆分出的新章节生成未使用的章节 ID。
func nextChapterID(project BookProject) string {
	used := map[string]bool{}
	for _, chapter := range project.Chapters {
		used[chapter.ID] = true
	}
	for next := len(project.Chapters) + 1; ; next++ {
		chapterID := fmt.Sprintf("chapter-%06d", next)
		if !used[chapterID] {
			return chapterID
		}
	}
}
