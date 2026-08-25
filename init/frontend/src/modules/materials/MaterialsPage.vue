<script setup lang="ts">
// MaterialsPage.vue 负责当前书小说内素材的本地读取、筛选、查看、编辑、删除和来源追溯。
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
  NPopconfirm,
  NSelect,
  NSpace,
  NSpin,
  NTag,
} from 'naive-ui'
import { BookOpen, Edit3, Eye, Library, MapPinned, Save, Search, Trash2 } from 'lucide-vue-next'
import { DeleteMaterial, OpenMaterialWorkspace, UpdateMaterial } from '@wails/go/main/App'
import type { library } from '@wails/go/models'
import type { BookProject } from '@/types/library'
import { formatDateTime, getErrorMessage } from '@/utils/format'

const props = defineProps<{
  project: BookProject | null
}>()

const emit = defineEmits<{
  backToSource: [payload?: { chapterId: string; sourceTextId: string; startOffset: number }]
}>()

const workspace = ref<library.MaterialWorkspace | null>(null)
const loading = ref(false)
const saving = ref(false)
const deleting = ref(false)
const errorMessage = ref('')
const noticeMessage = ref('')
const keyword = ref('')
const typeFilter = ref<string | null>(null)
const tagFilter = ref<string | null>(null)
const selectedMaterialId = ref('')
const detailOpen = ref(false)
const editOpen = ref(false)
const editTitle = ref('')
const editSummary = ref('')
const editContent = ref('')
const editTagInput = ref('')
const tabIndentSpaces = '    '

const activeBookPath = computed(() => props.project?.book.bookPath ?? '')
const materials = computed(() => workspace.value?.materials.items ?? [])
const materialContents = computed(() => workspace.value?.materialContents ?? {})

const selectedMaterial = computed(() => {
  return materials.value.find((material) => material.id === selectedMaterialId.value) ?? null
})

const sourceTextsById = computed(() => {
  const map = new Map<string, library.SourceText>()
  for (const sourceText of workspace.value?.sourceTexts.items ?? []) {
    map.set(sourceText.id, sourceText)
  }
  return map
})

const sortedMaterials = computed(() => {
  return [...materials.value].sort((left, right) => materialTimeValue(right) - materialTimeValue(left))
})

const filteredMaterials = computed(() => {
  const cleanKeyword = keyword.value.trim().toLowerCase()
  return sortedMaterials.value.filter((material) => {
    const content = materialContents.value[material.id] ?? ''
    const sourceText = material.sourceRefs.map((source) => source.snapshot.selectedText).join(' ')
    const sourceSummary = material.sourceRefs.map((source) => source.snapshot.sourceSummary).join(' ')
    const hitKeyword =
      !cleanKeyword ||
      material.title.toLowerCase().includes(cleanKeyword) ||
      material.summary.toLowerCase().includes(cleanKeyword) ||
      content.toLowerCase().includes(cleanKeyword) ||
      sourceText.toLowerCase().includes(cleanKeyword) ||
      sourceSummary.toLowerCase().includes(cleanKeyword) ||
      material.tagIds.some((tag) => tag.toLowerCase().includes(cleanKeyword)) ||
      deconstructionTypeName(material).toLowerCase().includes(cleanKeyword) ||
      dimensionName(material).toLowerCase().includes(cleanKeyword)
    return (
      hitKeyword &&
      (!typeFilter.value || material.materialType === typeFilter.value) &&
      (!tagFilter.value || material.tagIds.includes(tagFilter.value))
    )
  })
})

const typeOptions = computed(() => {
  return uniqueStringOptions(materials.value.map((material) => material.materialType), materialTypeLabel)
})

const tagOptions = computed(() => {
  return uniqueStringOptions(materials.value.flatMap((material) => material.tagIds))
})

const hasDirtyEdit = computed(() => {
  const material = selectedMaterial.value
  if (!material) {
    return false
  }
  return (
    editTitle.value !== material.title ||
    editSummary.value !== material.summary ||
    editContent.value !== (materialContents.value[material.id] ?? '') ||
    editTagInput.value !== material.tagIds.join(', ')
  )
})

/**
 * 打开小说内素材工作区；拆解说明从 TXT 读取，结构化信息从 materials.json 读取。
 */
