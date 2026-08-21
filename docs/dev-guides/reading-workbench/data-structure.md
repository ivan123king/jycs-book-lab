# 阅读工作台数据结构

本文定义阅读工作台需要读写的第一阶段数据结构。所有单本小说相关正式数据必须保存在该小说项目目录中；应用系统数据目录只保存全局偏好和最近打开项目入口。

## 目录结构

在书库模块定义的单本小说目录基础上，阅读工作台建议增加以下文件：

```text
books/{bookId}/
├── book.json
├── original.txt
├── chapters/
│   ├── chapter-000001.txt
│   └── chapter-000302.txt
├── reading/
│   └── reading-state.json
├── source-texts/
│   └── source-texts.json
└── deconstructions/
    └── deconstruction-cards.json
```

说明：

1. `chapters/` 保存用户修正后的章节正文 TXT。
2. `reading/reading-state.json` 保存当前书阅读位置。
3. `source-texts/source-texts.json` 保存选文来源片段的结构化位置。
4. `deconstructions/deconstruction-cards.json` 保存拆解卡片结构化数据。
5. 如果某个拆解或素材包含需要直接阅读的长文本，应额外保存为 TXT，并在 JSON 中记录相对路径。

## 阅读状态

`reading/reading-state.json` 是单本书的正式阅读位置来源。

```json
{
  "version": 1,
  "bookId": "book-20260821-001",
  "lastPosition": {
    "chapterId": "chapter-000302",
    "chapterOrder": 302,
    "charOffset": 1842,
    "paragraphIndex": 18,
    "renderedPageNumber": 6,
    "updatedAt": "2026-08-21T10:30:00+08:00"
  },
  "lastMode": "read",
  "readingProgressPercent": 63,
  "preferences": {
    "fontSize": 18
  }
}
```

字段说明：

| 字段 | 含义 |
| --- | --- |
| `version` | 阅读状态结构版本。 |
| `bookId` | 所属小说 ID。 |
| `lastPosition` | 最近阅读位置。 |
| `chapterId` | 最近阅读章节 ID。 |
| `chapterOrder` | 最近阅读章节序号，用于目录分页定位。 |
| `charOffset` | 最近阅读位置在章节正文中的字符偏移。 |
| `paragraphIndex` | 最近阅读段落序号，用于偏移失效时辅助恢复。 |
| `renderedPageNumber` | 最近渲染页码，只作为恢复体验参考。 |
| `updatedAt` | 阅读位置更新时间。 |
| `lastMode` | 最近使用的工作台模式，第一阶段打开时仍建议默认进入 `read`。 |
| `readingProgressPercent` | 当前书整体阅读进度百分比，用于书库列表展示；该值应可由 `book.json` 章节字数和 `lastPosition` 重新估算。 |
| `preferences` | 当前书独立的阅读界面偏好，属于该小说项目目录，不属于全局应用设置。 |
| `fontSize` | 当前书阅读正文字体大小，单位 px，默认 18。 |

恢复原则：

1. 优先使用 `chapterId + charOffset` 恢复阅读位置。
2. 如果正文编辑导致 `charOffset` 失效，使用 `paragraphIndex` 辅助定位。
3. 如果章节不存在，打开最接近的章节；仍失败则打开第一章。
4. 阅读区按整章滚动展示，恢复时根据 `charOffset` 映射到章节滚动位置。
5. `renderedPageNumber` 是旧分页阅读的兼容字段，滚动阅读不依赖它恢复位置。
6. `readingProgressPercent` 是展示摘要，不是恢复阅读位置的唯一依据；旧项目没有该字段时按章节顺序和字符偏移估算。
7. `preferences.fontSize` 随该书项目目录迁移；旧项目没有该字段时使用默认字号并在下次保存时补齐。
8. 阅读位置应在滚动、章节跳转、窗口关闭或应用退出前保存。

## 目录分页

目录分页不需要单独保存正式数据，应由 `book.json` 中的 `chapters` 和当前阅读位置计算。

计算规则：

```text
catalogPage = floor((chapterOrder - 1) / pageSize) + 1
```

示例：

1. 当前章节为第 302 章。
2. 目录每页 50 章。
3. `floor((302 - 1) / 50) + 1 = 7`。
4. 打开目录弹层时默认展示第 7 页。

