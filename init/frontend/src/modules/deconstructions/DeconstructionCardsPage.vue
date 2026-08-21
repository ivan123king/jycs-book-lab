<script setup lang="ts">
// DeconstructionCardsPage.vue 展示当前书拆解卡片原型，并保留回到原文的页面动作。
import { NButton, NCard, NInput, NSpace, NTag, NTimeline, NTimelineItem } from 'naive-ui'
import { MapPinned, SlidersHorizontal } from 'lucide-vue-next'
import type { DemoDeconstructionCard } from '@/types/deconstruction'

defineProps<{
  cards: DemoDeconstructionCard[]
}>()

const emit = defineEmits<{
  backToSource: []
}>()
</script>

<template>
  <section class="page-section">
    <div class="section-heading">
      <div>
        <h1>拆解卡片</h1>
        <p>集中管理当前书的拆解结果，并保留回到原文的位置。</p>
      </div>
      <n-space>
        <n-input placeholder="搜索拆解、章节、标签" />
        <n-button>
          <template #icon>
            <SlidersHorizontal :size="16" />
          </template>
          筛选
        </n-button>
      </n-space>
    </div>

    <div class="list-detail-layout">
      <div class="stack-list">
        <n-card v-for="card in cards" :key="card.title" :bordered="false" class="item-card">
          <div class="item-card-head">
            <div>
              <n-tag size="small" type="info">{{ card.type }}</n-tag>
              <h2>{{ card.title }}</h2>
            </div>
            <n-button text type="primary" @click="emit('backToSource')">回到原文</n-button>
          </div>
          <p>{{ card.summary }}</p>
          <div class="source-line">
            <MapPinned :size="15" />
            {{ card.chapter }} · {{ card.source }}
          </div>
          <n-space size="small">
            <n-tag v-for="tag in card.tags" :key="tag" size="small">{{ tag }}</n-tag>
          </n-space>
        </n-card>
      </div>

      <n-card :bordered="false" title="来源追溯" class="detail-panel">
        <n-timeline>
          <n-timeline-item type="success" title="书籍" content="雾港来信" />
          <n-timeline-item type="info" title="章节" content="第三章 雨夜码头" />
          <n-timeline-item type="warning" title="原文位置" content="第 18 段，字符 842-895" />
          <n-timeline-item title="关联素材" content="用天气压低重逢情绪" />
        </n-timeline>
      </n-card>
    </div>
  </section>
</template>
