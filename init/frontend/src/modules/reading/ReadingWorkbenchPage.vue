<script setup lang="ts">
// ReadingWorkbenchPage.vue 负责当前书阅读、正文编辑、选文来源捕获和方案驱动拆解创建。
import { computed, nextTick, onBeforeUnmount, onMounted, ref, shallowRef, watch, type CSSProperties } from 'vue'
import {
  NAlert,
  NButton,
  NButtonGroup,
  NDrawer,
  NDrawerContent,
  NDropdown,
  NForm,
  NFormItem,
  NInput,
  NInputNumber,
  NModal,
  NScrollbar,
  NSelect,
  NSpace,
  NSpin,
  NTag,
} from 'naive-ui'
import { BookOpen, ChevronLeft, ChevronRight, Copy, Highlighter, List, RotateCcw, Save, StickyNote, Type } from 'lucide-vue-next'
import {
  CreateDeconstructionCard,
  GetChapterContent,
  OpenLastReadingSession,
  OpenReadingSession,
  SaveChapterContent,
  SaveChapterMemo,
  SaveReadingPreferences,
  SaveReadingPosition,
  SplitChapterFromStart,
} from '@wails/go/main/App'
import type { library } from '@wails/go/models'
import type { BookProject } from '@/types/library'
import { getErrorMessage } from '@/utils/format'

const props = defineProps<{
  project: BookProject | null
  sourceFocus?: {
    chapterId: string
    sourceTextId: string
    startOffset: number
    requestKey: number
  } | null
}>()

const emit = defineEmits<{
  readingUpdated: [summary: string]
  projectRestored: [project: BookProject]
}>()

type ReaderMode = 'read' | 'edit'
type CapturedSelection = {
  chapterId: string
  chapterOrder: number
  startOffset: number
  endOffset: number
  selectedText: string
}

const catalogPageSize = 30
const defaultReaderFontSize = 18
const tabIndentSpaces = '    '
const readerMode = ref<ReaderMode>('read')
const session = ref<library.ReadingSession | null>(null)
const currentChapter = ref<library.BookChapter | null>(null)
const chapterContent = ref('')
const currentCharOffset = ref(0)
const loading = ref(false)
const savingPosition = ref(false)
const savingPreferences = ref(false)
const readingError = ref('')
const readingNotice = ref('')
const catalogOpen = ref(false)
const catalogSearch = ref('')
const catalogPage = ref(1)
const chapterContextMenuOpen = ref(false)
const chapterContextMenuX = ref(0)
const chapterContextMenuY = ref(0)
const chapterContextChapter = ref<library.BookChapter | null>(null)
const selectedSource = ref<CapturedSelection | null>(null)
const selectedRange = shallowRef<Range | null>(null)
const selectionToolbarRef = ref<HTMLElement | null>(null)
const selectionToolbarPosition = ref({ left: 0, top: 0 })
const deconstructionOpen = ref(false)
const splitChapterOpen = ref(false)
const splitChapterSaving = ref(false)
const splitChapterError = ref('')
const splitTargetChapter = ref<library.BookChapter | null>(null)
const splitForm = ref({ newTitle: '', newContent: '', remainingContent: '' })
const chapterMemoOpen = ref(false)
const chapterMemoSaving = ref(false)
const chapterMemoError = ref('')
const chapterMemoNotice = ref('')
const chapterMemoTarget = ref<library.BookChapter | null>(null)
const chapterMemoContent = ref('')
const selectedPlanTypeId = ref('')
const selectedDimensionId = ref('')
const deconstructionContent = ref('')
const tagInput = ref('')
const savingCard = ref(false)
const cardError = ref('')
const editTitle = ref('')
const editContent = ref('')
const editSaving = ref(false)
const editNotice = ref('')
const editError = ref('')
const readerFontSize = ref(defaultReaderFontSize)
const readerScrollRef = ref<HTMLElement | null>(null)
const readerContentRef = ref<HTMLElement | null>(null)
const restoringScroll = ref(false)
const scrollSaveTimer = ref<number | null>(null)

const chapters = computed(() => session.value?.project.chapters ?? [])

const activeBookPath = computed(() => props.project?.book.bookPath ?? session.value?.project.book.bookPath ?? '')

const hasReadingTarget = computed(() => Boolean(activeBookPath.value))

