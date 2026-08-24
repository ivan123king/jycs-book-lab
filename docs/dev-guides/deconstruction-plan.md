# 拆解方案与模板指南

本文定义拆解方案配置模块的产品流程、数据边界和第一阶段实现规则。拆解方案不是普通笔记模板，也不是 AI 自动分析配置；它负责让用户按自己的分析口径配置“类型 -> 维度 -> 看法”。

## 核心概念

拆解模板属于单本小说项目。用户可以为同一本书维护多条拆解路线，目前第一阶段先区分两类：

| `analysisKind` | 中文名称 | 使用场景 |
| --- | --- | --- |
| `writingTechnique` | 写作模板 | 阅读中选中文字后，分析写作手法、人物感情、结构设计等。 |
| `storyPlot` | 剧情模板 | 在剧情拆解页中新建剧情拆解时，作为剧情类型和初始维度提示。 |

写作模板的阅读侧流程：

```text
选中文字 -> 选择拆解类型 -> 选择拆解维度 -> 填写自己的看法
```

示例：

```text
表现手法 -> 环境描写 -> 这段用雨声和灯光压低重逢情绪。
人物感情 -> 情绪变化 -> 人物从防御到迟疑，再到被迫回应。
```

第一阶段不再把写作手法拆解字段设计成单行文本、多行文本、单选、多选、标签、数字、评分等表单字段。写作手法模板只负责提供“类型”和“维度”，用户的分析正文统一作为拆解内容保存。

剧情模板也只提供“类型”和“维度”的初始提示。用户新建某张剧情拆解卡片时，可以修改、新增或删除本卡片中的维度，这些调整不应自动改写模板文件。

## 项目目录结构

导入小说入库时，应在单本小说项目目录中创建当前方案引用文件和模板目录：

```text
books/{bookId}/
├── book.json
├── original.txt
├── chapters/
├── deconstruction-plan.json
└── deconstruction-templates/
    ├── 系统内置写作手法.json
    └── 系统内置剧情分析.json
```

说明：

1. `deconstruction-plan.json` 只保存当前书各类拆解路线正在使用哪个模板文件，不保存模板内容。
2. `deconstruction-templates/` 保存当前书所有可选拆解模板。
3. `系统内置写作手法.json` 和 `系统内置剧情分析.json` 是导入小说时创建的两份默认模板，用户可以修改。
4. 用户另存为模板时，按用户输入的模板名称在 `deconstruction-templates/` 下新增 JSON 文件。
5. 模板目录属于单本小说项目目录，复制该小说项目目录即可带走当前方案引用和该书的模板列表。

## 当前方案引用结构

`deconstruction-plan.json` 使用引用结构，方便后续继续增加新的拆解路线：

```json
{
  "version": 1,
  "bookId": "book-20260824-001",
  "currentTemplates": {
    "writingTechnique": "系统内置写作手法.json",
    "storyPlot": "系统内置剧情分析.json"
  },
  "updatedAt": "2026-08-24T10:30:00+08:00"
}
```

字段说明：

| 字段 | 含义 |
| --- | --- |
| `version` | 结构版本。 |
| `bookId` | 所属小说 ID。 |
| `currentTemplates` | 每类 `analysisKind` 当前使用的模板文件名。 |
| `updatedAt` | 最近保存当前模板引用的时间。 |

## 模板结构

模板 JSON 保存在 `deconstruction-templates/` 目录下：

```json
{
  "version": 1,
  "bookId": "book-20260824-001",
  "analysisKind": "writingTechnique",
  "templateName": "系统内置写作手法",
  "types": [
    {
      "id": "technique",
      "name": "表现手法",
      "order": 1,
      "dimensions": [
        { "id": "environment", "name": "环境描写", "order": 1 },
        { "id": "rhythm", "name": "节奏控制", "order": 2 }
      ]
    }
  ],
  "updatedAt": "2026-08-24T10:30:00+08:00"
}
```

字段说明：

| 字段 | 含义 |
| --- | --- |
| `analysisKind` | 模板所属拆解路线，目前是 `writingTechnique` 或 `storyPlot`。 |
| `templateName` | 展示给用户看的模板名称，让用户知道模板用途。 |
| `types` | 拆解类型列表，例如表现手法、人物感情、剧情分析。 |
| `dimensions` | 类型下的拆解维度，例如环境描写、情绪变化、核心事件、悬念设置。 |
| `order` | 页面展示顺序。 |

剧情分析默认类型可以包含装逼打脸、危机救场、误会反转、情感爆发、伏笔回收等。默认维度可以包含：剧情简介、铺垫、剧情推进、冲突变化、人物目标、人物关系变化、信息揭示、悬念设置、高潮、围观反应、余波、伏笔与回收、本段剧情作用。

## 模板页面

拆解方案页面应包含：

