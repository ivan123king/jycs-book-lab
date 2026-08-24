<script setup lang="ts">
// DeconstructionCardsPage.vue 负责当前书写作手法拆解卡片的本地读取、筛选、编辑和来源追溯。
import { computed, ref, watch } from 'vue'
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
import { BookOpen, Edit3, Eye, Library, MapPinned, Save, Search, Sparkles, Trash2 } from 'lucide-vue-next'
import {
  DeleteDeconstructionCard,
  OpenDeconstructionCardsWorkspace,
  UpdateDeconstructionCard,
} from '@wails/go/main/App'
import type { library } from '@wails/go/models'
import type { BookProject } from '@/types/library'
import { formatDateTime, getErrorMessage } from '@/utils/format'

const props = defineProps<{
  project: BookProject | null
}>()

const emit = defineEmits<{
  backToSource: [payload?: { chapterId: string; sourceTextId: string; startOffset: number }]
}>()

const workspace = ref<library.DeconstructionCardsWorkspace | null>(null)
const loading = ref(false)
const saving = ref(false)
const deleting = ref(false)
const errorMessage = ref('')
const noticeMessage = ref('')
const keyword = ref('')
const typeFilter = ref<string | null>(null)
const dimensionFilter = ref<string | null>(null)
const chapterFilter = ref<string | null>(null)
const tagFilter = ref<string | null>(null)
const selectedCardId = ref('')
const detailOpen = ref(false)
const editOpen = ref(false)
const editPlanTypeId = ref('')
const editDimensionId = ref('')
const editContent = ref('')
const editTagInput = ref('')

const activeBookPath = computed(() => props.project?.book.bookPath ?? '')

const cards = computed(() => workspace.value?.deconstructionCards.items ?? [])

const chaptersById = computed(() => {
  const map = new Map<string, library.BookChapter>()
  for (const chapter of workspace.value?.project.chapters ?? []) {
    map.set(chapter.id, chapter)
  }
  return map
})

const sourceTextsById = computed(() => {
  const map = new Map<string, library.SourceText>()
  for (const sourceText of workspace.value?.sourceTexts.items ?? []) {
    map.set(sourceText.id, sourceText)
  }
  return map
})

const selectedCard = computed(() => cards.value.find((card) => card.id === selectedCardId.value) ?? null)

const selectedSourceText = computed(() => {
  return selectedCard.value ? sourceTextsById.value.get(selectedCard.value.sourceTextId) ?? null : null
})

const selectedChapter = computed(() => {
  return selectedSourceText.value ? chaptersById.value.get(selectedSourceText.value.chapterId) ?? null : null
})

const editPlanType = computed(() => {
  const types = workspace.value?.plan.types ?? []
  return types.find((item) => item.id === editPlanTypeId.value) ?? types[0] ?? null
})

const editPlanTypeOptions = computed(() => {
  return (workspace.value?.plan.types ?? []).map((type) => ({
    label: type.name,
    value: type.id,
  }))
})

const editDimensionOptions = computed(() => {
  return (editPlanType.value?.dimensions ?? []).map((dimension) => ({
    label: dimension.name,
    value: dimension.id,
  }))
})

const sortedCards = computed(() => {
  return [...cards.value].sort((left, right) => {
    return cardTimeValue(right) - cardTimeValue(left)
  })
})

const filteredCards = computed(() => {
  const cleanKeyword = keyword.value.trim().toLowerCase()
  return sortedCards.value.filter((card) => {
    const sourceText = sourceTextsById.value.get(card.sourceTextId)
    const chapter = sourceText ? chaptersById.value.get(sourceText.chapterId) : null
    const hitKeyword =
      !cleanKeyword ||
      card.content.toLowerCase().includes(cleanKeyword) ||
      card.planTypeName.toLowerCase().includes(cleanKeyword) ||
      card.dimensionName.toLowerCase().includes(cleanKeyword) ||
      (sourceText?.selectedText ?? '').toLowerCase().includes(cleanKeyword) ||
      (chapter?.title ?? '').toLowerCase().includes(cleanKeyword) ||
      card.tagIds.some((tag) => tag.toLowerCase().includes(cleanKeyword))
    return (
      hitKeyword &&
      (!typeFilter.value || card.planTypeId === typeFilter.value) &&
      (!dimensionFilter.value || card.dimensionId === dimensionFilter.value) &&
      (!chapterFilter.value || sourceText?.chapterId === chapterFilter.value) &&
      (!tagFilter.value || card.tagIds.includes(tagFilter.value))
    )
  })
})

