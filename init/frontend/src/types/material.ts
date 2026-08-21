// material.ts 定义素材库页面原型使用的数据类型。

/**
 * DemoMaterial 是素材库列表的原型数据。
 * 正式素材必须保留来源书籍、章节、原文片段和来源拆解卡片引用。
 */
export type DemoMaterial = {
  /** 素材标题。 */
  title: string
  /** 素材所属分类展示名。 */
  category: string
  /** 素材正文摘要；长文本正式数据后续应保存为 TXT。 */
  text: string
  /** 来源链路展示文本。 */
  source: string
  /** 素材关联标签展示名。 */
  tags: string[]
  /** 是否收藏，用于素材库筛选和强调展示。 */
  favorite: boolean
}
