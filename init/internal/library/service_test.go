// service_test.go 验证书库导入服务会生成可迁移的小说项目目录。
package library

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestConfirmNovelImportCreatesPortableBookProject 覆盖导入闭环的正式落盘结果。
func TestConfirmNovelImportCreatesPortableBookProject(t *testing.T) {
	libraryPath := t.TempDir()
	sourcePath := filepath.Join(t.TempDir(), "雾港来信.txt")
	sourceText := "第一章 远处的灯\n雨在码头上铺开。\n\n第二章 借宿人\n她把信收进外套。\n"
	if err := os.WriteFile(sourcePath, []byte(sourceText), 0o644); err != nil {
		t.Fatalf("写入测试小说失败: %v", err)
	}

	result, err := ConfirmNovelImport(libraryPath, ConfirmImportRequest{
		BookID:      "book-test-001",
		SourcePath:  sourcePath,
		BookName:    "雾港来信",
		Description: "测试简介",
	})
	if err != nil {
		t.Fatalf("确认导入失败: %v", err)
	}

	if len(result.Library.Index.Books) != 1 {
		t.Fatalf("书库索引条目数 = %d, want 1", len(result.Library.Index.Books))
	}
	if len(result.Project.Chapters) != 2 {
		t.Fatalf("章节数 = %d, want 2", len(result.Project.Chapters))
	}

	bookPath := filepath.Join(libraryPath, "books", "book-test-001")
	for _, requiredPath := range []string{
		filepath.Join(libraryPath, libraryIndexName),
		filepath.Join(bookPath, bookJSONName),
		filepath.Join(bookPath, originalTextName),
		filepath.Join(bookPath, "chapters", "chapter-000001.txt"),
		filepath.Join(bookPath, "chapters", "chapter-000002.txt"),
	} {
		if _, err := os.Stat(requiredPath); err != nil {
			t.Fatalf("缺少导入产物 %s: %v", requiredPath, err)
		}
	}

	project, err := OpenBookProject(bookPath)
	if err != nil {
		t.Fatalf("重新打开小说项目失败: %v", err)
	}
	if project.Book.Name != "雾港来信" || project.Book.Description != "测试简介" {
		t.Fatalf("书籍元数据不正确: %+v", project.Book)
	}
	chapterContent, err := os.ReadFile(filepath.Join(bookPath, "chapters", "chapter-000001.txt"))
	if err != nil {
		t.Fatalf("读取章节正文失败: %v", err)
	}
	if string(chapterContent) != "雨在码头上铺开。\n\n" {
		t.Fatalf("章节正文 = %q, want %q", string(chapterContent), "雨在码头上铺开。\n\n")
	}
}

// TestRemoveBookFromLibraryDeletesProjectData 覆盖书库移除会删除单书项目目录并刷新轻量索引。
func TestRemoveBookFromLibraryDeletesProjectData(t *testing.T) {
	libraryPath := t.TempDir()
	sourcePath := filepath.Join(t.TempDir(), "雾港来信.txt")
	sourceText := "第一章 远处的灯\n雨在码头上铺开。\n"
	if err := os.WriteFile(sourcePath, []byte(sourceText), 0o644); err != nil {
		t.Fatalf("写入测试小说失败: %v", err)
	}
	result, err := ConfirmNovelImport(libraryPath, ConfirmImportRequest{
		BookID:     "book-remove-001",
		SourcePath: sourcePath,
		BookName:   "雾港来信",
	})
	if err != nil {
		t.Fatalf("确认导入失败: %v", err)
	}

	bookPath := result.Project.Book.BookPath
	state, err := RemoveBookFromLibrary(libraryPath, bookPath)
	if err != nil {
		t.Fatalf("移除书籍失败: %v", err)
	}
	if len(state.Index.Books) != 0 {
		t.Fatalf("移除后书库索引条目数 = %d, want 0", len(state.Index.Books))
	}
	if _, err := os.Stat(bookPath); !os.IsNotExist(err) {
		t.Fatalf("小说项目目录未删除，stat err = %v", err)
	}

	reloaded, err := ensureLibraryIndex(libraryPath)
	if err != nil {
		t.Fatalf("重新读取书库索引失败: %v", err)
	}
	if len(reloaded.Books) != 0 {
		t.Fatalf("落盘索引条目数 = %d, want 0", len(reloaded.Books))
	}
}