const typeOptions = computed(() => {
  return uniqueOptions(cards.value, (card) => card.planTypeId, (card) => card.planTypeName)
})

const dimensionOptions = computed(() => {
  return uniqueOptions(cards.value, (card) => card.dimensionId, (card) => card.dimensionName)
})

const chapterOptions = computed(() => {
  const sourceChapterIds = new Set(
    cards.value
      .map((card) => sourceTextsById.value.get(card.sourceTextId)?.chapterId)
      .filter((chapterId): chapterId is string => Boolean(chapterId)),
  )
  return (workspace.value?.project.chapters ?? [])
    .filter((chapter) => sourceChapterIds.has(chapter.id))
    .map((chapter) => ({ label: `${chapter.order}. ${chapter.title}`, value: chapter.id }))
})

const tagOptions = computed(() => {
  const tags = Array.from(new Set(cards.value.flatMap((card) => card.tagIds))).filter(Boolean)
  return tags.map((tag) => ({ label: tag, value: tag }))
})

const hasDirtyEdit = computed(() => {
  const card = selectedCard.value
  if (!card) {
    return false
  }
  return (
    editPlanTypeId.value !== card.planTypeId ||
    editDimensionId.value !== card.dimensionId ||
    editContent.value !== card.content ||
    editTagInput.value !== card.tagIds.join(', ')
  )
})

/**
 * 打开写作手法拆解工作区，所有卡片和来源都来自当前小说项目目录。
 */
async function loadWorkspace() {
  const bookPath = activeBookPath.value
  workspace.value = null
  selectedCardId.value = ''
  editPlanTypeId.value = ''
  editDimensionId.value = ''
  editContent.value = ''
  editTagInput.value = ''
  errorMessage.value = ''
  noticeMessage.value = ''
  if (!bookPath) {
    return
  }
  loading.value = true
  try {
    workspace.value = await OpenDeconstructionCardsWorkspace(bookPath)
    selectedCardId.value = filteredCards.value[0]?.id ?? cards.value[0]?.id ?? ''
    syncDetailEditor()
  } catch (error) {
    errorMessage.value = getErrorMessage(error)
  } finally {
    loading.value = false
  }
}

function openDetail(card: library.DeconstructionCard) {
  selectedCardId.value = card.id
  detailOpen.value = true
}

function openEdit(card: library.DeconstructionCard) {
  selectedCardId.value = card.id
  syncDetailEditor()
  editOpen.value = true
}

function syncDetailEditor() {
  editPlanTypeId.value = selectedCard.value?.planTypeId ?? workspace.value?.plan.types[0]?.id ?? ''
  syncEditDimension(selectedCard.value?.dimensionId)
  editContent.value = selectedCard.value?.content ?? ''
  editTagInput.value = selectedCard.value?.tagIds.join(', ') ?? ''
}

function syncEditDimension(preferredDimensionId = '') {
  const dimensions = editPlanType.value?.dimensions ?? []
  editDimensionId.value =
    dimensions.find((dimension) => dimension.id === preferredDimensionId)?.id ?? dimensions[0]?.id ?? ''
}

async function saveSelectedCard() {
  const bookPath = activeBookPath.value
  const card = selectedCard.value
  if (!bookPath || !card) {
    return
  }
  saving.value = true
  errorMessage.value = ''
  noticeMessage.value = ''
  try {
    const result = await UpdateDeconstructionCard({
      bookPath,
      cardId: card.id,
      planTypeId: editPlanTypeId.value,
      dimensionId: editDimensionId.value,
      content: editContent.value,
      tagIds: splitTags(editTagInput.value),
    } as library.UpdateDeconstructionCardRequest)
    if (workspace.value) {
      workspace.value.deconstructionCards = result.deconstructionCards
    }
    selectedCardId.value = result.card.id
    syncDetailEditor()
    editOpen.value = false
    noticeMessage.value = '写作手法拆解已保存。'
  } catch (error) {
    errorMessage.value = getErrorMessage(error)
  } finally {
    saving.value = false
  }
}

