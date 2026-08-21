<script setup lang="ts">
// App.vue 负责展示第一阶段页面结构原型，使用假数据模拟书库、导入、阅读拆解、素材和方案配置页面。
import { computed, h, ref, type Component } from 'vue'
import {
  NAlert,
  NBadge,
  NButton,
  NButtonGroup,
  NCard,
  NConfigProvider,
  NDataTable,
  NDescriptions,
  NDescriptionsItem,
  NDivider,
  NForm,
  NFormItem,
  NGrid,
  NGi,
  NInput,
  NLayout,
  NLayoutContent,
  NLayoutHeader,
  NLayoutSider,
  NList,
  NListItem,
  NMenu,
  NMessageProvider,
  NPagination,
  NProgress,
  NScrollbar,
  NSelect,
  NSpace,
  NSwitch,
  NTabPane,
  NTabs,
  NTag,
  NThing,
  NTimeline,
  NTimelineItem,
  NTooltip,
  NStatistic,
} from 'naive-ui'
import type { MenuOption } from 'naive-ui'
import {
  BarChart3,
  BookMarked,
  BookOpenText,
  Bookmark,
  CheckCircle2,
  CircleAlert,
  Copy,
  Database,
  FileInput,
  FolderOpen,
  Highlighter,
  Library,
  ListChecks,
  MapPinned,
  Pencil,
  Plus,
  RotateCcw,
  Save,
  Search,
  Settings,
  SlidersHorizontal,
} from 'lucide-vue-next'

type ViewKey =
  | 'library'
  | 'reading'
  | 'deconstructions'
  | 'materials'
  | 'plan'
  | 'analysis'
  | 'settings'

const activeView = ref<ViewKey>('library')
const libraryMode = ref<'list' | 'import'>('list')
const readerMode = ref<'read' | 'edit'>('read')
const selectedChapterKey = ref('chapter-03')
const selectedDeconstructionType = ref('scene')
const importChapterPage = ref(1)

