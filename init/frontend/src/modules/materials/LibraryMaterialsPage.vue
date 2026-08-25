<script setup lang="ts">
// LibraryMaterialsPage.vue 负责书库素材的分页聚合、筛选、自建素材维护和来源书跳转。
import { computed, nextTick, ref, watch } from 'vue'
import {
  NAlert,
  NButton,
  NCard,
  NEmpty,
  NForm,
  NFormItem,
  NInput,
  NModal,
  NPagination,
  NPopconfirm,
  NSelect,
  NSpace,
  NSpin,
  NTag,
} from 'naive-ui'
import { BookOpen, Edit3, Eye, Library, MapPinned, Save, Search, Trash2 } from 'lucide-vue-next'
import {
  CreateLibraryMaterial,
  DeleteLibraryMaterial,
  OpenBookProject,
  OpenLibraryMaterialPage,
  UpdateLibraryMaterial,
} from '@wails/go/main/App'
import type { library } from '@wails/go/models'
import type { BookProject } from '@/types/library'
import { formatDateTime, getErrorMessage } from '@/utils/format'

const emit = defineEmits<{
  openBookSource: [project: BookProject]
}>()

const pageData = ref<library.LibraryMaterialPage | null>(null)
const loading = ref(false)
const saving = ref(false)
const deleting = ref(false)
const openingBook = ref(false)
const errorMessage = ref('')
const noticeMessage = ref('')
const keyword = ref('')
const bookFilter = ref<string | null>(null)
const typeFilter = ref<string | null>(null)
const tagFilter = ref<string | null>(null)
const currentPage = ref(1)
const pageSize = ref(50)
const selectedItemId = ref('')
const detailOpen = ref(false)
const editOpen = ref(false)
const editorMode = ref<'create' | 'edit'>('create')
const editTitle = ref('')
const editSummary = ref('')
const editContent = ref('')
const editType = ref('writingTechnique')
const editPlanType = ref('')
const editDimension = ref('')
const editTagInput = ref('')
const tabIndentSpaces = '    '

const items = computed(() => pageData.value?.items ?? [])
const selectedItem = computed(() => items.value.find((item) => item.id === selectedItemId.value) ?? null)
const canEditSelected = computed(() => selectedItem.value ? canEditLibraryItem(selectedItem.value) : editorMode.value === 'create')

const bookOptions = computed(() => {
  return (pageData.value?.books ?? []).map((book) => ({
    label: book.bookName,
    value: book.bookId,
  }))
})

const typeOptions = computed(() => {
  const values = new Set([...(pageData.value?.types ?? []), editType.value].filter(Boolean))
  return Array.from(values).map((value) => ({
    label: materialTypeLabel(value),
    value,
  }))
})

const tagOptions = computed(() => {
  return (pageData.value?.tags ?? []).map((tag) => ({
    label: tag,
    value: tag,
  }))
})

const hasDirtyEdit = computed(() => {
  if (editorMode.value === 'create') {
    return Boolean(editTitle.value.trim() || editSummary.value.trim() || editContent.value.trim() || editTagInput.value.trim())
  }
  const item = selectedItem.value
  if (!item) {
    return false
  }
  return (
    editTitle.value !== item.title ||
    editSummary.value !== item.summary ||
    editContent.value !== item.content ||
    editType.value !== item.materialType ||
    editPlanType.value !== deconstructionTypeName(item) ||
    editDimension.value !== dimensionName(item) ||
    editTagInput.value !== item.tagIds.join(', ')
  )
})

/**
 * 书库素材可能有上千条，筛选和分页统一交给后端，前端只渲染当前页。
 */
async function loadPage(options: { preserveNotice?: boolean } = {}) {
  loading.value = true
  errorMessage.value = ''
  if (!options.preserveNotice) {
    noticeMessage.value = ''
  }
  try {
    pageData.value = await OpenLibraryMaterialPage({
      keyword: keyword.value,
      bookId: bookFilter.value ?? '',
      materialType: typeFilter.value ?? '',
      tagId: tagFilter.value ?? '',
      page: currentPage.value,
      pageSize: pageSize.value,
    } as library.LibraryMaterialPageRequest)
    currentPage.value = pageData.value.page
    pageSize.value = pageData.value.pageSize
    selectedItemId.value = items.value[0]?.id ?? ''
    syncEditorFromSelected()
  } catch (error) {
    errorMessage.value = getErrorMessage(error)
  } finally {
    loading.value = false
  }
}