async function deleteCard(card: library.DeconstructionCard) {
  const bookPath = activeBookPath.value
  if (!bookPath || !card) {
    return
  }
  selectedCardId.value = card.id
  deleting.value = true
  errorMessage.value = ''
  noticeMessage.value = ''
  try {
    const result = await DeleteDeconstructionCard({
      bookPath,
      cardId: card.id,
    } as library.DeleteDeconstructionCardRequest)
    if (workspace.value) {
      workspace.value.sourceTexts = result.sourceTexts
      workspace.value.deconstructionCards = result.deconstructionCards
    }
    selectedCardId.value = filteredCards.value[0]?.id ?? cards.value[0]?.id ?? ''
    syncDetailEditor()
    detailOpen.value = false
    editOpen.value = false
    noticeMessage.value = '写作手法拆解已删除。'
  } catch (error) {
    errorMessage.value = getErrorMessage(error)
  } finally {
    deleting.value = false
  }
}

function emitBackToSource() {
  if (!selectedSourceText.value) {
    emit('backToSource')
    return
  }
  emit('backToSource', {
    chapterId: selectedSourceText.value.chapterId,
    sourceTextId: selectedSourceText.value.id,
    startOffset: selectedSourceText.value.startOffset,
  })
}

function showMaterialNotice() {
  detailOpen.value = false
  noticeMessage.value = '小说内素材模块尚未实现，写作手法拆解已保留沉淀入口。'
}

function uniqueOptions<T>(items: T[], getValue: (item: T) => string, getLabel: (item: T) => string) {
  const seen = new Set<string>()
  const options: Array<{ label: string; value: string }> = []
  for (const item of items) {
    const value = getValue(item)
    if (!value || seen.has(value)) {
      continue
    }
    seen.add(value)
    options.push({ label: getLabel(item), value })
  }
  return options
}

function sourceSummary(card: library.DeconstructionCard) {
  const sourceText = sourceTextsById.value.get(card.sourceTextId)
  const chapter = sourceText ? chaptersById.value.get(sourceText.chapterId) : null
  if (!sourceText || !chapter) {
    return '来源已缺失'
  }
  return `${chapter.title} · 字符 ${sourceText.startOffset}-${sourceText.endOffset}`
}

function cardExcerpt(card: library.DeconstructionCard) {
  return card.content.trim().replace(/\s+/g, ' ').slice(0, 92)
}

function cardTimeValue(card: library.DeconstructionCard) {
  const value = Date.parse(card.updatedAt || card.createdAt)
  return Number.isFinite(value) ? value : 0
}

function splitTags(value: string) {
  return value
    .split(/[，,]/)
    .map((item) => item.trim())
    .filter(Boolean)
}

watch(activeBookPath, loadWorkspace, { immediate: true })

watch(selectedCard, syncDetailEditor)

watch(editPlanTypeId, () => {
  syncEditDimension(editDimensionId.value)
})
</script>

