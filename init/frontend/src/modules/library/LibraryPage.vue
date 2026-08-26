<script setup lang="ts">
// LibraryPage.vue 负责书库列表和书库内导入流程，正式数据读写通过 Wails 书库 API 完成。
import { computed, onMounted, ref } from 'vue'
import {
  NAlert,
  NButton,
  NCard,
  NDescriptions,
  NDescriptionsItem,
  NForm,
  NFormItem,
  NGi,
  NGrid,
  NInput,
  NPopconfirm,
  NSpace,
} from 'naive-ui'
import { FileInput, FolderOpen, RotateCcw, Save, Search, Trash2 } from 'lucide-vue-next'
import {
  ChooseLibraryDirectory,
  ConfirmNovelImport,
  GetLibraryState,
  OpenBookProject,
  PreviewNovelImport,
  ReDetectNovelImport,
  RemoveBookFromLibrary,
} from '@wails/go/main/App'
import type { BookProject, ImportDraft, LibraryBookEntry, LibraryState } from '@/types/library'
import { formatBytes, getErrorMessage } from '@/utils/format'

const props = defineProps<{
  mode: 'list' | 'import'
}>()

const emit = defineEmits<{
  'update:mode': [mode: 'list' | 'import']
  openBook: [project: BookProject]
  bookRemoved: [bookPath: string]
}>()

const libraryMode = computed({
  get: () => props.mode,
  set: (value: 'list' | 'import') => emit('update:mode', value),
})

const libraryState = ref<LibraryState | null>(null)
const libraryLoading = ref(false)
const libraryError = ref('')
const librarySearch = ref('')
const removingBookPath = ref('')
const importDraft = ref<ImportDraft | null>(null)
const importForm = ref({ bookName: '', description: '' })
const importLoading = ref(false)
const importError = ref('')
const importSuccess = ref('')
const importChapterPage = ref(1)
const importChapterPageSize = 10

/**
 * 为章节预览表格提供稳定行标识，避免快速分页时表格内部复用错误的行状态。
 */
function importChapterRowKey(row: { id?: string; order: number }) {
  return row.id || `chapter-order-${row.order}`
}

const filteredLibraryBooks = computed(() => {
  const keyword = librarySearch.value.trim().toLowerCase()
  const books = libraryState.value?.index.books ?? []
  if (!keyword) {
    return books
  }
  return books.filter((book) => book.bookName.toLowerCase().includes(keyword))
})

const pagedImportChapters = computed(() => {
  if (!importDraft.value) {
    return []
  }
  const start = (importChapterPage.value - 1) * importChapterPageSize
  return importDraft.value.chapters.slice(start, start + importChapterPageSize)
})

const importChapterPageCount = computed(() => {
  if (!importDraft.value?.chapterCount) {
    return 1
  }
  return Math.ceil(importDraft.value.chapterCount / importChapterPageSize)
})

function setImportChapterPage(page: number) {
  const nextPage = Math.min(importChapterPageCount.value, Math.max(1, page))
  importChapterPage.value = nextPage
}

const currentImportInfo = computed(() => ({
  sourcePath: importDraft.value?.sourcePath ?? '',
  libraryPath: importDraft.value?.libraryPath ?? libraryState.value?.libraryPath ?? '',
  targetPath: importDraft.value?.targetPath ?? '',
  size: formatBytes(importDraft.value?.bookSize ?? 0),
  chapterCount: importDraft.value?.chapterCount ?? 0,
}))

const importDiagnostics = computed(() => importDraft.value?.diagnostics)

const importDiagnosticStages = computed(() => importDiagnostics.value?.stages ?? [])

function formatDurationMs(durationMs?: number) {
  if (durationMs === undefined || durationMs === null) {
    return '-'
  }
  if (durationMs < 1000) {
    return `${durationMs} ms`
  }
  return `${(durationMs / 1000).toFixed(2)} s`
}

/**
 * 启动时加载书库索引，页面只展示后端返回的 index.json 轻量数据。
 */
async function loadLibrary() {
  libraryLoading.value = true
  libraryError.value = ''
  try {
    libraryState.value = await GetLibraryState()
  } catch (error) {
    libraryError.value = getErrorMessage(error)
  } finally {
    libraryLoading.value = false
  }
}

