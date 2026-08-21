<script setup lang="ts">
// ReadingWorkbenchPage.vue 负责当前书阅读、正文编辑、选文来源捕获和方案驱动拆解创建。
import { computed, nextTick, onMounted, ref, watch } from 'vue'
import {
  NAlert,
  NButton,
  NButtonGroup,
  NDrawer,
  NDrawerContent,
  NForm,
  NFormItem,
  NInput,
  NModal,
  NScrollbar,
  NSelect,
  NSpace,
  NSpin,
  NTag,
} from 'naive-ui'
import { BookOpen, ChevronLeft, ChevronRight, Copy, Highlighter, List, RotateCcw, Save } from 'lucide-vue-next'
import {
  CreateDeconstructionCard,
  GetChapterContent,
  OpenReadingSession,
  SaveChapterContent,
  SaveReadingPosition,
} from '@wails/go/main/App'
import type { library } from '@wails/go/models'
import type { BookProject } from '@/types/library'
import { getErrorMessage } from '@/utils/format'

const props = defineProps<{
  project: BookProject | null
}>()

const emit = defineEmits<{
  readingUpdated: [summary: string]
}>()

type ReaderMode = 'read' | 'edit'
type CapturedSelection = {
  chapterId: string
  chapterOrder: number
  startOffset: number
  endOffset: number
  selectedText: string
}

const pageSize = 3600
const catalogPageSize = 30
const readerMode = ref<ReaderMode>('read')
const session = ref<library.ReadingSession | null>(null)
const currentChapter = ref<library.BookChapter | null>(null)
const chapterContent = ref('')
const currentPageStart = ref(0)
const loading = ref(false)
const pageLoading = ref(false)
const savingPosition = ref(false)
const readingError = ref('')
const readingNotice = ref('')
const catalogOpen = ref(false)
const catalogSearch = ref('')
const catalogPage = ref(1)
const selectedSource = ref<CapturedSelection | null>(null)
const deconstructionOpen = ref(false)
const selectedPlanTypeId = ref('')
const fieldValues = ref<Record<string, string>>({})
const tagInput = ref('')
const savingCard = ref(false)
const cardError = ref('')
const editTitle = ref('')
const editContent = ref('')
const editSaving = ref(false)
const editNotice = ref('')
const editError = ref('')

const chapters = computed(() => session.value?.project.chapters ?? [])

const currentPageContent = computed(() => {
  if (!chapterContent.value) {
    return ''
  }
  return chapterContent.value.slice(currentPageStart.value, currentPageStart.value + pageSize)
})

const pageCount = computed(() => Math.max(1, Math.ceil(chapterContent.value.length / pageSize)))

const currentPageNumber = computed(() => Math.floor(currentPageStart.value / pageSize) + 1)

const readingProgressText = computed(() => {
  if (!currentChapter.value) {
    return '未打开章节'
  }
  return `${currentChapter.value.title} · 第 ${currentPageNumber.value}/${pageCount.value} 页`
})

const selectedPlanType = computed(() => {
  const types = session.value?.plan.types ?? []
  return types.find((item) => item.id === selectedPlanTypeId.value) ?? types[0] ?? null
})

const visiblePlanFields = computed(() => (selectedPlanType.value?.fields ?? []).filter((field) => field.visible))

const planTypeOptions = computed(() =>
  (session.value?.plan.types ?? []).map((type) => ({
    label: `${type.dimension} / ${type.name}`,
    value: type.id,
  })),
)

const sourceTextsById = computed(() => {
  const map = new Map<string, library.SourceText>()
  for (const source of session.value?.sourceTexts.items ?? []) {
    map.set(source.id, source)
  }
  return map
})

const deconstructionCountByChapter = computed(() => {
  const counts = new Map<string, number>()
  for (const card of session.value?.deconstructionCards.items ?? []) {
    const source = sourceTextsById.value.get(card.sourceTextId)
    if (!source) {
      continue
    }
    counts.set(source.chapterId, (counts.get(source.chapterId) ?? 0) + 1)
  }
  return counts
})