<template>
  <section class="page-section">
    <div class="section-heading">
      <div>
        <h1>写作手法拆解</h1>
        <p>管理当前小说从原文选段沉淀出的写作手法卡片，并保留来源追溯。</p>
      </div>
      <n-space>
        <n-button @click="loadWorkspace">
          <template #icon>
            <Search :size="16" />
          </template>
          刷新
        </n-button>
      </n-space>
    </div>

    <n-alert v-if="!activeBookPath" type="info">
      请先从书库打开一本小说。写作手法拆解卡片会保存在该小说项目目录中。
    </n-alert>

    <n-spin v-else :show="loading">
      <n-alert v-if="errorMessage" type="error" class="cards-alert">{{ errorMessage }}</n-alert>
      <n-alert v-else-if="noticeMessage" type="success" class="cards-alert">{{ noticeMessage }}</n-alert>

      <div class="cards-filter-bar">
        <n-input v-model:value="keyword" clearable placeholder="搜索拆解内容、原文、章节、标签">
          <template #prefix>
            <Search :size="15" />
          </template>
        </n-input>
        <div class="filter-selects">
          <n-select v-model:value="chapterFilter" clearable :options="chapterOptions" placeholder="章节" />
          <n-select v-model:value="typeFilter" clearable :options="typeOptions" placeholder="类型" />
          <n-select v-model:value="dimensionFilter" clearable :options="dimensionOptions" placeholder="维度" />
          <n-select v-model:value="tagFilter" clearable :options="tagOptions" placeholder="标签" />
        </div>
      </div>

      <div class="cards-grid">
        <n-empty v-if="filteredCards.length === 0" description="暂无符合条件的写作手法拆解卡片" />
        <n-card
          v-for="card in filteredCards"
          :key="card.id"
          :bordered="false"
          class="item-card writing-card"
        >
          <div class="writing-card-head">
            <n-space size="small">
              <n-tag size="small" type="info">{{ card.planTypeName }}</n-tag>
              <n-tag size="small">{{ card.dimensionName }}</n-tag>
            </n-space>
            <span>{{ formatDateTime(card.updatedAt) }}</span>
          </div>
          <h2>{{ cardExcerpt(card) || '未填写拆解内容' }}</h2>
          <div class="source-line">
            <MapPinned :size="15" />
            {{ sourceSummary(card) }}
          </div>
          <n-space v-if="card.tagIds.length" size="small" class="card-tags">
            <n-tag v-for="tag in card.tagIds.slice(0, 3)" :key="tag" size="small">{{ tag }}</n-tag>
            <n-tag v-if="card.tagIds.length > 3" size="small">+{{ card.tagIds.length - 3 }}</n-tag>
          </n-space>
          <div class="card-actions">
            <n-button size="small" secondary @click="openDetail(card)">
              <template #icon>
                <Eye :size="15" />
              </template>
              详情
            </n-button>
            <n-button size="small" secondary @click="openEdit(card)">
              <template #icon>
                <Edit3 :size="15" />
              </template>
              编辑
            </n-button>
            <n-popconfirm @positive-click="deleteCard(card)">
              <template #trigger>
                <n-button size="small" type="error" secondary :loading="deleting && selectedCardId === card.id">
                  <template #icon>
                    <Trash2 :size="15" />
                  </template>
                  删除
                </n-button>
              </template>
              删除后会移除这张卡片；若来源选文没有被其他卡片引用，也会同步清理。
            </n-popconfirm>
          </div>
        </n-card>
      </div>

      <div class="cards-footer">
        <Library :size="15" />
        共 {{ cards.length }} 张写作手法拆解卡片，当前显示 {{ filteredCards.length }} 张
      </div>
    </n-spin>

    <n-modal
      v-model:show="detailOpen"
      preset="card"
      title="写作手法拆解详情"
      class="writing-card-modal"
      :bordered="false"
    >
      <template v-if="selectedCard">
        <div class="detail-meta-grid">
          <div>
            <span>小说章节</span>
            <strong>{{ selectedChapter?.title ?? '来源章节已缺失' }}</strong>
          </div>
          <div>
            <span>更新时间</span>
            <strong>{{ formatDateTime(selectedCard.updatedAt) }}</strong>
          </div>
          <div>
            <span>拆解类型</span>
            <strong>{{ selectedCard.planTypeName }}</strong>
          </div>
          <div>
            <span>拆解维度</span>
            <strong>{{ selectedCard.dimensionName }}</strong>
          </div>
        </div>
        <section class="modal-block">
          <h3>拆解内容</h3>
          <p>{{ selectedCard.content }}</p>
        </section>
        <section class="modal-block">
          <h3>标签</h3>
          <n-space v-if="selectedCard.tagIds.length" size="small">
            <n-tag v-for="tag in selectedCard.tagIds" :key="tag" size="small">{{ tag }}</n-tag>
          </n-space>
          <span v-else class="empty-text">暂无标签</span>
        </section>
      </template>
      <template #footer>
        <n-space justify="end">
          <n-button @click="detailOpen = false">关闭</n-button>
          <n-button secondary @click="emitBackToSource">
            <template #icon>
              <BookOpen :size="16" />
            </template>
            回到原文
          </n-button>
          <n-button type="primary" secondary @click="showMaterialNotice">
            <template #icon>
              <Sparkles :size="16" />
            </template>
            沉淀为素材
          </n-button>
        </n-space>
      </template>
    </n-modal>

    <n-modal
      v-model:show="editOpen"
      preset="card"
      title="编辑写作手法拆解"
      class="writing-card-modal"
      :bordered="false"
    >
      <template v-if="selectedCard">
        <div class="detail-meta-grid edit-meta-grid">
          <div>
            <span>小说章节</span>
            <strong>{{ selectedChapter?.title ?? '来源章节已缺失' }}</strong>
          </div>
        </div>
        <n-form label-placement="top">
          <n-form-item label="拆解类型">
            <n-select v-model:value="editPlanTypeId" :options="editPlanTypeOptions" />
          </n-form-item>
          <n-form-item label="拆解维度">
            <n-select v-model:value="editDimensionId" :options="editDimensionOptions" />
          </n-form-item>
          <n-form-item label="拆解内容">
            <n-input v-model:value="editContent" type="textarea" :autosize="{ minRows: 8, maxRows: 14 }" />
          </n-form-item>
          <n-form-item label="标签">
            <n-input v-model:value="editTagInput" placeholder="用逗号分隔，例如：氛围, 重逢" />
          </n-form-item>
        </n-form>
      </template>
      <template #footer>
        <n-space justify="end">
          <n-button @click="editOpen = false">取消</n-button>
          <n-button type="primary" :disabled="!hasDirtyEdit" :loading="saving" @click="saveSelectedCard">
            <template #icon>
              <Save :size="16" />
            </template>
            保存修改
          </n-button>
        </n-space>
      </template>
    </n-modal>
  </section>