// TestRemoveBookFromLibraryRejectsPathOutsideBooks 确保异常索引不会导致删除书库外部数据。
func TestRemoveBookFromLibraryRejectsPathOutsideBooks(t *testing.T) {
	libraryPath := t.TempDir()
	outsideDir := filepath.Join(t.TempDir(), "outside-book")
	if err := os.MkdirAll(outsideDir, 0o755); err != nil {
		t.Fatalf("创建外部目录失败: %v", err)
	}
	index := LibraryIndex{
		Version: 1,
		Books: []LibraryBookEntry{
			{
				BookName: "异常条目",
				BookPath: outsideDir,
				BookSize: 12,
			},
		},
	}
	if err := writeJSONFileAtomic(filepath.Join(libraryPath, libraryIndexName), index); err != nil {
		t.Fatalf("写入异常索引失败: %v", err)
	}

	if _, err := RemoveBookFromLibrary(libraryPath, outsideDir); err == nil {
		t.Fatal("移除书库外部路径成功，want error")
	}
	if _, err := os.Stat(outsideDir); err != nil {
		t.Fatalf("外部目录不应被删除: %v", err)
	}
}

// TestDetectChaptersFallsBackToSingleChapter 确保无章节标题的 TXT 也能形成可打开项目。
func TestDetectChaptersFallsBackToSingleChapter(t *testing.T) {
	chapters := DetectChapters("没有章节标题的正文")
	if len(chapters) != 1 {
		t.Fatalf("章节数 = %d, want 1", len(chapters))
	}
	if chapters[0].Title != unknownChapterTitle {
		t.Fatalf("默认章节标题 = %s, want %s", chapters[0].Title, unknownChapterTitle)
	}
}

// TestDetectChaptersKeepsContentBeforeFirstRecognizedTitle 确保第一个标题前的正文不会被导入丢弃。
func TestDetectChaptersKeepsContentBeforeFirstRecognizedTitle(t *testing.T) {
	text := "楔子里先交代世界观。\n这里还不是章节标题。\n\n第四部 风暴之前\n真正被识别到的正文。\n"
	chapters := DetectChapters(text)
	if len(chapters) != 2 {
		t.Fatalf("章节数 = %d, want 2", len(chapters))
	}
	if chapters[0].Title != unknownChapterTitle {
		t.Fatalf("第一段未识别正文标题 = %s, want %s", chapters[0].Title, unknownChapterTitle)
	}
	if chapters[0].StartOffset != 0 || chapters[0].EndOffset != len("楔子里先交代世界观。\n这里还不是章节标题。\n\n") {
		t.Fatalf("未知标题章节范围不正确: %+v", chapters[0])
	}
	if chapters[1].Title != "第四部 风暴之前" || chapters[1].Order != 2 {
		t.Fatalf("识别标题章节不正确: %+v", chapters[1])
	}
}

// TestConfirmNovelImportWritesUnknownTitleChapterContent 确保未知标题章节写入完整正文，不剥掉首行。
func TestConfirmNovelImportWritesUnknownTitleChapterContent(t *testing.T) {
	libraryPath := t.TempDir()
	sourcePath := filepath.Join(t.TempDir(), "前言小说.txt")
	sourceText := "楔子里先交代世界观。\n这里还不是章节标题。\n\n第四部 风暴之前\n真正被识别到的正文。\n"
	if err := os.WriteFile(sourcePath, []byte(sourceText), 0o644); err != nil {
		t.Fatalf("写入测试小说失败: %v", err)
	}

	result, err := ConfirmNovelImport(libraryPath, ConfirmImportRequest{
		BookID:     "book-prefix-001",
		SourcePath: sourcePath,
		BookName:   "前言小说",
	})
	if err != nil {
		t.Fatalf("确认导入失败: %v", err)
	}
	if len(result.Project.Chapters) != 2 {
		t.Fatalf("章节数 = %d, want 2", len(result.Project.Chapters))
	}
	if result.Project.Chapters[0].Title != unknownChapterTitle {
		t.Fatalf("第一章标题 = %s, want %s", result.Project.Chapters[0].Title, unknownChapterTitle)
	}

	bookPath := result.Project.Book.BookPath
	unknownContent, err := os.ReadFile(filepath.Join(bookPath, "chapters", "chapter-000001.txt"))
	if err != nil {
		t.Fatalf("读取未知标题章节失败: %v", err)
	}
	if string(unknownContent) != "楔子里先交代世界观。\n这里还不是章节标题。\n\n" {
		t.Fatalf("未知标题章节正文 = %q", string(unknownContent))
	}
	recognizedContent, err := os.ReadFile(filepath.Join(bookPath, "chapters", "chapter-000002.txt"))
	if err != nil {
		t.Fatalf("读取识别标题章节失败: %v", err)
	}
	if string(recognizedContent) != "真正被识别到的正文。\n" {
		t.Fatalf("识别标题章节正文 = %q", string(recognizedContent))
	}
}