/**
 * 选择或更换书库目录；正式小说数据仍写在用户选择的目录中。
 */
async function chooseLibraryDirectory() {
  libraryLoading.value = true
  libraryError.value = ''
  try {
    libraryState.value = await ChooseLibraryDirectory()
  } catch (error) {
    libraryError.value = getErrorMessage(error)
  } finally {
    libraryLoading.value = false
  }
}

function showLibraryImport() {
  libraryMode.value = 'import'
  importError.value = ''
  importSuccess.value = ''
}

/**
 * 选择 TXT 小说文件并生成导入草稿，草稿阶段不写入 book.json。
 */
async function chooseImportFile() {
  importLoading.value = true
  importError.value = ''
  importSuccess.value = ''
  try {
    const draft = await PreviewNovelImport()
    if (!draft) {
      return
    }
    applyImportDraft(draft)
  } catch (error) {
    importError.value = getErrorMessage(error)
  } finally {
    importLoading.value = false
  }
}

/**
 * 对当前导入文件重新执行章节识别，便于导入前刷新章节标题预览。
 */
async function reDetectImport() {
  if (!importDraft.value) {
    await chooseImportFile()
    return
  }
  importLoading.value = true
  importError.value = ''
  importSuccess.value = ''
  try {
    const draft = await ReDetectNovelImport(importDraft.value.sourcePath)
    if (draft) {
      applyImportDraft(draft)
    }
  } catch (error) {
    importError.value = getErrorMessage(error)
  } finally {
    importLoading.value = false
  }
}

/**
 * 确认导入会触发后端复制原文、写章节 TXT、写 book.json 并更新书库 index.json。
 */
async function confirmImport() {
  if (!importDraft.value) {
    importError.value = '请先选择 TXT 小说文件。'
    return
  }
  if (!importForm.value.bookName.trim()) {
    importError.value = '小说名不能为空。'
    return
  }

  importLoading.value = true
  importError.value = ''
  importSuccess.value = ''
  try {
    const result = await ConfirmNovelImport({
      bookId: importDraft.value.bookId,
      sourcePath: importDraft.value.sourcePath,
      bookName: importForm.value.bookName,
      description: importForm.value.description,
    })
    libraryState.value = result.library
    console.info('确认导入性能诊断', result.diagnostics)
    importSuccess.value = '导入成功，已写入书库目录。'
    resetImportDraft()
    libraryMode.value = 'list'
    emit('openBook', result.project)
  } catch (error) {
    importError.value = getErrorMessage(error)
  } finally {
    importLoading.value = false
  }
}

async function openBook(book: LibraryBookEntry) {
  libraryError.value = ''
  try {
    const project = await OpenBookProject(book.bookPath)
    promoteLibraryBookInView(book.bookPath)
    emit('openBook', project)
  } catch (error) {
    libraryError.value = getErrorMessage(error)
  }
}

function promoteLibraryBookInView(bookPath: string) {
  if (!libraryState.value) {
    return
  }
  const books = libraryState.value.index.books ?? []
  const target = books.find((book) => book.bookPath === bookPath)
  if (!target) {
    return
  }
  libraryState.value = {
    ...libraryState.value,
    index: {
      ...libraryState.value.index,
      books: [target, ...books.filter((book) => book.bookPath !== bookPath)],
    },
  }
}

/**
 * 移除书库条目会删除该书完整项目目录，前端只负责确认和刷新索引。
 */
async function removeBook(book: LibraryBookEntry) {
  removingBookPath.value = book.bookPath
  libraryError.value = ''
  try {
    libraryState.value = await RemoveBookFromLibrary(book.bookPath)
    emit('bookRemoved', book.bookPath)
  } catch (error) {
    libraryError.value = getErrorMessage(error)
  } finally {
    removingBookPath.value = ''
  }
}

function applyImportDraft(draft: ImportDraft) {
  importDraft.value = draft
  importForm.value = {
    bookName: draft.bookName,
    description: draft.description ?? '',
  }
  importChapterPage.value = 1
}

function resetImportDraft() {
  importDraft.value = null
  importForm.value = { bookName: '', description: '' }
  importChapterPage.value = 1
}

function cancelImport() {
  resetImportDraft()
  importError.value = ''
  importSuccess.value = ''
  libraryMode.value = 'list'
}

onMounted(loadLibrary)
</script>