async function loadWorkspace() {
  const bookPath = activeBookPath.value
  workspace.value = null
  selectedMaterialId.value = ''
  errorMessage.value = ''
  noticeMessage.value = ''
  if (!bookPath) {
    return
  }
  loading.value = true
  try {
    workspace.value = await OpenMaterialWorkspace(bookPath)
    selectedMaterialId.value = filteredMaterials.value[0]?.id ?? materials.value[0]?.id ?? ''
    syncDetailEditor()
  } catch (error) {
    errorMessage.value = getErrorMessage(error)
  } finally {
    loading.value = false
  }
}

function openDetail(material: library.Material) {
  selectedMaterialId.value = material.id
  detailOpen.value = true
}

function openEdit(material: library.Material) {
  selectedMaterialId.value = material.id
  syncDetailEditor()
  editOpen.value = true
}

function syncDetailEditor() {
  const material = selectedMaterial.value
  editTitle.value = material?.title ?? ''
  editSummary.value = material?.summary ?? ''
  editContent.value = material ? materialContents.value[material.id] ?? '' : ''
  editTagInput.value = material?.tagIds.join(', ') ?? ''
}

async function saveSelectedMaterial() {
  const bookPath = activeBookPath.value
  const material = selectedMaterial.value
  if (!bookPath || !material) {
    return
  }
  saving.value = true
  errorMessage.value = ''
  noticeMessage.value = ''
  try {
    const result = await UpdateMaterial({
      bookPath,
      materialId: material.id,
      title: editTitle.value,
      summary: editSummary.value,
      content: editContent.value,
      categoryIds: [],
      tagIds: splitList(editTagInput.value),
    } as library.UpdateMaterialRequest)
    if (workspace.value) {
      workspace.value.materials = result.materials
      workspace.value.materialContents = result.materialContents
    }
    selectedMaterialId.value = result.material.id
    syncDetailEditor()
    editOpen.value = false
    noticeMessage.value = '素材已保存。'
  } catch (error) {
    errorMessage.value = getErrorMessage(error)
  } finally {
    saving.value = false
  }
}

async function deleteSelectedMaterial(material: library.Material) {
  const bookPath = activeBookPath.value
  if (!bookPath) {
    return
  }
  selectedMaterialId.value = material.id
  deleting.value = true
  errorMessage.value = ''
  noticeMessage.value = ''
  try {
    const result = await DeleteMaterial({
      bookPath,
      materialId: material.id,
    } as library.DeleteMaterialRequest)
    if (workspace.value) {
      workspace.value.materials = result.materials
      workspace.value.materialContents = result.materialContents
      workspace.value.deconstructionCards = result.deconstructionCards
    }
    selectedMaterialId.value = filteredMaterials.value[0]?.id ?? materials.value[0]?.id ?? ''
    detailOpen.value = false
    editOpen.value = false
    syncDetailEditor()
    noticeMessage.value = '素材已删除。'
  } catch (error) {
    errorMessage.value = getErrorMessage(error)
  } finally {
    deleting.value = false
  }
}

function emitBackToSource(material: library.Material | null = selectedMaterial.value) {
  const ref = material?.sourceRefs.find((source) => source.sourceTextId)
  const sourceText = ref ? sourceTextsById.value.get(ref.sourceTextId) : null
  if (!sourceText) {
    emit('backToSource')
    return
  }
  emit('backToSource', {
    chapterId: sourceText.chapterId,
    sourceTextId: sourceText.id,
    startOffset: sourceText.startOffset,
  })
}

/**
 * 在素材原文编辑区拦截 Tab，在当前光标位置插入 4 个普通空格。
 */
function handleEditSummaryKeydown(event: KeyboardEvent) {
  if (event.key !== 'Tab' || !(event.target instanceof HTMLTextAreaElement)) {
    return
  }
  event.preventDefault()
  event.stopPropagation()
  const textarea = event.target
  const { selectionStart, selectionEnd } = textarea
  editSummary.value = `${editSummary.value.slice(0, selectionStart)}${tabIndentSpaces}${editSummary.value.slice(selectionEnd)}`
  restoreTextareaSelection(textarea, selectionStart + tabIndentSpaces.length)
}

/**
 * 在素材拆解说明编辑区拦截 Tab，在当前光标位置插入 4 个普通空格。
 */
function handleEditContentKeydown(event: KeyboardEvent) {
  if (event.key !== 'Tab' || !(event.target instanceof HTMLTextAreaElement)) {
    return
  }
  event.preventDefault()
  event.stopPropagation()
  const textarea = event.target
  const { selectionStart, selectionEnd } = textarea
  editContent.value = `${editContent.value.slice(0, selectionStart)}${tabIndentSpaces}${editContent.value.slice(selectionEnd)}`
  restoreTextareaSelection(textarea, selectionStart + tabIndentSpaces.length)
}

