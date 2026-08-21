// workbenchFixtures.ts 集中保存第一阶段页面原型数据，避免 App.vue 和页面组件混杂静态样例。
import type { DemoAnalysisRow } from '@/types/analysis'
import type { CurrentBookSummary } from '@/types/currentBook'
import type { DemoDeconstructionCard } from '@/types/deconstruction'
import type { DemoMaterial } from '@/types/material'
import type { DemoPlanField } from '@/types/plan'
import type { DemoChapter } from '@/types/reading'

export const defaultCurrentBook: CurrentBookSummary = {
  title: '雾港来信',
  author: '林岚',
  path: 'E:/Novels/雾港来信',
  description: '一部以雾港旧案和人物重逢为线索的长篇小说。',
  chapters: 42,
  deconstructions: 128,
  materials: 36,
  updatedAt: '今天 09:18',
  readingPosition: '第三章 雨夜码头 · 68%',
}

export const demoChapters: DemoChapter[] = [
  { key: 'chapter-01', title: '第一章 远处的灯', count: 4, progress: 100 },
  { key: 'chapter-02', title: '第二章 借宿人', count: 7, progress: 100 },
  { key: 'chapter-03', title: '第三章 雨夜码头', count: 12, progress: 68 },
  { key: 'chapter-04', title: '第四章 空房间', count: 3, progress: 0 },
  { key: 'chapter-05', title: '第五章 回声', count: 0, progress: 0 },
]

export const demoDeconstructionCards: DemoDeconstructionCard[] = [
  {
    title: '雨夜码头的重逢场景',
    type: '场景拆解',
    chapter: '第三章 雨夜码头',
    source: '从仓库檐下到渡轮鸣笛',
    summary: '通过潮湿环境、延迟回应和灯光遮挡，把人物重逢写成带有试探性的对峙。',
    tags: ['氛围', '重逢', '悬念'],
  },
  {
    title: '沈知遥的防御性对白',
    type: '对白拆解',
    chapter: '第三章 雨夜码头',
    source: '“你来晚了。”',
    summary: '台词表面责备，实际暴露等待已久的心理；短句降低解释感。',
    tags: ['对白', '人物关系'],
  },
  {
    title: '码头旧灯的意象复用',
    type: '意象拆解',
    chapter: '第一章 远处的灯',
    source: '雾里那盏忽明忽暗的灯',
    summary: '用同一个物象连接开篇伏笔和第三章情绪回收。',
    tags: ['意象', '伏笔'],
  },
]

export const demoMaterials: DemoMaterial[] = [
  {
    title: '用天气压低重逢情绪',
    category: '场景写法',
    text: '雨声、锈味和码头广播先于人物登场，让重逢不是拥抱，而是被环境压住的确认。',
    source: '雾港来信 / 第三章 雨夜码头 / 场景拆解',
    tags: ['氛围', '重逢'],
    favorite: true,
  },
  {
    title: '短句对白里的旧关系',
    category: '对白技巧',
    text: '旧识之间不需要解释全部前史，一句责备可以同时包含等待、失望和试探。',
    source: '雾港来信 / 第三章 雨夜码头 / 对白拆解',
    tags: ['对白', '人物关系'],
    favorite: false,
  },
  {
    title: '重复意象的回收方式',
    category: '结构技巧',
    text: '开篇出现的灯在关键章节再次出现，读者会自然把情绪线索连起来。',
    source: '雾港来信 / 第一章 远处的灯 / 意象拆解',
    tags: ['意象', '伏笔'],
    favorite: true,
  },
]

export const demoPlanFields: DemoPlanField[] = [
  { key: '1', dimension: '场景', type: '场景拆解', field: '场景目的', fieldType: '短文本', required: '是', visible: '显示' },
  { key: '2', dimension: '场景', type: '场景拆解', field: '情绪变化', fieldType: '长文本', required: '是', visible: '显示' },
  { key: '3', dimension: '人物', type: '对白拆解', field: '潜台词', fieldType: '长文本', required: '否', visible: '显示' },
  { key: '4', dimension: '结构', type: '意象拆解', field: '前后呼应', fieldType: '长文本', required: '否', visible: '显示' },
]

export const demoAnalysisRows: DemoAnalysisRow[] = [
  { type: '场景拆解', count: 46, percent: 36 },
  { type: '对白拆解', count: 32, percent: 25 },
  { type: '人物弧光', count: 27, percent: 21 },
  { type: '意象拆解', count: 23, percent: 18 },
]