// TestOpenBookProjectRecoversLeadingUnknownChapter 确保旧项目可从 original.txt 修复第一个标题前的正文。
func TestOpenBookProjectRecoversLeadingUnknownChapter(t *testing.T) {
	bookPath := filepath.Join(t.TempDir(), "book-old-prefix")
	if err := os.MkdirAll(filepath.Join(bookPath, "chapters"), 0o755); err != nil {
		t.Fatalf("创建测试小说目录失败: %v", err)
	}
	prefixText := "楔子里先交代世界观。\n这里还不是章节标题。\n\n"
	sourceText := prefixText + "第四部 风暴之前\n真正被识别到的正文。\n"
	if err := os.WriteFile(filepath.Join(bookPath, originalTextName), []byte(sourceText), 0o644); err != nil {
		t.Fatalf("写入原始文本失败: %v", err)
	}
	if err := os.WriteFile(filepath.Join(bookPath, "chapters", "chapter-000001.txt"), []byte("真正被识别到的正文。\n"), 0o644); err != nil {
		t.Fatalf("写入旧章节正文失败: %v", err)
	}
	project := BookProject{
		Version: 1,
		Book: BookMeta{
			ID:             "book-old-prefix",
			Name:           "旧项目",
			SourceTextPath: originalTextName,
			BookPath:       bookPath,
			CreatedAt:      "2026-08-21T10:00:00+08:00",
			UpdatedAt:      "2026-08-21T10:00:00+08:00",
		},
		Chapters: []BookChapter{
			{
				ID:          "chapter-000001",
				Title:       "第四部 风暴之前",
				Order:       1,
				ContentPath: "chapters/chapter-000001.txt",
				StartOffset: len(prefixText),
				EndOffset:   len(sourceText),
				WordCount:   countTextRunes("真正被识别到的正文。\n"),
			},
		},
	}
	if err := writeJSONFileAtomic(filepath.Join(bookPath, bookJSONName), project); err != nil {
		t.Fatalf("写入旧 book.json 失败: %v", err)
	}

	recovered, err := OpenBookProject(bookPath)
	if err != nil {
		t.Fatalf("打开旧项目失败: %v", err)
	}
	if len(recovered.Chapters) != 2 {
		t.Fatalf("修复后章节数 = %d, want 2", len(recovered.Chapters))
	}
	if recovered.Chapters[0].Title != unknownChapterTitle || recovered.Chapters[0].ID != "chapter-000000" {
		t.Fatalf("修复出的未知标题章节不正确: %+v", recovered.Chapters[0])
	}
	if recovered.Chapters[1].ID != "chapter-000001" || recovered.Chapters[1].Order != 2 {
		t.Fatalf("旧章节顺序未正确后移: %+v", recovered.Chapters[1])
	}
	recoveredContent, err := os.ReadFile(filepath.Join(bookPath, "chapters", "chapter-000000.txt"))
	if err != nil {
		t.Fatalf("读取修复章节正文失败: %v", err)
	}
	if string(recoveredContent) != prefixText {
		t.Fatalf("修复章节正文 = %q, want %q", string(recoveredContent), prefixText)
	}

	reopened, err := OpenBookProject(bookPath)
	if err != nil {
		t.Fatalf("再次打开旧项目失败: %v", err)
	}
	if len(reopened.Chapters) != 2 {
		t.Fatalf("再次打开后章节数 = %d, want 2", len(reopened.Chapters))
	}
}

