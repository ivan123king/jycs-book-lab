# 开发指南索引

后续任何 AI 或开发者修改本项目时，必须先阅读本文，再按任务涉及的模块阅读对应文档。

## 阅读顺序

1. 先读：`docs/ai-guides/product-brief.md`
2. 再读：本文
3. 根据任务选择模块文档：
   - 工程创建、初始化、首次运行：`docs/dev-guides/project-initialization.md`
   - 页面、交互、组件：`docs/dev-guides/frontend-pages.md`
   - 页面设计、页面结构、功能模块拆分：`docs/dev-guides/page-design/overall-page-structure.md`
   - 书库、小说导入、书库数据结构：`docs/dev-guides/book-library/design-principles.md`、`docs/dev-guides/book-library/data-structure.md`、`docs/dev-guides/book-library/page-structure.md`
   - 阅读工作台、阅读/编辑 TAB、目录分页、选文拆解弹框、阅读位置：`docs/dev-guides/reading-workbench/design-principles.md`、`docs/dev-guides/reading-workbench/page-structure.md`、`docs/dev-guides/reading-workbench/data-structure.md`
   - 拆解方案、拆解模板、`analysisKind`、当前模板引用、阅读中创建拆解的类型和维度配置：`docs/dev-guides/deconstruction-plan.md`
   - 写作手法拆解、选文拆解弹框、写作手法卡片、沉淀素材入口：`docs/dev-guides/writing-technique-deconstruction/implementation.md`
   - 剧情拆解、非连续章节范围、剧情简介、剧情维度输入、剧情草稿：`docs/dev-guides/story-plot-deconstruction/implementation.md`
   - 小说内素材、拆解沉淀素材、单本小说素材正式数据：`docs/dev-guides/book-material-library/implementation.md`
   - 书库素材库、单书素材引用、书库自有素材、移除书籍前素材迁移：`docs/dev-guides/global-material-library.md`
   - 后端、本地数据、业务服务：`docs/dev-guides/backend-local-data.md`
   - 技术选型、进程边界、存储原则：`docs/dev-guides/tech-stack.md`
   - 注释、文件组织、协作规则：`docs/dev-guides/coding-rules.md`

## 文档索引维护

后续新增任何 `docs` 文档时，必须同步更新本文索引。

新增文档需要写清：

1. 文档路径。
2. 文档用途。
3. 哪类任务需要先阅读该文档。

不要让新的指南文档脱离索引单独存在。后续 AI 只应从本文开始寻找需要阅读的开发文档。

## 当前文档索引