const currentBook = {
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

const projectCards = [
  {
    title: '雾港来信',
    author: '林岚',
    path: 'E:/BookLibrary/books/book-20260821-001',
    size: '178 KB',
    chapters: 42,
    deconstructions: 128,
    updatedAt: '今天 09:18',
    status: '可打开',
  },
  {
    title: '长明灯',
    author: '周弈',
    path: 'E:/BookLibrary/books/book-20260821-002',
    size: '251 KB',
    chapters: 31,
    deconstructions: 73,
    updatedAt: '昨天 22:41',
    status: '继续阅读',
  },
  {
    title: '旧城雪',
    author: '未知作者',
    path: 'E:/BookLibrary/books/book-20260821-003',
    size: '0 KB',
    chapters: 0,
    deconstructions: 0,
    updatedAt: '路径异常',
    status: '需要恢复',
  },
]

const chapters = [
  { key: 'chapter-01', title: '第一章 远处的灯', count: 4, progress: 100 },
  { key: 'chapter-02', title: '第二章 借宿人', count: 7, progress: 100 },
  { key: 'chapter-03', title: '第三章 雨夜码头', count: 12, progress: 68 },
  { key: 'chapter-04', title: '第四章 空房间', count: 3, progress: 0 },
  { key: 'chapter-05', title: '第五章 回声', count: 0, progress: 0 },
]

const importChapters = [
  { order: 1, title: '第一章 远处的灯' },
  { order: 2, title: '第二章 借宿人' },
  { order: 3, title: '第三章 雨夜码头' },
  { order: 4, title: '第四章 空房间' },
  { order: 5, title: '第五章 回声' },
  { order: 6, title: '第六章 潮汐表' },
  { order: 7, title: '第七章 没有署名的信' },
  { order: 8, title: '第八章 旧仓库' },
  { order: 9, title: '第九章 海雾里的证词' },
  { order: 10, title: '第十章 灯塔下' },
]

const importInfo = {
  bookName: '雾港来信',
  description: '一部以雾港旧案和人物重逢为线索的长篇小说。',
  sourcePath: 'D:/Downloads/雾港来信.txt',
  libraryPath: 'E:/BookLibrary',
  targetPath: 'E:/BookLibrary/books/book-20260821-001',
  size: '178 KB',
  chapterCount: 3142,
}

const importChapterColumns = [
  { title: '序号', key: 'order', width: 90 },
  { title: '章节标题', key: 'title' },
]

const deconstructionCards = [
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

const materials = [
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

const planFields = [
  { key: '1', dimension: '场景', type: '场景拆解', field: '场景目的', fieldType: '短文本', required: '是', visible: '显示' },
  { key: '2', dimension: '场景', type: '场景拆解', field: '情绪变化', fieldType: '长文本', required: '是', visible: '显示' },
  { key: '3', dimension: '人物', type: '对白拆解', field: '潜台词', fieldType: '长文本', required: '否', visible: '显示' },
  { key: '4', dimension: '结构', type: '意象拆解', field: '前后呼应', fieldType: '长文本', required: '否', visible: '显示' },
]

const planColumns = [
  { title: '维度', key: 'dimension' },
  { title: '拆解类型', key: 'type' },
  { title: '字段名称', key: 'field' },
  { title: '字段类型', key: 'fieldType' },
  { title: '必填', key: 'required' },
  { title: '状态', key: 'visible' },
]

const analysisRows = [
  { type: '场景拆解', count: 46, percent: 36 },
  { type: '对白拆解', count: 32, percent: 25 },
  { type: '人物弧光', count: 27, percent: 21 },
  { type: '意象拆解', count: 23, percent: 18 },
]

const navOptions = computed<MenuOption[]>(() => [
  {
    label: '书库',
    key: 'library',
    icon: renderMenuIcon(Library),
  },
  {
    type: 'group',
    label: '当前书',
    key: 'workspace',
    children: [
      { label: '阅读', key: 'reading', icon: renderMenuIcon(BookOpenText) },
      { label: '拆解', key: 'deconstructions', icon: renderMenuIcon(Highlighter) },
      { label: '素材', key: 'materials', icon: renderMenuIcon(Bookmark) },
      { label: '方案', key: 'plan', icon: renderMenuIcon(ListChecks) },
      { label: '分析', key: 'analysis', icon: renderMenuIcon(BarChart3) },
    ],
  },
  {
    label: '设置',
    key: 'settings',
    icon: renderMenuIcon(Settings),
  },
])

const activeTitle = computed(() => {
  const titles: Record<ViewKey, string> = {
    library: '书库',
    reading: '阅读与拆解',
    deconstructions: '拆解卡片',
    materials: '素材库',
    plan: '拆解方案',
    analysis: '基础分析',
    settings: '全局设置',
  }
  return activeView.value === 'library' && libraryMode.value === 'import' ? '书库 / 导入小说' : titles[activeView.value]
})

/**
 * 为 Naive UI 菜单渲染统一尺寸的 lucide 图标，避免各页面重复写图标结构。
 */
function renderMenuIcon(icon: Component) {
  return () => h(icon, { size: 18, strokeWidth: 1.8 })
}

function goTo(view: ViewKey) {
  activeView.value = view
}

function showLibraryImport() {
  activeView.value = 'library'
  libraryMode.value = 'import'
}
</script>

<template>
  <n-config-provider>
    <n-message-provider>
      <n-layout class="app-shell" has-sider>
        <n-layout-sider
          bordered
          collapse-mode="width"
          :collapsed-width="72"
          :width="236"
          class="sidebar"
        >
          <div class="brand-block">
            <div class="brand-mark">
              <BookMarked :size="26" />
            </div>
            <div>
              <div class="brand-title">佳依拆书</div>
              <div class="brand-subtitle">本地小说拆解工作台</div>
            </div>
          </div>

          <n-menu
            :value="activeView"
            :options="navOptions"
            :root-indent="20"
            :indent="18"
            @update:value="(key) => goTo(key as ViewKey)"
          />

          <div class="sidebar-footer">
            <n-alert type="info" :show-icon="false" class="storage-alert">
              <div class="storage-title">当前数据边界</div>
              <div class="storage-text">单本书正式数据保存在小说项目目录中。</div>
            </n-alert>
          </div>
        </n-layout-sider>

        <n-layout>
          <n-layout-header bordered class="topbar">
            <div>
              <div class="page-title">{{ activeTitle }}</div>
              <div class="page-context">
                当前书：{{ currentBook.title }} · {{ currentBook.readingPosition }}
              </div>
            </div>
            <div class="topbar-actions">
              <n-tag round type="success">
                <template #icon>
                  <CheckCircle2 :size="14" />
                </template>
                已保存
              </n-tag>
              <n-button tertiary>
                <template #icon>
                  <Search :size="16" />
                </template>
                搜索
              </n-button>
              <n-button type="primary" @click="showLibraryImport">
                <template #icon>
                  <Plus :size="16" />
                </template>
                导入
              </n-button>
            </div>
          </n-layout-header>

          <n-layout-content class="content-shell">
            <n-scrollbar class="content-scroll">
              <section v-if="activeView === 'library'" class="page-section">
                <div class="section-heading">
                  <div>
                    <h1>{{ libraryMode === 'list' ? '书库' : '导入小说' }}</h1>
                    <p>
                      {{ libraryMode === 'list'
                        ? '书库目录、小说列表和导入入口集中在这里。'
                        : '导入只展示小说信息、导入信息和分页章节标题。' }}
                    </p>
                  </div>
                  <n-space>
                    <n-button v-if="libraryMode === 'list'">
                      <template #icon>
                        <FolderOpen :size="16" />
                      </template>
                      更换书库
                    </n-button>
                    <n-button v-if="libraryMode === 'list'" type="primary" @click="showLibraryImport">
                      <template #icon>
                        <FileInput :size="16" />
                      </template>
                      导入 TXT
                    </n-button>
                    <n-button v-else @click="libraryMode = 'list'">返回书库</n-button>
                  </n-space>
                </div>

                <template v-if="libraryMode === 'list'">
                  <n-card :bordered="false" class="workspace-card library-toolbar">
                    <div>
                      <div class="card-kicker">当前书库目录</div>
                      <div class="library-path">E:/BookLibrary</div>
                    </div>
                    <n-input placeholder="搜索小说名" style="max-width: 260px" />
                  </n-card>

                  <n-grid :cols="3" :x-gap="16" :y-gap="16" responsive="screen">
                    <n-gi v-for="book in projectCards" :key="book.path">
                      <n-card class="project-card" :bordered="false">
                        <div class="card-kicker">{{ book.status }}</div>
                        <h2>{{ book.title }}</h2>
                        <div class="muted">作者：{{ book.author }}</div>
                        <div class="project-path">{{ book.path }}</div>
                        <n-divider />
                        <div class="metric-row">
                          <n-statistic label="大小" :value="book.size" />
                          <n-statistic label="章节" :value="book.chapters" />
                        </div>
                        <div class="card-footer">
                          <span>{{ book.updatedAt }}</span>
                          <n-button size="small" type="primary" @click="goTo('reading')">打开</n-button>
                        </div>
                      </n-card>
                    </n-gi>
                  </n-grid>

                  <n-alert type="warning" class="page-alert">
                    书库 `index.json` 只记录小说名、小说目录和大小；章节、拆解、素材读取单本小说目录的数据。
                  </n-alert>
                </template>

                <template v-else>
                  <div class="import-review-layout">
                    <n-card :bordered="false" title="小说信息" class="workspace-card">
                      <n-form label-placement="top">
                        <n-form-item label="小说名">
                          <n-input :value="importInfo.bookName" />
                        </n-form-item>
                        <n-form-item label="简介">
                          <n-input
                            type="textarea"
                            :value="importInfo.description"
                            :autosize="{ minRows: 4, maxRows: 6 }"
                          />
                        </n-form-item>
                      </n-form>
                    </n-card>

                    <n-card :bordered="false" title="导入信息" class="workspace-card">
                      <n-descriptions label-placement="left" :column="1" bordered>
                        <n-descriptions-item label="原路径">{{ importInfo.sourcePath }}</n-descriptions-item>
                        <n-descriptions-item label="书库目录">{{ importInfo.libraryPath }}</n-descriptions-item>
                        <n-descriptions-item label="入库路径">{{ importInfo.targetPath }}</n-descriptions-item>
                        <n-descriptions-item label="大小">{{ importInfo.size }}</n-descriptions-item>
                        <n-descriptions-item label="章节数">{{ importInfo.chapterCount }}</n-descriptions-item>
                      </n-descriptions>
                    </n-card>
                  </div>

                  <n-card :bordered="false" title="章节信息" class="workspace-card chapter-table-card">
                    <n-data-table
                      :columns="importChapterColumns"
                      :data="importChapters"
                      :bordered="false"
                      :pagination="false"
                    />
                    <div class="chapter-page-footer">
                      <span>共识别 {{ importInfo.chapterCount }} 章，仅展示标题用于快速检查。</span>
                      <n-pagination
                        v-model:page="importChapterPage"
                        :page-count="315"
                        :page-slot="7"
                      />
                    </div>
                  </n-card>

                  <n-space justify="end" class="card-actions">
                    <n-button @click="libraryMode = 'list'">取消</n-button>
                    <n-button>
                      <template #icon>
                        <RotateCcw :size="16" />
                      </template>
                      重新识别
                    </n-button>
                    <n-button type="primary" @click="goTo('reading')">
                      <template #icon>
                        <Save :size="16" />
                      </template>
                      确认导入
                    </n-button>
                  </n-space>
                </template>
              </section>

              <section v-else-if="activeView === 'reading'" class="reading-page">
                <aside class="chapter-panel">
                  <div class="panel-title">章节目录</div>
                  <n-scrollbar style="max-height: calc(100vh - 180px)">
                    <button
                      v-for="chapter in chapters"
                      :key="chapter.key"
                      class="chapter-item"
                      :class="{ active: selectedChapterKey === chapter.key }"
                      @click="selectedChapterKey = chapter.key"
                    >
                      <span>{{ chapter.title }}</span>
                      <small>{{ chapter.count }} 张卡片</small>
                      <n-progress
                        type="line"
                        :percentage="chapter.progress"
                        :show-indicator="false"
                        :height="4"
                      />
                    </button>
                  </n-scrollbar>
                </aside>

                <main class="reader-pane">
                  <div class="reader-toolbar">
                    <n-space align="center">
                      <n-button-group>
                        <n-button :type="readerMode === 'read' ? 'primary' : 'default'" @click="readerMode = 'read'">
                          阅读
                        </n-button>
                        <n-button :type="readerMode === 'edit' ? 'primary' : 'default'" @click="readerMode = 'edit'">
                          编辑
                        </n-button>
                      </n-button-group>
                      <n-tag type="info">第三章 雨夜码头</n-tag>
                    </n-space>
                    <n-space>
                      <n-button quaternary circle>
                        <template #icon>
                          <Copy :size="16" />
                        </template>
                      </n-button>
                      <n-button quaternary circle>
                        <template #icon>
                          <Pencil :size="16" />
                        </template>
                      </n-button>
                    </n-space>
                  </div>

                  <article v-if="readerMode === 'read'" class="novel-text">
                    <p>雨在码头上铺开，像一层被反复揉皱的灰布。旧仓库檐下的灯忽明忽暗，照见铁轨尽头慢慢停下来的货车。</p>
                    <p>
                      沈知遥站在阴影里，手指压着信封的边角。她听见身后有人喊她的名字，却没有立刻回头。
                      <span class="selected-source">那声音比三年前低了一点，像被潮气磨过。</span>
                    </p>
                    <p>渡轮鸣笛时，周围所有人都短暂地看向江面。只有她仍看着仓库的玻璃窗，看里面倒映出的那个人影。</p>
                    <div class="floating-toolbar">
                      <n-button size="small" type="primary">
                        <template #icon>
                          <Highlighter :size="14" />
                        </template>
                        创建拆解
                      </n-button>
                      <n-button size="small">标记</n-button>
                      <n-button size="small">复制</n-button>
                    </div>
                  </article>

                  <div v-else class="edit-pane">
                    <n-form label-placement="top">
                      <n-form-item label="章节标题">
                        <n-input class="edit-title-input" value="第三章 雨夜码头" />
                      </n-form-item>
                      <n-form-item label="章节正文">
                        <n-input
                          type="textarea"
                          value="雨在码头上铺开，像一层被反复揉皱的灰布。旧仓库檐下的灯忽明忽暗，照见铁轨尽头慢慢停下来的货车。

沈知遥站在阴影里，手指压着信封的边角。她听见身后有人喊她的名字，却没有立刻回头。那声音比三年前低了一点，像被潮气磨过。

渡轮鸣笛时，周围所有人都短暂地看向江面。只有她仍看着仓库的玻璃窗，看里面倒映出的那个人影。"
                          :autosize="{ minRows: 15, maxRows: 24 }"
                        />
                      </n-form-item>
                    </n-form>
                    <n-alert type="info" :show-icon="false">
                      章节标题和正文修改放在阅读页编辑模式中，导入阶段只展示识别结果。
                    </n-alert>
                    <n-space justify="end" class="card-actions">
                      <n-button>放弃修改</n-button>
                      <n-button type="primary">
                        <template #icon>
                          <Save :size="16" />
                        </template>
                        保存章节
                      </n-button>
                    </n-space>
                  </div>
                </main>

                <aside class="workbench-panel">
                  <div class="panel-title-row">
                    <div>
                      <div class="panel-title">拆解工作台</div>
                      <div class="panel-subtitle">当前选文 · 可创建多张拆解</div>
                    </div>
                    <n-badge :value="3" type="success" />
                  </div>
                  <n-alert type="default" :show-icon="false" class="source-box">
                    那声音比三年前低了一点，像被潮气磨过。
                  </n-alert>
                  <n-form label-placement="top">
                    <n-form-item label="拆解类型">
                      <n-select
                        v-model:value="selectedDeconstructionType"
                        :options="[
                          { label: '场景拆解', value: 'scene' },
                          { label: '对白拆解', value: 'dialogue' },
                          { label: '人物弧光', value: 'character' },
                        ]"
                      />
                    </n-form-item>
                    <n-form-item label="场景目的">
                      <n-input value="让重逢带有迟疑和对峙感" />
                    </n-form-item>
                    <n-form-item label="情绪变化">
                      <n-input
                        type="textarea"
                        value="从确认对方出现，到压住回应，再到被环境迫使面对。"
                        :autosize="{ minRows: 4, maxRows: 5 }"
                      />
                    </n-form-item>
                    <n-form-item label="标签">
                      <n-select
                        multiple
                        :value="['氛围', '重逢']"
                        :options="[
                          { label: '氛围', value: '氛围' },
                          { label: '重逢', value: '重逢' },
                          { label: '悬念', value: '悬念' },
                        ]"
                      />
                    </n-form-item>
                  </n-form>
                  <n-space justify="end">
                    <n-button>继续建卡</n-button>
                    <n-button type="primary">保存拆解</n-button>
                  </n-space>
                </aside>
              </section>

              <section v-else-if="activeView === 'deconstructions'" class="page-section">
                <div class="section-heading">
                  <div>
                    <h1>拆解卡片</h1>
                    <p>集中管理当前书的拆解结果，并保留回到原文的位置。</p>
                  </div>
                  <n-space>
                    <n-input placeholder="搜索拆解、章节、标签" />
                    <n-button>
                      <template #icon>
                        <SlidersHorizontal :size="16" />
                      </template>
                      筛选
                    </n-button>
                  </n-space>
                </div>

                <div class="list-detail-layout">
                  <div class="stack-list">
                    <n-card v-for="card in deconstructionCards" :key="card.title" :bordered="false" class="item-card">
                      <div class="item-card-head">
                        <div>
                          <n-tag size="small" type="info">{{ card.type }}</n-tag>
                          <h2>{{ card.title }}</h2>
                        </div>
                        <n-button text type="primary" @click="goTo('reading')">回到原文</n-button>
                      </div>
                      <p>{{ card.summary }}</p>
                      <div class="source-line">
                        <MapPinned :size="15" />
                        {{ card.chapter }} · {{ card.source }}
                      </div>
                      <n-space size="small">
                        <n-tag v-for="tag in card.tags" :key="tag" size="small">{{ tag }}</n-tag>
                      </n-space>
                    </n-card>
                  </div>

                  <n-card :bordered="false" title="来源追溯" class="detail-panel">
                    <n-timeline>
                      <n-timeline-item type="success" title="书籍" content="雾港来信" />
                      <n-timeline-item type="info" title="章节" content="第三章 雨夜码头" />
                      <n-timeline-item type="warning" title="原文位置" content="第 18 段，字符 842-895" />
                      <n-timeline-item title="关联素材" content="用天气压低重逢情绪" />
                    </n-timeline>
                  </n-card>
                </div>
              </section>

              <section v-else-if="activeView === 'materials'" class="page-section">
                <div class="section-heading">
                  <div>
                    <h1>素材库</h1>
                    <p>从拆解中沉淀可复用写作知识，并保留来源链路。</p>
                  </div>
                  <n-space>
                    <n-select
                      value="all"
                      style="width: 140px"
                      :options="[
                        { label: '全部分类', value: 'all' },
                        { label: '场景写法', value: 'scene' },
                        { label: '对白技巧', value: 'dialogue' },
                      ]"
                    />
                    <n-input placeholder="搜索素材" />
                  </n-space>
                </div>

                <n-grid :cols="3" :x-gap="16" :y-gap="16" responsive="screen">
                  <n-gi v-for="material in materials" :key="material.title">
                    <n-card :bordered="false" class="material-card">
                      <div class="item-card-head">
                        <n-tag type="success" size="small">{{ material.category }}</n-tag>
                        <n-tooltip>
                          <template #trigger>
                            <n-button quaternary circle>
                              <template #icon>
                                <Bookmark :size="16" :fill="material.favorite ? 'currentColor' : 'none'" />
                              </template>
                            </n-button>
                          </template>
                          收藏状态
                        </n-tooltip>
                      </div>
                      <h2>{{ material.title }}</h2>
                      <p>{{ material.text }}</p>
                      <div class="source-line">
                        <MapPinned :size="15" />
                        {{ material.source }}
                      </div>
                      <n-space size="small">
                        <n-tag v-for="tag in material.tags" :key="tag" size="small">{{ tag }}</n-tag>
                      </n-space>
                    </n-card>
                  </n-gi>
                </n-grid>
              </section>

              <section v-else-if="activeView === 'plan'" class="page-section">
                <div class="section-heading">
                  <div>
                    <h1>拆解方案</h1>
                    <p>当前书独立配置拆解维度、类型和字段。</p>
                  </div>
                  <n-space>
                    <n-button>从模板创建</n-button>
                    <n-button type="primary">另存为模板</n-button>
                  </n-space>
                </div>

                <div class="plan-layout">
                  <n-card :bordered="false" title="维度与类型" class="dimension-card">
                    <n-tabs type="line" animated>
                      <n-tab-pane name="scene" tab="场景">
                        <n-list>
                          <n-list-item>
                            <n-thing title="场景拆解" description="场景目的、冲突、情绪变化、环境调度" />
                          </n-list-item>
                          <n-list-item>
                            <n-thing title="氛围拆解" description="感官描写、节奏、意象组合" />
                          </n-list-item>
                        </n-list>
                      </n-tab-pane>
                      <n-tab-pane name="character" tab="人物">
                        <n-list>
                          <n-list-item>
                            <n-thing title="人物弧光" description="欲望、阻力、选择、变化" />
                          </n-list-item>
                          <n-list-item>
                            <n-thing title="对白拆解" description="表层台词、潜台词、关系变化" />
                          </n-list-item>
                        </n-list>
                      </n-tab-pane>
                      <n-tab-pane name="structure" tab="结构">
                        <n-list>
                          <n-list-item>
                            <n-thing title="伏笔回收" description="埋设位置、回收章节、情绪收益" />
                          </n-list-item>
                        </n-list>
                      </n-tab-pane>
                    </n-tabs>
                  </n-card>

                  <n-card :bordered="false" title="字段配置" class="fields-card">
                    <n-alert type="warning" class="page-alert">
                      字段调整需要保留旧卡片数据，并在真实实现中提供迁移提示。
                    </n-alert>
                    <n-data-table :columns="planColumns" :data="planFields" :bordered="false" />
                  </n-card>
                </div>
              </section>

              <section v-else-if="activeView === 'analysis'" class="page-section">
                <div class="section-heading">
                  <div>
                    <h1>基础分析</h1>
                    <p>第一阶段只展示当前书的轻量统计。</p>
                  </div>
                  <n-tag type="default">不含跨书比较</n-tag>
                </div>

                <n-grid :cols="4" :x-gap="16" :y-gap="16" responsive="screen">
                  <n-gi>
                    <n-card :bordered="false"><n-statistic label="章节" :value="currentBook.chapters" /></n-card>
                  </n-gi>
                  <n-gi>
                    <n-card :bordered="false"><n-statistic label="拆解卡片" :value="currentBook.deconstructions" /></n-card>
                  </n-gi>
                  <n-gi>
                    <n-card :bordered="false"><n-statistic label="素材" :value="currentBook.materials" /></n-card>
                  </n-gi>
                  <n-gi>
                    <n-card :bordered="false"><n-statistic label="已拆章节" :value="19" /></n-card>
                  </n-gi>
                </n-grid>

                <div class="analysis-grid">
                  <n-card :bordered="false" title="按拆解类型">
                    <div v-for="row in analysisRows" :key="row.type" class="progress-row">
                      <div>
                        <span>{{ row.type }}</span>
                        <strong>{{ row.count }}</strong>
                      </div>
                      <n-progress type="line" :percentage="row.percent" :height="8" />
                    </div>
                  </n-card>

                  <n-card :bordered="false" title="章节拆解分布">
                    <div v-for="chapter in chapters" :key="chapter.key" class="chapter-stat-row">
                      <span>{{ chapter.title }}</span>
                      <n-progress type="line" :percentage="Math.min(chapter.count * 8, 100)" :height="8" />
                      <strong>{{ chapter.count }}</strong>
                    </div>
                  </n-card>
                </div>
              </section>

              <section v-else class="page-section">
                <div class="section-heading">
                  <div>
                    <h1>全局设置</h1>
                    <p>只管理应用级偏好，不保存某一本书的正式业务数据。</p>
                  </div>
                </div>

                <div class="settings-grid">
                  <n-card :bordered="false" title="界面偏好">
                    <n-form label-placement="left" label-width="110">
                      <n-form-item label="阅读字号">
                        <n-select
                          value="18"
                          :options="[
                            { label: '16px', value: '16' },
                            { label: '18px', value: '18' },
                            { label: '20px', value: '20' },
                          ]"
                        />
                      </n-form-item>
                      <n-form-item label="紧凑布局">
                        <n-switch />
                      </n-form-item>
                      <n-form-item label="最近项目数">
                        <n-select
                          value="10"
                          :options="[
                            { label: '5', value: '5' },
                            { label: '10', value: '10' },
                            { label: '20', value: '20' },
                          ]"
                        />
                      </n-form-item>
                    </n-form>
                  </n-card>

                  <n-card :bordered="false" title="本地数据">
                    <n-list>
                      <n-list-item>
                        <n-thing title="项目目录" description="单本书正式数据随项目目录迁移" />
                      </n-list-item>
                      <n-list-item>
                        <n-thing title="缓存索引" description=".cache/index.sqlite 可删除并重建" />
                      </n-list-item>
                      <n-list-item>
                        <n-space>
                          <n-button>
                            <template #icon>
                              <Database :size="16" />
                            </template>
                            清理缓存
                          </n-button>
                          <n-button>
                            <template #icon>
                              <CircleAlert :size="16" />
                            </template>
                            检查项目
                          </n-button>
                        </n-space>
                      </n-list-item>
                    </n-list>
                  </n-card>
                </div>
              </section>
            </n-scrollbar>
          </n-layout-content>
        </n-layout>
      </n-layout>
    </n-message-provider>
  </n-config-provider>
</template>