function searchMaterials() {
  currentPage.value = 1
  void loadPage()
}

function openDetail(item: library.LibraryMaterialListItem) {
  selectedItemId.value = item.id
  detailOpen.value = true
}

function openCreate() {
  editorMode.value = 'create'
  selectedItemId.value = ''
  resetEditor()
  editOpen.value = true
}

function openEdit(item: library.LibraryMaterialListItem) {
  if (!canEditLibraryItem(item)) {
    return
  }
  editorMode.value = 'edit'
  selectedItemId.value = item.id
  syncEditorFromSelected()
  editOpen.value = true
}

function resetEditor() {
  editTitle.value = ''
  editSummary.value = ''
  editContent.value = ''
  editType.value = 'writingTechnique'
  editPlanType.value = ''
  editDimension.value = ''
  editTagInput.value = ''
}

function syncEditorFromSelected() {
  const item = selectedItem.value
  if (!item) {
    if (editorMode.value === 'create') {
      resetEditor()
    }
    return
  }
  editTitle.value = item.title
  editSummary.value = item.summary
  editContent.value = item.content
  editType.value = item.materialType || 'writingTechnique'
  editPlanType.value = deconstructionTypeName(item)
  editDimension.value = dimensionName(item)
  editTagInput.value = item.tagIds.join(', ')
}

async function saveLibraryMaterial() {
  saving.value = true
  errorMessage.value = ''
  noticeMessage.value = ''
  try {
    if (editorMode.value === 'create') {
      await CreateLibraryMaterial({
        title: editTitle.value,
        summary: editSummary.value,
        content: editContent.value,
        materialType: editType.value,
        planTypeName: editPlanType.value,
        dimensionName: editDimension.value,
        tagIds: splitList(editTagInput.value),
      } as library.CreateLibraryMaterialRequest)
      noticeMessage.value = '书库素材已新增。'
    } else {
      const item = selectedItem.value
      if (!item || !canEditLibraryItem(item)) {
        return
      }
      await UpdateLibraryMaterial({
        libraryMaterialId: item.libraryMaterialId,
        title: editTitle.value,
        summary: editSummary.value,
        content: editContent.value,
        materialType: editType.value,
        planTypeName: editPlanType.value,
        dimensionName: editDimension.value,
        tagIds: splitList(editTagInput.value),
      } as library.UpdateLibraryMaterialRequest)
      noticeMessage.value = '书库素材已保存。'
    }
    editOpen.value = false
    await loadPage({ preserveNotice: true })
  } catch (error) {
    errorMessage.value = getErrorMessage(error)
  } finally {
    saving.value = false
  }
}

async function deleteLibraryMaterial(item: library.LibraryMaterialListItem) {
  if (!canEditLibraryItem(item)) {
    return
  }
  deleting.value = true
  selectedItemId.value = item.id
  errorMessage.value = ''
  noticeMessage.value = ''
  try {
    await DeleteLibraryMaterial({
      libraryMaterialId: item.libraryMaterialId,
    } as library.DeleteLibraryMaterialRequest)
    detailOpen.value = false
    editOpen.value = false
    noticeMessage.value = '书库素材已删除。'
    await loadPage({ preserveNotice: true })
  } catch (error) {
    errorMessage.value = getErrorMessage(error)
  } finally {
    deleting.value = false
  }
}

async function openSourceBook(item: library.LibraryMaterialListItem | null = selectedItem.value) {
  if (!item?.bookPath || !item.sourceAvailable) {
    return
  }
  openingBook.value = true
  errorMessage.value = ''
  try {
    const project = await OpenBookProject(item.bookPath)
    emit('openBookSource', project as BookProject)
  } catch (error) {
    errorMessage.value = getErrorMessage(error)
  } finally {
    openingBook.value = false
  }
}

/**
 * 多行编辑框按 Tab 时插入 4 个空格，保持中文写作首行缩进习惯。
 */
