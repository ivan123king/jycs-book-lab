<script setup lang="ts">
// StoryPlotPreviewPanel.vue 提供剧情拆解正式页面，负责卡片读写、临时类型维度编辑和章节内容辅助查看。
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
  NScrollbar,
  NSelect,
  NSpace,
  NSpin,
  NTag,
} from 'naive-ui'
import {
  BookOpen,
  ChevronDown,
  ChevronRight,
  Edit3,
  Eye,
  Library,
  MapPinned,
  Plus,
  Save,
  Search,
  Sparkles,
  Trash2,
} from 'lucide-vue-next'
import {
  BuildMaterialDraftFromCard,
  CreateMaterialFromCard,
  CreateStoryPlotCard,
  DeleteStoryPlotCard,
  GetChapterContent,
  OpenStoryPlotWorkspace,
  UpdateStoryPlotCard,
} from '@wails/go/main/App'
import type { library } from '@wails/go/models'
import type { BookProject } from '@/types/library'
import { formatDateTime, getErrorMessage } from '@/utils/format'

type StoryStatus = 'draft' | 'completed'

type StoryChapterRef = {
  chapterId: string
  chapterOrder: number
  chapterTitleSnapshot: string
}

type StoryDimensionDraft = {
  id: string
  name: string
  order: number
  content: string
  collapsed: boolean
}

type StoryTypeDraft = {
  id: string
  name: string
  order: number
  dimensions: StoryDimensionDraft[]
}

type StoryPlotCardDraft = {
  id: string
  status: StoryStatus
  title: string
  summary: string
  chapterRangeInput: string
  chapterRefs: StoryChapterRef[]
  planTypeId: string
  planTypeName: string
  dimensions: StoryDimensionDraft[]
  typeSnapshots: StoryTypeDraft[]
  parseErrors: string[]
  templateFileName: string
  templateName: string
  tagIds: string[]
  materialIds: string[]
  createdAt: string
  updatedAt: string
}

const props = defineProps<{
  project: BookProject | null
}>()

const loading = ref(false)
const saving = ref(false)
const templateName = ref('')
const currentTemplateFileName = ref('')
const templateTypes = ref<StoryTypeDraft[]>([])
const cards = ref<StoryPlotCardDraft[]>([])
const pageError = ref('')
const pageNotice = ref('')
const keyword = ref('')
const typeFilter = ref<string | null>(null)
const chapterFilter = ref<string | null>(null)
const statusFilter = ref<StoryStatus | null>(null)
const editOpen = ref(false)
const detailOpen = ref(false)
const chapterContentOpen = ref(false)
const closeConfirmOpen = ref(false)
const materialOpen = ref(false)
const dialogMode = ref<'create' | 'edit'>('create')
const selectedCardId = ref('')
const draft = ref<StoryPlotCardDraft | null>(null)
const draftBaseline = ref('')
const tagInput = ref('')
const chapterContent = ref('')
const chapterContentError = ref('')
const chapterLoading = ref(false)
const activeChapterId = ref('')
const materialTitle = ref('')
const materialSummary = ref('')
const materialContent = ref('')
const materialTagInput = ref('')
const materialSaving = ref(false)
const tabIndentSpaces = '    '

const activeBookPath = computed(() => props.project?.book.bookPath ?? '')
const chapters = computed(() => props.project?.chapters ?? [])

const chaptersByOrder = computed(() => {
  const map = new Map<number, library.BookChapter>()
  for (const chapter of chapters.value) {
    map.set(chapter.order, chapter)
  }
  return map
})

const selectedCard = computed(() => cards.value.find((card) => card.id === selectedCardId.value) ?? null)

const activeType = computed(() => {
  if (!draft.value) {
    return null
  }
  return draft.value.typeSnapshots.find((item) => item.id === draft.value?.planTypeId) ?? draft.value.typeSnapshots[0] ?? null
})

const typeOptions = computed(() => {
  const seen = new Set<string>()
  return cards.value
    .filter((card) => {
      if (seen.has(card.planTypeId)) {
        return false
      }
      seen.add(card.planTypeId)
      return true
    })
    .map((card) => ({ label: card.planTypeName, value: card.planTypeId }))
})

const chapterOptions = computed(() => {
  return chapters.value.map((chapter) => ({ label: `${chapter.order}. ${chapter.title}`, value: chapter.id }))
})

const statusOptions = [
  { label: '草稿', value: 'draft' },
  { label: '已完成', value: 'completed' },
]

const filteredCards = computed(() => {
  const cleanKeyword = keyword.value.trim().toLowerCase()
  return [...cards.value]
    .sort((left, right) => Date.parse(right.updatedAt) - Date.parse(left.updatedAt))
    .filter((card) => {
      const searchText = [
        card.title,
        card.summary,
        card.chapterRangeInput,
        card.planTypeName,
        ...card.tagIds,
        ...card.dimensions.flatMap((dimension) => [dimension.name, dimension.content]),
        ...card.chapterRefs.map((chapter) => chapter.chapterTitleSnapshot),
      ]
        .join('\n')
        .toLowerCase()
      return (
        (!cleanKeyword || searchText.includes(cleanKeyword)) &&
        (!typeFilter.value || card.planTypeId === typeFilter.value) &&
        (!chapterFilter.value || card.chapterRefs.some((chapter) => chapter.chapterId === chapterFilter.value)) &&
        (!statusFilter.value || card.status === statusFilter.value)
      )
    })
})

