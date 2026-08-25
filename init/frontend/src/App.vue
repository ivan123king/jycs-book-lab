<script setup lang="ts">
// App.vue 负责第一阶段应用壳层路由和当前书摘要，不承载具体页面业务实现。
import { computed, onMounted, ref } from 'vue'
import { NButton, NConfigProvider, NLayout, NLayoutContent, NMessageProvider, NScrollbar } from 'naive-ui'
import { Plus, RefreshCw } from 'lucide-vue-next'
import { OpenLastReadingSession } from '@wails/go/main/App'
import AppSidebar from '@/components/app/AppSidebar.vue'
import AppTopbar from '@/components/app/AppTopbar.vue'
import {
  defaultCurrentBook,
  demoAnalysisRows,
  demoChapters,
} from '@/data/workbenchFixtures'
import AnalysisPage from '@/modules/analysis/AnalysisPage.vue'
import DeconstructionCardsPage from '@/modules/deconstructions/DeconstructionCardsPage.vue'
import LibraryPage from '@/modules/library/LibraryPage.vue'
import LibraryMaterialsPage from '@/modules/materials/LibraryMaterialsPage.vue'
import MaterialsPage from '@/modules/materials/MaterialsPage.vue'
import DeconstructionPlanPage from '@/modules/plan/DeconstructionPlanPage.vue'
import ReadingWorkbenchPage from '@/modules/reading/ReadingWorkbenchPage.vue'
import SettingsPage from '@/modules/settings/SettingsPage.vue'
import type { CurrentBookSummary } from '@/types/currentBook'
import type { BookProject } from '@/types/library'
import type { ViewKey } from '@/types/navigation'
import { formatDateTime } from '@/utils/format'

const activeView = ref<ViewKey>('library')
const libraryMode = ref<'list' | 'import'>('list')
const currentBook = ref<CurrentBookSummary>({ ...defaultCurrentBook })
const currentProject = ref<BookProject | null>(null)
const readingSourceFocus = ref<ReadingSourceFocus | null>(null)
const materialsPageRef = ref<InstanceType<typeof MaterialsPage> | null>(null)
const libraryMaterialsPageRef = ref<InstanceType<typeof LibraryMaterialsPage> | null>(null)

type ApplyCurrentBookOptions = {
  navigateToReading?: boolean
  readingPosition?: string
}

type ReadingSourceFocus = {
  chapterId: string
  sourceTextId: string
  startOffset: number
  requestKey: number
}

const activeTitle = computed(() => {
  const titles: Record<ViewKey, string> = {
    library: '书库',
    libraryMaterials: '书库素材',
    reading: '阅读与拆解',
    deconstructions: '写作手法拆解',
    materials: '素材库',
    plan: '拆解方案',
    analysis: '基础分析',
    settings: '全局设置',
  }
  return activeView.value === 'library' && libraryMode.value === 'import' ? '书库 / 导入小说' : titles[activeView.value]
})

function goTo(view: ViewKey) {
  activeView.value = view
}

/**
 * 用打开的小说项目刷新壳层摘要；章节正文和拆解数据仍由具体页面按项目目录加载。
 */
function applyCurrentBook(project: BookProject, options: ApplyCurrentBookOptions = {}) {
  currentProject.value = project
  currentBook.value = {
    title: project.book.name,
    author: '未知作者',
    path: project.book.bookPath,
    description: project.book.description,
    chapters: project.chapters.length,
    deconstructions: 0,
    materials: 0,
    updatedAt: formatDateTime(project.book.updatedAt),
    readingPosition:
      options.readingPosition ?? (project.chapters[0]?.title ? `${project.chapters[0].title} · 0%` : '尚未开始阅读'),
  }
  if (options.navigateToReading ?? true) {
    activeView.value = 'reading'
  }
}

/**
 * 从应用级最近打开入口恢复当前书，只刷新壳层上下文，不强制跳转当前页面。
 */
async function restoreLastOpenedBook() {
  try {
    const session = await OpenLastReadingSession()
    if (!session) {
      return
    }
    applyCurrentBook(session.project, {
      navigateToReading: false,
      readingPosition: buildReadingPositionSummary(session),
    })
  } catch (error) {
    console.warn('恢复最近打开小说失败', error)
  }
}

/**
 * 用后端恢复出的阅读状态生成顶部当前书摘要，供所有模块共享。
 */