// TestReadingWorkbenchPersistsStateAndDeconstruction 覆盖阅读位置、章节编辑和选文拆解的本地闭环。
func TestReadingWorkbenchPersistsStateAndDeconstruction(t *testing.T) {
	libraryPath := t.TempDir()
	sourcePath := filepath.Join(t.TempDir(), "雾港来信.txt")
	sourceText := "第一章 远处的灯\n雨在码头上铺开。\n\n第二章 借宿人\n她把信收进外套。\n"
	if err := os.WriteFile(sourcePath, []byte(sourceText), 0o644); err != nil {
		t.Fatalf("写入测试小说失败: %v", err)
	}

	result, err := ConfirmNovelImport(libraryPath, ConfirmImportRequest{
		BookID:     "book-reading-001",
		SourcePath: sourcePath,
		BookName:   "雾港来信",
	})
	if err != nil {
		t.Fatalf("确认导入失败: %v", err)
	}
	bookPath := result.Project.Book.BookPath

	session, err := OpenReadingSession(bookPath)
	if err != nil {
		t.Fatalf("打开阅读会话失败: %v", err)
	}
	if session.CurrentChapter.ID != "chapter-000001" {
		t.Fatalf("默认章节 = %s, want chapter-000001", session.CurrentChapter.ID)
	}
	if session.State.Preferences.FontSize != defaultReaderFontSize {
		t.Fatalf("默认阅读字号 = %d, want %d", session.State.Preferences.FontSize, defaultReaderFontSize)
	}
	for _, requiredPath := range []string{
		filepath.Join(bookPath, "reading", "reading-state.json"),
		filepath.Join(bookPath, "source-texts", "source-texts.json"),
		filepath.Join(bookPath, "deconstructions", "deconstruction-cards.json"),
		filepath.Join(bookPath, "deconstruction-plan.json"),
	} {
		if _, err := os.Stat(requiredPath); err != nil {
			t.Fatalf("缺少阅读数据文件 %s: %v", requiredPath, err)
		}
	}

	state, err := SaveReadingPosition(SaveReadingPositionRequest{
		BookPath:               bookPath,
		ChapterID:              "chapter-000002",
		CharOffset:             2,
		ParagraphIndex:         0,
		RenderedPageNumber:     1,
		LastMode:               "read",
		ReadingProgressPercent: 50,
	})
	if err != nil {
		t.Fatalf("保存阅读位置失败: %v", err)
	}
	if state.LastPosition.ChapterID != "chapter-000002" {
		t.Fatalf("阅读位置章节 = %s, want chapter-000002", state.LastPosition.ChapterID)
	}
	if state.ReadingProgressPercent != 50 {
		t.Fatalf("阅读进度 = %d, want 50", state.ReadingProgressPercent)
	}
	preferenceState, err := SaveReadingPreferences(SaveReadingPreferencesRequest{
		BookPath: bookPath,
		FontSize: 22,
	})
	if err != nil {
		t.Fatalf("保存阅读偏好失败: %v", err)
	}
	if preferenceState.Preferences.FontSize != 22 {
		t.Fatalf("阅读字号 = %d, want 22", preferenceState.Preferences.FontSize)
	}
	progressIndex := enrichLibraryReadingProgress(LibraryIndex{
		Version: 1,
		Books: []LibraryBookEntry{
			{BookName: "雾港来信", BookPath: bookPath, BookSize: result.Project.Book.BookSize},
		},
	})
	if progressIndex.Books[0].ReadingProgressPercent != 50 {
		t.Fatalf("书库阅读进度 = %d, want 50", progressIndex.Books[0].ReadingProgressPercent)
	}
	restoredSession, err := OpenReadingSession(bookPath)
	if err != nil {
		t.Fatalf("重新打开阅读会话失败: %v", err)
	}
	if restoredSession.CurrentChapter.ID != "chapter-000002" {
		t.Fatalf("恢复阅读章节 = %s, want chapter-000002", restoredSession.CurrentChapter.ID)
	}
	if restoredSession.State.Preferences.FontSize != 22 {
		t.Fatalf("恢复阅读字号 = %d, want 22", restoredSession.State.Preferences.FontSize)
	}

	savedChapter, err := SaveChapterContent(SaveChapterContentRequest{
		BookPath:  bookPath,
		ChapterID: "chapter-000002",
		Title:     "第二章 新标题",
		Content:   "她把信收进外套，又看了一眼窗外。",
	})
	if err != nil {
		t.Fatalf("保存章节失败: %v", err)
	}
	if savedChapter.Chapter.Title != "第二章 新标题" {
		t.Fatalf("章节标题 = %s, want 第二章 新标题", savedChapter.Chapter.Title)
	}
	reopened, err := OpenBookProject(bookPath)
	if err != nil {
		t.Fatalf("重新打开项目失败: %v", err)
	}
	if reopened.Chapters[1].Title != "第二章 新标题" {
		t.Fatalf("book.json 章节标题未更新: %s", reopened.Chapters[1].Title)
	}

	created, err := CreateDeconstructionCard(CreateDeconstructionRequest{
		BookPath: bookPath,
		Selection: SourceSelection{
			ChapterID:    "chapter-000002",
			ChapterOrder: 2,
			StartOffset:  0,
			EndOffset:    6,
			SelectedText: "她把信收进外套",
		},
		PlanTypeID: "scene",
		FieldValues: map[string]string{
			"scenePurpose": "制造迟疑感",
		},
		TagIDs: []string{"氛围", "重逢"},
	})
	if err != nil {
		t.Fatalf("创建拆解失败: %v", err)
	}
	if created.Card.SourceTextID == "" || created.SourceText.ID == "" {
		t.Fatalf("拆解未保留来源引用: %+v / %+v", created.Card, created.SourceText)
	}
	if len(created.SourceTexts.Items) != 1 || len(created.DeconstructionCards.Items) != 1 {
		t.Fatalf("选文或拆解数量不正确: %+v / %+v", created.SourceTexts.Items, created.DeconstructionCards.Items)
	}
}

