<script setup lang="ts">
// AppTopbar.vue 展示当前页面标题、当前书摘要和壳层级快捷入口。
import type { CurrentBookSummary } from '@/types/currentBook'
import { NLayoutHeader } from 'naive-ui'

defineProps<{
  title: string
  currentBook: CurrentBookSummary
}>()
</script>

<template>
  <n-layout-header bordered class="topbar">
    <div>
      <div class="page-title">{{ title }}</div>
      <div class="page-context">
        当前书：{{ currentBook.title }} · {{ currentBook.readingPosition }}
      </div>
    </div>
    <div class="topbar-actions">
      <slot name="actions" />
    </div>
  </n-layout-header>
</template>
