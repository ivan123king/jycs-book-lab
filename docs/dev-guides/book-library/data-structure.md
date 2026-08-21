# 书库数据结构

本文定义书库目录、小说目录、书库索引和单本小说 JSON 的第一阶段结构。

## 目录结构

用户选择的书库目录建议如下：

```text
library-root/
├── index.json
└── books/
    └── book-20260821-001/
        ├── book.json
        ├── original.txt
        └── chapters/
            ├── chapter-000001.txt
            └── chapter-000002.txt
```

说明：

1. `library-root` 是用户设置的书库目录。
2. `books/{bookId}` 是每本小说的独立目录。
3. `original.txt` 是导入小说的复制文件。
4. `chapters/` 保存识别并允许用户修正后的章节正文。
5. `book.json` 是该小说后续使用的正式结构化数据。
6. 后续拆解、素材等文件继续放在该小说目录下。

## 书库索引

`index.json` 只用于快速展示书库列表，保持轻量。

```json
{
  "version": 1,
  "books": [
    {
      "bookName": "雾港来信",
      "bookPath": "E:/BookLibrary/books/book-20260821-001",
      "bookSize": 182430
    }
  ]
}
```

字段说明：

| 字段 | 含义 |
| --- | --- |
| `version` | 书库索引结构版本。 |
| `books` | 书库中的小说入口列表。 |
| `bookName` | 小说名，用于书库列表展示。 |
| `bookPath` | 小说目录路径，不是原始导入文件路径。 |
| `bookSize` | 导入副本文件大小，单位字节。 |

## 小说数据

每本小说目录下保存自己的 `book.json`。

```json
{
  "version": 1,
  "book": {
    "id": "book-20260821-001",
    "name": "雾港来信",
    "description": "一部以雾港旧案和人物重逢为线索的长篇小说。",
    "sourceFileName": "雾港来信.txt",
    "sourceTextPath": "original.txt",
    "bookPath": "E:/BookLibrary/books/book-20260821-001",
    "bookSize": 182430,
    "createdAt": "2026-08-21T10:30:00+08:00",
    "updatedAt": "2026-08-21T10:30:00+08:00"
  },
  "chapters": [
    {
      "id": "chapter-000001",
      "title": "第一章 远处的灯",
      "order": 1,
      "contentPath": "chapters/chapter-000001.txt",
      "startOffset": 0,
      "endOffset": 4230,
      "wordCount": 4230
    }
  ]
}
```

字段原则：

1. `sourceTextPath` 指向小说目录内的导入副本。
2. `contentPath` 指向该章节修正后的 TXT 正文。
3. `description` 是用户导入时可填写或修改的小说简介。
4. `chapters` 保存章节识别结果、正文路径和原始位置范围。
5. 导入阶段只展示章节标题分页，不修改章节标题或正文。
6. 用户在阅读工作台编辑 TAB 修改章节标题或正文后，必须更新章节 TXT 和 `book.json` 中对应章节信息。
7. 后续拆解方案、拆解卡片和素材可以继续扩展到该小说目录，不写入书库 `index.json`。
