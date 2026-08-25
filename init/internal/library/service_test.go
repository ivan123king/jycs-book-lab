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

// TestLibraryMaterialPageKeepsMaterialsAfterBookRemoval 覆盖书库素材聚合、自建素材和移除书籍前的素材迁移。
func TestLibraryMaterialPageKeepsMaterialsAfterBookRemoval(t *testing.T) {
	t.Setenv("AppData", t.TempDir())
	libraryPath := t.TempDir()
	sourcePath := filepath.Join(t.TempDir(), "素材迁移小说.txt")
	sourceText := "第一章 雨夜\n她把信收进外套，又看了一眼窗外。\n"
	if err := os.WriteFile(sourcePath, []byte(sourceText), 0o644); err != nil {
		t.Fatalf("写入测试小说失败: %v", err)
	}
	result, err := ConfirmNovelImport(libraryPath, ConfirmImportRequest{
		BookID:     "book-material-migrate-001",
		SourcePath: sourcePath,
		BookName:   "素材迁移小说",
	})
	if err != nil {
		t.Fatalf("确认导入失败: %v", err)
	}
	bookPath := result.Project.Book.BookPath

	if _, err := OpenReadingSession(bookPath); err != nil {
		t.Fatalf("打开阅读会话失败: %v", err)
	}
	createdCard, err := CreateDeconstructionCard(CreateDeconstructionRequest{
		BookPath: bookPath,
		Selection: SourceSelection{
			ChapterID:    "chapter-000001",
			ChapterOrder: 1,
			StartOffset:  0,
			EndOffset:    len("她把信收进外套"),
			SelectedText: "她把信收进外套",
		},
		PlanTypeID:  "technique",
		DimensionID: "environment",
		Content:     "动作先于情绪说明，可以压住人物表达。",
		TagIDs:      []string{"动作"},
	})
	if err != nil {
		t.Fatalf("创建拆解失败: %v", err)
	}
	bookMaterial, err := CreateMaterialFromCard(CreateMaterialFromCardRequest{
		BookPath: bookPath,
		CardID:   createdCard.Card.ID,
		Title:    "动作压住情绪",
		Summary:  "她把信收进外套",
		Content:  "动作先于情绪说明，可以压住人物表达。",
		TagIDs:   []string{"动作", "克制"},
	})
	if err != nil {
		t.Fatalf("创建本书素材失败: %v", err)
	}
	if _, err := CreateLibraryMaterial(CreateLibraryMaterialRequest{
		LibraryPath:   libraryPath,
		Title:         "书库自建素材",
		Summary:       "不依赖任何单本小说的原文摘录",
		Content:       "这是书库直接沉淀的写作经验。",
		MaterialType:  analysisKindWritingTechnique,
		PlanTypeName:  "手动整理",
		DimensionName: "经验",
		TagIDs:        []string{"自建"},
	}); err != nil {
		t.Fatalf("创建书库自建素材失败: %v", err)
	}

	page, err := OpenLibraryMaterialPage(LibraryMaterialPageRequest{LibraryPath: libraryPath, Page: 1, PageSize: 50})
	if err != nil {
		t.Fatalf("打开书库素材页失败: %v", err)
	}
	if page.Total != 2 {
		t.Fatalf("书库素材总数 = %d, want 2: %+v", page.Total, page.Items)
	}
	foundReference := false
	for _, item := range page.Items {
		if item.SourceMode == libraryMaterialSourceReference && item.BookMaterialID == bookMaterial.Material.ID {
			foundReference = true
		}
	}
	if !foundReference {
		t.Fatalf("书库素材页未展示本书素材引用: %+v", page.Items)
	}

	if _, err := RemoveBookFromLibrary(libraryPath, bookPath); err != nil {
		t.Fatalf("移除书籍失败: %v", err)
	}
	page, err = OpenLibraryMaterialPage(LibraryMaterialPageRequest{LibraryPath: libraryPath, Page: 1, PageSize: 50})
	if err != nil {
		t.Fatalf("移除后打开书库素材页失败: %v", err)
	}
	if page.Total != 2 {
		t.Fatalf("移除后书库素材总数 = %d, want 2: %+v", page.Total, page.Items)
	}
	foundMigrated := false
	for _, item := range page.Items {
		if item.SourceMode == libraryMaterialSourceMigrated && item.BookMaterialID == bookMaterial.Material.ID &&
			item.Content == "动作先于情绪说明，可以压住人物表达。" {
			foundMigrated = true
		}
		if item.SourceMode == libraryMaterialSourceReference {
			t.Fatalf("移除后不应再保留实时本书引用: %+v", item)
		}
	}
	if !foundMigrated {
		t.Fatalf("移除后未找到迁移素材: %+v", page.Items)
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
	t.Setenv("AppData", t.TempDir())
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
	if len(session.ChapterMemos.Items) != 0 {
		t.Fatalf("默认章节小记数量 = %d, want 0", len(session.ChapterMemos.Items))
	}
	for _, requiredPath := range []string{
		filepath.Join(bookPath, "reading", "reading-state.json"),
		filepath.Join(bookPath, "reading", "chapter-memos.json"),
		filepath.Join(bookPath, "source-texts", "source-texts.json"),
		filepath.Join(bookPath, "deconstructions", "writing-technique-cards.json"),
		filepath.Join(bookPath, "deconstruction-plan.json"),
		filepath.Join(bookPath, "deconstruction-templates", systemWritingTemplateFileName),
		filepath.Join(bookPath, "deconstruction-templates", systemStoryTemplateFileName),
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
	memoResult, err := SaveChapterMemo(SaveChapterMemoRequest{
		BookPath:  bookPath,
		ChapterID: "chapter-000002",
		Content:   "　　这一章可以记录人物进入新场景后的防备感。",
	})
	if err != nil {
		t.Fatalf("保存章节小记失败: %v", err)
	}
	if len(memoResult.ChapterMemos.Items) != 1 {
		t.Fatalf("章节小记数量 = %d, want 1", len(memoResult.ChapterMemos.Items))
	}
	if memoResult.Memo.ChapterID != "chapter-000002" || memoResult.Memo.ChapterOrder != 2 {
		t.Fatalf("章节小记来源不正确: %+v", memoResult.Memo)
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
	if len(restoredSession.ChapterMemos.Items) != 1 || restoredSession.ChapterMemos.Items[0].Content != "　　这一章可以记录人物进入新场景后的防备感。" {
		t.Fatalf("恢复章节小记不正确: %+v", restoredSession.ChapterMemos.Items)
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
		PlanTypeID:  "technique",
		DimensionID: "environment",
		Content:     "制造迟疑感",
		TagIDs:      []string{"氛围", "重逢"},
	})
	if err != nil {
		t.Fatalf("创建拆解失败: %v", err)
	}
	if created.Card.SourceTextID == "" || created.SourceText.ID == "" {
		t.Fatalf("拆解未保留来源引用: %+v / %+v", created.Card, created.SourceText)
	}
	if created.Card.PlanTypeName != "表现手法" || created.Card.DimensionName != "环境描写" || created.Card.Content != "制造迟疑感" {
		t.Fatalf("拆解卡片未保存类型、维度和内容快照: %+v", created.Card)
	}
	if created.Card.AnalysisKind != analysisKindWritingTechnique {
		t.Fatalf("拆解卡片类型 = %s, want %s", created.Card.AnalysisKind, analysisKindWritingTechnique)
	}
	if len(created.SourceTexts.Items) != 1 || len(created.DeconstructionCards.Items) != 1 {
		t.Fatalf("选文或拆解数量不正确: %+v / %+v", created.SourceTexts.Items, created.DeconstructionCards.Items)
	}
	workspace, err := OpenDeconstructionCardsWorkspace(bookPath)
	if err != nil {
		t.Fatalf("打开写作手法拆解工作区失败: %v", err)
	}
	if len(workspace.DeconstructionCards.Items) != 1 || len(workspace.SourceTexts.Items) != 1 {
		t.Fatalf("写作手法拆解工作区数据不正确: %+v / %+v", workspace.DeconstructionCards.Items, workspace.SourceTexts.Items)
	}
	updatedCard, err := UpdateDeconstructionCard(UpdateDeconstructionCardRequest{
		BookPath:    bookPath,
		CardID:      created.Card.ID,
		PlanTypeID:  "emotion",
		DimensionID: "emotion-change",
		Content:     "制造迟疑感，并用动作压住情绪",
		TagIDs:      []string{"氛围", "氛围", "动作"},
	})
	if err != nil {
		t.Fatalf("更新写作手法拆解失败: %v", err)
	}
	if updatedCard.Card.PlanTypeName != "人物感情" || updatedCard.Card.DimensionName != "情绪变化" ||
		updatedCard.Card.Content != "制造迟疑感，并用动作压住情绪" || len(updatedCard.Card.TagIDs) != 2 {
		t.Fatalf("写作手法拆解更新结果不正确: %+v", updatedCard.Card)
	}
	draft, err := BuildMaterialDraftFromCard(BuildMaterialDraftFromCardRequest{
		BookPath: bookPath,
		CardID:   created.Card.ID,
	})
	if err != nil {
		t.Fatalf("生成素材草稿失败: %v", err)
	}
	if draft.MaterialType != analysisKindWritingTechnique || len(draft.SourceRefs) != 1 ||
		draft.Summary != "她把信收进外套" || draft.Content != "制造迟疑感，并用动作压住情绪" ||
		draft.SourceRefs[0].Snapshot.PlanTypeName != "人物感情" || draft.SourceRefs[0].Snapshot.DimensionName != "情绪变化" {
		t.Fatalf("素材草稿未按原文和拆解说明生成: %+v", draft)
	}
	materialResult, err := CreateMaterialFromCard(CreateMaterialFromCardRequest{
		BookPath: bookPath,
		CardID:   created.Card.ID,
		Title:    "动作压住情绪的迟疑写法",
		Summary:  "    她把信收进外套",
		Content:  "    动作先于情绪说明，可以压住人物的直接表达。",
		TagIDs:   []string{"动作", "防备"},
	})
	if err != nil {
		t.Fatalf("创建素材失败: %v", err)
	}
	if len(materialResult.Materials.Items) != 1 || materialResult.Material.ContentPath == "" {
		t.Fatalf("素材落盘元数据不正确: %+v", materialResult)
	}
	if materialResult.Material.Summary != "    她把信收进外套" || len(materialResult.Material.CategoryIDs) != 0 {
		t.Fatalf("素材原文或分类不正确: %+v", materialResult.Material)
	}
	materialTextPath := filepath.Join(bookPath, filepath.FromSlash(materialResult.Material.ContentPath))
	materialContent, err := os.ReadFile(materialTextPath)
	if err != nil {
		t.Fatalf("读取素材正文失败: %v", err)
	}
	if string(materialContent) != "    动作先于情绪说明，可以压住人物的直接表达。" {
		t.Fatalf("素材正文 TXT = %q", string(materialContent))
	}
	if len(materialResult.DeconstructionCards.Items) != 1 || len(materialResult.DeconstructionCards.Items[0].MaterialIDs) != 1 {
		t.Fatalf("拆解卡片未同步素材引用: %+v", materialResult.DeconstructionCards.Items)
	}
	materialWorkspace, err := OpenMaterialWorkspace(bookPath)
	if err != nil {
		t.Fatalf("打开素材库失败: %v", err)
	}
	if materialWorkspace.MaterialContents[materialResult.Material.ID] == "" || len(materialWorkspace.SourceTexts.Items) != 1 {
		t.Fatalf("素材工作区缺少正文或来源: %+v", materialWorkspace)
	}
	updatedMaterial, err := UpdateMaterial(UpdateMaterialRequest{
		BookPath:   bookPath,
		MaterialID: materialResult.Material.ID,
		Title:      "动作压住情绪",
		Summary:    "    她把信收进外套，又看了一眼窗外。",
		Content:    "    用动作压住情绪，可以让人物显得更克制。",
		TagIDs:     []string{"动作", "动作"},
	})
	if err != nil {
		t.Fatalf("更新素材失败: %v", err)
	}
	if updatedMaterial.Material.Summary != "    她把信收进外套，又看了一眼窗外。" || len(updatedMaterial.Material.TagIDs) != 1 {
		t.Fatalf("素材更新结果不正确: %+v", updatedMaterial.Material)
	}
	deletedMaterial, err := DeleteMaterial(DeleteMaterialRequest{
		BookPath:   bookPath,
		MaterialID: materialResult.Material.ID,
	})
	if err != nil {
		t.Fatalf("删除素材失败: %v", err)
	}
	if len(deletedMaterial.Materials.Items) != 0 || len(deletedMaterial.DeconstructionCards.Items[0].MaterialIDs) != 0 {
		t.Fatalf("删除素材后引用未清理: %+v", deletedMaterial)
	}
	if _, err := os.Stat(materialTextPath); !os.IsNotExist(err) {
		t.Fatalf("素材正文 TXT 未删除，stat err = %v", err)
	}
	deletedCard, err := DeleteDeconstructionCard(DeleteDeconstructionCardRequest{
		BookPath: bookPath,
		CardID:   created.Card.ID,
	})
	if err != nil {
		t.Fatalf("删除写作手法拆解失败: %v", err)
	}
	if len(deletedCard.DeconstructionCards.Items) != 0 || len(deletedCard.SourceTexts.Items) != 0 {
		t.Fatalf("删除后卡片或无人引用选文未清理: %+v / %+v", deletedCard.DeconstructionCards.Items, deletedCard.SourceTexts.Items)
	}

	clearedMemo, err := SaveChapterMemo(SaveChapterMemoRequest{
		BookPath:  bookPath,
		ChapterID: "chapter-000002",
		Content:   "",
	})
	if err != nil {
		t.Fatalf("清空章节小记失败: %v", err)
	}
	if len(clearedMemo.ChapterMemos.Items) != 0 {
		t.Fatalf("清空后章节小记数量 = %d, want 0", len(clearedMemo.ChapterMemos.Items))
	}
}

// TestDeconstructionPlanTemplatesPersistPerBook 覆盖当前书模板目录、另存模板和设为当前模板流程。
func TestDeconstructionPlanTemplatesPersistPerBook(t *testing.T) {
	libraryPath := t.TempDir()
	sourcePath := filepath.Join(t.TempDir(), "模板小说.txt")
	sourceText := "第一章 雨夜\nA 段正文。\n"
	if err := os.WriteFile(sourcePath, []byte(sourceText), 0o644); err != nil {
		t.Fatalf("写入测试小说失败: %v", err)
	}
	result, err := ConfirmNovelImport(libraryPath, ConfirmImportRequest{
		BookID:     "book-template-001",
		SourcePath: sourcePath,
		BookName:   "模板小说",
	})
	if err != nil {
		t.Fatalf("确认导入失败: %v", err)
	}
	bookPath := result.Project.Book.BookPath

	workspace, err := OpenDeconstructionPlanWorkspace(bookPath)
	if err != nil {
		t.Fatalf("打开拆解方案工作区失败: %v", err)
	}
	if workspace.CurrentPlan.TemplateName != systemWritingTemplateName {
		t.Fatalf("当前模板名 = %s, want %s", workspace.CurrentPlan.TemplateName, systemWritingTemplateName)
	}
	if len(workspace.Templates) != 2 {
		t.Fatalf("模板数量 = %d, want 2: %+v", len(workspace.Templates), workspace.Templates)
	}
	if workspace.Templates[0].FileName != systemWritingTemplateFileName || workspace.Templates[1].FileName != systemStoryTemplateFileName {
		t.Fatalf("模板列表不正确: %+v", workspace.Templates)
	}
	if workspace.CurrentTemplateFileName != systemWritingTemplateFileName {
		t.Fatalf("当前模板文件 = %s, want %s", workspace.CurrentTemplateFileName, systemWritingTemplateFileName)
	}
	if workspace.CurrentTemplateFileNames[analysisKindStoryPlot] != systemStoryTemplateFileName {
		t.Fatalf("当前剧情模板文件 = %s, want %s", workspace.CurrentTemplateFileNames[analysisKindStoryPlot], systemStoryTemplateFileName)
	}

	customPlan := workspace.CurrentPlan
	customPlan.TemplateName = "感情线模板"
	customPlan.Types = []DeconstructionType{
		{
			ID:    "emotion-line",
			Name:  "人物感情",
			Order: 1,
			Dimensions: []DeconstructionDimension{
				{ID: "hesitation", Name: "迟疑", Order: 1},
				{ID: "defense", Name: "防御", Order: 2},
			},
		},
	}
	workspace, err = SaveDeconstructionTemplate(SaveDeconstructionTemplateRequest{
		BookPath:     bookPath,
		AnalysisKind: analysisKindWritingTechnique,
		TemplateName: "感情线模板",
		Plan:         customPlan,
	})
	if err != nil {
		t.Fatalf("另存模板失败: %v", err)
	}
	if len(workspace.Templates) != 3 {
		t.Fatalf("另存后模板数量 = %d, want 3", len(workspace.Templates))
	}

	workspace, err = SetCurrentDeconstructionTemplate(SetCurrentDeconstructionTemplateRequest{
		BookPath:     bookPath,
		AnalysisKind: analysisKindWritingTechnique,
		FileName:     "感情线模板.json",
	})
	if err != nil {
		t.Fatalf("设为当前模板失败: %v", err)
	}
	if workspace.CurrentPlan.TemplateName != "感情线模板" || workspace.CurrentPlan.Types[0].ID != "emotion-line" {
		t.Fatalf("当前模板未替换为自定义模板: %+v", workspace.CurrentPlan)
	}
	if workspace.CurrentTemplateFileName != "感情线模板.json" {
		t.Fatalf("当前模板文件 = %s, want 感情线模板.json", workspace.CurrentTemplateFileName)
	}

	customPlan.Types[0].Dimensions[0].Name = "情绪迟疑"
	workspace, err = SaveDeconstructionTemplate(SaveDeconstructionTemplateRequest{
		BookPath:     bookPath,
		AnalysisKind: analysisKindWritingTechnique,
		TemplateName: "感情线模板",
		FileName:     "感情线模板.json",
		Plan:         customPlan,
		Overwrite:    true,
	})
	if err != nil {
		t.Fatalf("保存当前模板失败: %v", err)
	}
	if workspace.CurrentPlan.Types[0].Dimensions[0].Name != "情绪迟疑" {
		t.Fatalf("保存当前模板后未同步当前方案: %+v", workspace.CurrentPlan.Types[0].Dimensions[0])
	}
	planStatePath := filepath.Join(bookPath, "deconstruction-plan.json")
	var planState DeconstructionPlanState
	if err := readJSONFile(planStatePath, &planState); err != nil {
		t.Fatalf("读取当前方案引用失败: %v", err)
	}
	if len(planState.CurrentTemplates) == 0 || len(planState.CurrentTemplates[analysisKindWritingTechnique]) == 0 {
		t.Fatalf("当前方案不应再保存模板内容: %+v", planState)
	}
}

// TestStoryPlotCardsPersistChapterRangesAndSnapshots 覆盖剧情拆解从模板加载、草稿、正式保存到删除的本地闭环。
func TestStoryPlotCardsPersistChapterRangesAndSnapshots(t *testing.T) {
	libraryPath := t.TempDir()
	sourcePath := filepath.Join(t.TempDir(), "剧情线小说.txt")
	sourceText := "第一章 开端\nA 段正文。\n\n第二章 误会\nB 段正文。\n\n第三章 爆发\nC 段正文。\n"
	if err := os.WriteFile(sourcePath, []byte(sourceText), 0o644); err != nil {
		t.Fatalf("写入测试小说失败: %v", err)
	}
	result, err := ConfirmNovelImport(libraryPath, ConfirmImportRequest{
		BookID:     "book-story-plot-001",
		SourcePath: sourcePath,
		BookName:   "剧情线小说",
	})
	if err != nil {
		t.Fatalf("确认导入失败: %v", err)
	}
	bookPath := result.Project.Book.BookPath

	workspace, err := OpenStoryPlotWorkspace(bookPath)
	if err != nil {
		t.Fatalf("打开剧情工作区失败: %v", err)
	}
	if workspace.Plan.AnalysisKind != analysisKindStoryPlot || workspace.CurrentTemplateFileName != systemStoryTemplateFileName {
		t.Fatalf("剧情模板加载不正确: %+v", workspace)
	}
	if len(workspace.StoryPlotCards.Items) != 0 {
		t.Fatalf("初始剧情卡片数量 = %d, want 0", len(workspace.StoryPlotCards.Items))
	}
	if _, err := os.Stat(filepath.Join(bookPath, filepath.FromSlash(storyPlotCardsRelativePath))); err != nil {
		t.Fatalf("缺少剧情卡片数据文件: %v", err)
	}

	draftResult, err := CreateStoryPlotCard(SaveStoryPlotCardRequest{
		BookPath:          bookPath,
		Status:            storyPlotStatusDraft,
		ChapterRangeInput: "1-9",
		TemplateFileName:  workspace.CurrentTemplateFileName,
		TemplateName:      workspace.Plan.TemplateName,
		PlanTypeID:        "plot",
		PlanTypeName:      "剧情分析",
		TypeSnapshots: []StoryPlotTypeSnapshot{
			{
				ID:    "plot",
				Name:  "剧情分析",
				Order: 1,
				Dimensions: []StoryPlotDimension{
					{ID: "core-event", Name: "核心事件", Order: 1},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("保存剧情草稿失败: %v", err)
	}
	if draftResult.Card.Status != storyPlotStatusDraft || len(draftResult.Card.ParseErrors) == 0 {
		t.Fatalf("草稿应允许无效章节范围并保存解析错误: %+v", draftResult.Card)
	}

	updatedResult, err := UpdateStoryPlotCard(SaveStoryPlotCardRequest{
		BookPath:          bookPath,
		CardID:            draftResult.Card.ID,
		Status:            storyPlotStatusCompleted,
		Title:             "误会到爆发",
		Summary:           "误会在前两章铺开，第三章完成情绪爆发。",
		ChapterRangeInput: "1-2, 2, 3",
		TemplateFileName:  draftResult.Card.TemplateFileName,
		TemplateName:      draftResult.Card.TemplateName,
		PlanTypeID:        "custom-plot",
		PlanTypeName:      "误会反转",
		Dimensions: []StoryPlotDimension{
			{ID: "relationship", Name: "人物关系", Order: 1, Content: "关系由误判转向正面冲突。"},
			{ID: "turn", Name: "转折", Order: 2, Content: "第三章用行动揭开误会。"},
			{ID: "empty", Name: "空维度", Order: 3, Content: ""},
		},
		TypeSnapshots: []StoryPlotTypeSnapshot{
			{
				ID:    "custom-plot",
				Name:  "误会反转",
				Order: 1,
				Dimensions: []StoryPlotDimension{
					{ID: "relationship", Name: "人物关系", Order: 1, Content: "关系由误判转向正面冲突。"},
					{ID: "turn", Name: "转折", Order: 2, Content: "第三章用行动揭开误会。"},
					{ID: "empty", Name: "空维度", Order: 3, Content: ""},
				},
			},
		},
		TagIDs: []string{"误会", "误会", "情绪爆发"},
	})
	if err != nil {
		t.Fatalf("正式保存剧情卡片失败: %v", err)
	}
	if updatedResult.Card.Status != storyPlotStatusCompleted || len(updatedResult.Card.ParseErrors) != 0 {
		t.Fatalf("正式剧情卡片状态或解析错误不正确: %+v", updatedResult.Card)
	}
	if len(updatedResult.Card.ChapterRefs) != 3 || updatedResult.Card.ChapterRefs[0].ChapterOrder != 1 || updatedResult.Card.ChapterRefs[2].ChapterOrder != 3 {
		t.Fatalf("章节范围未按序去重解析: %+v", updatedResult.Card.ChapterRefs)
	}
	if updatedResult.Card.PlanTypeName != "误会反转" || updatedResult.Card.Dimensions[0].Name != "人物关系" ||
		updatedResult.Card.Dimensions[0].Content != "关系由误判转向正面冲突。" || len(updatedResult.Card.TagIDs) != 2 {
		t.Fatalf("剧情类型、维度快照或标签未正确保存: %+v", updatedResult.Card)
	}

	reopened, err := OpenStoryPlotWorkspace(bookPath)
	if err != nil {
		t.Fatalf("重新打开剧情工作区失败: %v", err)
	}
	if len(reopened.StoryPlotCards.Items) != 1 || reopened.StoryPlotCards.Items[0].Title != "误会到爆发" {
		t.Fatalf("剧情卡片未持久化: %+v", reopened.StoryPlotCards.Items)
	}
	if reopened.Plan.Types[0].Dimensions[0].ID != "core-event" {
		t.Fatalf("剧情卡片临时维度不应改写模板: %+v", reopened.Plan.Types[0].Dimensions)
	}
	materialDraft, err := BuildMaterialDraftFromCard(BuildMaterialDraftFromCardRequest{
		BookPath: bookPath,
		CardID:   updatedResult.Card.ID,
	})
	if err != nil {
		t.Fatalf("生成剧情素材草稿失败: %v", err)
	}
	if materialDraft.MaterialType != analysisKindStoryPlot || materialDraft.Summary != "误会在前两章铺开，第三章完成情绪爆发。" {
		t.Fatalf("剧情素材草稿基础信息不正确: %+v", materialDraft)
	}
	if strings.Contains(materialDraft.Content, "来源章节") || strings.Contains(materialDraft.Content, "剧情简介") ||
		strings.Contains(materialDraft.Content, "未填写") || strings.Contains(materialDraft.Content, "空维度") {
		t.Fatalf("剧情素材可复用方法不应混入来源信息或空维度: %q", materialDraft.Content)
	}
	if !strings.Contains(materialDraft.Content, "人物关系") || !strings.Contains(materialDraft.Content, "转折") {
		t.Fatalf("剧情素材可复用方法应保留已填写维度: %q", materialDraft.Content)
	}

	_, err = SplitChapterFromStart(SplitChapterRequest{
		BookPath:         bookPath,
		ChapterID:        updatedResult.Card.ChapterRefs[0].ChapterID,
		NewTitle:         "第一章 A",
		NewContent:       "A 段正文。\n",
		RemainingContent: "剩余正文。",
	})
	if err == nil || !strings.Contains(err.Error(), "剧情拆解引用") {
		t.Fatalf("已有剧情引用的章节拆分错误 = %v, want 包含 剧情拆解引用", err)
	}

	deleted, err := DeleteStoryPlotCard(DeleteStoryPlotCardRequest{BookPath: bookPath, CardID: updatedResult.Card.ID})
	if err != nil {
		t.Fatalf("删除剧情卡片失败: %v", err)
	}
	if len(deleted.StoryPlotCards.Items) != 0 {
		t.Fatalf("删除后剧情卡片数量 = %d, want 0", len(deleted.StoryPlotCards.Items))
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
		PlanTypeID:  "technique",
		DimensionID: "environment",
		Content:     "测试来源保护",
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