目录页状态属于页面交互状态，除非后续有明确体验要求，否则不需要写入正式数据文件。

## 原文片段

用户选中文字创建拆解时，应生成或复用原文片段记录。

`source-texts/source-texts.json` 示例：

```json
{
  "version": 1,
  "items": [
    {
      "id": "source-20260821-0001",
      "bookId": "book-20260821-001",
      "chapterId": "chapter-000302",
      "chapterOrder": 302,
      "startOffset": 1842,
      "endOffset": 1896,
      "selectedText": "那声音比三年前低了一点，像被潮气磨过。",
      "createdAt": "2026-08-21T10:32:00+08:00"
    }
  ]
}
```

字段原则：

1. `startOffset` 和 `endOffset` 指向用户修正后的章节正文。
2. `selectedText` 保存创建拆解时的选文快照，用于正文后续修改后的追溯。
3. 如果选文很长，需要给用户直接阅读，应保存为 TXT，并在 JSON 中记录 `selectedTextPath`。
4. 同一选文可以关联多个拆解卡片。
5. 如果正文编辑导致偏移变化，后续迁移逻辑应尽量按选文快照重新匹配，并保留旧偏移作为历史信息。

## 拆解卡片引用

拆解卡片必须引用原文片段，而不是只保存一段孤立笔记。

`deconstructions/deconstruction-cards.json` 示例：

```json
{
  "version": 1,
  "items": [
    {
      "id": "deconstruction-20260821-0001",
      "bookId": "book-20260821-001",
      "sourceTextId": "source-20260821-0001",
      "planTypeId": "scene",
      "fieldValues": {
        "scenePurpose": "让重逢带有迟疑和对峙感",
        "emotionChange": "从确认对方出现，到压住回应，再到被环境迫使面对。"
      },
      "tagIds": ["tag-atmosphere", "tag-reunion"],
      "createdAt": "2026-08-21T10:35:00+08:00",
      "updatedAt": "2026-08-21T10:35:00+08:00"
    }
  ]
}
```

字段原则：

1. `sourceTextId` 负责来源追溯。
2. `planTypeId` 指向当前书拆解方案中的拆解类型。
3. `fieldValues` 必须按当前书拆解方案字段保存，不能写死固定字段。
4. 一个 `sourceTextId` 可以被多个拆解卡片引用。
5. 删除拆解卡片不应直接删除原文片段；只有确认没有任何引用时才可清理。

## 与书库索引的关系

书库根目录 `index.json` 不应成为阅读位置的唯一数据源。

允许做法：

1. 书库列表只保存小说名、路径、大小等轻量入口信息。
2. 打开小说时读取单本小说目录中的 `reading/reading-state.json`。
3. 如需在书库列表展示“继续阅读到第几章”，可以从单本小说目录读取阅读状态后展示。
4. 书库列表展示阅读进度百分比时，优先从单本小说目录的 `reading/reading-state.json` 读取 `readingProgressPercent`。
5. 如为了性能在 `index.json` 中缓存阅读摘要，该字段必须可从单本小说目录重新生成。

禁止做法：

1. 不把阅读位置只保存到应用系统数据目录。
2. 不把阅读位置只保存到书库 `index.json`。
3. 不把 SQLite 缓存作为恢复阅读位置的唯一依据。

## 数据迁移和恢复

新增阅读工作台数据时必须考虑旧数据：

1. 如果旧小说目录没有 `reading/reading-state.json`，打开第一章第一处内容，并在首次保存阅读位置时创建文件。
2. 如果旧阅读状态没有 `readingProgressPercent`，读取书库或阅读工作台时按 `book.json` 章节字数和 `lastPosition` 估算，不应阻断打开。
3. 如果旧阅读状态没有 `preferences.fontSize`，打开时使用默认字号 18，用户调整或保存阅读状态时写回。
4. 如果旧拆解卡片没有 `sourceTextId`，迁移时应根据旧的书籍、章节、原文位置创建 `SourceText`。
5. 如果 `source-texts.json` 损坏，应提示用户修复或从拆解卡片中的来源快照重建。
6. 如果 SQLite 索引损坏，删除缓存后必须能从 `book.json`、章节 TXT、阅读状态、原文片段和拆解 JSON 重建索引。