const filteredCatalogChapters = computed(() => {
  const keyword = catalogSearch.value.trim().toLowerCase()
  if (!keyword) {
    return chapters.value
  }
  return chapters.value.filter((chapter) => chapter.title.toLowerCase().includes(keyword))
})

const catalogPageCount = computed(() => Math.max(1, Math.ceil(filteredCatalogChapters.value.length / catalogPageSize)))

const catalogPageChapters = computed(() => {
  const page = Math.min(catalogPage.value, catalogPageCount.value)
  const start = (page - 1) * catalogPageSize
  return filteredCatalogChapters.value.slice(start, start + catalogPageSize)
})

const editDirty = computed(() => {
  return currentChapter.value ? editTitle.value !== currentChapter.value.title || editContent.value !== chapterContent.value : false
})

/**
 * 打开当前书阅读会补齐旧项目缺失的阅读数据文件，并恢复上次阅读位置。
 */
async function loadReadingSession() {
  if (!props.project?.book.bookPath) {
    session.value = null
    return
  }
  loading.value = true
  readingError.value = ''
  readingNotice.value = ''
  try {
    const loaded = await OpenReadingSession(props.project.book.bookPath)
    session.value = loaded
    currentChapter.value = loaded.currentChapter
    chapterContent.value = loaded.chapterContent
    currentPageStart.value = clampPageStart(loaded.state.lastPosition.charOffset)
    selectedPlanTypeId.value = loaded.plan.types[0]?.id ?? ''
    syncEditor()
    emitReadingSummary()
  } catch (error) {
    readingError.value = getErrorMessage(error)
  } finally {
    loading.value = false
  }
}

/**
 * 按章节加载正文并立即保存阅读位置，目录跳转和跨章节翻页共用这条路径。
 */
async function loadChapter(chapterId: string, charOffset = 0) {
  if (!props.project?.book.bookPath) {
    return
  }
  pageLoading.value = true
  readingError.value = ''
  readingNotice.value = ''
  try {
    const result = await GetChapterContent({ bookPath: props.project.book.bookPath, chapterId })
    currentChapter.value = result.chapter
    chapterContent.value = result.content
    currentPageStart.value = clampPageStart(charOffset)
    syncEditor()
    clearSelection()
    await persistReadingPosition()
    emitReadingSummary()
  } catch (error) {
    readingError.value = getErrorMessage(error)
  } finally {
    pageLoading.value = false
  }
}

async function goPreviousPage() {
  if (!currentChapter.value) {
    return
  }
  if (currentPageStart.value > 0) {
    currentPageStart.value = Math.max(0, currentPageStart.value - pageSize)
    clearSelection()
    await persistReadingPosition()
    emitReadingSummary()
    return
  }
  const index = chapters.value.findIndex((chapter) => chapter.id === currentChapter.value?.id)
  const previous = chapters.value[index - 1]
  if (previous) {
    await loadChapter(previous.id, Number.MAX_SAFE_INTEGER)
  }
}

async function goNextPage() {
  if (!currentChapter.value) {
    return
  }
  if (currentPageStart.value + pageSize < chapterContent.value.length) {
    currentPageStart.value += pageSize
    clearSelection()
    await persistReadingPosition()
    emitReadingSummary()
    return
  }
  const index = chapters.value.findIndex((chapter) => chapter.id === currentChapter.value?.id)
  const next = chapters.value[index + 1]
  if (next) {
    await loadChapter(next.id, 0)
  }
}

async function persistReadingPosition() {
  if (!props.project?.book.bookPath || !currentChapter.value) {
    return
  }
  savingPosition.value = true
  try {
    const state = await SaveReadingPosition({
      bookPath: props.project.book.bookPath,
      chapterId: currentChapter.value.id,
      charOffset: currentPageStart.value,
      paragraphIndex: estimateParagraphIndex(chapterContent.value, currentPageStart.value),
      renderedPageNumber: currentPageNumber.value,
      lastMode: readerMode.value,
    })
    if (session.value) {
      session.value.state = state
    }
  } catch (error) {
    readingError.value = getErrorMessage(error)
  } finally {
    savingPosition.value = false
  }
}

function openCatalog() {
  if (currentChapter.value) {
    catalogPage.value = Math.floor((currentChapter.value.order - 1) / catalogPageSize) + 1
  }
  catalogOpen.value = true
}

