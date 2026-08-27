<script setup lang="ts">
// App.vue 负责第一阶段应用壳层路由、Wails 原生菜单事件和当前书摘要，不承载具体页面业务实现。
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { NButton, NConfigProvider, NLayout, NLayoutContent, NMessageProvider, NModal, NScrollbar } from 'naive-ui'
import { Mail, Plus, RefreshCw, X } from 'lucide-vue-next'
import { GetApplicationInfo, OpenLastReadingSession } from '@wails/go/main/App'
import { EventsOn } from '@wails/runtime/runtime'
import aboutBgUrl from '@/assets/images/about-bg.jpg'
import AppSidebar from '@/components/app/AppSidebar.vue'
import AppTopbar from '@/components/app/AppTopbar.vue'
import { defaultCurrentBook } from '@/data/workbenchFixtures'
import AnalysisPage from '@/modules/analysis/AnalysisPage.vue'
import DeconstructionCardsPage from '@/modules/deconstructions/DeconstructionCardsPage.vue'
import StoryPlotPreviewPanel from '@/modules/deconstructions/StoryPlotPreviewPanel.vue'
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
const deconstructionCardsPageRef = ref<InstanceType<typeof DeconstructionCardsPage> | null>(null)
const storyPlotPageRef = ref<InstanceType<typeof StoryPlotPreviewPanel> | null>(null)
const materialsPageRef = ref<InstanceType<typeof MaterialsPage> | null>(null)
const libraryMaterialsPageRef = ref<InstanceType<typeof LibraryMaterialsPage> | null>(null)
const aboutDialogOpen = ref(false)
const applicationInfo = ref<ApplicationInfo>({
  name: 'jiayi-book-lab',
  displayName: '佳依拆书',
  version: 'v1.0',
  description: '面向小说作者、写作爱好者和文学研究者的本地小说拆解与写作学习工具，帮助用户从阅读原文、选文拆解到沉淀个人写作素材。',
  authors: ['GPT5.5', '刘伟'],
  contactEmail: 'lw18751836671@163.com',
})
let stopAboutMenuListener: (() => void) | null = null

type ApplyCurrentBookOptions = {
  navigateToReading?: boolean
  readingPosition?: string
}