/**
 * 读取当前书剧情模板和正式剧情卡片，所有卡片数据来自小说项目目录。
 */
async function loadStoryTemplate() {
  pageError.value = ''
  pageNotice.value = ''
  templateTypes.value = []
  templateName.value = ''
  currentTemplateFileName.value = ''
  cards.value = []
  if (!activeBookPath.value) {
    return
  }
  loading.value = true
  try {
    const workspace = await OpenStoryPlotWorkspace(activeBookPath.value)
    templateName.value = workspace.plan?.templateName || '当前剧情模板'
    currentTemplateFileName.value = workspace.currentTemplateFileName || ''
    templateTypes.value = cloneTemplateTypes(workspace.plan?.types ?? [])
    cards.value = (workspace.storyPlotCards?.items ?? []).map(toStoryPlotCardDraft)
  } catch (error) {
    pageError.value = getErrorMessage(error)
  } finally {
    loading.value = false
  }
}

function openCreateDialog() {
  dialogMode.value = 'create'
  draft.value = createEmptyDraft()
  tagInput.value = ''
  draftBaseline.value = serializeDraftState()
  editOpen.value = true
}

function openEditDialog(card: StoryPlotCardDraft) {
  dialogMode.value = 'edit'
  selectedCardId.value = card.id
  draft.value = cloneCard(card)
  tagInput.value = card.tagIds.join(', ')
  draftBaseline.value = serializeDraftState()
  editOpen.value = true
}

function openDetail(card: StoryPlotCardDraft) {
  selectedCardId.value = card.id
  draft.value = null
  detailOpen.value = true
}

function createEmptyDraft(): StoryPlotCardDraft {
  const now = new Date().toISOString()
  const typeSnapshots = cloneTypeSnapshots(templateTypes.value)
  if (typeSnapshots.length === 0) {
    typeSnapshots.push(createStoryTypeDraft('剧情分析', 1))
  }
  const firstType = typeSnapshots[0]
  return {
    id: '',
    status: 'draft',
    title: '',
    summary: '',
    chapterRangeInput: '',
    chapterRefs: [],
    planTypeId: firstType.id,
    planTypeName: firstType.name,
    dimensions: cloneDimensionDrafts(firstType.dimensions),
    typeSnapshots,
    parseErrors: [],
    templateFileName: currentTemplateFileName.value,
    templateName: templateName.value,
    tagIds: [],
    materialIds: [],
    createdAt: now,
    updatedAt: now,
  }
}

function selectStoryType(typeId: string) {
  if (!draft.value) {
    return
  }
  const nextType = draft.value.typeSnapshots.find((item) => item.id === typeId)
  if (!nextType) {
    return
  }
  draft.value.planTypeId = nextType.id
  draft.value.planTypeName = nextType.name
  draft.value.dimensions = cloneDimensionDrafts(nextType.dimensions)
}

function addStoryType() {
  if (!draft.value) {
    return
  }
  const nextType = createStoryTypeDraft(`剧情类型 ${draft.value.typeSnapshots.length + 1}`, draft.value.typeSnapshots.length + 1)
  draft.value.typeSnapshots.push(nextType)
  selectStoryType(nextType.id)
}

function removeStoryType(typeId: string) {
  if (!draft.value || draft.value.typeSnapshots.length <= 1) {
    pageError.value = '剧情拆解至少需要保留一个类型。'
    return
  }
  draft.value.typeSnapshots = draft.value.typeSnapshots
    .filter((item) => item.id !== typeId)
    .map((item, index) => ({ ...item, order: index + 1 }))
  if (draft.value.planTypeId === typeId) {
    selectStoryType(draft.value.typeSnapshots[0]?.id ?? '')
  }
}

function addDimension() {
  if (!activeType.value) {
    return
  }
  const nextDimension = createDimensionDraft(`剧情维度 ${activeType.value.dimensions.length + 1}`, 1)
  activeType.value.dimensions = [nextDimension, ...activeType.value.dimensions].map((dimension, index) => ({
    ...dimension,
    order: index + 1,
  }))
  draft.value!.dimensions = cloneDimensionDrafts(activeType.value.dimensions)
}

function removeDimension(dimensionId: string) {
  if (!activeType.value || activeType.value.dimensions.length <= 1) {
    pageError.value = '每个剧情类型至少需要保留一个维度。'
    return
  }
  activeType.value.dimensions = activeType.value.dimensions
    .filter((item) => item.id !== dimensionId)
    .map((item, index) => ({ ...item, order: index + 1 }))
  draft.value!.dimensions = cloneDimensionDrafts(activeType.value.dimensions)
}

