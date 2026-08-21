// analysis.ts 定义基础分析页面原型使用的数据类型。

/**
 * DemoAnalysisRow 是按拆解类型统计的原型数据。
 * 第一阶段只做当前书轻量统计，不包含跨书比较。
 */
export type DemoAnalysisRow = {
  /** 拆解类型展示名。 */
  type: string
  /** 该拆解类型下的卡片数量。 */
  count: number
  /** 用于进度条展示的占比。 */
  percent: number
}
