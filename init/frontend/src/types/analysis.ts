// analysis.ts 定义基础分析页面从当前书正式数据计算出的展示类型。

/**
 * AnalysisBreakdownRow 是基础分析页中按类型或标签聚合后的统计行。
 * 数据来源必须是当前小说项目目录中的拆解卡片和素材正式数据。
 */
export type AnalysisBreakdownRow = {
  /** 统计项展示名。 */
  label: string
  /** 该统计项下的数据数量。 */
  count: number
  /** 用于进度条展示的占比。 */
  percent: number
}

/**
 * AnalysisChapterRow 展示某章关联的写作手法拆解和剧情拆解数量。
 */
export type AnalysisChapterRow = {
  /** 章节 ID，用于稳定渲染。 */
  chapterId: string
  /** 页面展示用章节标题。 */
  title: string
  /** 章节顺序。 */
  order: number
  /** 写作手法拆解数量。 */
  writingCount: number
  /** 剧情拆解引用数量。 */
  storyCount: number
  /** 两类拆解合计数量。 */
  totalCount: number
  /** 用于进度条展示的占比。 */
  percent: number
}