function syncActiveTypeName() {
  if (!draft.value || !activeType.value) {
    return
  }
  draft.value.planTypeName = activeType.value.name
}

function syncActiveDimensions() {
  if (!draft.value || !activeType.value) {
    return
  }
  draft.value.dimensions = cloneDimensionDrafts(activeType.value.dimensions)
}

function setAllDimensionsCollapsed(collapsed: boolean) {
  if (!activeType.value) {
    return
  }
  activeType.value.dimensions = activeType.value.dimensions.map((dimension) => ({ ...dimension, collapsed }))
  syncActiveDimensions()
}

function toggleDimension(dimension: StoryDimensionDraft) {
  dimension.collapsed = !dimension.collapsed
  syncActiveDimensions()
}

async function saveStoryPlot(status: StoryStatus) {
  if (!draft.value) {
    return
  }
  const parsedRange = parseChapterRange(draft.value.chapterRangeInput)
  if (status === 'completed' && parsedRange.errors.length > 0) {
    pageError.value = parsedRange.errors[0]
    return
  }
  if (status === 'completed' && !draft.value.title.trim()) {
    pageError.value = '正式保存前需要填写剧情名称。'
    return
  }
  saving.value = true
  pageError.value = ''
  pageNotice.value = ''
  try {
    syncActiveTypeName()
    syncActiveDimensions()
    const request = buildSaveRequest(status)
    const result =
      dialogMode.value === 'edit'
        ? await UpdateStoryPlotCard(request as library.SaveStoryPlotCardRequest)
        : await CreateStoryPlotCard(request as library.SaveStoryPlotCardRequest)
    cards.value = (result.storyPlotCards?.items ?? []).map(toStoryPlotCardDraft)
    selectedCardId.value = result.card.id
    editOpen.value = false
    closeConfirmOpen.value = false
    draftBaseline.value = ''
    pageNotice.value = status === 'draft' ? '剧情草稿已保存到当前小说项目目录。' : '剧情已保存到当前小说项目目录。'
  } catch (error) {
    pageError.value = getErrorMessage(error)
  } finally {
    saving.value = false
  }
}

async function deleteStoryPlot(card: StoryPlotCardDraft) {
  if (!activeBookPath.value) {
    return
  }
  pageError.value = ''
  pageNotice.value = ''
  try {
    const result = await DeleteStoryPlotCard({
      bookPath: activeBookPath.value,
      cardId: card.id,
    } as library.DeleteStoryPlotCardRequest)
    cards.value = (result.storyPlotCards?.items ?? []).map(toStoryPlotCardDraft)
    if (selectedCardId.value === card.id) {
      selectedCardId.value = cards.value[0]?.id ?? ''
      detailOpen.value = false
    }
    pageNotice.value = '剧情已删除。'
  } catch (error) {
    pageError.value = getErrorMessage(error)
  }
}

/**
 * 从剧情拆解卡片生成可编辑素材草稿；可复用方法只带入已填写的维度内容。
 */
async function openMaterialDraft() {
  const bookPath = activeBookPath.value
  const card = selectedCard.value
  if (!bookPath || !card) {
    return
  }
  pageError.value = ''
  pageNotice.value = ''
  try {
    const materialDraft = await BuildMaterialDraftFromCard({
      bookPath,
      cardId: card.id,
    } as library.BuildMaterialDraftFromCardRequest)
    materialTitle.value = materialDraft.title
    materialSummary.value = materialDraft.summary
    materialContent.value = materialDraft.content
    materialTagInput.value = materialDraft.tagIds.join(', ')
    materialOpen.value = true
  } catch (error) {
    pageError.value = getErrorMessage(error)
  }
}

/**
 * 保存剧情素材到当前小说项目目录；剧情卡片只记录素材 ID，素材自身保留来源快照。
 */
async function saveMaterialDraft() {
  const bookPath = activeBookPath.value
  const card = selectedCard.value
  if (!bookPath || !card) {
    return
  }
  materialSaving.value = true
  pageError.value = ''
  pageNotice.value = ''
  try {
    await CreateMaterialFromCard({
      bookPath,
      cardId: card.id,
      title: materialTitle.value,
      summary: materialSummary.value,
      content: materialContent.value,
      categoryIds: [],
      tagIds: splitTags(materialTagInput.value),
    } as library.CreateMaterialFromCardRequest)
    materialOpen.value = false
    detailOpen.value = false
    await loadStoryTemplate()
    selectedCardId.value = card.id
    pageNotice.value = '已沉淀为剧情结构素材。'
  } catch (error) {
    pageError.value = getErrorMessage(error)
  } finally {
    materialSaving.value = false
  }
}

function hasUnsavedDraftChanges() {
  return editOpen.value && draft.value !== null && serializeDraftState() !== draftBaseline.value
}

function requestCloseEditDialog() {
  if (hasUnsavedDraftChanges()) {
    closeConfirmOpen.value = true
    return
  }
  closeEditDialog()
}

function handleEditModalShowUpdate(nextShow: boolean) {
  if (nextShow) {
    editOpen.value = true
    return
  }
  requestCloseEditDialog()
}

