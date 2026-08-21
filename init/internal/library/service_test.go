// service_test.go 验证书库导入服务会生成可迁移的小说项目目录。
package library

import (
	"os"
	"path/filepath"
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
}

// TestDetectChaptersFallsBackToSingleChapter 确保无章节标题的 TXT 也能形成可打开项目。
func TestDetectChaptersFallsBackToSingleChapter(t *testing.T) {
	chapters := DetectChapters("没有章节标题的正文")
	if len(chapters) != 1 {
		t.Fatalf("章节数 = %d, want 1", len(chapters))
	}
	if chapters[0].Title != "正文" {
		t.Fatalf("默认章节标题 = %s, want 正文", chapters[0].Title)
	}
}