function restoreTextareaSelection(textarea: HTMLTextAreaElement, nextPosition: number) {
  void nextTick(() => {
    textarea.focus()
    textarea.setSelectionRange(nextPosition, nextPosition)
  })
}

function materialContentExcerpt(material: library.Material, limit = 96) {
  const content = originalText(material)
  return content.trim().replace(/\s+/g, ' ').slice(0, limit)
}

function firstSourceSummary(material: library.Material) {
  const source = material.sourceRefs[0]
  if (!source) {
    return '暂无来源'
  }
  const chapter = source.chapterRefs[0]?.chapterTitleSnapshot || source.snapshot.chapterTitle
  return [chapter, source.snapshot.deconstructionTitle].filter(Boolean).join(' / ') || source.snapshot.sourceSummary
}

function firstSource(material: library.Material) {
  return material.sourceRefs[0] ?? null
}

function deconstructionTypeName(material: library.Material) {
  const source = firstSource(material)
  return source?.snapshot.planTypeName || splitDeconstructionTitle(source?.snapshot.deconstructionTitle ?? '')[0] || materialTypeLabel(material.materialType)
}

function dimensionName(material: library.Material) {
  const source = firstSource(material)
  return source?.snapshot.dimensionName || splitDeconstructionTitle(source?.snapshot.deconstructionTitle ?? '')[1] || '未记录维度'
}

function originalText(material: library.Material) {
  return material.summary || firstSource(material)?.snapshot.selectedText || ''
}

