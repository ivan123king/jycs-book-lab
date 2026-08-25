<script setup lang="ts">
// AnalysisPage.vue 读取当前书正式数据并展示第一阶段轻量统计，不扩展到跨书比较。
import { computed, ref, watch } from 'vue'
import {
  NAlert,
  NButton,
  NCard,
  NEmpty,
  NGi,
  NGrid,
  NPagination,
  NProgress,
  NSpace,
  NSpin,
  NStatistic,
  NTag,
} from 'naive-ui'
import { RefreshCw } from 'lucide-vue-next'
import {
  OpenDeconstructionCardsWorkspace,
  OpenMaterialWorkspace,
  OpenStoryPlotWorkspace,
} from '@wails/go/main/App'
import type { library } from '@wails/go/models'
import type { AnalysisBreakdownRow, AnalysisChapterRow } from '@/types/analysis'
import type { BookProject } from '@/types/library'
import { formatDateTime, getErrorMessage } from '@/utils/format'

const props = defineProps<{
  project: BookProject | null
}>()

const writingWorkspace = ref<library.DeconstructionCardsWorkspace | null>(null)
const storyWorkspace = ref<library.StoryPlotWorkspace | null>(null)
const materialWorkspace = ref<library.MaterialWorkspace | null>(null)
const loading = ref(false)
const errorMessage = ref('')
const chapterPage = ref(1)
const chapterPageSize = ref(50)

const activeBookPath = computed(() => props.project?.book.bookPath ?? '')
const chapters = computed(() => props.project?.chapters ?? [])
const writingCards = computed(() => writingWorkspace.value?.deconstructionCards.items ?? [])
const sourceTexts = computed(() => writingWorkspace.value?.sourceTexts.items ?? [])
const storyCards = computed(() => storyWorkspace.value?.storyPlotCards.items ?? [])
const materials = computed(() => materialWorkspace.value?.materials.items ?? [])

const sourceTextsById = computed(() => {
  const map = new Map<string, library.SourceText>()
  for (const sourceText of sourceTexts.value) {
    map.set(sourceText.id, sourceText)
  }
  return map
})

const typeRows = computed(() => {
  const counts = new Map<string, number>()
  for (const card of writingCards.value) {
    addCount(counts, `写作手法 / ${card.planTypeName || '未记录类型'}`)
  }
  for (const card of storyCards.value) {
    addCount(counts, `剧情拆解 / ${card.planTypeName || '未记录类型'}`)
  }
  return buildBreakdownRows(counts, writingCards.value.length + storyCards.value.length)
})

const tagRows = computed(() => {
  const counts = new Map<string, number>()
  for (const card of writingCards.value) {
    for (const tag of card.tagIds ?? []) {
      addCount(counts, tag)
    }
  }
  for (const card of storyCards.value) {
    for (const tag of card.tagIds ?? []) {
      addCount(counts, tag)
    }
  }
  for (const material of materials.value) {
    for (const tag of material.tagIds ?? []) {
      addCount(counts, tag)
    }
  }
  return buildBreakdownRows(counts, Array.from(counts.values()).reduce((sum, count) => sum + count, 0), 8)
})

const chapterRows = computed(() => {
  const rowMap = new Map<string, AnalysisChapterRow>()
  for (const chapter of chapters.value) {
    rowMap.set(chapter.id, {
      chapterId: chapter.id,
      title: `${chapter.order}. ${chapter.title}`,
      order: chapter.order,
      writingCount: 0,
      storyCount: 0,
      totalCount: 0,
      percent: 0,
    })
  }

  // 写作手法拆解通过 sourceTextId 回到原文片段，再定位到章节。
  for (const card of writingCards.value) {
    const sourceText = sourceTextsById.value.get(card.sourceTextId)
    if (sourceText) {
      incrementChapterRow(rowMap, sourceText.chapterId, 'writing')
    }
  }

  // 剧情拆解以章节范围为来源，一张剧情卡片可以同时计入多个章节。
  for (const card of storyCards.value) {
    const chapterIds = new Set((card.chapterRefs ?? []).map((chapter) => chapter.chapterId).filter(Boolean))
    for (const chapterId of chapterIds) {
      incrementChapterRow(rowMap, chapterId, 'story')
    }
  }

  const rows = Array.from(rowMap.values()).filter((row) => row.totalCount > 0)
  const maxCount = Math.max(...rows.map((row) => row.totalCount), 0)
  return rows
    .map((row) => ({
      ...row,
      percent: maxCount > 0 ? Math.max(4, Math.round((row.totalCount / maxCount) * 100)) : 0,
    }))
    .sort((left, right) => left.order - right.order)
})