</template>

<style scoped>
.cards-alert {
  margin-bottom: 14px;
}

.cards-filter-bar {
  display: grid;
  grid-template-columns: minmax(260px, 1fr);
  gap: 8px;
  margin-bottom: 16px;
}

.filter-selects {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 8px;
}

.cards-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(292px, 1fr));
  gap: 16px;
}

.writing-card {
  min-height: 250px;
  border: 1px solid transparent;
  transition: border-color 0.16s ease, transform 0.16s ease;
}

.writing-card:hover {
  border-color: rgba(42, 110, 245, 0.36);
  transform: translateY(-1px);
}

.writing-card :deep(.n-card__content) {
  display: flex;
  flex-direction: column;
  height: 100%;
}

.writing-card-head {
  display: flex;
  gap: 10px;
  align-items: flex-start;
  justify-content: space-between;
}

.writing-card-head span {
  flex: 0 0 auto;
  color: var(--text-secondary);
  font-size: 12px;
}

.writing-card h2 {
  display: -webkit-box;
  min-height: 50px;
  margin: 12px 0 10px;
  overflow: hidden;
  font-size: 16px;
  line-height: 1.45;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
}

.card-tags {
  min-height: 24px;
  margin-top: 10px;
}

.card-actions {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 10px;
  margin-top: auto;
  padding-top: 14px;
}

.card-actions :deep(.n-button) {
  min-width: 0;
}

.detail-meta-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px;
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

.edit-meta-grid {
  grid-template-columns: 1fr 1fr;
}

.modal-block {
  margin-top: 14px;
}

.modal-block h3 {
  margin: 0 0 8px;
  font-size: 14px;
}

.modal-block p {
  margin: 0;
  white-space: pre-wrap;
}

.empty-text {
  color: var(--text-secondary);
  font-size: 13px;
}

.cards-footer {
  display: flex;
  gap: 6px;
  align-items: center;
  margin-top: 14px;
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
  .edit-meta-grid {
    grid-template-columns: 1fr;
  }
}
</style>
