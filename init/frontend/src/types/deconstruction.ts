// deconstruction.ts 定义拆解卡片页面原型使用的数据类型。

/**
 * DemoDeconstructionCard 是拆解卡片列表的原型数据。
 * 后续正式结构应保留 sourceTextId，确保同一原文片段可创建多张拆解卡片。
 */
export type DemoDeconstructionCard = {
  /** 拆解卡片标题。 */
  title: string
  /** 当前书拆解方案中的拆解类型展示名。 */
  type: string
  /** 来源章节展示名。 */
  chapter: string
  /** 来源原文片段或位置摘要。 */
  source: string
  /** 用户沉淀的拆解摘要。 */
  summary: string
  /** 卡片关联标签展示名。 */
  tags: string[]
}
