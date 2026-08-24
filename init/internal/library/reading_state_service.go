// reading_state_service.go 负责单本小说阅读位置和阅读偏好的正式数据维护。
package library

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

// SaveReadingPosition 将阅读位置保存到单本小说项目目录，作为恢复阅读的正式来源。
func SaveReadingPosition(request SaveReadingPositionRequest) (ReadingState, error) {
	project, err := OpenBookProject(request.BookPath)
	if err != nil {
		return ReadingState{}, err
	}
	bookPath := projectRootPath(request.BookPath)
	chapter := chooseChapterForPosition(project, request.ChapterID)
	mode := strings.TrimSpace(request.LastMode)
	if mode != "edit" {
		mode = "read"
	}
	now := time.Now().Format(time.RFC3339)
	position := ReadingPosition{
		ChapterID:          chapter.ID,
		ChapterOrder:       chapter.Order,
		CharOffset:         maxInt(0, request.CharOffset),
		ParagraphIndex:     maxInt(0, request.ParagraphIndex),
		RenderedPageNumber: maxInt(1, request.RenderedPageNumber),
		UpdatedAt:          now,
	}
	progressPercent := clampPercent(request.ReadingProgressPercent)
	if progressPercent == 0 && (position.ChapterOrder > 1 || position.CharOffset > 0) {
		progressPercent = estimateReadingProgressPercent(project, position)
	}
	existingState, err := loadReadingState(bookPath, project)
	if err != nil {
		return ReadingState{}, err
	}
	state := ReadingState{
		Version:                1,
		BookID:                 project.Book.ID,
		LastPosition:           position,
		LastMode:               mode,
		ReadingProgressPercent: progressPercent,
		Preferences:            normalizeReadingPreferences(existingState.Preferences),
	}
	if err := writeJSONFileAtomic(filepath.Join(bookPath, filepath.FromSlash(readingStateRelativePath)), state); err != nil {
		return ReadingState{}, fmt.Errorf("保存阅读位置失败: %w", err)
	}
	return state, nil
}

// SaveReadingPreferences 将单本小说自己的阅读界面偏好保存到项目目录。
func SaveReadingPreferences(request SaveReadingPreferencesRequest) (ReadingState, error) {
	project, err := OpenBookProject(request.BookPath)
	if err != nil {
		return ReadingState{}, err
	}
	bookPath := projectRootPath(request.BookPath)
	if err := ensureReadingSupportFiles(bookPath, project); err != nil {
		return ReadingState{}, err
	}
	state, err := loadReadingState(bookPath, project)
	if err != nil {
		return ReadingState{}, err
	}
	state.Preferences = normalizeReadingPreferences(ReadingPreferences{FontSize: request.FontSize})
	if err := writeJSONFileAtomic(filepath.Join(bookPath, filepath.FromSlash(readingStateRelativePath)), state); err != nil {
		return ReadingState{}, fmt.Errorf("保存阅读偏好失败: %w", err)
	}
	return state, nil
}

// defaultReadingState 创建旧项目缺失阅读状态时使用的默认位置。
func defaultReadingState(bookID string, chapter BookChapter) ReadingState {
	return ReadingState{
		Version:                1,
		BookID:                 bookID,
		ReadingProgressPercent: 0,
		Preferences:            ReadingPreferences{FontSize: defaultReaderFontSize},
		LastPosition: ReadingPosition{
			ChapterID:          chapter.ID,
			ChapterOrder:       chapter.Order,
			CharOffset:         0,
			ParagraphIndex:     0,
			RenderedPageNumber: 1,
			UpdatedAt:          time.Now().Format(time.RFC3339),
		},
		LastMode: "read",
	}
}

// loadReadingState 读取并归一化单书正式阅读状态。
func loadReadingState(bookPath string, project BookProject) (ReadingState, error) {
	var state ReadingState
	if err := readJSONFile(filepath.Join(bookPath, filepath.FromSlash(readingStateRelativePath)), &state); err != nil {
		return ReadingState{}, fmt.Errorf("读取阅读位置失败: %w", err)
	}
	if state.Version == 0 {
		state.Version = 1
	}
	if state.BookID == "" {
		state.BookID = project.Book.ID
	}
	if state.LastMode == "" {
		state.LastMode = "read"
	}
	state.Preferences = normalizeReadingPreferences(state.Preferences)
	if state.LastPosition.ChapterID == "" {
		defaultState := defaultReadingState(project.Book.ID, chooseChapterForPosition(project, ""))
		defaultState.Preferences = state.Preferences
		return defaultState, nil
	}
	if state.ReadingProgressPercent == 0 && (state.LastPosition.ChapterOrder > 1 || state.LastPosition.CharOffset > 0) {
		state.ReadingProgressPercent = estimateReadingProgressPercent(project, state.LastPosition)
	} else {
		state.ReadingProgressPercent = clampPercent(state.ReadingProgressPercent)
	}
	return state, nil
}

// normalizeReadingPreferences 补齐并限制当前书阅读偏好范围。
func normalizeReadingPreferences(preferences ReadingPreferences) ReadingPreferences {
	if preferences.FontSize == 0 {
		preferences.FontSize = defaultReaderFontSize
	}
	preferences.FontSize = clampInt(preferences.FontSize, minReaderFontSize, maxReaderFontSize)
	return preferences
}

// estimateReadingProgressPercent 根据章节字数和阅读位置估算书库展示进度。
func estimateReadingProgressPercent(project BookProject, position ReadingPosition) int {
	totalWordCount := 0
	completedWordCount := 0
	for _, chapter := range project.Chapters {
		chapterWordCount := maxInt(0, chapter.WordCount)
		totalWordCount += chapterWordCount
		if chapter.ID == position.ChapterID {
			completedWordCount += minInt(chapterWordCount, maxInt(0, position.CharOffset))
			continue
		}
		if chapter.Order < position.ChapterOrder {
			completedWordCount += chapterWordCount
		}
	}
	if totalWordCount <= 0 {
		return 0
	}
	progress := int(float64(completedWordCount) / float64(totalWordCount) * 100)
	return clampPercent(progress)
}

// clampPercent 将百分比限制在 0 到 100 之间。
func clampPercent(value int) int {
	if value < 0 {
		return 0
	}
	if value > 100 {
		return 100
	}
	return value
}

// clampInt 将数值限制在给定范围内。
func clampInt(value int, minValue int, maxValue int) int {
	if value < minValue {
		return minValue
	}
	if value > maxValue {
		return maxValue
	}
	return value
}