async function jumpToChapter(chapter: library.BookChapter) {
  catalogOpen.value = false
  await loadChapter(chapter.id, 0)
}

function captureSelection() {
  if (readerMode.value !== 'read' || !currentChapter.value) {
    return
  }
  const selection = window.getSelection()
  const selectedText = selection?.toString().trim() ?? ''
  if (!selectedText) {
    clearSelection()
    return
  }
  const startInPage = currentPageContent.value.indexOf(selectedText)
  if (startInPage < 0) {
    selectedSource.value = {
      chapterId: currentChapter.value.id,
      chapterOrder: currentChapter.value.order,
      startOffset: currentPageStart.value,
      endOffset: currentPageStart.value + selectedText.length,
      selectedText,
    }
    return
  }
  selectedSource.value = {
    chapterId: currentChapter.value.id,
    chapterOrder: currentChapter.value.order,
    startOffset: currentPageStart.value + startInPage,
    endOffset: currentPageStart.value + startInPage + selectedText.length,
    selectedText,
  }
}

async function copySelection() {
  if (!selectedSource.value?.selectedText) {
    return
  }
  await navigator.clipboard.writeText(selectedSource.value.selectedText)
  readingNotice.value = '已复制选文。'
}

function openDeconstructionDialog() {
  if (!selectedSource.value) {
    readingError.value = '请先在正文中选中文字。'
    return
  }
  cardError.value = ''
  fieldValues.value = {}
  tagInput.value = ''
  selectedPlanTypeId.value = selectedPlanTypeId.value || session.value?.plan.types[0]?.id || ''
  deconstructionOpen.value = true
}

async function saveDeconstruction(keepDialogOpen: boolean) {
  if (!props.project?.book.bookPath || !selectedSource.value || !selectedPlanType.value) {
    return
  }
  const missingField = visiblePlanFields.value.find((field) => field.required && !fieldValues.value[field.id]?.trim())
  if (missingField) {
    cardError.value = `请填写必填字段：${missingField.label}`
    return
  }
  savingCard.value = true
  cardError.value = ''
  try {
    const result = await CreateDeconstructionCard({
      bookPath: props.project.book.bookPath,
      selection: selectedSource.value,
      planTypeId: selectedPlanType.value.id,
      fieldValues: fieldValues.value,
      tagIds: splitTags(tagInput.value),
    } as library.CreateDeconstructionRequest)
    if (session.value) {
      session.value.sourceTexts = result.sourceTexts
      session.value.deconstructionCards = result.deconstructionCards
    }
    readingNotice.value = '拆解已保存。'
    fieldValues.value = {}
    tagInput.value = ''
    if (!keepDialogOpen) {
      deconstructionOpen.value = false
      clearSelection()
    }
  } catch (error) {
    cardError.value = getErrorMessage(error)
  } finally {
    savingCard.value = false
  }
}

async function saveChapterEdit() {
  if (!props.project?.book.bookPath || !currentChapter.value) {
    return
  }
  editSaving.value = true
  editError.value = ''
  editNotice.value = ''
  try {
    const result = await SaveChapterContent({
      bookPath: props.project.book.bookPath,
      chapterId: currentChapter.value.id,
      title: editTitle.value,
      content: editContent.value,
    })
    currentChapter.value = result.chapter
    chapterContent.value = result.content
    currentPageStart.value = clampPageStart(currentPageStart.value)
    updateSessionChapter(result.chapter)
    syncEditor()
    await persistReadingPosition()
    editNotice.value = '章节已保存。'
  } catch (error) {
    editError.value = getErrorMessage(error)
  } finally {
    editSaving.value = false
  }
}

function discardChapterEdit() {
  syncEditor()
  editNotice.value = '已放弃未保存修改。'
  editError.value = ''
}

function syncEditor() {
  editTitle.value = currentChapter.value?.title ?? ''
  editContent.value = chapterContent.value
}