const pagedChapterRows = computed(() => {
  const start = (chapterPage.value - 1) * chapterPageSize.value
  return chapterRows.value.slice(start, start + chapterPageSize.value)
})

const deconstructedChapterCount = computed(() => chapterRows.value.length)
const totalDeconstructionCount = computed(() => writingCards.value.length + storyCards.value.length)
const latestActivityText = computed(() => {
  const values = [
    ...writingCards.value.map((card) => card.updatedAt || card.createdAt),
    ...storyCards.value.map((card) => card.updatedAt || card.createdAt),
    ...materials.value.map((material) => material.updatedAt || material.createdAt),
  ]
  const latest = values
    .map((value) => Date.parse(value))
    .filter((value) => Number.isFinite(value))
    .sort((left, right) => right - left)[0]
  return latest ? formatDateTime(new Date(latest).toISOString()) : '暂无记录'
})

/**
 * 读取当前书的拆解、剧情和素材工作区，基础分析页只做聚合计算，不写入业务数据。
 */
async function loadAnalysisData() {
  const bookPath = activeBookPath.value
  writingWorkspace.value = null
  storyWorkspace.value = null
  materialWorkspace.value = null
  errorMessage.value = ''
  if (!bookPath) {
    return
  }

  loading.value = true
  try {
    // 两个拆解工作区都会补齐当前书的拆解方案引用文件；Windows 下并发读写同一 JSON 容易触发文件占用。
    writingWorkspace.value = await OpenDeconstructionCardsWorkspace(bookPath)
    storyWorkspace.value = await OpenStoryPlotWorkspace(bookPath)
    materialWorkspace.value = await OpenMaterialWorkspace(bookPath)
  } catch (error) {
    errorMessage.value = getErrorMessage(error)
  } finally {
    loading.value = false
  }
}

function addCount(counts: Map<string, number>, rawLabel: string) {
  const label = rawLabel.trim()
  if (!label) {
    return
  }
  counts.set(label, (counts.get(label) ?? 0) + 1)
}

function buildBreakdownRows(counts: Map<string, number>, total: number, limit = 12): AnalysisBreakdownRow[] {
  return Array.from(counts.entries())
    .map(([label, count]) => ({
      label,
      count,
      percent: total > 0 ? Math.round((count / total) * 100) : 0,
    }))
    .sort((left, right) => right.count - left.count || left.label.localeCompare(right.label, 'zh-CN'))
    .slice(0, limit)
}

function incrementChapterRow(rowMap: Map<string, AnalysisChapterRow>, chapterId: string, kind: 'writing' | 'story') {
  const row = rowMap.get(chapterId)
  if (!row) {
    return
  }
  if (kind === 'story') {
    row.storyCount += 1
  } else {
    row.writingCount += 1
  }
  row.totalCount += 1
}

watch(activeBookPath, loadAnalysisData, { immediate: true })

watch([chapterRows, chapterPageSize], () => {
  const maxPage = Math.max(1, Math.ceil(chapterRows.value.length / chapterPageSize.value))
  if (chapterPage.value > maxPage) {
    chapterPage.value = maxPage
  }
})

watch(activeBookPath, () => {
  chapterPage.value = 1
})

defineExpose({
  refresh: loadAnalysisData,
})
</script>

