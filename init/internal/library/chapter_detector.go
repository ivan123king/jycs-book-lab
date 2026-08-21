// chapter_detector.go 负责把导入的小说 TXT 拆分为章节候选。
package library

import (
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"
)

var chapterTitlePattern = regexp.MustCompile(`(?i)^\s*((第[零〇一二三四五六七八九十百千万两0-9]{1,12}\s*[章节回卷部集][^\r\n]{0,80})|(卷[零〇一二三四五六七八九十百千万两0-9]{1,12}[^\r\n]{0,80})|(chapter\s+[0-9ivxlcdm]+[^\r\n]{0,80}))\s*$`)

type chapterMarker struct {
	title          string
	titleByteStart int
	titleByteEnd   int
	bodyByteStart  int
}

// DetectChapters 将导入原文拆分为可编辑的章节候选。
// 识别结果只是导入草稿，用户确认后才会写入单本小说项目目录。
func DetectChapters(text string) []BookChapter {
	markers := collectChapterMarkers(text)
	if len(markers) == 0 {
		return []BookChapter{
			{
				ID:          "chapter-000001",
				Title:       "正文",
				Order:       1,
				ContentPath: "chapters/chapter-000001.txt",
				StartOffset: 0,
				EndOffset:   len(text),
				WordCount:   countTextRunes(text),
			},
		}
	}

	chapters := make([]BookChapter, 0, len(markers))
	for index, marker := range markers {
		endOffset := len(text)
		if index+1 < len(markers) {
			endOffset = markers[index+1].titleByteStart
		}

		bodyStart := marker.bodyByteStart
		if bodyStart > endOffset {
			bodyStart = marker.titleByteEnd
		}

		chapterID := fmt.Sprintf("chapter-%06d", index+1)
		chapters = append(chapters, BookChapter{
			ID:          chapterID,
			Title:       marker.title,
			Order:       index + 1,
			ContentPath: fmt.Sprintf("chapters/%s.txt", chapterID),
			StartOffset: marker.titleByteStart,
			EndOffset:   endOffset,
			WordCount:   countTextRunes(text[bodyStart:endOffset]),
		})
	}

	return chapters
}

func collectChapterMarkers(text string) []chapterMarker {
	lines := strings.SplitAfter(text, "\n")
	markers := make([]chapterMarker, 0)
	offset := 0

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		lineEnd := offset + len(line)
		if chapterTitlePattern.MatchString(trimmed) {
			markers = append(markers, chapterMarker{
				title:          trimmed,
				titleByteStart: offset,
				titleByteEnd:   offset + len(strings.TrimRight(line, "\r\n")),
				bodyByteStart:  lineEnd,
			})
		}
		offset = lineEnd
	}

	return markers
}

func countTextRunes(text string) int {
	count := 0
	for len(text) > 0 {
		r, size := utf8.DecodeRuneInString(text)
		if size == 0 {
			break
		}
		if !isIgnoredTextRune(r) {
			count++
		}
		text = text[size:]
	}
	return count
}

func isIgnoredTextRune(r rune) bool {
	return r == '\r' || r == '\n' || r == '\t' || r == ' ' || r == '\u3000'
}
