// reading.ts 定义阅读工作台页面原型使用的章节展示类型。

/**
 * DemoChapter 是阅读工作台原型中的章节目录行。
 * 真实实现时应由 book.json 中的章节元数据和阅读状态计算生成。
 */
export type DemoChapter = {
  /** 前端渲染列表时使用的稳定键。 */
  key: string
  /** 章节标题。 */
  title: string
  /** 当前章节关联的拆解卡片数量摘要。 */
  count: number
  /** 当前章节阅读进度百分比。 */
  progress: number
}