const readingProgressText = computed(() => {
  if (!currentChapter.value) {
    return '未打开章节'
  }
  return `${currentChapter.value.title} · ${overallReadingPercent.value}%`
})

const overallReadingPercent = computed(() => {
  if (!currentChapter.value || chapters.value.length === 0) {
    return 0
  }
  const totalWordCount = chapters.value.reduce((sum, chapter) => sum + Math.max(0, chapter.wordCount), 0)
  if (totalWordCount <= 0) {
    return 0
  }
  const visibleEnd = Math.min(chapterContent.value.length, currentCharOffset.value)
  const completedWordCount = chapters.value.reduce((sum, chapter) => {
    if (chapter.id === currentChapter.value?.id) {
      return sum + Math.min(Math.max(chapter.wordCount, chapterContent.value.length), visibleEnd)
    }
    if (chapter.order < (currentChapter.value?.order ?? 0)) {
      return sum + Math.max(0, chapter.wordCount)
    }
    return sum
  }, 0)
  return clampPercent(Math.round((completedWordCount / totalWordCount) * 100))
})

const selectedPlanType = computed(() => {
  const types = session.value?.plan.types ?? []
  return types.find((item) => item.id === selectedPlanTypeId.value) ?? types[0] ?? null
})

const selectedDimension = computed(() => {
  const dimensions = selectedPlanType.value?.dimensions ?? []
  return dimensions.find((item) => item.id === selectedDimensionId.value) ?? dimensions[0] ?? null
})

const planTypeOptions = computed(() =>
  (session.value?.plan.types ?? []).map((type) => ({
    label: type.name,
    value: type.id,
  })),
)