function handleTextareaTab(event: KeyboardEvent, field: 'summary' | 'content') {
  if (event.key !== 'Tab' || !(event.target instanceof HTMLTextAreaElement)) {
    return
  }
  event.preventDefault()
  event.stopPropagation()
  const textarea = event.target
  const { selectionStart, selectionEnd } = textarea
  const targetRef = field === 'summary' ? editSummary : editContent
  targetRef.value = `${targetRef.value.slice(0, selectionStart)}${tabIndentSpaces}${targetRef.value.slice(selectionEnd)}`
  void nextTick(() => {
    textarea.focus()
    textarea.setSelectionRange(selectionStart + tabIndentSpaces.length, selectionStart + tabIndentSpaces.length)
  })
}

function canEditLibraryItem(item: library.LibraryMaterialListItem) {
  return item.sourceMode !== 'bookReference' && Boolean(item.libraryMaterialId)
}

function sourceModeLabel(item: library.LibraryMaterialListItem) {
  const labels: Record<string, string> = {
    bookReference: '本书引用',
    libraryOwned: '书库自建',
    migratedBookCopy: '已迁移',
  }
  return labels[item.sourceMode] ?? item.sourceMode
}

function sourceSummary(item: library.LibraryMaterialListItem) {
  if (item.bookName) {
    return item.sourceAvailable ? item.bookName : `${item.bookName} / 来源书已移除`
  }
  return '书库自建素材'
}

function materialContentExcerpt(item: library.LibraryMaterialListItem, limit = 96) {
  return originalText(item).trim().replace(/\s+/g, ' ').slice(0, limit)
}

function deconstructionTypeName(item: library.LibraryMaterialListItem) {
  return item.planTypeName || splitDeconstructionTitle(item.sourceRefs[0]?.snapshot.deconstructionTitle ?? '')[0] || materialTypeLabel(item.materialType)
}

function dimensionName(item: library.LibraryMaterialListItem) {
  return item.dimensionName || splitDeconstructionTitle(item.sourceRefs[0]?.snapshot.deconstructionTitle ?? '')[1] || '未记录维度'
}

function originalText(item: library.LibraryMaterialListItem) {
  return item.summary || item.sourceRefs[0]?.snapshot.selectedText || ''
}

function explanationText(item: library.LibraryMaterialListItem) {
  return item.content || item.sourceRefs[0]?.snapshot.sourceSummary || ''
}

function splitDeconstructionTitle(value: string) {
  return value.split('/').map((item) => item.trim()).filter(Boolean)
}

function materialTypeLabel(value: string) {
  const labels: Record<string, string> = {
    writingTechnique: '写作手法',
    storyPlot: '剧情结构',
  }
  return labels[value] ?? value
}

function splitList(value: string) {
  return value
    .split(/[，,]/)
    .map((item) => item.trim())
    .filter(Boolean)
}

watch([bookFilter, typeFilter, tagFilter], searchMaterials)
watch(selectedItem, syncEditorFromSelected)
void loadPage()

defineExpose({
  refresh: loadPage,
  openCreate,
})
</script>