function closeEditDialog() {
  editOpen.value = false
  closeConfirmOpen.value = false
  draft.value = null
  tagInput.value = ''
  draftBaseline.value = ''
}

function discardEditChanges() {
  closeEditDialog()
}

function parseAndPreviewRange() {
  if (!draft.value) {
    return []
  }
  const parsedRange = parseChapterRange(draft.value.chapterRangeInput)
  draft.value.chapterRefs = parsedRange.chapterRefs
  draft.value.parseErrors = parsedRange.errors
  return parsedRange.errors
}

async function openChapterContentDialog() {
  const errors = parseAndPreviewRange()
  if (errors.length > 0) {
    pageError.value = errors[0]
    return
  }
  chapterContentOpen.value = true
  const firstChapterId = draft.value?.chapterRefs[0]?.chapterId || selectedCard.value?.chapterRefs[0]?.chapterId || ''
  if (firstChapterId) {
    await loadChapterContent(firstChapterId)
  }
}

async function loadChapterContent(chapterId: string) {
  if (!activeBookPath.value) {
    return
  }
  activeChapterId.value = chapterId
  chapterContent.value = ''
  chapterContentError.value = ''
  chapterLoading.value = true
  try {
    const result = await GetChapterContent({ bookPath: activeBookPath.value, chapterId } as library.ChapterContentRequest)
    chapterContent.value = result.content
  } catch (error) {
    chapterContentError.value = getErrorMessage(error)
  } finally {
    chapterLoading.value = false
  }
}

function parseChapterRange(input: string) {
  const errors: string[] = []
  const orderSet = new Set<number>()
  const cleanInput = input.trim()
  if (!cleanInput) {
    return { chapterRefs: [] as StoryChapterRef[], errors: ['请输入章节范围，例如：1-3, 8, 12-15。'] }
  }
  for (const segment of cleanInput.split(/[，,]/)) {
    const cleanSegment = segment.trim()
    if (!cleanSegment) {
      continue
    }
    const match = cleanSegment.match(/^(\d+)(?:\s*-\s*(\d+))?$/)
    if (!match) {
      errors.push(`章节范围 "${cleanSegment}" 格式不正确。`)
      continue
    }
    const startOrder = Number(match[1])
    const endOrder = Number(match[2] ?? match[1])
    if (startOrder > endOrder) {
      errors.push(`章节范围 "${cleanSegment}" 的起始章节不能大于结束章节。`)
      continue
    }
    for (let order = startOrder; order <= endOrder; order += 1) {
      if (!chaptersByOrder.value.has(order)) {
        errors.push(`第 ${order} 章不存在。`)
        continue
      }
      orderSet.add(order)
    }
  }
  const chapterRefs = Array.from(orderSet)
    .sort((left, right) => left - right)
    .map((order) => {
      const chapter = chaptersByOrder.value.get(order)!
      return {
        chapterId: chapter.id,
        chapterOrder: chapter.order,
        chapterTitleSnapshot: chapter.title,
      }
    })
  if (chapterRefs.length === 0 && errors.length === 0) {
    errors.push('请输入有效的章节序号。')
  }
  return { chapterRefs, errors }
}

function chapterRangeLabel(card: StoryPlotCardDraft) {
  if (card.chapterRefs.length === 0) {
    return card.chapterRangeInput || '未设置章节范围'
  }
  const first = card.chapterRefs[0]
  const last = card.chapterRefs[card.chapterRefs.length - 1]
  const prefix =
    first.chapterOrder === last.chapterOrder ? `第 ${first.chapterOrder} 章` : `第 ${first.chapterOrder}-${last.chapterOrder} 章`
  return `${prefix} · 共 ${card.chapterRefs.length} 章`
}

function dimensionSummary(dimension: StoryDimensionDraft) {
  return dimension.content.trim().replace(/\s+/g, ' ').slice(0, 72) || '未填写'
}

function splitTags(value: string) {
  return value
    .split(/[，,]/)
    .map((item) => item.trim())
    .filter(Boolean)
}

function buildSaveRequest(status: StoryStatus) {
  if (!draft.value) {
    throw new Error('剧情草稿不存在')
  }
  return {
    bookPath: activeBookPath.value,
    cardId: dialogMode.value === 'edit' ? draft.value.id : '',
    status,
    title: draft.value.title,
    summary: draft.value.summary,
    chapterRangeInput: draft.value.chapterRangeInput,
    templateFileName: draft.value.templateFileName || currentTemplateFileName.value,
    templateName: draft.value.templateName || templateName.value,
    planTypeId: draft.value.planTypeId,
    planTypeName: draft.value.planTypeName,
    dimensions: stripCollapsedFromDimensions(draft.value.dimensions),
    typeSnapshots: stripCollapsedFromTypes(draft.value.typeSnapshots),
    tagIds: splitTags(tagInput.value),
  }
}