<template>
  <section class="page-section library-page">
    <div v-if="libraryMode === 'list'" class="library-page-header">
      <div>
        <h1>书库</h1>
        <p class="library-heading-path">
          {{ libraryState?.libraryPath || '尚未设置书库目录' }}
        </p>
      </div>
      <div class="library-page-actions">
        <n-input
          v-model:value="librarySearch"
          clearable
          placeholder="搜索小说名"
          class="library-page-search"
        >
          <template #prefix>
            <Search :size="16" />
          </template>
        </n-input>
        <n-button :loading="libraryLoading" @click="chooseLibraryDirectory">
          <template #icon>
            <FolderOpen :size="16" />
          </template>
          更换书库
        </n-button>
        <n-button type="primary" @click="showLibraryImport">
          <template #icon>
            <FileInput :size="16" />
          </template>
          导入
        </n-button>
      </div>
    </div>

    <div v-else class="section-heading">
      <div>
        <h1>导入小说</h1>
        <p>导入只展示小说信息、导入信息和分页章节标题。</p>
      </div>
      <n-space>
        <n-button @click="cancelImport">返回书库</n-button>
      </n-space>
    </div>

    <template v-if="libraryMode === 'list'">
      <n-alert v-if="libraryError" type="error" class="page-alert">
        {{ libraryError }}
      </n-alert>

      <n-alert v-if="!libraryState?.libraryPath" type="info" class="page-alert">
        请先设置一个本地书库目录。后续导入的每本小说都会在该目录下生成独立项目目录。
      </n-alert>

      <n-alert
        v-else-if="!libraryLoading && filteredLibraryBooks.length === 0"
        type="default"
        class="page-alert"
      >
        当前书库没有匹配的小说，可以点击“导入”把小说复制入库。
      </n-alert>

      <n-grid
        v-if="filteredLibraryBooks.length > 0"
        :cols="3"
        :x-gap="16"
        :y-gap="16"
        responsive="screen"
      >
        <n-gi v-for="book in filteredLibraryBooks" :key="book.bookPath">
          <n-card class="project-card" :bordered="false">
            <h2>{{ book.bookName }}</h2>
            <div class="book-card-meta">
              <div>
                <span>大小</span>
                <strong>{{ formatBytes(book.bookSize) }}</strong>
              </div>
              <div>
                <span>阅读进度</span>
                <strong>{{ book.readingProgressPercent ?? 0 }}%</strong>
              </div>
            </div>
            <div class="book-progress-track" aria-hidden="true">
              <div
                class="book-progress-value"
                :style="{ width: `${Math.min(100, Math.max(0, book.readingProgressPercent ?? 0))}%` }"
              />
            </div>
            <div class="card-footer book-card-footer">
              <n-popconfirm
                positive-text="移除"
                negative-text="取消"
                @positive-click="removeBook(book)"
              >
                <template #trigger>
                  <n-button
                    size="small"
                    type="error"
                    secondary
                    :loading="removingBookPath === book.bookPath"
                  >
                    <template #icon>
                      <Trash2 :size="15" />
                    </template>
                    移除
                  </n-button>
                </template>
                移除后会删除此书在书库中的项目目录，包括原文、章节、拆解和素材数据。此操作不可撤销。
              </n-popconfirm>
              <n-button size="small" type="primary" @click="openBook(book)">打开</n-button>
            </div>
          </n-card>
        </n-gi>
      </n-grid>
    </template>

    <template v-else>
      <n-alert v-if="importError" type="error" class="page-alert">
        {{ importError }}
      </n-alert>
      <n-alert v-if="importSuccess" type="success" class="page-alert">
        {{ importSuccess }}
      </n-alert>

      <n-card v-if="!libraryState?.libraryPath" :bordered="false" class="workspace-card">
        <n-alert type="warning" :show-icon="false">
          导入前需要先设置书库目录，小说正式数据会写入该目录下的 books 子目录。
        </n-alert>
        <n-space justify="end" class="card-actions">
          <n-button type="primary" :loading="libraryLoading" @click="chooseLibraryDirectory">
            <template #icon>
              <FolderOpen :size="16" />
            </template>
            设置书库目录
          </n-button>
        </n-space>
      </n-card>

      <n-card v-else-if="!importDraft" :bordered="false" class="workspace-card">
        <n-alert type="info" :show-icon="false">
          请选择一个 TXT 小说文件。系统会读取文件、识别章节，并在确认前只生成预览草稿。
        </n-alert>
        <n-space justify="end" class="card-actions">
          <n-button type="primary" :loading="importLoading" @click="chooseImportFile">
            <template #icon>
              <FileInput :size="16" />
            </template>
            选择 TXT 文件
          </n-button>
        </n-space>
      </n-card>

      <template v-else>
        <div class="import-review-layout">
          <n-card :bordered="false" title="小说信息" class="workspace-card">
            <n-form label-placement="top">
              <n-form-item label="小说名">
                <n-input v-model:value="importForm.bookName" />
              </n-form-item>
              <n-form-item label="简介">
                <n-input
                  v-model:value="importForm.description"
                  type="textarea"
                  :autosize="{ minRows: 4, maxRows: 6 }"
                />
              </n-form-item>
            </n-form>
          </n-card>

          <n-card :bordered="false" title="导入信息" class="workspace-card">
            <n-descriptions label-placement="left" :column="1" bordered>
              <n-descriptions-item label="原路径">{{ currentImportInfo.sourcePath }}</n-descriptions-item>
              <n-descriptions-item label="书库目录">{{ currentImportInfo.libraryPath }}</n-descriptions-item>
              <n-descriptions-item label="入库路径">{{ currentImportInfo.targetPath }}</n-descriptions-item>
              <n-descriptions-item label="大小">{{ currentImportInfo.size }}</n-descriptions-item>
              <n-descriptions-item label="章节数">{{ currentImportInfo.chapterCount }}</n-descriptions-item>
              <n-descriptions-item v-if="importDiagnostics" label="解析耗时">
                <div class="import-diagnostics">
                  <strong>总计 {{ formatDurationMs(importDiagnostics.totalMs) }}</strong>
                  <span
                    v-for="stage in importDiagnosticStages"
                    :key="stage.name"
                  >
                    {{ stage.label }} {{ formatDurationMs(stage.durationMs) }}
                  </span>
                </div>
              </n-descriptions-item>
            </n-descriptions>
          </n-card>
        </div>

        <n-card :bordered="false" title="章节信息" class="workspace-card chapter-table-card">
          <table class="import-chapter-table">
            <thead>
              <tr>
                <th>序号</th>
                <th>章节标题</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="chapter in pagedImportChapters" :key="importChapterRowKey(chapter)">
                <td>{{ chapter.order }}</td>
                <td>{{ chapter.title }}</td>
              </tr>
            </tbody>
          </table>
          <div class="chapter-page-footer">
            <span>共识别 {{ currentImportInfo.chapterCount }} 章，仅展示标题用于快速检查。</span>
            <div class="chapter-page-actions">
              <n-button size="small" secondary :disabled="importChapterPage <= 1" @click="setImportChapterPage(1)">
                首页
              </n-button>
              <n-button
                size="small"
                secondary
                :disabled="importChapterPage <= 1"
                @click="setImportChapterPage(importChapterPage - 1)"
              >
                上一页
              </n-button>
              <strong>{{ importChapterPage }} / {{ importChapterPageCount }}</strong>
              <n-button
                size="small"
                secondary
                :disabled="importChapterPage >= importChapterPageCount"
                @click="setImportChapterPage(importChapterPage + 1)"
              >
                下一页
              </n-button>
              <n-button
                size="small"
                secondary
                :disabled="importChapterPage >= importChapterPageCount"
                @click="setImportChapterPage(importChapterPageCount)"
              >
                末页
              </n-button>
            </div>
          </div>
        </n-card>

        <n-space justify="end" class="card-actions">
          <n-button @click="cancelImport">取消</n-button>
          <n-button :loading="importLoading" @click="chooseImportFile">
            <template #icon>
              <FileInput :size="16" />
            </template>
            重新选择
          </n-button>
          <n-button :loading="importLoading" @click="reDetectImport">
            <template #icon>
              <RotateCcw :size="16" />
            </template>
            重新识别
          </n-button>
          <n-button type="primary" :loading="importLoading" @click="confirmImport">
            <template #icon>
              <Save :size="16" />
            </template>
            确认导入
          </n-button>
        </n-space>
      </template>
    </template>
  </section>
</template>
