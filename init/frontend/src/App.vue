<script setup lang="ts">
// App.vue 负责第一阶段应用壳层路由和当前书摘要，不承载具体页面业务实现。
import { computed, ref } from 'vue'
import { NConfigProvider, NLayout, NLayoutContent, NMessageProvider, NScrollbar } from 'naive-ui'
import AppSidebar from '@/components/app/AppSidebar.vue'
import AppTopbar from '@/components/app/AppTopbar.vue'
import {
  defaultCurrentBook,
  demoAnalysisRows,
  demoChapters,
  demoDeconstructionCards,
  demoMaterials,
  demoPlanFields,
} from '@/data/workbenchFixtures'
import AnalysisPage from '@/modules/analysis/AnalysisPage.vue'
import DeconstructionCardsPage from '@/modules/deconstructions/DeconstructionCardsPage.vue'
import LibraryPage from '@/modules/library/LibraryPage.vue'
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

function goTo(view: ViewKey) {
  activeView.value = view
}

/**
 * 用打开的小说项目刷新壳层摘要；章节正文和拆解数据仍由具体页面按项目目录加载。
 */
function applyCurrentBook(project: BookProject) {
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
    readingPosition: project.chapters[0]?.title ? `${project.chapters[0].title} · 0%` : '尚未开始阅读',
  }
  activeView.value = 'reading'
}

function updateReadingPosition(summary: string) {
  currentBook.value = {
    ...currentBook.value,
    readingPosition: summary,
  }
}
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
          />

          <n-layout-content :class="['content-shell', { 'content-shell-full': activeView === 'library' }]">
            <n-scrollbar :class="['content-scroll', { 'content-scroll-full': activeView === 'library' }]">
              <LibraryPage
                v-if="activeView === 'library'"
                v-model:mode="libraryMode"
                @open-book="applyCurrentBook"
              />
              <ReadingWorkbenchPage
                v-else-if="activeView === 'reading'"
                :project="currentProject"
                @reading-updated="updateReadingPosition"
              />
              <DeconstructionCardsPage
                v-else-if="activeView === 'deconstructions'"
                :cards="demoDeconstructionCards"
                @back-to-source="goTo('reading')"
              />
              <MaterialsPage
                v-else-if="activeView === 'materials'"
                :materials="demoMaterials"
              />
              <DeconstructionPlanPage
                v-else-if="activeView === 'plan'"
                :fields="demoPlanFields"
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