// TestSplitChapterFromStartSupportsRepeatedUnknownChapterSplit 覆盖未知章节按顺序拆成多个章节的目录修正流程。
func TestSplitChapterFromStartSupportsRepeatedUnknownChapterSplit(t *testing.T) {
	libraryPath := t.TempDir()
	sourcePath := filepath.Join(t.TempDir(), "未知章节小说.txt")
	sourceText := "A 段正文。\n\nB 段正文。\n\nC 段正文。"
	if err := os.WriteFile(sourcePath, []byte(sourceText), 0o644); err != nil {
		t.Fatalf("写入测试小说失败: %v", err)
	}

	result, err := ConfirmNovelImport(libraryPath, ConfirmImportRequest{
		BookID:     "book-split-001",
		SourcePath: sourcePath,
		BookName:   "未知章节小说",
	})
	if err != nil {
		t.Fatalf("确认导入失败: %v", err)
	}
	if len(result.Project.Chapters) != 1 {
		t.Fatalf("初始章节数 = %d, want 1", len(result.Project.Chapters))
	}
	bookPath := result.Project.Book.BookPath
	unknownChapterID := result.Project.Chapters[0].ID

	firstSplit, err := SplitChapterFromStart(SplitChapterRequest{
		BookPath:         bookPath,
		ChapterID:        unknownChapterID,
		NewTitle:         "第一章 A",
		NewContent:       "A 段正文。\n\n",
		RemainingContent: "B 段正文。\n\nC 段正文。",
	})
	if err != nil {
		t.Fatalf("第一次拆分失败: %v", err)
	}
	if len(firstSplit.Project.Chapters) != 2 {
		t.Fatalf("第一次拆分后章节数 = %d, want 2", len(firstSplit.Project.Chapters))
	}
	if firstSplit.Project.Chapters[0].Title != "第一章 A" || firstSplit.Project.Chapters[0].Order != 1 {
		t.Fatalf("新章节 A 不正确: %+v", firstSplit.Project.Chapters[0])
	}
	if firstSplit.Project.Chapters[1].ID != unknownChapterID || firstSplit.Project.Chapters[1].Order != 2 {
		t.Fatalf("剩余未知章节不正确: %+v", firstSplit.Project.Chapters[1])
	}

	secondSplit, err := SplitChapterFromStart(SplitChapterRequest{
		BookPath:         bookPath,
		ChapterID:        unknownChapterID,
		NewTitle:         "第二章 B",
		NewContent:       "B 段正文。\n\n",
		RemainingContent: "C 段正文。",
	})
	if err != nil {
		t.Fatalf("第二次拆分失败: %v", err)
	}
	if len(secondSplit.Project.Chapters) != 3 {
		t.Fatalf("第二次拆分后章节数 = %d, want 3", len(secondSplit.Project.Chapters))
	}
	if secondSplit.Project.Chapters[0].Title != "第一章 A" || secondSplit.Project.Chapters[0].Order != 1 {
		t.Fatalf("章节 A 顺序不正确: %+v", secondSplit.Project.Chapters[0])
	}
	if secondSplit.Project.Chapters[1].Title != "第二章 B" || secondSplit.Project.Chapters[1].Order != 2 {
		t.Fatalf("章节 B 顺序不正确: %+v", secondSplit.Project.Chapters[1])
	}
	if secondSplit.Project.Chapters[2].ID != unknownChapterID || secondSplit.Project.Chapters[2].Order != 3 {
		t.Fatalf("剩余章节 C 顺序不正确: %+v", secondSplit.Project.Chapters[2])
	}

	reopened, err := OpenBookProject(bookPath)
	if err != nil {
		t.Fatalf("重新打开拆分后项目失败: %v", err)
	}
	if reopened.Chapters[0].Title != "第一章 A" || reopened.Chapters[1].Title != "第二章 B" || reopened.Chapters[2].ID != unknownChapterID {
		t.Fatalf("落盘章节顺序不正确: %+v", reopened.Chapters)
	}
	remainingContent, err := os.ReadFile(filepath.Join(bookPath, filepath.FromSlash(reopened.Chapters[2].ContentPath)))
	if err != nil {
		t.Fatalf("读取剩余章节正文失败: %v", err)
	}
	if string(remainingContent) != "C 段正文。" {
		t.Fatalf("剩余章节正文 = %q, want C 段正文。", string(remainingContent))
	}
}

