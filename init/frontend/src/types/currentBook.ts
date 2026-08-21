// currentBook.ts 定义应用壳层展示当前打开小说所需的轻量摘要类型。

/**
 * CurrentBookSummary 是顶部栏和轻量统计使用的当前书摘要。
 * 章节正文、拆解卡片、素材等正式数据仍应从单本小说项目目录读取。
 */
export type CurrentBookSummary = {
  /** 当前打开的小说名称。 */
  title: string
  /** 作者信息；第一阶段可能来自元数据或页面占位。 */
  author: string
  /** 单本小说项目目录路径，复制该目录即可迁移书籍数据。 */
  path: string
  /** 小说简介，用于书库或顶部上下文展示。 */
  description: string
  /** 当前书章节总数。 */
  chapters: number
  /** 当前书拆解卡片数量摘要。 */
  deconstructions: number
  /** 当前书素材数量摘要。 */
  materials: number
  /** 当前书元数据最后更新时间的展示文本。 */
  updatedAt: string
  /** 当前书最近阅读位置的展示文本。 */
  readingPosition: string
}