function serializeDraftState() {
  if (!draft.value) {
    return ''
  }
  return JSON.stringify({
    draft: {
      ...draft.value,
      dimensions: stripCollapsedFromDimensions(draft.value.dimensions),
      typeSnapshots: stripCollapsedFromTypes(draft.value.typeSnapshots),
    },
    tagInput: tagInput.value,
  })
}

function handleSummaryTab(event: KeyboardEvent) {
  if (!draft.value) {
    return
  }
  handleTextareaTab(event, draft.value.summary, (nextValue) => {
    draft.value!.summary = nextValue
  })
}

function handleDimensionTab(event: KeyboardEvent, dimension: StoryDimensionDraft) {
  handleTextareaTab(event, dimension.content, (nextValue) => {
    dimension.content = nextValue
    syncActiveDimensions()
  })
}

/**
 * 在剧情素材简介编辑区拦截 Tab，保持多行文本内插入 4 个普通空格。
 */
function handleMaterialSummaryTab(event: KeyboardEvent) {
  handleTextareaTab(event, materialSummary.value, (nextValue) => {
    materialSummary.value = nextValue
  })
}

/**
 * 在剧情素材可复用方法编辑区拦截 Tab，保持用户整理长文本时的缩进习惯。
 */
function handleMaterialContentTab(event: KeyboardEvent) {
  handleTextareaTab(event, materialContent.value, (nextValue) => {
    materialContent.value = nextValue
  })
}

function handleTextareaTab(event: KeyboardEvent, value: string, updateValue: (nextValue: string) => void) {
  if (event.key !== 'Tab' || !(event.target instanceof HTMLTextAreaElement)) {
    return
  }
  event.preventDefault()
  event.stopPropagation()
  const textarea = event.target
  const { selectionStart, selectionEnd } = textarea
  updateValue(`${value.slice(0, selectionStart)}${tabIndentSpaces}${value.slice(selectionEnd)}`)
  const nextPosition = selectionStart + tabIndentSpaces.length
  void nextTick(() => {
    textarea.focus()
    textarea.setSelectionRange(nextPosition, nextPosition)
  })
}

function createStoryTypeDraft(name: string, order: number): StoryTypeDraft {
  return {
    id: `story-type-${Date.now()}-${order}`,
    name,
    order,
    dimensions: [createDimensionDraft('剧情推进', 1)],
  }
}

function createDimensionDraft(name: string, order: number): StoryDimensionDraft {
  return {
    id: `story-dimension-${Date.now()}-${order}`,
    name,
    order,
    content: '',
    collapsed: false,
  }
}

function cloneTemplateTypes(types: library.DeconstructionType[]) {
  return types.map((type, typeIndex) => ({
    id: type.id || `story-type-${typeIndex + 1}`,
    name: type.name || `剧情类型 ${typeIndex + 1}`,
    order: type.order || typeIndex + 1,
    dimensions: (type.dimensions?.length ? type.dimensions : [{ id: `${type.id}-dimension`, name: '剧情推进', order: 1 }]).map(
      (dimension, dimensionIndex) => ({
        id: dimension.id || `story-dimension-${typeIndex + 1}-${dimensionIndex + 1}`,
        name: dimension.name || `剧情维度 ${dimensionIndex + 1}`,
        order: dimension.order || dimensionIndex + 1,
        content: '',
        collapsed: false,
      }),
    ),
  }))
}

function toStoryPlotCardDraft(card: library.StoryPlotCard): StoryPlotCardDraft {
  return {
    id: card.id,
    status: card.status === 'completed' ? 'completed' : 'draft',
    title: card.title,
    summary: card.summary,
    chapterRangeInput: card.chapterRangeInput,
    chapterRefs: (card.chapterRefs ?? []).map((chapter) => ({ ...chapter })),
    planTypeId: card.planTypeId,
    planTypeName: card.planTypeName,
    dimensions: (card.dimensions ?? []).map(toStoryDimensionDraft),
    typeSnapshots: (card.typeSnapshots ?? []).map((type, index) => ({
      id: type.id || `story-type-${index + 1}`,
      name: type.name || `剧情类型 ${index + 1}`,
      order: type.order || index + 1,
      dimensions: (type.dimensions ?? []).map(toStoryDimensionDraft),
    })),
    parseErrors: [...(card.parseErrors ?? [])],
    templateFileName: card.templateFileName,
    templateName: card.templateName,
    tagIds: [...(card.tagIds ?? [])],
    materialIds: [...(card.materialIds ?? [])],
    createdAt: card.createdAt,
    updatedAt: card.updatedAt,
  }
}

function toStoryDimensionDraft(dimension: library.StoryPlotDimension): StoryDimensionDraft {
  return {
    id: dimension.id,
    name: dimension.name,
    order: dimension.order,
    content: dimension.content,
    collapsed: false,
  }
}

function stripCollapsedFromDimensions(dimensions: StoryDimensionDraft[]) {
  return dimensions.map((dimension) => ({
    id: dimension.id,
    name: dimension.name,
    order: dimension.order,
    content: dimension.content,
  }))
}

