// navigation.ts 定义应用壳层的导航类型，不包含具体页面业务数据。

/**
 * ViewKey 表示应用壳层可切换的一级视图。
 * 该类型只用于前端路由和菜单状态，不对应单本小说的正式数据文件。
 */
export type ViewKey =
  | 'library'
  | 'reading'
  | 'deconstructions'
  | 'materials'
  | 'plan'
  | 'analysis'
  | 'settings'