| 路径 | 用途 | 适用任务 |
| --- | --- | --- |
| `docs/ai-guides/product-brief.md` | AI 产品简述，说明产品定位、核心对象、产品原则和 MVP 边界。 | 所有产品理解、功能设计、开发任务。 |
| `docs/dev-guides/index.md` | 开发指南总索引，说明阅读顺序、模块边界和检查清单。 | 所有代码修改和文档修改任务。 |
| `docs/dev-guides/project-initialization.md` | 工程创建指南，说明如何在仓库根目录下创建 `init/` Wails + Vue 3 工程并完成首次运行和构建验证。 | 首次创建工程、复现初始化流程、检查初始化命令和本地开发环境任务。 |
| `docs/dev-guides/tech-stack.md` | 技术选型与本地 PC 产品约束。 | 技术架构、依赖选择、本地存储、进程边界相关任务。 |
| `docs/dev-guides/frontend-pages.md` | 页面模块、交互和组件开发指南。 | 书库、导入、阅读、拆解、素材、分析等前端页面任务。 |
| `docs/dev-guides/page-design/overall-page-structure.md` | 页面整体结构与模块拆分文档，说明页面导航层级、模块职责、复用组件和第一阶段页面推进顺序。 | 页面设计评审、页面功能拆分、前端编码前的结构确认任务。 |
| `docs/dev-guides/book-library/design-principles.md` | 书库模块设计原则，说明书库目的、入库规则和边界。 | 书库、小说导入、书库目录设计相关任务。 |
| `docs/dev-guides/book-library/data-structure.md` | 书库数据结构，说明 `index.json`、单本小说 `book.json` 和入库目录结构。 | 书库索引、小说导入、章节识别结果保存相关任务。 |
| `docs/dev-guides/book-library/page-structure.md` | 书库页面结构，说明书库页和书库内导入面板的布局、流程和页面状态。 | 书库页、书库内导入面板、章节分页标题展示相关任务。 |
| `docs/dev-guides/reading-workbench/design-principles.md` | 阅读工作台设计原则，说明阅读优先、阅读/编辑 TAB、目录按钮、选文拆解弹框和功能边界。 | 阅读工作台、正文阅读、正文编辑、选文创建拆解相关任务。 |
| `docs/dev-guides/reading-workbench/page-structure.md` | 阅读工作台页面结构，说明上一页/下一页阅读区、右侧目录分页弹层、拆解弹框和页面状态。 | 阅读工作台布局、目录交互、阅读位置恢复、拆解弹框页面任务。 |
| `docs/dev-guides/reading-workbench/data-structure.md` | 阅读工作台数据结构，说明阅读位置、目录分页计算、原文片段和拆解卡片来源引用。 | 阅读位置保存、选文来源追溯、拆解卡片数据结构和旧数据迁移任务。 |
| `docs/dev-guides/deconstruction-plan.md` | 拆解方案与模板指南，说明当前方案引用、模板目录、`analysisKind`、类型、维度、模板切换和卡片快照保存规则。 | 拆解方案页、拆解模板、阅读拆解表单、拆解卡片保存结构相关任务。 |
| `docs/dev-guides/writing-technique-deconstruction/implementation.md` | 写作手法拆解模块实现指南，说明选文创建、来源追溯、卡片页、数据结构和沉淀素材入口。 | 写作手法拆解弹框、写作手法拆解卡片页、选文来源追溯和素材沉淀入口相关任务。 |
| `docs/dev-guides/story-plot-deconstruction/implementation.md` | 剧情拆解模块实现指南，说明剧情拆解与写作手法拆解的区别、非连续章节范围、剧情简介、维度折叠、章节内容查看和草稿保存。 | 剧情拆解页、新建剧情拆解弹框、剧情卡片详情、剧情素材沉淀相关任务。 |
| `docs/dev-guides/book-material-library/implementation.md` | 小说内素材模块实现指南，说明单本小说素材正式数据、来源追溯、拆解沉淀素材和书库素材刷新关系。 | 小说内素材页、素材数据结构、素材来源追溯、从拆解沉淀素材相关任务。 |
| `docs/dev-guides/global-material-library.md` | 书库素材库指南，说明单书素材引用、书库自有素材、分页检索和移除书籍前素材迁移规则。 | 书库素材库页、跨书素材检索、书库自建素材、素材迁移和来源项目移除处理任务。 |
| `docs/dev-guides/backend-local-data.md` | 本地后端、数据模型、存储和业务服务指南。 | 文件导入、章节识别、数据库、搜索、标签、素材、来源追溯相关任务。 |
| `docs/dev-guides/coding-rules.md` | 编码、注释、命名和协作规则。 | 新增文件、模块、函数、复杂业务逻辑、数据结构变更任务。 |

## 模块边界

本项目功能分为两大类：

1. 页面模块：负责 PC 端界面、阅读交互、表单配置、素材管理和分析呈现。
2. 后端模块：负责本地文件导入、章节识别、本地存储、数据模型、业务规则和搜索查询。

页面模块不要直接承担复杂数据持久化逻辑。后端模块不要绑定具体 UI 表现。

## 第一阶段功能模块

P0 只围绕完整闭环开发：

1. 书库模块
2. 小说导入与章节识别模块
3. 阅读与正文编辑模块
4. 写作手法拆解模块
5. 剧情拆解模块
6. 拆解方案配置模块
7. 小说内素材库模块
8. 书库素材库模块
9. 分类与标签模块
10. 来源追溯模块

暂不主动开发跨书比较、复杂图表、AI 自动分析、云同步、账号系统。

## 修改前检查

每次开发前确认：

1. 任务属于哪个模块。
2. 是否影响书籍、章节、原文片段、写作手法拆解卡片、剧情拆解卡片、素材之间的关系。
3. 是否需要更新本地数据结构。
4. 是否需要迁移旧数据。
5. 是否会增加用户填写成本。
6. 是否破坏阅读优先和选文优先原则。

## 提交前检查

每次修改后确认：

1. 页面和后端职责没有混在一起。
2. 新增文件有文件注释。
3. 新增模块有模块注释。
4. 新增导出函数或复杂函数有函数注释。
5. 关键业务逻辑有功能注释。
6. 新增字段能解释其业务含义。
7. 本地数据读写有错误处理。
8. 不需要的抽象、框架和依赖没有被引入。