function updateSessionChapter(chapter: library.BookChapter) {
  if (!session.value) {
    return
  }
  const index = session.value.project.chapters.findIndex((item) => item.id === chapter.id)
  if (index >= 0) {
    session.value.project.chapters[index] = chapter
  }
  session.value.currentChapter = chapter
  session.value.chapterContent = chapterContent.value
}

function clearSelection() {
  selectedSource.value = null
  window.getSelection()?.removeAllRanges()
}

function clampPageStart(offset: number) {
  if (!chapterContent.value) {
    return 0
  }
  const maxStart = Math.max(0, Math.floor((chapterContent.value.length - 1) / pageSize) * pageSize)
  return Math.min(Math.max(0, Math.floor(offset / pageSize) * pageSize), maxStart)
}

function estimateParagraphIndex(content: string, offset: number) {
  const before = content.slice(0, Math.max(0, offset))
  return before.split(/\n\s*\n/).length - 1
}

function splitTags(value: string) {
  return value
    .split(/[，,]/)
    .map((item) => item.trim())
    .filter(Boolean)
}

function chapterCardCount(chapterId: string) {
  return deconstructionCountByChapter.value.get(chapterId) ?? 0
}

function emitReadingSummary() {
  nextTick(() => {
    emit('readingUpdated', readingProgressText.value)
  })
}

watch(
  () => props.project?.book.bookPath,
  () => {
    void loadReadingSession()
  },
)

watch(readerMode, () => {
  void persistReadingPosition()
})

watch(catalogSearch, () => {
  catalogPage.value = 1
})

onMounted(loadReadingSession)
</script>

