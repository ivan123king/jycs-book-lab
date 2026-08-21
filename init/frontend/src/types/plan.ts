// plan.ts 定义拆解方案页面原型使用的数据类型。

/**
 * DemoPlanField 是拆解方案字段配置表的原型行。
 * 真实字段属于单本书的拆解方案，不能写成全局固定模板。
 */
export type DemoPlanField = {
  /** 表格渲染使用的稳定键。 */
  key: string
  /** 字段所属拆解维度。 */
  dimension: string
  /** 字段所属拆解类型。 */
  type: string
  /** 用户可配置的字段名称。 */
  field: string
  /** 字段输入类型展示名。 */
  fieldType: string
  /** 字段是否必填的展示状态。 */
  required: string
  /** 字段是否在表单中显示的展示状态。 */
  visible: string
}
