// reading_session_service.go 负责打开阅读工作台和读取章节正文。
package library

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// OpenReadingSession 读取当前书阅读页所需的正式项目数据，并为旧项目补齐缺失文件。
func OpenReadingSession(bookPath string) (ReadingSession, error) {
	project, err := OpenBookProject(bookPath)
	if err != nil {
		return ReadingSession{}, err
	}
	bookPath = projectRootPath(bookPath)
	if err := ensureReadingSupportFiles(bookPath, project); err != nil {
		return ReadingSession{}, err
	}

	state, err := loadReadingState(bookPath, project)
	if err != nil {
		return ReadingSession{}, err
	}
	currentChapter := chooseChapterForPosition(project, state.LastPosition.ChapterID)
	content, err := readChapterText(bookPath, currentChapter)
	if err != nil {
		return ReadingSession{}, err
	}
	plan, err := loadDeconstructionPlan(bookPath, project.Book.ID)
	if err != nil {
		return ReadingSession{}, err
	}
	sourceTexts, err := loadSourceTexts(bookPath)
	if err != nil {
		return ReadingSession{}, err
	}
	cards, err := loadDeconstructionCards(bookPath)
	if err != nil {
		return ReadingSession{}, err
	}
	chapterMemos, err := loadChapterMemos(bookPath)
	if err != nil {
		return ReadingSession{}, err
	}

	state.LastPosition.ChapterID = currentChapter.ID
	state.LastPosition.ChapterOrder = currentChapter.Order
	return ReadingSession{
		Project:             project,
		State:               state,
		CurrentChapter:      currentChapter,
		ChapterContent:      content,
		Plan:                plan,
		SourceTexts:         sourceTexts,
		DeconstructionCards: cards,
		ChapterMemos:        chapterMemos,
	}, nil
}

// GetChapterContent 读取用户修正后的章节 TXT，供翻页和目录跳章使用。
func GetChapterContent(request ChapterContentRequest) (ChapterContentResult, error) {
	project, err := OpenBookProject(request.BookPath)
	if err != nil {
		return ChapterContentResult{}, err
	}
	chapter, ok := findChapterByID(project, request.ChapterID)
	if !ok {
		return ChapterContentResult{}, fmt.Errorf("未找到章节: %s", request.ChapterID)
	}
	content, err := readChapterText(projectRootPath(request.BookPath), chapter)
	if err != nil {
		return ChapterContentResult{}, err
	}
	return ChapterContentResult{Chapter: chapter, Content: content}, nil
}

// ensureReadingSupportFiles 为旧项目补齐阅读、选文、拆解和模板相关正式数据文件。
func ensureReadingSupportFiles(bookPath string, project BookProject) error {
	for _, relativeDir := range []string{"reading", "source-texts", "deconstructions", "materials", materialTextsDir} {
		if err := os.MkdirAll(filepath.Join(bookPath, relativeDir), 0o755); err != nil {
			return fmt.Errorf("创建阅读数据目录失败: %w", err)
		}
	}
	if _, err := os.Stat(filepath.Join(bookPath, filepath.FromSlash(readingStateRelativePath))); errors.Is(err, os.ErrNotExist) {
		chapter := chooseChapterForPosition(project, "")
		state := defaultReadingState(project.Book.ID, chapter)
		if err := writeJSONFileAtomic(filepath.Join(bookPath, filepath.FromSlash(readingStateRelativePath)), state); err != nil {
			return err
		}
	}
	if _, err := os.Stat(filepath.Join(bookPath, filepath.FromSlash(sourceTextsRelativePath))); errors.Is(err, os.ErrNotExist) {
		if err := writeJSONFileAtomic(filepath.Join(bookPath, filepath.FromSlash(sourceTextsRelativePath)), SourceTextCollection{Version: 1, Items: []SourceText{}}); err != nil {
			return err
		}
	}
	if _, err := os.Stat(filepath.Join(bookPath, filepath.FromSlash(deconstructionCardsRelativePath))); errors.Is(err, os.ErrNotExist) {
		legacyCards := DeconstructionCardCollection{Version: 1, Items: []DeconstructionCard{}}
		legacyPath := filepath.Join(bookPath, filepath.FromSlash(legacyDeconstructionCardsPath))
		if legacyErr := readJSONFile(legacyPath, &legacyCards); legacyErr == nil {
			legacyCards = normalizeWritingTechniqueCardCollection(legacyCards)
			if err := writeJSONFileAtomic(filepath.Join(bookPath, filepath.FromSlash(deconstructionCardsRelativePath)), legacyCards); err != nil {
				return err
			}
		} else {
			if err := writeJSONFileAtomic(filepath.Join(bookPath, filepath.FromSlash(deconstructionCardsRelativePath)), DeconstructionCardCollection{Version: 1, Items: []DeconstructionCard{}}); err != nil {
				return err
			}
		}
	}
	if _, err := os.Stat(filepath.Join(bookPath, filepath.FromSlash(chapterMemosRelativePath))); errors.Is(err, os.ErrNotExist) {
		if err := writeJSONFileAtomic(filepath.Join(bookPath, filepath.FromSlash(chapterMemosRelativePath)), ChapterMemoCollection{Version: 1, Items: []ChapterMemo{}}); err != nil {
			return err
		}
	}
	if _, err := os.Stat(filepath.Join(bookPath, filepath.FromSlash(materialsRelativePath))); errors.Is(err, os.ErrNotExist) {
		if err := writeJSONFileAtomic(filepath.Join(bookPath, filepath.FromSlash(materialsRelativePath)), MaterialCollection{Version: 1, Items: []Material{}}); err != nil {
			return err
		}
	}
	return ensureDeconstructionPlanFiles(bookPath, project.Book.ID)
}

// readChapterText 读取用户修正后的章节正文 TXT。
func readChapterText(bookPath string, chapter BookChapter) (string, error) {
	contentPath, err := safeProjectPath(bookPath, chapter.ContentPath)
	if err != nil {
		return "", err
	}
	content, err := os.ReadFile(contentPath)
	if err != nil {
		return "", fmt.Errorf("读取章节正文失败: %w", err)
	}
	return string(trimUTF8BOM(content)), nil
}

// chooseChapterForPosition 根据阅读位置选择章节，失效时回退到第一章。
func chooseChapterForPosition(project BookProject, chapterID string) BookChapter {
	if chapter, ok := findChapterByID(project, chapterID); ok {
		return chapter
	}
	if len(project.Chapters) == 0 {
		return BookChapter{ID: "chapter-000001", Title: "正文", Order: 1, ContentPath: "chapters/chapter-000001.txt"}
	}
	return project.Chapters[0]
}

// findChapterByID 在项目章节列表中查找指定章节。
func findChapterByID(project BookProject, chapterID string) (BookChapter, bool) {
	for _, chapter := range project.Chapters {
		if chapter.ID == chapterID {
			return chapter, true
		}
	}
	return BookChapter{}, false
}
