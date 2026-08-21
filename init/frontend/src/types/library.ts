// library.ts 定义前端调用 Wails 书库 API 时复用的数据类型别名。
import type { library } from '@wails/go/models'

/**
 * LibraryBookEntry 是书库 index.json 中的轻量书籍入口。
 * 它只用于书库列表展示，不承载章节、拆解或素材详情。
 */
export type LibraryBookEntry = library.LibraryBookEntry

/**
 * LibraryState 表示当前书库目录和轻量索引。
 * 其中 index 来自书库根目录 index.json，单本书正式数据仍在 bookPath 指向的项目目录中。
 */
export type LibraryState = library.LibraryState

/**
 * ImportDraft 是 TXT 小说确认入库前的导入预览草稿。
 * 草稿用于展示小说名、简介、目标目录和章节标题，不代表正式数据已写入。
 */
export type ImportDraft = library.ImportDraft

/**
 * BookProject 对应单本小说项目目录中的 book.json。
 * 前端打开书籍后应以它作为当前书工作区的数据入口。
 */
export type BookProject = library.BookProject