function buildReadingPositionSummary(session: Awaited<ReturnType<typeof OpenLastReadingSession>>) {
  if (!session) {
    return '尚未开始阅读'
  }
  const chapterTitle = session.currentChapter?.title || session.project.chapters[0]?.title
  const progressPercent = Math.min(100, Math.max(0, Math.round(session.state?.readingProgressPercent ?? 0)))
  return chapterTitle ? `${chapterTitle} · ${progressPercent}%` : '尚未开始阅读'
}

function updateReadingPosition(summary: string) {
  currentBook.value = {
    ...currentBook.value,
    readingPosition: summary,
  }
}

/**
 * 从拆解卡片详情回到来源章节，requestKey 用来保证重复点击同一来源也会触发阅读页定位。
 */
function openReadingSource(payload?: { chapterId: string; sourceTextId: string; startOffset: number }) {
  if (payload) {
    readingSourceFocus.value = {
      ...payload,
      requestKey: Date.now(),
    }
  }
  activeView.value = 'reading'
}

function clearCurrentBookIfRemoved(bookPath: string) {
  if (currentProject.value?.book.bookPath !== bookPath) {
    return
  }
  currentProject.value = null
  currentBook.value = { ...defaultCurrentBook }
}

function refreshMaterialsPage() {
  materialsPageRef.value?.refresh()
}

function refreshLibraryMaterialsPage() {
  libraryMaterialsPageRef.value?.refresh()
}

function openLibraryMaterialCreate() {
  libraryMaterialsPageRef.value?.openCreate()
}

onMounted(restoreLastOpenedBook)
</script>

<template>
  <n-config-provider>
    <n-message-provider>
      <n-layout class="app-shell" has-sider>
        <AppSidebar :active-view="activeView" @navigate="goTo" />

        <n-layout>
          <AppTopbar
            v-if="activeView !== 'library'"
            :title="activeTitle"
            :current-book="currentBook"
            :show-book-context="activeView !== 'libraryMaterials'"
          >
            <template v-if="activeView === 'libraryMaterials'" #actions>
              <n-button type="primary" secondary @click="openLibraryMaterialCreate">
                <template #icon>
                  <Plus :size="16" />
                </template>
                新增素材
              </n-button>
              <n-button secondary @click="refreshLibraryMaterialsPage">
                <template #icon>
                  <RefreshCw :size="16" />
                </template>
                刷新
              </n-button>
            </template>
            <template v-else-if="activeView === 'materials'" #actions>
              <n-button secondary @click="refreshMaterialsPage">
                <template #icon>
                  <RefreshCw :size="16" />
                </template>
                刷新
              </n-button>
            </template>
          </AppTopbar>

          <n-layout-content :class="['content-shell', { 'content-shell-full': activeView === 'library' }]">
            <n-scrollbar :class="['content-scroll', { 'content-scroll-full': activeView === 'library' }]">
              <LibraryPage
                v-if="activeView === 'library'"
                v-model:mode="libraryMode"
                @open-book="applyCurrentBook"
                @book-removed="clearCurrentBookIfRemoved"
              />
              <LibraryMaterialsPage
                v-else-if="activeView === 'libraryMaterials'"
                ref="libraryMaterialsPageRef"
                @open-book-source="applyCurrentBook"
              />
              <ReadingWorkbenchPage
                v-else-if="activeView === 'reading'"
                :project="currentProject"
                :source-focus="readingSourceFocus"
                @reading-updated="updateReadingPosition"
                @project-restored="applyCurrentBook"
              />
              <DeconstructionCardsPage
                v-else-if="activeView === 'deconstructions'"
                :project="currentProject"
                @back-to-source="openReadingSource"
              />
              <MaterialsPage
                v-else-if="activeView === 'materials'"
                ref="materialsPageRef"
                :project="currentProject"
                @back-to-source="openReadingSource"
              />
              <DeconstructionPlanPage
                v-else-if="activeView === 'plan'"
                :project="currentProject"
              />
              <AnalysisPage
                v-else-if="activeView === 'analysis'"
                :current-book="currentBook"
                :rows="demoAnalysisRows"
                :chapters="demoChapters"
              />
              <SettingsPage v-else />
            </n-scrollbar>
          </n-layout-content>
        </n-layout>
      </n-layout>
    </n-message-provider>
  </n-config-provider>
</template>