// TestSplitChapterFromStartRejectsChapterWithSourceTexts 确保章节已有选文来源时不会静默破坏拆解追溯。
func TestSplitChapterFromStartRejectsChapterWithSourceTexts(t *testing.T) {
	libraryPath := t.TempDir()
	sourcePath := filepath.Join(t.TempDir(), "已有拆解小说.txt")
	sourceText := "第一章 雨夜\nA 段正文。\n\nB 段正文。"
	if err := os.WriteFile(sourcePath, []byte(sourceText), 0o644); err != nil {
		t.Fatalf("写入测试小说失败: %v", err)
	}

	result, err := ConfirmNovelImport(libraryPath, ConfirmImportRequest{
		BookID:     "book-split-source-001",
		SourcePath: sourcePath,
		BookName:   "已有拆解小说",
	})
	if err != nil {
		t.Fatalf("确认导入失败: %v", err)
	}
	bookPath := result.Project.Book.BookPath
	chapterID := result.Project.Chapters[0].ID

	if _, err := OpenReadingSession(bookPath); err != nil {
		t.Fatalf("打开阅读会话失败: %v", err)
	}
	if _, err := CreateDeconstructionCard(CreateDeconstructionRequest{
		BookPath: bookPath,
		Selection: SourceSelection{
			ChapterID:    chapterID,
			ChapterOrder: 1,
			StartOffset:  0,
			EndOffset:    len("A 段正文"),
			SelectedText: "A 段正文",
		},
		PlanTypeID: "scene",
		FieldValues: map[string]string{
			"scenePurpose": "测试来源保护",
		},
	}); err != nil {
		t.Fatalf("创建拆解失败: %v", err)
	}

	_, err = SplitChapterFromStart(SplitChapterRequest{
		BookPath:         bookPath,
		ChapterID:        chapterID,
		NewTitle:         "第一章 A",
		NewContent:       "A 段正文。\n\n",
		RemainingContent: "B 段正文。",
	})
	if err == nil {
		t.Fatal("已有选文来源的章节拆分成功，want error")
	}
	if !strings.Contains(err.Error(), "已有选文来源") {
		t.Fatalf("错误信息 = %v, want 包含 已有选文来源", err)
	}
}