type ApplicationInfo = {
  name: string
  displayName: string
  version: string
  description: string
  authors: string[]
  contactEmail: string
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
    deconstructions: '写作手法',
    storyPlot: '剧情',
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

function refreshDeconstructionCardsPage() {
  deconstructionCardsPageRef.value?.refresh()
}

function refreshStoryPlotPage() {
  storyPlotPageRef.value?.refresh()
}

function openStoryPlotCreate() {
  storyPlotPageRef.value?.openCreate()
}

function refreshLibraryMaterialsPage() {
  libraryMaterialsPageRef.value?.refresh()
}

function openLibraryMaterialCreate() {
  libraryMaterialsPageRef.value?.openCreate()
}

async function loadApplicationInfo() {
  try {
    applicationInfo.value = await GetApplicationInfo()
  } catch (error) {
    console.warn('读取应用信息失败', error)
  }
}

function openAboutDialog() {
  aboutDialogOpen.value = true
}

onMounted(() => {
  restoreLastOpenedBook()
  loadApplicationInfo()
  stopAboutMenuListener = EventsOn('app:open-about-dialog', openAboutDialog)
})

onUnmounted(() => {
  stopAboutMenuListener?.()
})
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
            <template v-else-if="activeView === 'deconstructions'" #actions>
              <n-button secondary @click="refreshDeconstructionCardsPage">
                <template #icon>
                  <RefreshCw :size="16" />
                </template>
                刷新
              </n-button>
            </template>
            <template v-else-if="activeView === 'storyPlot'" #actions>
              <n-button type="primary" secondary @click="openStoryPlotCreate">
                <template #icon>
                  <Plus :size="16" />
                </template>
                新建剧情
              </n-button>
              <n-button secondary @click="refreshStoryPlotPage">
                <template #icon>
                  <RefreshCw :size="16" />
                </template>
                刷新模板
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
                ref="deconstructionCardsPageRef"
                :project="currentProject"
                @back-to-source="openReadingSource"
              />
              <StoryPlotPreviewPanel
                v-else-if="activeView === 'storyPlot'"
                ref="storyPlotPageRef"
                :project="currentProject"
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
                :project="currentProject"
              />
              <SettingsPage v-else />
            </n-scrollbar>
          </n-layout-content>
        </n-layout>
      </n-layout>

      <n-modal v-model:show="aboutDialogOpen" :bordered="false" :show-icon="false">
        <section
          class="about-panel"
          :style="{
            backgroundImage: `linear-gradient(120deg, rgba(16, 20, 28, 0.88), rgba(16, 20, 28, 0.58)), url(${aboutBgUrl})`,
          }"
          aria-label="关于佳依拆书"
        >
          <n-button class="about-close" circle tertiary type="default" aria-label="关闭关于窗口" @click="aboutDialogOpen = false">
            <template #icon>
              <X :size="18" />
            </template>
          </n-button>

          <div class="about-content">
            <div class="about-kicker">本地小说拆解与写作学习工具</div>
            <h2>{{ applicationInfo.displayName }}</h2>
            <div class="about-version">{{ applicationInfo.version }}</div>
            <p>{{ applicationInfo.description }}</p>

            <dl class="about-meta">
              <div>
                <dt>作者</dt>
                <dd>{{ applicationInfo.authors.join('、') }}</dd>
              </div>
              <div>
                <dt>联系方式</dt>
                <dd>
                  <Mail :size="15" />
                  <span>{{ applicationInfo.contactEmail }}</span>
                </dd>
              </div>
            </dl>
          </div>
        </section>
      </n-modal>
    </n-message-provider>
  </n-config-provider>
</template>

<style scoped>
.about-panel {
  position: relative;
  width: min(680px, calc(100vw - 48px));
  min-height: 420px;
  overflow: hidden;
  border: 1px solid rgba(255, 255, 255, 0.18);
  border-radius: 8px;
  background-position: center;
  background-size: cover;
  box-shadow: 0 24px 80px rgba(0, 0, 0, 0.38);
  color: #f8fafc;
}

.about-close {
  position: absolute;
  top: 16px;
  right: 16px;
  z-index: 2;
  color: #f8fafc;
  background: rgba(15, 23, 42, 0.35);
  backdrop-filter: blur(10px);
}

.about-content {
  display: flex;
  min-height: 420px;
  width: min(430px, 100%);
  flex-direction: column;
  justify-content: flex-end;
  padding: 48px;
}

.about-kicker {
  margin-bottom: 12px;
  color: rgba(226, 232, 240, 0.82);
  font-size: 13px;
}

.about-content h2 {
  margin: 0;
  color: #ffffff;
  font-size: 36px;
  font-weight: 700;
  line-height: 1.15;
}

.about-version {
  margin-top: 10px;
  color: rgba(226, 232, 240, 0.82);
  font-size: 14px;
}

.about-content p {
  margin: 22px 0 0;
  color: rgba(248, 250, 252, 0.9);
  font-size: 15px;
  line-height: 1.8;
}

.about-meta {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 16px;
  margin: 28px 0 0;
}

.about-meta div {
  min-width: 0;
}

.about-meta dt {
  margin-bottom: 6px;
  color: rgba(226, 232, 240, 0.68);
  font-size: 12px;
}

.about-meta dd {
  display: flex;
  min-width: 0;
  align-items: center;
  gap: 7px;
  margin: 0;
  color: #ffffff;
  font-size: 14px;
  overflow-wrap: anywhere;
}

@media (max-width: 640px) {
  .about-panel {
    width: calc(100vw - 28px);
    min-height: 480px;
  }

  .about-content {
    min-height: 480px;
    padding: 34px 26px;
  }

  .about-content h2 {
    font-size: 30px;
  }

  .about-meta {
    grid-template-columns: 1fr;
  }
}
</style>