<template>
  <section class="page-section">
    <n-spin :show="loading">
      <n-alert v-if="errorMessage" type="error" class="materials-alert">{{ errorMessage }}</n-alert>
      <n-alert v-else-if="noticeMessage" type="success" class="materials-alert">{{ noticeMessage }}</n-alert>

      <div class="materials-filter-bar">
        <n-input v-model:value="keyword" clearable placeholder="搜索书名、标题、原文、拆解说明、类型、维度、标签" @keyup.enter="searchMaterials">
          <template #prefix>
            <Search :size="15" />
          </template>
        </n-input>
        <n-button type="primary" secondary @click="searchMaterials">搜索</n-button>
        <div class="filter-selects">
          <n-select v-model:value="bookFilter" clearable :options="bookOptions" placeholder="书籍" />
          <n-select v-model:value="typeFilter" clearable :options="typeOptions" placeholder="类型" />
          <n-select v-model:value="tagFilter" clearable :options="tagOptions" placeholder="标签" />
        </div>
      </div>

      <div class="materials-grid">
        <n-empty v-if="items.length === 0" description="暂无符合条件的素材" />
        <n-card
          v-for="item in items"
          :key="item.id"
          :bordered="false"
          class="item-card material-card"
        >
          <div class="material-card-head">
            <n-space size="small">
              <n-tag type="success" size="small">{{ deconstructionTypeName(item) }}</n-tag>
              <n-tag size="small">{{ dimensionName(item) }}</n-tag>
            </n-space>
            <n-tag size="small" :type="item.sourceMode === 'bookReference' ? 'info' : 'warning'">
              {{ sourceModeLabel(item) }}
            </n-tag>
          </div>
          <h2>{{ item.title || '未命名素材' }}</h2>
          <p>{{ materialContentExcerpt(item) || '暂无原文' }}</p>
          <div class="source-line">
            <MapPinned :size="15" />
            {{ sourceSummary(item) }}
          </div>
          <n-space v-if="item.tagIds.length" size="small" class="material-tags">
            <n-tag v-for="tag in item.tagIds.slice(0, 3)" :key="tag" size="small">{{ tag }}</n-tag>
            <n-tag v-if="item.tagIds.length > 3" size="small">+{{ item.tagIds.length - 3 }}</n-tag>
          </n-space>
          <div :class="['material-actions', { 'material-actions-two': !canEditLibraryItem(item) }]">
            <n-button size="small" secondary @click="openDetail(item)">
              <template #icon>
                <Eye :size="15" />
              </template>
              详情
            </n-button>
            <n-button v-if="canEditLibraryItem(item)" size="small" secondary @click="openEdit(item)">
              <template #icon>
                <Edit3 :size="15" />
              </template>
              编辑
            </n-button>
            <n-popconfirm v-if="canEditLibraryItem(item)" @positive-click="deleteLibraryMaterial(item)">
              <template #trigger>
                <n-button size="small" type="error" secondary :loading="deleting && selectedItemId === item.id">
                  <template #icon>
                    <Trash2 :size="15" />
                  </template>
                  删除
                </n-button>
              </template>
              删除后会移除这条书库素材和拆解说明 TXT。
            </n-popconfirm>
            <n-button v-else size="small" secondary :disabled="!item.sourceAvailable" @click="openSourceBook(item)">
              <template #icon>
                <BookOpen :size="15" />
              </template>
              来源书
            </n-button>
          </div>
        </n-card>
      </div>

      <div class="materials-footer">
        <div>
          <Library :size="15" />
          共 {{ pageData?.total ?? 0 }} 条素材，当前第 {{ pageData?.page ?? 1 }} 页
        </div>
        <n-pagination
          v-model:page="currentPage"
          v-model:page-size="pageSize"
          :item-count="pageData?.total ?? 0"
          :page-sizes="[20, 50, 100]"
          show-size-picker
          @update:page="() => loadPage()"
          @update:page-size="searchMaterials"
        />
      </div>
    </n-spin>

    <n-modal v-model:show="detailOpen" preset="card" title="素材详情" class="material-modal" :bordered="false">
      <template v-if="selectedItem">
        <div class="detail-meta-grid">
          <div>
            <span>拆解类型</span>
            <strong>{{ deconstructionTypeName(selectedItem) }}</strong>
          </div>
          <div>
            <span>维度</span>
            <strong>{{ dimensionName(selectedItem) }}</strong>
          </div>
        </div>
        <section class="material-field">
          <strong>原文</strong>
          <p>{{ originalText(selectedItem) || '暂无原文' }}</p>
        </section>
        <section class="material-field">
          <strong>拆解说明</strong>
          <p>{{ explanationText(selectedItem) || '暂无拆解说明' }}</p>
        </section>
        <section class="material-field">
          <strong>标签</strong>
          <n-space v-if="selectedItem.tagIds.length" size="small">
            <n-tag v-for="tag in selectedItem.tagIds" :key="tag" size="small">{{ tag }}</n-tag>
          </n-space>
          <span v-else class="empty-text">暂无标签</span>
        </section>
      </template>
      <template #footer>
        <n-space justify="end">
          <n-button @click="detailOpen = false">关闭</n-button>
          <n-button secondary :disabled="!selectedItem?.sourceAvailable" :loading="openingBook" @click="openSourceBook()">
            <template #icon>
              <BookOpen :size="16" />
            </template>
            打开来源书
          </n-button>
          <n-button v-if="selectedItem && canEditLibraryItem(selectedItem)" type="primary" secondary @click="openEdit(selectedItem)">
            <template #icon>
              <Edit3 :size="16" />
            </template>
            编辑素材
          </n-button>
        </n-space>
      </template>
    </n-modal>

    <n-modal v-model:show="editOpen" preset="card" :title="editorMode === 'create' ? '新增书库素材' : '编辑书库素材'" class="material-modal" :bordered="false">
      <n-alert v-if="editorMode === 'edit' && !canEditSelected" type="info" class="materials-alert">
        这条素材是本书素材引用，请到对应小说的素材页编辑。
      </n-alert>
      <n-form v-else label-placement="top">
        <n-form-item label="标题">
          <n-input v-model:value="editTitle" placeholder="素材标题" />
        </n-form-item>
        <div class="edit-grid">
          <n-form-item label="拆解类型">
            <n-input v-model:value="editPlanType" placeholder="例如：表现手法" />
          </n-form-item>
          <n-form-item label="维度">
            <n-input v-model:value="editDimension" placeholder="例如：环境描写" />
          </n-form-item>
        </div>
        <n-form-item label="原文">
          <n-input
            v-model:value="editSummary"
            type="textarea"
            :autosize="{ minRows: 3, maxRows: 5 }"
            @keydown="(event) => handleTextareaTab(event, 'summary')"
          />
        </n-form-item>
        <n-form-item label="拆解说明">
          <n-input
            v-model:value="editContent"
            type="textarea"
            :autosize="{ minRows: 9, maxRows: 16 }"
            @keydown="(event) => handleTextareaTab(event, 'content')"
          />
        </n-form-item>
        <n-form-item label="标签">
          <n-input v-model:value="editTagInput" placeholder="用逗号分隔，例如：氛围, 重逢" />
        </n-form-item>
      </n-form>
      <template #footer>
        <n-space justify="end">
          <n-button @click="editOpen = false">取消</n-button>
          <n-button type="primary" :disabled="!hasDirtyEdit || !canEditSelected" :loading="saving" @click="saveLibraryMaterial">
            <template #icon>
              <Save :size="16" />
            </template>
            保存素材
          </n-button>
        </n-space>
      </template>
    </n-modal>
  </section>