<template>
  <section class="page-section">
    <n-alert v-if="!activeBookPath" type="info">
      请先从书库打开一本小说。分析页只统计当前小说项目目录中的拆解、素材和章节数据。
    </n-alert>

    <n-spin v-else :show="loading">
      <n-alert v-if="errorMessage" type="error" class="analysis-alert">{{ errorMessage }}</n-alert>

      <div class="analysis-toolbar">
        <p>第一阶段只展示当前书的轻量统计。</p>
        <n-space align="center">
          <n-tag type="default">不含跨书比较</n-tag>
          <n-button secondary :loading="loading" @click="loadAnalysisData">
            <template #icon>
              <RefreshCw :size="16" />
            </template>
            刷新
          </n-button>
        </n-space>
      </div>

      <n-grid :cols="4" :x-gap="16" :y-gap="16" responsive="screen">
        <n-gi>
          <n-card :bordered="false"><n-statistic label="章节" :value="chapters.length" /></n-card>
        </n-gi>
        <n-gi>
          <n-card :bordered="false"><n-statistic label="拆解卡片" :value="totalDeconstructionCount" /></n-card>
        </n-gi>
        <n-gi>
          <n-card :bordered="false"><n-statistic label="素材" :value="materials.length" /></n-card>
        </n-gi>
        <n-gi>
          <n-card :bordered="false"><n-statistic label="已拆章节" :value="deconstructedChapterCount" /></n-card>
        </n-gi>
      </n-grid>

      <div class="analysis-grid">
        <n-card :bordered="false" title="按拆解类型">
          <n-empty v-if="typeRows.length === 0" description="暂无拆解卡片" />
          <div v-for="row in typeRows" :key="row.label" class="progress-row">
            <div>
              <span>{{ row.label }}</span>
              <strong>{{ row.count }}</strong>
            </div>
            <n-progress type="line" :percentage="row.percent" :height="8" />
          </div>
        </n-card>

        <n-card :bordered="false" title="按标签">
          <n-empty v-if="tagRows.length === 0" description="暂无标签数据" />
          <div v-for="row in tagRows" :key="row.label" class="progress-row">
            <div>
              <span>{{ row.label }}</span>
              <strong>{{ row.count }}</strong>
            </div>
            <n-progress type="line" :percentage="row.percent" :height="8" />
          </div>
        </n-card>

        <n-card :bordered="false" title="章节拆解分布" class="analysis-wide-card">
          <n-empty v-if="chapterRows.length === 0" description="暂无章节拆解分布" />
          <template v-else>
            <div v-for="chapter in pagedChapterRows" :key="chapter.chapterId" class="chapter-stat-row">
              <span>{{ chapter.title }}</span>
              <n-progress type="line" :percentage="chapter.percent" :height="8" />
              <strong>{{ chapter.totalCount }}</strong>
              <small>写作 {{ chapter.writingCount }} / 剧情 {{ chapter.storyCount }}</small>
            </div>
            <div class="analysis-pagination">
              <span>
                共 {{ chapterRows.length }} 个已拆章节，当前显示
                {{ (chapterPage - 1) * chapterPageSize + 1 }} -
                {{ Math.min(chapterPage * chapterPageSize, chapterRows.length) }}
              </span>
              <n-pagination
                v-model:page="chapterPage"
                v-model:page-size="chapterPageSize"
                :item-count="chapterRows.length"
                :page-sizes="[30, 50, 100]"
                show-size-picker
              />
            </div>
          </template>
        </n-card>

        <n-card :bordered="false" title="最近进展" class="analysis-wide-card">
          <div class="analysis-summary-grid">
            <div>
              <span>写作手法拆解</span>
              <strong>{{ writingCards.length }}</strong>
            </div>
            <div>
              <span>剧情拆解</span>
              <strong>{{ storyCards.length }}</strong>
            </div>
            <div>
              <span>最近更新时间</span>
              <strong>{{ latestActivityText }}</strong>
            </div>
          </div>
        </n-card>
      </div>
    </n-spin>
  </section>
</template>

<style scoped>
.analysis-alert {
  margin-bottom: 14px;
}

.analysis-toolbar {
  display: flex;
  gap: 16px;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}

.analysis-toolbar p {
  margin: 0;
  color: var(--text-secondary);
}

.analysis-wide-card {
  grid-column: 1 / -1;
}

.chapter-stat-row {
  grid-template-columns: minmax(180px, 1.4fr) minmax(160px, 2fr) 52px minmax(112px, auto);
}

.chapter-stat-row small {
  color: var(--text-secondary);
  font-size: 12px;
  white-space: nowrap;
}

.analysis-pagination {
  display: flex;
  gap: 12px;
  align-items: center;
  justify-content: space-between;
  margin-top: 16px;
  color: var(--text-secondary);
  font-size: 13px;
}

.analysis-summary-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 12px;
}

.analysis-summary-grid div {
  padding: 12px 14px;
  background: rgba(15, 23, 42, 0.04);
  border-radius: 8px;
}

.analysis-summary-grid span {
  display: block;
  margin-bottom: 8px;
  color: var(--text-secondary);
  font-size: 13px;
}

.analysis-summary-grid strong {
  display: block;
  overflow-wrap: anywhere;
  font-size: 18px;
}

@media (max-width: 720px) {
  .analysis-toolbar,
  .analysis-pagination {
    align-items: flex-start;
    flex-direction: column;
  }

  .chapter-stat-row,
  .analysis-summary-grid {
    grid-template-columns: 1fr;
  }

  .chapter-stat-row small {
    white-space: normal;
  }
}
</style>