1. 模板类型切换：写作模板、剧情模板。
2. 当前模板列表：读取 `deconstruction-templates/` 下所有 JSON，并按 `analysisKind` 过滤展示。
3. 当前使用模板：按 `deconstruction-plan.json` 中当前 `analysisKind` 的引用高亮。
4. 模板名称：用户可修改当前模板的展示名称。
5. 类型管理：新增、改名、删除或调整顺序。
6. 维度管理：在某个类型下新增、改名、删除或调整顺序。
7. 保存模板：保存正在编辑的模板文件。
8. 另存为模板：用新模板名称创建一份同 `analysisKind` 的新模板 JSON。
9. 设为当前模板：只更新 `deconstruction-plan.json` 中对应 `analysisKind` 的模板引用。

第一阶段删除类型或维度时，应先提示已有拆解卡片不会被改写。已有卡片保存的是创建时的快照，不依赖当前模板继续显示。

## 当前模板应用规则

“设为当前模板”的含义是：

1. 读取用户选中的模板 JSON。
2. 校验模板 `analysisKind` 与当前页面类型一致。
3. 将 `deconstruction-plan.json` 中对应 `analysisKind` 的引用改为该模板文件名。
4. 不把模板内容复制到 `deconstruction-plan.json`。
5. 后续打开阅读工作台时，写作手法拆解读取当前 `writingTechnique` 引用指向的模板。
6. 后续打开剧情拆解页时，新建剧情拆解读取当前 `storyPlot` 引用指向的模板。

如果当前书已有拆解卡片，仍允许切换当前模板，因为历史卡片保存了分析类型、类型名、维度名和分析内容快照，不依赖当前模板结构。

## 拆解卡片保存规则

拆解卡片必须保留来源追溯，也必须保存创建时的模板快照信息：

```json
{
  "id": "deconstruction-20260824-0001",
  "bookId": "book-20260824-001",
  "sourceTextId": "source-20260824-0001",
  "analysisKind": "writingTechnique",
  "planTypeId": "technique",
  "planTypeName": "表现手法",
  "dimensionId": "environment",
  "dimensionName": "环境描写",
  "content": "这段用雨声和灯光压低重逢情绪。",
  "tagIds": ["氛围"],
  "createdAt": "2026-08-24T10:35:00+08:00",
  "updatedAt": "2026-08-24T10:35:00+08:00"
}
```

原则：

1. `sourceTextId` 负责回到原文片段。
2. `analysisKind` 保留创建拆解时使用的分析路线。
3. `planTypeId`、`dimensionId` 保留创建时的结构标识。
4. `planTypeName`、`dimensionName` 保存创建时展示名称，避免后续模板改名导致旧卡片显示变化。
5. `content` 保存用户自己的分析看法。
6. 修改模板不能批量改写历史卡片。

## 与阅读工作台关系

阅读工作台创建选文拆解时：

1. 读取当前书 `deconstruction-plan.json`。
2. 从 `currentTemplates.writingTechnique` 找到当前写作模板。
3. 用户选中文字。
4. 弹框展示原文片段。
5. 用户选择拆解类型。
6. 用户选择该类型下的拆解维度。
7. 用户填写自己的看法。
8. 保存拆解卡片和来源片段。

阅读工作台只负责写作手法拆解入口。写作手法拆解详细规则见 `docs/dev-guides/writing-technique-deconstruction/implementation.md`。剧情拆解入口在当前书工作区的剧情拆解页中实现，详细规则见 `docs/dev-guides/story-plot-deconstruction/implementation.md`。

## 与剧情拆解关系

剧情拆解页新建剧情拆解时：

1. 读取当前书 `deconstruction-plan.json`。
2. 从 `currentTemplates.storyPlot` 找到当前剧情模板。
3. 用户输入非连续章节范围，例如 `1-3, 8, 12-15`。
4. 用户填写剧情名称和剧情简介。
5. 用户选择剧情类型。
6. 页面加载该类型下的初始维度。
7. 用户可以修改、新增或删除本次剧情拆解中的维度。
8. 用户为每个维度填写描述。
9. 保存剧情拆解卡片，保留章节范围、模板快照和本卡片维度内容。

剧情拆解卡片不依赖模板后续变化。模板只作为新建时的提示，不是历史剧情卡片的展示事实来源。

## 旧数据兼容

旧项目可能只有早期的 `deconstruction-plan.json`，其中保存的是完整模板内容，或者 `DeconstructionType` 直接带 `dimension` 和 `fields`：

1. 打开旧方案时，应先把旧 `deconstruction-plan.json` 中的完整模板内容迁移为 `deconstruction-templates/系统内置写作手法.json`。
2. 然后把 `deconstruction-plan.json` 改写为当前模板引用结构。
3. 旧 `dimension` 可作为该类型下的一个维度。
4. 旧 `fields` 不再作为创建拆解的表单字段；如需保留，可只用于迁移参考，不进入第一阶段 UI。
5. 旧拆解卡片的 `fieldValues` 不应丢失。后续展示旧卡片时可把 `fieldValues` 作为历史内容摘要。

## 边界

第一阶段不做：

1. 跨书模板中心。
2. 多人协作或权限。
3. 复杂版本回滚。
4. AI 自动生成模板。
5. 根据章节自动切换模板。
6. 把剧情卡片中的临时维度修改自动同步回模板。

不同章节可能需要不同分析维度，用户可以在拆解方案页切换当前写作模板，再回到阅读工作台继续创建选文拆解。