</template>

<style scoped>
.materials-alert {
  margin-bottom: 14px;
}

.materials-filter-bar {
  display: grid;
  grid-template-columns: minmax(260px, 1fr) auto;
  gap: 8px;
  margin-bottom: 16px;
}

.filter-selects {
  display: grid;
  grid-column: 1 / -1;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 8px;
}

.materials-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(292px, 1fr));
  gap: 16px;
}

.material-card {
  min-height: 286px;
  border: 1px solid transparent;
  transition: border-color 0.16s ease, transform 0.16s ease;
}

.material-card:hover {
  border-color: rgba(42, 110, 245, 0.36);
  transform: translateY(-1px);
}

.material-card :deep(.n-card__content) {
  display: flex;
  flex-direction: column;
  height: 100%;
}

.material-card-head {
  display: flex;
  gap: 10px;
  align-items: flex-start;
  justify-content: space-between;
}

.material-tags {
  min-height: 24px;
  margin-top: 10px;
}

.material-actions {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 10px;
  margin-top: auto;
  padding-top: 14px;
}

.material-actions-two {
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.materials-footer {
  display: flex;
  gap: 16px;
  align-items: center;
  justify-content: space-between;
  margin-top: 14px;
  color: var(--text-secondary);
  font-size: 13px;
}

.materials-footer > div {
  display: flex;
  gap: 6px;
  align-items: center;
}

.detail-meta-grid,
.edit-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px;
}

.detail-meta-grid {
  margin-bottom: 16px;
}

.detail-meta-grid div {
  padding: 10px 12px;
  background: rgba(15, 23, 42, 0.04);
  border-radius: 8px;
}

.detail-meta-grid span {
  display: block;
  margin-bottom: 6px;
  color: var(--text-secondary);
  font-size: 12px;
}

.detail-meta-grid strong {
  display: block;
  overflow-wrap: anywhere;
  font-size: 14px;
}

.material-field {
  margin-top: 14px;
}

.material-field strong {
  display: block;
  margin-bottom: 8px;
  font-size: 14px;
}

.material-field p {
  margin: 0;
  white-space: pre-wrap;
}

.empty-text {
  color: var(--text-secondary);
  font-size: 13px;
}

@media (max-width: 1180px) {
  .filter-selects {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 720px) {
  .materials-filter-bar,
  .filter-selects,
  .detail-meta-grid,
  .edit-grid {
    grid-template-columns: 1fr;
  }
}
</style>