const dimensionOptions = computed(() =>
  (selectedPlanType.value?.dimensions ?? []).map((dimension) => ({
    label: dimension.name,
    value: dimension.id,
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

const chapterMemoByChapter = computed(() => {
  const map = new Map<string, library.ChapterMemo>()
  for (const memo of session.value?.chapterMemos?.items ?? []) {
    if (memo.content.trim()) {
      map.set(memo.chapterId, memo)
    }
  }
  return map
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

const chapterContextMenuOptions = [{ label: '拆分章节', key: 'split' }]

const currentChapterIndex = computed(() => {
  return chapters.value.findIndex((chapter) => chapter.id === currentChapter.value?.id)
})

const previousChapter = computed(() => {
  const index = currentChapterIndex.value
  return index > 0 ? chapters.value[index - 1] : null
})

const nextChapter = computed(() => {
  const index = currentChapterIndex.value
  return index >= 0 && index < chapters.value.length - 1 ? chapters.value[index + 1] : null
})

const editDirty = computed(() => {
  return currentChapter.value ? editTitle.value !== currentChapter.value.title || editContent.value !== chapterContent.value : false
})

const selectionToolbarStyle = computed<CSSProperties>(() => ({
  left: `${selectionToolbarPosition.value.left}px`,
  top: `${selectionToolbarPosition.value.top}px`,
}))

/**
 * 打开当前书阅读会补齐旧项目缺失的阅读数据文件，并恢复上次阅读位置。
 */
async function loadReadingSession() {
  const requestedBookPath = props.project?.book.bookPath ?? ''
  if (requestedBookPath && session.value?.project.book.bookPath === requestedBookPath) {
    return
  }
  if (!requestedBookPath) {
    session.value = null
  }
  loading.value = true
  readingError.value = ''
  readingNotice.value = ''
  try {
    const loaded = requestedBookPath ? await OpenReadingSession(requestedBookPath) : await OpenLastReadingSession()
    if (!loaded) {
      readingNotice.value = '还没有最近阅读记录，请先从书库打开一本小说。'
      return
    }
    session.value = loaded
    currentChapter.value = loaded.currentChapter
    chapterContent.value = loaded.chapterContent
    currentCharOffset.value = clampCharOffset(loaded.state.lastPosition.charOffset)
    readerFontSize.value = normalizeFontSize(loaded.state.preferences?.fontSize)
    selectedPlanTypeId.value = loaded.plan.types[0]?.id ?? ''
    syncSelectedDimension()
    syncEditor()
    await restoreScrollPosition()
    emitReadingSummary()
    if (!requestedBookPath) {
      emit('projectRestored', loaded.project)
    }
  } catch (error) {
    readingError.value = getErrorMessage(error)
  } finally {
    loading.value = false
  }
}

/**
 * 按章节加载整章正文并立即保存阅读位置，目录跳转会共用这条路径。
 */
async function loadChapter(chapterId: string, charOffset = 0) {
  const bookPath = activeBookPath.value
  if (!bookPath) {
    return
  }
  readingError.value = ''
  readingNotice.value = ''
  try {
    const result = await GetChapterContent({ bookPath, chapterId })
    currentChapter.value = result.chapter
    chapterContent.value = result.content
    currentCharOffset.value = clampCharOffset(charOffset)
    syncEditor()
    clearSelection()
    await restoreScrollPosition()
    await persistReadingPosition()
    emitReadingSummary()
  } catch (error) {
    readingError.value = getErrorMessage(error)
  }
}

/**
 * 从写作手法拆解卡片回跳原文时，定位到来源章节和大致字符位置。
 */
async function focusSourceTextFromCard() {
  const focus = props.sourceFocus
  if (!focus || !activeBookPath.value) {
    return
  }
  await loadReadingSession()
  readerMode.value = 'read'
  if (currentChapter.value?.id !== focus.chapterId) {
    await loadChapter(focus.chapterId, focus.startOffset)
  } else {
    currentCharOffset.value = clampCharOffset(focus.startOffset)
    await restoreScrollPosition()
    await persistReadingPosition()
    emitReadingSummary()
  }
  readingNotice.value = '已回到拆解来源位置。'
}

async function persistReadingPosition() {
  const bookPath = activeBookPath.value
  if (!bookPath || !currentChapter.value) {
    return
  }
  savingPosition.value = true
  try {
    const state = await SaveReadingPosition({
      bookPath,
      chapterId: currentChapter.value.id,
      charOffset: currentCharOffset.value,
      paragraphIndex: estimateParagraphIndex(chapterContent.value, currentCharOffset.value),
      renderedPageNumber: 1,
      lastMode: readerMode.value,
      readingProgressPercent: overallReadingPercent.value,
    })
    if (session.value) {
      session.value.state = state
      readerFontSize.value = normalizeFontSize(state.preferences?.fontSize)
    }
  } catch (error) {
    readingError.value = getErrorMessage(error)
  } finally {
    savingPosition.value = false
  }
}

async function persistReadingPreferences() {
  const bookPath = activeBookPath.value
  if (!bookPath) {
    return
  }
  savingPreferences.value = true
  readingError.value = ''
  try {
    const state = await SaveReadingPreferences({
      bookPath,
      fontSize: readerFontSize.value,
    })
    if (session.value) {
      session.value.state = state
      readerFontSize.value = normalizeFontSize(state.preferences?.fontSize)
    }
  } catch (error) {
    readingError.value = getErrorMessage(error)
  } finally {
    savingPreferences.value = false
  }
}

async function updateReaderFontSize(value: number | null) {
  readerFontSize.value = normalizeFontSize(value ?? defaultReaderFontSize)
  await restoreScrollPosition()
  await persistReadingPreferences()
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

function openChapterContextMenu(event: MouseEvent, chapter: library.BookChapter) {
  chapterContextMenuOpen.value = false
  chapterContextChapter.value = chapter
  chapterContextMenuX.value = event.clientX
  chapterContextMenuY.value = event.clientY
  nextTick(() => {
    chapterContextMenuOpen.value = true
  })
}

function closeChapterContextMenu() {
  chapterContextMenuOpen.value = false
}

async function handleChapterContextSelect(key: string) {
  chapterContextMenuOpen.value = false
  if (key !== 'split' || !chapterContextChapter.value) {
    return
  }
  await openSplitChapterDialog(chapterContextChapter.value)
}

async function openSplitChapterDialog(chapter: library.BookChapter) {
  const bookPath = activeBookPath.value
  if (!bookPath) {
    return
  }
  splitChapterError.value = ''
  splitTargetChapter.value = chapter
  splitForm.value = { newTitle: chapter.title, newContent: '', remainingContent: '' }
  try {
    const result = await GetChapterContent({ bookPath, chapterId: chapter.id })
    splitTargetChapter.value = result.chapter
    splitForm.value = {
      newTitle: result.chapter.title,
      newContent: result.content,
      remainingContent: '',
    }
    splitChapterOpen.value = true
  } catch (error) {
    readingError.value = getErrorMessage(error)
  }
}

function openChapterMemoDialog(chapter: library.BookChapter) {
  const memo = chapterMemoByChapter.value.get(chapter.id)
  chapterMemoTarget.value = chapter
  chapterMemoContent.value = memo?.content ?? ''
  chapterMemoError.value = ''
  chapterMemoNotice.value = ''
  chapterMemoOpen.value = true
}

async function saveChapterMemo() {
  const bookPath = activeBookPath.value
  if (!bookPath || !chapterMemoTarget.value) {
    return
  }
  chapterMemoSaving.value = true
  chapterMemoError.value = ''
  chapterMemoNotice.value = ''
  try {
    const result = await SaveChapterMemo({
      bookPath,
      chapterId: chapterMemoTarget.value.id,
      content: chapterMemoContent.value,
    })
    if (session.value) {
      session.value.chapterMemos = result.chapterMemos
    }
    chapterMemoNotice.value = chapterMemoContent.value.trim() ? '小记已保存。' : '小记已清空。'
  } catch (error) {
    chapterMemoError.value = getErrorMessage(error)
  } finally {
    chapterMemoSaving.value = false
  }
}

async function saveSplitChapter() {
  const bookPath = activeBookPath.value
  if (!bookPath || !splitTargetChapter.value) {
    return
  }
  if (!splitForm.value.newTitle.trim()) {
    splitChapterError.value = '请填写新章节标题。'
    return
  }
  if (!splitForm.value.newContent.trim()) {
    splitChapterError.value = '请填写新章节正文。'
    return
  }
  if (!splitForm.value.remainingContent.trim()) {
    splitChapterError.value = '请填写当前章节剩余正文。'
    return
  }
  splitChapterSaving.value = true
  splitChapterError.value = ''
  try {
    const result = await SplitChapterFromStart({
      bookPath,
      chapterId: splitTargetChapter.value.id,
      newTitle: splitForm.value.newTitle,
      newContent: splitForm.value.newContent,
      remainingContent: splitForm.value.remainingContent,
    })
    if (session.value) {
      session.value.project = result.project
      session.value.currentChapter = result.remainingChapter
      session.value.chapterContent = result.remainingContent
    }
    currentChapter.value = result.remainingChapter
    chapterContent.value = result.remainingContent
    currentCharOffset.value = 0
    catalogPage.value = Math.floor((result.remainingChapter.order - 1) / catalogPageSize) + 1
    syncEditor()
    clearSelection()
    splitChapterOpen.value = false
    await restoreScrollPosition()
    await persistReadingPosition()
    emitReadingSummary()
    readingNotice.value = '章节已拆分，新章节已插入到当前章节之前。'
  } catch (error) {
    splitChapterError.value = getErrorMessage(error)
  } finally {
    splitChapterSaving.value = false
  }
}

async function goPreviousChapter() {
  if (!previousChapter.value) {
    return
  }
  await loadChapter(previousChapter.value.id, 0)
}

async function goNextChapter() {
  if (!nextChapter.value) {
    return
  }
  await loadChapter(nextChapter.value.id, 0)
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
  const range = selection?.rangeCount ? selection.getRangeAt(0) : null
  if (!range || !isRangeInsideReader(range)) {
    clearSelection()
    return
  }
  const selectionOffset = range ? getSelectionOffsetInReader(range) : null
  const startOffset = selectionOffset ?? chapterContent.value.indexOf(selectedText)
  const safeStartOffset = startOffset >= 0 ? startOffset : currentCharOffset.value
  selectedRange.value = range.cloneRange()
  selectedSource.value = {
    chapterId: currentChapter.value.id,
    chapterOrder: currentChapter.value.order,
    startOffset: safeStartOffset,
    endOffset: safeStartOffset + selectedText.length,
    selectedText,
  }
  updateSelectionToolbar(range)
  nextTick(() => {
    updateSelectionToolbar(selectedRange.value)
  })
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
  deconstructionContent.value = ''
  tagInput.value = ''
  selectedPlanTypeId.value = selectedPlanTypeId.value || session.value?.plan.types[0]?.id || ''
  syncSelectedDimension()
  deconstructionOpen.value = true
}

async function saveDeconstruction(keepDialogOpen: boolean) {
  const bookPath = activeBookPath.value
  if (!bookPath || !selectedSource.value || !selectedPlanType.value || !selectedDimension.value) {
    return
  }
  if (!deconstructionContent.value.trim()) {
    cardError.value = '请填写你的拆解看法。'
    return
  }
  savingCard.value = true
  cardError.value = ''
  try {
    const result = await CreateDeconstructionCard({
      bookPath,
      selection: selectedSource.value,
      planTypeId: selectedPlanType.value.id,
      dimensionId: selectedDimension.value.id,
      content: deconstructionContent.value,
      fieldValues: {},
      tagIds: splitTags(tagInput.value),
    } as library.CreateDeconstructionRequest)
    if (session.value) {
      session.value.sourceTexts = result.sourceTexts
      session.value.deconstructionCards = result.deconstructionCards
    }
    readingNotice.value = '拆解已保存。'
    deconstructionContent.value = ''
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
  const bookPath = activeBookPath.value
  if (!bookPath || !currentChapter.value) {
    return
  }
  editSaving.value = true
  editError.value = ''
  editNotice.value = ''
  try {
    const result = await SaveChapterContent({
      bookPath,
      chapterId: currentChapter.value.id,
      title: editTitle.value,
      content: editContent.value,
    })
    currentChapter.value = result.chapter
    chapterContent.value = result.content
    currentCharOffset.value = clampCharOffset(currentCharOffset.value)
    updateSessionChapter(result.chapter)
    syncEditor()
    await restoreScrollPosition()
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

/**
 * 在章节正文编辑区拦截 Tab，在当前光标位置插入 4 个普通空格。
 */
function handleChapterContentKeydown(event: KeyboardEvent) {
  handleTextareaTabKeydown(event, editContent.value, (nextValue) => {
    editContent.value = nextValue
    editNotice.value = ''
  })
}

/**
 * 在章节小记编辑区拦截 Tab，保持与章节正文一致的 4 空格缩进体验。
 */
function handleChapterMemoKeydown(event: KeyboardEvent) {
  handleTextareaTabKeydown(event, chapterMemoContent.value, (nextValue) => {
    chapterMemoContent.value = nextValue
    chapterMemoNotice.value = ''
  })
}

/**
 * 在写作手法拆解看法输入区拦截 Tab，避免焦点跳到弹框里的下一个控件。
 */
function handleDeconstructionContentKeydown(event: KeyboardEvent) {
  handleTextareaTabKeydown(event, deconstructionContent.value, (nextValue) => {
    deconstructionContent.value = nextValue
    cardError.value = ''
  })
}

function handleTextareaTabKeydown(
  event: KeyboardEvent,
  content: string,
  updateContent: (nextValue: string) => void,
) {
  if (event.key !== 'Tab' || !(event.target instanceof HTMLTextAreaElement)) {
    return
  }
  event.preventDefault()
  event.stopPropagation()
  insertTabSpacesIntoTextarea(event.target, content, updateContent)
}

function insertTabSpacesIntoTextarea(
  textarea: HTMLTextAreaElement,
  content: string,
  updateContent: (nextValue: string) => void,
) {
  const { selectionStart, selectionEnd } = textarea
  const nextValue = `${content.slice(0, selectionStart)}${tabIndentSpaces}${content.slice(selectionEnd)}`
  updateContent(nextValue)
  const nextPosition = selectionStart + tabIndentSpaces.length
  restoreTextareaSelection(textarea, nextPosition, nextPosition)
}

function restoreTextareaSelection(textarea: HTMLTextAreaElement, selectionStart: number, selectionEnd: number) {
  void nextTick(() => {
    textarea.focus()
    textarea.setSelectionRange(selectionStart, selectionEnd)
  })
}

function syncEditor() {
  editTitle.value = currentChapter.value?.title ?? ''
  editContent.value = chapterContent.value
}

function syncSelectedDimension() {
  selectedDimensionId.value = selectedPlanType.value?.dimensions[0]?.id ?? ''
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
  selectedRange.value = null
  window.getSelection()?.removeAllRanges()
}

function getSelectionOffsetInReader(range: Range) {
  const contentElement = readerContentRef.value
  if (!contentElement || !isRangeInsideReader(range)) {
    return null
  }
  const prefixRange = range.cloneRange()
  prefixRange.selectNodeContents(contentElement)
  prefixRange.setEnd(range.startContainer, range.startOffset)
  return prefixRange.toString().length
}

function isRangeInsideReader(range: Range) {
  const contentElement = readerContentRef.value
  return Boolean(contentElement && contentElement.contains(range.commonAncestorContainer))
}

function getSelectionRect(range: Range) {
  const rects = Array.from(range.getClientRects()).filter((rect) => rect.width > 0 && rect.height > 0)
  if (rects.length === 0) {
    const fallbackRect = range.getBoundingClientRect()
    return fallbackRect.width > 0 && fallbackRect.height > 0 ? fallbackRect : null
  }
  const left = Math.min(...rects.map((rect) => rect.left))
  const top = Math.min(...rects.map((rect) => rect.top))
  const right = Math.max(...rects.map((rect) => rect.right))
  const bottom = Math.max(...rects.map((rect) => rect.bottom))
  return { left, top, right, bottom, width: right - left, height: bottom - top }
}

function updateSelectionToolbar(range: Range | null) {
  const scrollElement = readerScrollRef.value
  if (!scrollElement || !range || !isRangeInsideReader(range)) {
    return
  }
  const selectionRect = getSelectionRect(range)
  if (!selectionRect) {
    return
  }
  const containerRect = scrollElement.getBoundingClientRect()
  const toolbarWidth = selectionToolbarRef.value?.offsetWidth ?? 170
  const toolbarHeight = selectionToolbarRef.value?.offsetHeight ?? 42
  const gap = 10
  const edgeGap = 8
  const minLeft = scrollElement.scrollLeft + toolbarWidth / 2 + edgeGap
  const maxLeft = scrollElement.scrollLeft + scrollElement.clientWidth - toolbarWidth / 2 - edgeGap
  const rawLeft = selectionRect.left + selectionRect.width / 2 - containerRect.left + scrollElement.scrollLeft
  const aboveTop = selectionRect.top - containerRect.top + scrollElement.scrollTop - toolbarHeight - gap
  const belowTop = selectionRect.bottom - containerRect.top + scrollElement.scrollTop + gap
  const rawTop = aboveTop >= scrollElement.scrollTop + edgeGap ? aboveTop : belowTop
  const maxTop = scrollElement.scrollHeight - toolbarHeight - edgeGap
  selectionToolbarPosition.value = {
    left: clampNumber(rawLeft, minLeft, Math.max(minLeft, maxLeft)),
    top: clampNumber(rawTop, edgeGap, Math.max(edgeGap, maxTop)),
  }
}

function handleReaderScroll() {
  if (restoringScroll.value) {
    return
  }
  currentCharOffset.value = estimateCharOffsetFromScroll()
  emitReadingSummary()
  schedulePersistReadingPosition()
}

function estimateCharOffsetFromScroll() {
  const scrollElement = readerScrollRef.value
  if (!scrollElement || chapterContent.value.length === 0) {
    return 0
  }
  const maxScrollTop = Math.max(1, scrollElement.scrollHeight - scrollElement.clientHeight)
  const scrollRatio = scrollElement.scrollTop / maxScrollTop
  return clampCharOffset(Math.round(chapterContent.value.length * scrollRatio))
}

function schedulePersistReadingPosition() {
  if (scrollSaveTimer.value !== null) {
    window.clearTimeout(scrollSaveTimer.value)
  }
  scrollSaveTimer.value = window.setTimeout(() => {
    scrollSaveTimer.value = null
    void persistReadingPosition()
  }, 500)
}

async function restoreScrollPosition() {
  await nextTick()
  const scrollElement = readerScrollRef.value
  if (!scrollElement) {
    return
  }
  restoringScroll.value = true
  const maxScrollTop = Math.max(0, scrollElement.scrollHeight - scrollElement.clientHeight)
  const ratio = chapterContent.value.length > 0 ? currentCharOffset.value / chapterContent.value.length : 0
  scrollElement.scrollTop = maxScrollTop * ratio
  window.setTimeout(() => {
    restoringScroll.value = false
  }, 0)
}

function clampCharOffset(offset: number) {
  if (!Number.isFinite(offset) || offset < 0) {
    return 0
  }
  return Math.min(Math.round(offset), chapterContent.value.length)
}

function normalizeFontSize(value?: number) {
  if (!Number.isFinite(value)) {
    return defaultReaderFontSize
  }
  return Math.min(28, Math.max(14, Math.round(value ?? defaultReaderFontSize)))
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

function clampPercent(value: number) {
  if (!Number.isFinite(value)) {
    return 0
  }
  return Math.min(100, Math.max(0, value))
}

function clampNumber(value: number, minValue: number, maxValue: number) {
  if (!Number.isFinite(value)) {
    return minValue
  }
  return Math.min(maxValue, Math.max(minValue, value))
}

function chapterCardCount(chapterId: string) {
  return deconstructionCountByChapter.value.get(chapterId) ?? 0
}

function hasChapterMemo(chapterId: string) {
  return Boolean(chapterMemoByChapter.value.get(chapterId))
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

watch(
  () => props.sourceFocus?.requestKey,
  () => {
    void focusSourceTextFromCard()
  },
)

watch(readerMode, () => {
  void persistReadingPosition()
})

watch(selectedPlanTypeId, () => {
  syncSelectedDimension()
})

watch(catalogSearch, () => {
  catalogPage.value = 1
})

onMounted(async () => {
  await loadReadingSession()
  await focusSourceTextFromCard()
})

onBeforeUnmount(() => {
  if (scrollSaveTimer.value !== null) {
    window.clearTimeout(scrollSaveTimer.value)
    scrollSaveTimer.value = null
  }
  void persistReadingPosition()
})
</script>

<template>
  <section class="reading-page">
    <n-alert v-if="!hasReadingTarget && !loading" type="info" class="reader-empty-state">
      {{ readingNotice || '请先从书库打开一本小说。阅读位置、拆解卡片和选文来源都会保存在该小说项目目录中。' }}
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
            <div class="reader-font-control">
              <Type :size="16" />
              <n-input-number
                v-model:value="readerFontSize"
                size="small"
                :min="14"
                :max="28"
                :step="1"
                :show-button="true"
                @update:value="updateReaderFontSize"
              />
            </div>
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
            <article
              ref="readerScrollRef"
              class="novel-text"
              :style="{ fontSize: `${readerFontSize}px` }"
              @scroll="handleReaderScroll"
              @mouseup="captureSelection"
              @keyup="captureSelection"
            >
              <div v-if="!chapterContent" class="reader-placeholder">
                当前章节没有正文，可切换到编辑 TAB 补充内容。
              </div>
              <div v-else ref="readerContentRef" class="reader-content-text">{{ chapterContent }}</div>

              <div
                v-if="selectedSource"
                ref="selectionToolbarRef"
                class="floating-toolbar"
                :style="selectionToolbarStyle"
                @mousedown.prevent.stop
                @mouseup.stop
                @click.stop
              >
                <n-button size="small" type="primary" @click="openDeconstructionDialog">
                  <template #icon>
                    <Highlighter :size="14" />
                  </template>
                  创建写作手法拆解
                </n-button>
                <n-button size="small" @click="copySelection">
                  <template #icon>
                    <Copy :size="14" />
                  </template>
                  复制
                </n-button>
              </div>
            </article>
          </div>

          <div class="reader-statusbar">
            <n-button size="small" :disabled="!previousChapter" @click="goPreviousChapter">
              <template #icon>
                <ChevronLeft :size="15" />
              </template>
              上一章
            </n-button>
            <div class="reader-status-summary">
              <span>{{ currentChapter?.title ?? '未打开章节' }}</span>
              <span>阅读位置 {{ currentCharOffset + 1 }} / {{ Math.max(chapterContent.length, 1) }}</span>
              <span v-if="selectedSource">已选 {{ selectedSource.selectedText.length }} 字</span>
            </div>
            <n-button size="small" type="primary" secondary :disabled="!nextChapter" @click="goNextChapter">
              下一章
              <template #icon>
                <ChevronRight :size="15" />
              </template>
            </n-button>
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
                @keydown="handleChapterContentKeydown"
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
          <div class="catalog-list-inner">
            <div
              v-for="chapter in catalogPageChapters"
              :key="chapter.id"
              class="chapter-item"
              :class="{ active: currentChapter?.id === chapter.id }"
              role="button"
              tabindex="0"
              @contextmenu.prevent.stop="openChapterContextMenu($event, chapter)"
              @click="jumpToChapter(chapter)"
              @keydown.enter.prevent="jumpToChapter(chapter)"
              @keydown.space.prevent="jumpToChapter(chapter)"
            >
              <div class="chapter-item-head">
                <span>{{ chapter.title }}</span>
                <n-button
                  size="tiny"
                  quaternary
                  class="chapter-memo-button"
                  :type="hasChapterMemo(chapter.id) ? 'primary' : 'default'"
                  @click.stop="openChapterMemoDialog(chapter)"
                  @keydown.stop
                >
                  小记
                </n-button>
              </div>
              <small>
                第 {{ chapter.order }} 章 · {{ chapterCardCount(chapter.id) }} 张卡片
                <span v-if="hasChapterMemo(chapter.id)"> · 已有小记</span>
                · 右键可拆分
              </small>
            </div>
          </div>
        </n-scrollbar>
        <div class="catalog-actions">
          <n-button :disabled="catalogPage <= 1" @click="catalogPage -= 1">上一目录页</n-button>
          <n-button :disabled="catalogPage >= catalogPageCount" @click="catalogPage += 1">下一目录页</n-button>
        </div>
      </n-drawer-content>
    </n-drawer>

    <n-dropdown
      trigger="manual"
      placement="bottom-start"
      :show="chapterContextMenuOpen"
      :x="chapterContextMenuX"
      :y="chapterContextMenuY"
      :options="chapterContextMenuOptions"
      @select="handleChapterContextSelect"
      @clickoutside="closeChapterContextMenu"
    />

    <n-modal
      v-model:show="chapterMemoOpen"
      preset="card"
      :title="`小记 · ${chapterMemoTarget?.title ?? '未选择章节'}`"
      class="chapter-memo-modal"
      :bordered="false"
    >
      <n-alert v-if="chapterMemoError" type="error" class="reader-inline-alert split-inline-alert">
        {{ chapterMemoError }}
      </n-alert>
      <n-alert v-else-if="chapterMemoNotice" type="success" class="reader-inline-alert split-inline-alert">
        {{ chapterMemoNotice }}
      </n-alert>
      <n-input
        v-model:value="chapterMemoContent"
        type="textarea"
        :autosize="{ minRows: 12, maxRows: 18 }"
        @keydown="handleChapterMemoKeydown"
      />
      <template #footer>
        <n-space justify="end">
          <n-button @click="chapterMemoOpen = false">关闭</n-button>
          <n-button type="primary" :loading="chapterMemoSaving" @click="saveChapterMemo">
            <template #icon>
              <Save :size="16" />
            </template>
            保存小记
          </n-button>
        </n-space>
      </template>
    </n-modal>

    <n-modal
      v-model:show="splitChapterOpen"
      preset="card"
      title="拆分章节"
      class="chapter-split-modal"
      :bordered="false"
    >
      <n-alert type="info" :show-icon="false" class="reader-inline-alert split-inline-alert">
        保存后会把“新章节正文”作为新章节插入到当前章节之前，当前章节保留“剩余正文”。适合把未知章节按顺序拆成 A、B、C。
      </n-alert>
      <n-alert v-if="splitChapterError" type="error" class="reader-inline-alert split-inline-alert">
        {{ splitChapterError }}
      </n-alert>
      <n-form label-placement="top">
        <n-form-item label="新章节标题">
          <n-input v-model:value="splitForm.newTitle" />
        </n-form-item>
        <n-form-item label="新章节正文">
          <n-input
            v-model:value="splitForm.newContent"
            type="textarea"
            :autosize="{ minRows: 8, maxRows: 14 }"
          />
        </n-form-item>
        <n-form-item label="当前章节剩余正文">
          <n-input
            v-model:value="splitForm.remainingContent"
            type="textarea"
            :autosize="{ minRows: 8, maxRows: 14 }"
          />
        </n-form-item>
      </n-form>
      <template #footer>
        <n-space justify="end">
          <n-button @click="splitChapterOpen = false">取消</n-button>
          <n-button type="primary" :loading="splitChapterSaving" @click="saveSplitChapter">
            保存拆分
          </n-button>
        </n-space>
      </template>
    </n-modal>

    <n-modal
      v-model:show="deconstructionOpen"
      preset="card"
      title="创建写作手法拆解"
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
        <n-form-item label="拆解维度">
          <n-select v-model:value="selectedDimensionId" :options="dimensionOptions" />
        </n-form-item>
        <n-form-item label="我的看法 *">
          <n-input
            v-model:value="deconstructionContent"
            type="textarea"
            placeholder="写下你对这段原文的分析，例如表现手法、人物感情或结构作用。"
            :autosize="{ minRows: 5, maxRows: 8 }"
            @keydown="handleDeconstructionContentKeydown"
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