<template>
  <section class="reading-page">
    <n-alert v-if="!project" type="info" class="reader-empty-state">
      请先从书库打开一本小说。阅读位置、拆解卡片和选文来源都会保存在该小说项目目录中。
    </n-alert>

    <n-spin v-else :show="loading" class="reader-spin">
      <main class="reader-pane reader-pane-full">
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
            <n-tag type="info">{{ readingProgressText }}</n-tag>
            <span class="reader-save-state">{{ savingPosition ? '保存阅读位置中' : '阅读位置已本地保存' }}</span>
          </n-space>
          <n-space>
            <n-button quaternary @click="openCatalog">
              <template #icon>
                <List :size="16" />
              </template>
              目录
            </n-button>
          </n-space>
        </div>

        <n-alert v-if="readingError" type="error" class="reader-inline-alert">
          {{ readingError }}
        </n-alert>
        <n-alert v-else-if="readingNotice" type="success" class="reader-inline-alert">
          {{ readingNotice }}
        </n-alert>

        <template v-if="readerMode === 'read'">
          <div class="reader-layout">
            <button class="reader-page-button" :disabled="pageLoading" @click="goPreviousPage">
              <ChevronLeft :size="26" />
              <span>上一页</span>
            </button>

            <article class="novel-text" @mouseup="captureSelection" @keyup="captureSelection">
              <div v-if="!currentPageContent" class="reader-placeholder">
                当前章节没有正文，可切换到编辑 TAB 补充内容。
              </div>
              <div v-else class="reader-content-text">{{ currentPageContent }}</div>

              <div v-if="selectedSource" class="floating-toolbar">
                <n-button size="small" type="primary" @click="openDeconstructionDialog">
                  <template #icon>
                    <Highlighter :size="14" />
                  </template>
                  创建拆解
                </n-button>
                <n-button size="small" @click="copySelection">
                  <template #icon>
                    <Copy :size="14" />
                  </template>
                  复制
                </n-button>
              </div>
            </article>

            <button class="reader-page-button" :disabled="pageLoading" @click="goNextPage">
              <ChevronRight :size="26" />
              <span>下一页</span>
            </button>
          </div>

          <div class="reader-statusbar">
            <span>{{ currentChapter?.title ?? '未打开章节' }}</span>
            <span>页内位置 {{ currentPageStart + 1 }} / {{ Math.max(chapterContent.length, 1) }}</span>
            <span v-if="selectedSource">已选 {{ selectedSource.selectedText.length }} 字</span>
          </div>
        </template>

        <div v-else class="edit-pane">
          <n-alert v-if="editError" type="error" class="reader-inline-alert">
            {{ editError }}
          </n-alert>
          <n-alert v-else-if="editNotice" type="success" class="reader-inline-alert">
            {{ editNotice }}
          </n-alert>
          <n-form label-placement="top">
            <n-form-item label="章节标题">
              <n-input v-model:value="editTitle" class="edit-title-input" />
            </n-form-item>
            <n-form-item label="章节正文">
              <n-input
                v-model:value="editContent"
                type="textarea"
                :autosize="{ minRows: 18, maxRows: 26 }"
              />
            </n-form-item>
          </n-form>
          <n-space justify="end" class="card-actions">
            <n-button :disabled="!editDirty" @click="discardChapterEdit">
              <template #icon>
                <RotateCcw :size="16" />
              </template>
              放弃修改
            </n-button>
            <n-button type="primary" :disabled="!editDirty" :loading="editSaving" @click="saveChapterEdit">
              <template #icon>
                <Save :size="16" />
              </template>
              保存章节
            </n-button>
          </n-space>
        </div>
      </main>
    </n-spin>

    <n-drawer v-model:show="catalogOpen" :width="420" placement="right">
      <n-drawer-content title="章节目录">
        <div class="catalog-meta">
          <span>共 {{ chapters.length }} 章</span>
          <span>第 {{ Math.min(catalogPage, catalogPageCount) }} / {{ catalogPageCount }} 页</span>
        </div>
        <n-input v-model:value="catalogSearch" clearable placeholder="搜索章节标题" class="catalog-search" />
        <n-scrollbar style="max-height: calc(100vh - 250px)">
          <button
            v-for="chapter in catalogPageChapters"
            :key="chapter.id"
            class="chapter-item"
            :class="{ active: currentChapter?.id === chapter.id }"
            @click="jumpToChapter(chapter)"
          >
            <span>{{ chapter.title }}</span>
            <small>第 {{ chapter.order }} 章 · {{ chapterCardCount(chapter.id) }} 张卡片</small>
          </button>
        </n-scrollbar>
        <div class="catalog-actions">
          <n-button :disabled="catalogPage <= 1" @click="catalogPage -= 1">上一目录页</n-button>
          <n-button :disabled="catalogPage >= catalogPageCount" @click="catalogPage += 1">下一目录页</n-button>
        </div>
      </n-drawer-content>
    </n-drawer>

    <n-modal
      v-model:show="deconstructionOpen"
      preset="card"
      title="创建拆解"
      class="deconstruction-modal"
      :bordered="false"
    >
      <n-alert v-if="selectedSource" type="default" :show-icon="false" class="source-box">
        {{ selectedSource.selectedText }}
      </n-alert>
      <n-alert v-if="cardError" type="error" class="reader-inline-alert">
        {{ cardError }}
      </n-alert>
      <n-form label-placement="top">
        <n-form-item label="拆解类型">
          <n-select v-model:value="selectedPlanTypeId" :options="planTypeOptions" />
        </n-form-item>
        <n-form-item
          v-for="field in visiblePlanFields"
          :key="field.id"
          :label="field.required ? `${field.label} *` : field.label"
        >
          <n-select
            v-if="field.type === 'select'"
            v-model:value="fieldValues[field.id]"
            :options="field.options.map((option) => ({ label: option, value: option }))"
          />
          <n-input
            v-else
            v-model:value="fieldValues[field.id]"
            :type="field.type === 'text' ? 'text' : 'textarea'"
            :autosize="field.type === 'text' ? undefined : { minRows: 3, maxRows: 6 }"
          />
        </n-form-item>
        <n-form-item label="标签">
          <n-input v-model:value="tagInput" placeholder="用逗号分隔，例如：氛围, 重逢" />
        </n-form-item>
      </n-form>
      <template #footer>
        <n-space justify="end">
          <n-button @click="deconstructionOpen = false">取消</n-button>
          <n-button :loading="savingCard" @click="saveDeconstruction(true)">保存并继续创建</n-button>
          <n-button type="primary" :loading="savingCard" @click="saveDeconstruction(false)">
            <template #icon>
              <BookOpen :size="16" />
            </template>
            保存拆解
          </n-button>
        </n-space>
      </template>
    </n-modal>
  </section>
</template>