function explanationText(material: library.Material) {
  return materialContents.value[material.id] || firstSource(material)?.snapshot.sourceSummary || ''
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

function materialTimeValue(material: library.Material) {
  const value = Date.parse(material.createdAt)
  return Number.isFinite(value) ? value : 0
}

function uniqueStringOptions(values: string[], labeler: (value: string) => string = (value) => value) {
  return Array.from(new Set(values.map((value) => value.trim()).filter(Boolean))).map((value) => ({
    label: labeler(value),
    value,
  }))
}

function splitList(value: string) {
  return value
    .split(/[，,]/)
    .map((item) => item.trim())
    .filter(Boolean)
}

watch(activeBookPath, loadWorkspace, { immediate: true })
watch(selectedMaterial, syncDetailEditor)

defineExpose({
  refresh: loadWorkspace,
})
</script>

<template>
  <section class="page-section">
    <n-alert v-if="!activeBookPath" type="info">
      请先从书库打开一本小说。小说内素材会保存在该小说项目目录中。
    </n-alert>

    <n-spin v-else :show="loading">
      <n-alert v-if="errorMessage" type="error" class="materials-alert">{{ errorMessage }}</n-alert>
      <n-alert v-else-if="noticeMessage" type="success" class="materials-alert">{{ noticeMessage }}</n-alert>

      <div class="materials-filter-bar">
        <n-input v-model:value="keyword" clearable placeholder="搜索标题、原文、拆解说明、类型、维度、标签">
          <template #prefix>
            <Search :size="15" />
          </template>
        </n-input>
        <div class="filter-selects">
          <n-select v-model:value="typeFilter" clearable :options="typeOptions" placeholder="类型" />
          <n-select v-model:value="tagFilter" clearable :options="tagOptions" placeholder="标签" />
        </div>
      </div>

      <div class="materials-grid">
        <n-empty v-if="filteredMaterials.length === 0" description="暂无符合条件的素材" />
        <n-card
          v-for="material in filteredMaterials"
          :key="material.id"
          :bordered="false"
          class="item-card material-card"
        >
          <div class="material-card-head">
            <n-space size="small">
              <n-tag type="success" size="small">{{ deconstructionTypeName(material) }}</n-tag>
              <n-tag size="small">{{ dimensionName(material) }}</n-tag>
            </n-space>
          </div>
          <h2>{{ material.title || '未命名素材' }}</h2>
          <p>{{ materialContentExcerpt(material) || '暂无原文' }}</p>
          <div class="source-line">
            <MapPinned :size="15" />
            {{ firstSourceSummary(material) }}
          </div>
          <n-space v-if="material.tagIds.length" size="small" class="material-tags">
            <n-tag v-for="tag in material.tagIds.slice(0, 3)" :key="tag" size="small">{{ tag }}</n-tag>
            <n-tag v-if="material.tagIds.length > 3" size="small">+{{ material.tagIds.length - 3 }}</n-tag>
          </n-space>
          <div class="material-actions">
            <n-button size="small" secondary @click="openDetail(material)">
              <template #icon>
                <Eye :size="15" />
              </template>
              详情
            </n-button>
            <n-button size="small" secondary @click="openEdit(material)">
              <template #icon>
                <Edit3 :size="15" />
              </template>
              编辑
            </n-button>
            <n-popconfirm @positive-click="deleteSelectedMaterial(material)">
              <template #trigger>
                <n-button size="small" type="error" secondary :loading="deleting && selectedMaterialId === material.id">
                  <template #icon>
                    <Trash2 :size="15" />
                  </template>
                  删除
                </n-button>
              </template>
              删除后会移除这条素材和拆解说明 TXT；来源拆解卡片不会被删除。
            </n-popconfirm>
          </div>
        </n-card>
      </div>

      <div class="materials-footer">
        <Library :size="15" />
        共 {{ materials.length }} 条素材，当前显示 {{ filteredMaterials.length }} 条
      </div>
    </n-spin>

    <n-modal v-model:show="detailOpen" preset="card" title="素材详情" class="material-modal" :bordered="false">
      <template v-if="selectedMaterial">
        <div class="detail-meta-grid">
          <div>
            <span>拆解类型</span>
            <strong>{{ deconstructionTypeName(selectedMaterial) }}</strong>
          </div>
          <div>
            <span>维度</span>
            <strong>{{ dimensionName(selectedMaterial) }}</strong>
          </div>
        </div>
        <section class="material-field">
          <strong>原文</strong>
          <p>{{ originalText(selectedMaterial) || '暂无原文' }}</p>
        </section>
        <section class="material-field">
          <strong>拆解说明</strong>
          <p>{{ explanationText(selectedMaterial) || '暂无拆解说明' }}</p>
        </section>
        <section class="material-field">
          <strong>标签</strong>
          <n-space v-if="selectedMaterial.tagIds.length" size="small">
            <n-tag v-for="tag in selectedMaterial.tagIds" :key="tag" size="small">{{ tag }}</n-tag>
          </n-space>
          <span v-else class="empty-text">暂无标签</span>
        </section>
      </template>
      <template #footer>
        <n-space justify="end">
          <n-button @click="detailOpen = false">关闭</n-button>
          <n-button secondary @click="emitBackToSource()">
            <template #icon>
              <BookOpen :size="16" />
            </template>
            回到原文
          </n-button>
          <n-button type="primary" secondary @click="selectedMaterial && openEdit(selectedMaterial)">
            <template #icon>
              <Edit3 :size="16" />
            </template>
            编辑素材
          </n-button>
        </n-space>
      </template>
    </n-modal>

    <n-modal v-model:show="editOpen" preset="card" title="编辑素材" class="material-modal" :bordered="false">
      <template v-if="selectedMaterial">
        <n-form label-placement="top">
          <n-form-item label="标题">
            <n-input v-model:value="editTitle" placeholder="素材标题" />
          </n-form-item>
          <n-form-item label="原文">
            <n-input
              v-model:value="editSummary"
              type="textarea"
              :autosize="{ minRows: 3, maxRows: 5 }"
              @keydown="handleEditSummaryKeydown"
            />
          </n-form-item>
          <n-form-item label="拆解说明">
            <n-input
              v-model:value="editContent"
              type="textarea"
              :autosize="{ minRows: 9, maxRows: 16 }"
              @keydown="handleEditContentKeydown"
            />
          </n-form-item>
          <n-form-item label="标签">
            <n-input v-model:value="editTagInput" placeholder="用逗号分隔，例如：氛围, 重逢" />
          </n-form-item>
        </n-form>
      </template>
      <template #footer>
        <n-space justify="end">
          <n-button @click="editOpen = false">取消</n-button>
          <n-button type="primary" :disabled="!hasDirtyEdit" :loading="saving" @click="saveSelectedMaterial">
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
  grid-template-columns: minmax(260px, 1fr);
  gap: 8px;
  margin-bottom: 16px;
}

.filter-selects {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 8px;
}

.materials-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(292px, 1fr));
  gap: 16px;
}

.material-card {
  min-height: 264px;
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

.materials-footer {
  display: flex;
  gap: 6px;
  align-items: center;
  margin-top: 14px;
  color: var(--text-secondary);
  font-size: 13px;
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

.source-text {
  margin-top: 8px !important;
  color: var(--text-secondary);
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
  .filter-selects,
  .detail-meta-grid,
  .edit-grid {
    grid-template-columns: 1fr;
  }
}
</style>