function stripCollapsedFromTypes(types: StoryTypeDraft[]) {
  return types.map((type) => ({
    id: type.id,
    name: type.name,
    order: type.order,
    dimensions: stripCollapsedFromDimensions(type.dimensions),
  }))
}

function cloneDimensionDrafts(dimensions: StoryDimensionDraft[]) {
  return dimensions.map((dimension) => ({ ...dimension }))
}

function cloneTypeSnapshots(types: StoryTypeDraft[]) {
  return types.map((type) => ({ ...type, dimensions: cloneDimensionDrafts(type.dimensions) }))
}

function cloneCard(card: StoryPlotCardDraft): StoryPlotCardDraft {
  return {
    ...card,
    chapterRefs: card.chapterRefs.map((chapter) => ({ ...chapter })),
    dimensions: cloneDimensionDrafts(card.dimensions),
    typeSnapshots: cloneTypeSnapshots(card.typeSnapshots),
    parseErrors: [...card.parseErrors],
    tagIds: [...card.tagIds],
    materialIds: [...card.materialIds],
  }
}

watch(activeBookPath, loadStoryTemplate, { immediate: true })

defineExpose({
  refresh: loadStoryTemplate,
  openCreate: openCreateDialog,
})
</script>

<template>
  <section class="page-section story-panel">
    <n-alert v-if="!activeBookPath" type="info">
      请先从书库打开一本小说。剧情会围绕该书章节范围整理故事线。
    </n-alert>

    <n-spin v-else :show="loading">
      <n-alert v-if="pageError" type="error" class="cards-alert">{{ pageError }}</n-alert>
      <n-alert v-else-if="pageNotice" type="success" class="cards-alert">{{ pageNotice }}</n-alert>

      <div class="story-template-note">
        当前参考模板：{{ templateName || '未读取到剧情模板' }}；本次增删改类型和维度只作用于当前剧情，不会改写模板。
      </div>

      <div class="cards-filter-bar">
        <n-input v-model:value="keyword" clearable placeholder="搜索剧情名称、简介、维度内容、章节、标签">
          <template #prefix>
            <Search :size="15" />
          </template>
        </n-input>
        <div class="story-filter-selects">
          <n-select v-model:value="chapterFilter" clearable :options="chapterOptions" placeholder="章节" />
          <n-select v-model:value="typeFilter" clearable :options="typeOptions" placeholder="剧情类型" />
          <n-select v-model:value="statusFilter" clearable :options="statusOptions" placeholder="状态" />
        </div>
      </div>

      <div class="cards-grid">
        <n-empty v-if="filteredCards.length === 0" description="暂无剧情卡片，先新建一条故事线" />
        <n-card v-for="card in filteredCards" :key="card.id" :bordered="false" class="item-card writing-card story-card">
          <div class="writing-card-head">
            <n-space size="small">
              <n-tag size="small" type="success">{{ card.planTypeName }}</n-tag>
              <n-tag size="small" :type="card.status === 'completed' ? 'info' : 'warning'">
                {{ card.status === 'completed' ? '已完成' : '草稿' }}
              </n-tag>
            </n-space>
            <span>{{ formatDateTime(card.updatedAt) }}</span>
          </div>
          <h2>{{ card.title || '未命名剧情拆解' }}</h2>
          <div class="source-line">
            <MapPinned :size="15" />
            {{ chapterRangeLabel(card) }}
          </div>
          <div class="card-actions">
            <n-button size="small" secondary @click="openDetail(card)">
              <template #icon>
                <Eye :size="15" />
              </template>
              详情
            </n-button>
            <n-button size="small" secondary @click="openEditDialog(card)">
              <template #icon>
                <Edit3 :size="15" />
              </template>
              编辑
            </n-button>
            <n-popconfirm @positive-click="deleteStoryPlot(card)">
              <template #trigger>
                <n-button size="small" type="error" secondary>
                  <template #icon>
                    <Trash2 :size="15" />
                  </template>
                  删除
                </n-button>
              </template>
              删除这条剧情拆解卡片。
            </n-popconfirm>
          </div>
        </n-card>
      </div>

      <div class="cards-footer">
        <Library :size="15" />
        共 {{ cards.length }} 张剧情卡片，当前显示 {{ filteredCards.length }} 张
      </div>
    </n-spin>

    <n-modal
      :show="editOpen"
      preset="card"
      title="剧情"
      class="story-plot-modal"
      :bordered="false"
      :mask-closable="false"
      @update:show="handleEditModalShowUpdate"
    >
      <template v-if="draft">
        <div class="story-edit-layout">
          <aside class="story-type-panel">
            <div class="panel-title-row">
              <strong>类型</strong>
              <n-button size="small" @click="addStoryType">
                <template #icon>
                  <Plus :size="14" />
                </template>
                新增
              </n-button>
            </div>
            <button
              v-for="type in draft.typeSnapshots"
              :key="type.id"
              class="type-list-item"
              :class="{ active: draft.planTypeId === type.id }"
              @click="selectStoryType(type.id)"
            >
              {{ type.name }}
            </button>
          </aside>

          <div class="story-form-panel">
            <n-form label-placement="top">
              <n-form-item label="章节范围">
                <div class="story-inline-control story-range-control">
                  <n-input v-model:value="draft.chapterRangeInput" placeholder="例如：1-3, 8, 12-15" @blur="parseAndPreviewRange" />
                  <n-button class="chapter-content-button" secondary @click="openChapterContentDialog">
                    <template #icon>
                      <BookOpen :size="16" />
                    </template>
                    章节内容
                  </n-button>
                </div>
              </n-form-item>
              <div v-if="draft.chapterRefs.length" class="story-range-summary">
                已解析 {{ draft.chapterRefs.length }} 个章节，点击“章节内容”按目录快速浏览。
              </div>
              <n-alert v-if="draft.parseErrors.length" type="warning" class="story-range-warning">
                {{ draft.parseErrors[0] }}
              </n-alert>

              <n-form-item label="剧情名称">
                <n-input v-model:value="draft.title" placeholder="例如：第一次误会反转、主线危机救场" />
              </n-form-item>

              <n-form-item label="剧情简介">
                <n-input
                  v-model:value="draft.summary"
                  type="textarea"
                  :autosize="{ minRows: 4, maxRows: 8 }"
                  placeholder="先用自己的话概括这段故事线"
                  @keydown="handleSummaryTab"
                />
              </n-form-item>

              <n-card v-if="activeType" :bordered="false" class="story-type-editor">
                <n-form-item label="当前剧情类型" class="story-type-name-item">
                  <div class="story-inline-control story-type-control">
                    <n-input v-model:value="activeType.name" placeholder="剧情类型名称" @update:value="syncActiveTypeName" />
                    <n-button tertiary type="error" @click="removeStoryType(activeType.id)">
                      <template #icon>
                        <Trash2 :size="15" />
                      </template>
                      删除类型
                    </n-button>
                  </div>
                </n-form-item>
                <div class="panel-title-row dimension-heading">
                  <strong>维度</strong>
                  <n-space size="small">
                    <n-button size="small" secondary @click="setAllDimensionsCollapsed(false)">全部展开</n-button>
                    <n-button size="small" secondary @click="setAllDimensionsCollapsed(true)">全部收起</n-button>
                    <n-button size="small" @click="addDimension">
                      <template #icon>
                        <Plus :size="14" />
                      </template>
                      新增维度
                    </n-button>
                  </n-space>
                </div>

                <div class="story-dimension-list">
                  <div v-for="dimension in activeType.dimensions" :key="dimension.id" class="story-dimension-item">
                    <div class="story-dimension-head">
                      <n-button quaternary circle @click="toggleDimension(dimension)">
                        <template #icon>
                          <ChevronRight v-if="dimension.collapsed" :size="16" />
                          <ChevronDown v-else :size="16" />
                        </template>
                      </n-button>
                      <n-input v-model:value="dimension.name" placeholder="维度名称" @update:value="syncActiveDimensions" />
                      <n-button quaternary type="error" @click="removeDimension(dimension.id)">
                        <template #icon>
                          <Trash2 :size="15" />
                        </template>
                      </n-button>
                    </div>
                    <div v-if="dimension.collapsed" class="dimension-collapsed-summary">
                      {{ dimensionSummary(dimension) }}
                    </div>
                    <n-input
                      v-else
                      v-model:value="dimension.content"
                      type="textarea"
                      :autosize="{ minRows: 4, maxRows: 9 }"
                      placeholder="输入这一维度下的剧情分析"
                      @keydown="(event) => handleDimensionTab(event, dimension)"
                      @update:value="syncActiveDimensions"
                    />
                  </div>
                </div>
              </n-card>

              <n-form-item label="标签">
                <n-input v-model:value="tagInput" placeholder="用逗号分隔，例如：误会反转, 人物关系" />
              </n-form-item>
            </n-form>
          </div>
        </div>
      </template>
      <template #footer>
        <n-space justify="space-between">
          <n-tag :bordered="false" type="info">保存到当前小说项目目录，不改写剧情模板</n-tag>
          <n-space>
            <n-button @click="requestCloseEditDialog">取消</n-button>
            <n-button :loading="saving" secondary @click="saveStoryPlot('draft')">
              <template #icon>
                <Save :size="16" />
              </template>
              保存草稿
            </n-button>
            <n-button type="primary" :loading="saving" @click="saveStoryPlot('completed')">
              <template #icon>
                <Save :size="16" />
              </template>
              保存
            </n-button>
          </n-space>
        </n-space>
      </template>
    </n-modal>

    <n-modal v-model:show="closeConfirmOpen" preset="card" title="保存剧情草稿？" class="story-close-modal" :bordered="false">
      <p class="close-confirm-text">当前剧情拆解有未保存修改，关闭前可以先保存为草稿。</p>
      <template #footer>
        <n-space justify="end">
          <n-button @click="closeConfirmOpen = false">取消</n-button>
          <n-button secondary @click="discardEditChanges">不保存并关闭</n-button>
          <n-button type="primary" :loading="saving" @click="saveStoryPlot('draft')">保存草稿</n-button>
        </n-space>
      </template>
    </n-modal>

    <n-modal v-model:show="detailOpen" preset="card" title="剧情详情" class="story-detail-modal" :bordered="false">
      <template v-if="selectedCard">
        <div class="detail-meta-grid">
          <div>
            <span>章节范围</span>
            <strong>{{ selectedCard.chapterRangeInput || '未设置' }}</strong>
          </div>
          <div>
            <span>剧情类型</span>
            <strong>{{ selectedCard.planTypeName }}</strong>
          </div>
          <div>
            <span>状态</span>
            <strong>{{ selectedCard.status === 'completed' ? '已完成' : '草稿' }}</strong>
          </div>
          <div>
            <span>更新时间</span>
            <strong>{{ formatDateTime(selectedCard.updatedAt) }}</strong>
          </div>
        </div>
        <section class="modal-block">
          <h3>剧情简介</h3>
          <p>{{ selectedCard.summary || '未填写剧情简介' }}</p>
        </section>
        <section class="modal-block">
          <h3>维度拆解</h3>
          <div class="story-detail-dimensions">
            <div v-for="dimension in selectedCard.dimensions" :key="dimension.id">
              <strong>{{ dimension.name }}</strong>
              <p>{{ dimension.content || '未填写' }}</p>
            </div>
          </div>
        </section>
      </template>
      <template #footer>
        <n-space justify="end">
          <n-button @click="detailOpen = false">关闭</n-button>
          <n-button secondary @click="openChapterContentDialog">
            <template #icon>
              <BookOpen :size="16" />
            </template>
            章节内容
          </n-button>
          <n-button type="primary" secondary @click="openMaterialDraft">
            <template #icon>
              <Sparkles :size="16" />
            </template>
            沉淀为素材
          </n-button>
          <n-button secondary @click="openEditDialog(selectedCard!)">
            <template #icon>
              <Edit3 :size="16" />
            </template>
            编辑
          </n-button>
        </n-space>
      </template>
    </n-modal>

    <n-modal v-model:show="chapterContentOpen" preset="card" title="章节内容" class="chapter-content-modal" :bordered="false">
      <div class="chapter-content-layout">
        <aside class="chapter-content-list">
          <button
            v-for="chapter in draft?.chapterRefs ?? selectedCard?.chapterRefs ?? []"
            :key="chapter.chapterId"
            class="chapter-item"
            :class="{ active: activeChapterId === chapter.chapterId }"
            @click="loadChapterContent(chapter.chapterId)"
          >
            <span>{{ chapter.chapterOrder }}. {{ chapter.chapterTitleSnapshot }}</span>
          </button>
        </aside>
        <n-spin :show="chapterLoading">
          <n-alert v-if="chapterContentError" type="error">{{ chapterContentError }}</n-alert>
          <n-scrollbar v-else class="chapter-content-text">
            <pre>{{ chapterContent || '选择左侧章节查看正文。' }}</pre>
          </n-scrollbar>
        </n-spin>
      </div>
    </n-modal>

    <n-modal v-model:show="materialOpen" preset="card" title="沉淀为剧情结构素材" class="story-detail-modal" :bordered="false">
      <template v-if="selectedCard">
        <div class="detail-meta-grid story-material-source-grid">
          <div>
            <span class="detail_label">剧情名称</span>
            <strong>{{ selectedCard.title || '未命名剧情拆解' }}</strong>
          </div>
          <div>
            <span class="detail_label">剧情类型</span>
            <strong>{{ selectedCard.planTypeName || '剧情结构' }}</strong>
          </div>
          <div>
            <span class="detail_label">章节范围</span>
            <strong>{{ chapterRangeLabel(selectedCard) }}</strong>
          </div>
        </div>
        <n-form label-placement="top">
          <n-form-item label="素材标题">
            <n-input v-model:value="materialTitle" placeholder="例如：装逼打脸桥段的四段式推进" />
          </n-form-item>
          <n-form-item label="剧情简介">
            <n-input
              v-model:value="materialSummary"
              type="textarea"
              :autosize="{ minRows: 3, maxRows: 6 }"
              @keydown="handleMaterialSummaryTab"
            />
          </n-form-item>
          <n-form-item label="可复用方法">
            <n-input
              v-model:value="materialContent"
              type="textarea"
              :autosize="{ minRows: 9, maxRows: 16 }"
              @keydown="handleMaterialContentTab"
            />
          </n-form-item>
          <n-form-item label="标签">
            <n-input v-model:value="materialTagInput" placeholder="用逗号分隔，例如：装逼打脸, 爽点, 围观反应" />
          </n-form-item>
        </n-form>
      </template>
      <template #footer>
        <n-space justify="end">
          <n-button @click="materialOpen = false">取消</n-button>
          <n-button type="primary" :loading="materialSaving" @click="saveMaterialDraft">
            <template #icon>
              <Sparkles :size="16" />
            </template>
            保存素材
          </n-button>
        </n-space>
      </template>
    </n-modal>
  </section>
</template>
