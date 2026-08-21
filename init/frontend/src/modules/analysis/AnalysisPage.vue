<script setup lang="ts">
// AnalysisPage.vue 展示当前书第一阶段轻量统计，不扩展到跨书比较。
import { NCard, NGi, NGrid, NProgress, NStatistic, NTag } from 'naive-ui'
import type { DemoAnalysisRow } from '@/types/analysis'
import type { CurrentBookSummary } from '@/types/currentBook'
import type { DemoChapter } from '@/types/reading'

defineProps<{
  currentBook: CurrentBookSummary
  rows: DemoAnalysisRow[]
  chapters: DemoChapter[]
}>()
</script>

<template>
  <section class="page-section">
    <div class="section-heading">
      <div>
        <h1>基础分析</h1>
        <p>第一阶段只展示当前书的轻量统计。</p>
      </div>
      <n-tag type="default">不含跨书比较</n-tag>
    </div>

    <n-grid :cols="4" :x-gap="16" :y-gap="16" responsive="screen">
      <n-gi>
        <n-card :bordered="false"><n-statistic label="章节" :value="currentBook.chapters" /></n-card>
      </n-gi>
      <n-gi>
        <n-card :bordered="false"><n-statistic label="拆解卡片" :value="currentBook.deconstructions" /></n-card>
      </n-gi>
      <n-gi>
        <n-card :bordered="false"><n-statistic label="素材" :value="currentBook.materials" /></n-card>
      </n-gi>
      <n-gi>
        <n-card :bordered="false"><n-statistic label="已拆章节" :value="19" /></n-card>
      </n-gi>
    </n-grid>

    <div class="analysis-grid">
      <n-card :bordered="false" title="按拆解类型">
        <div v-for="row in rows" :key="row.type" class="progress-row">
          <div>
            <span>{{ row.type }}</span>
            <strong>{{ row.count }}</strong>
          </div>
          <n-progress type="line" :percentage="row.percent" :height="8" />
        </div>
      </n-card>

      <n-card :bordered="false" title="章节拆解分布">
        <div v-for="chapter in chapters" :key="chapter.key" class="chapter-stat-row">
          <span>{{ chapter.title }}</span>
          <n-progress type="line" :percentage="Math.min(chapter.count * 8, 100)" :height="8" />
          <strong>{{ chapter.count }}</strong>
        </div>
      </n-card>
    </div>
  </section>
</template>
