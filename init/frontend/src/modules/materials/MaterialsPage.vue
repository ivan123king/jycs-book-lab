<script setup lang="ts">
// MaterialsPage.vue 展示当前书素材库原型，素材始终带来源说明。
import { NButton, NCard, NGi, NGrid, NInput, NSelect, NSpace, NTag, NTooltip } from 'naive-ui'
import { Bookmark, MapPinned } from 'lucide-vue-next'
import type { DemoMaterial } from '@/types/material'

defineProps<{
  materials: DemoMaterial[]
}>()
</script>

<template>
  <section class="page-section">
    <div class="section-heading">
      <div>
        <h1>素材库</h1>
        <p>从拆解中沉淀可复用写作知识，并保留来源链路。</p>
      </div>
      <n-space>
        <n-select
          value="all"
          style="width: 140px"
          :options="[
            { label: '全部分类', value: 'all' },
            { label: '场景写法', value: 'scene' },
            { label: '对白技巧', value: 'dialogue' },
          ]"
        />
        <n-input placeholder="搜索素材" />
      </n-space>
    </div>

    <n-grid :cols="3" :x-gap="16" :y-gap="16" responsive="screen">
      <n-gi v-for="material in materials" :key="material.title">
        <n-card :bordered="false" class="material-card">
          <div class="item-card-head">
            <n-tag type="success" size="small">{{ material.category }}</n-tag>
            <n-tooltip>
              <template #trigger>
                <n-button quaternary circle>
                  <template #icon>
                    <Bookmark :size="16" :fill="material.favorite ? 'currentColor' : 'none'" />
                  </template>
                </n-button>
              </template>
              收藏状态
            </n-tooltip>
          </div>
          <h2>{{ material.title }}</h2>
          <p>{{ material.text }}</p>
          <div class="source-line">
            <MapPinned :size="15" />
            {{ material.source }}
          </div>
          <n-space size="small">
            <n-tag v-for="tag in material.tags" :key="tag" size="small">{{ tag }}</n-tag>
          </n-space>
        </n-card>
      </n-gi>
    </n-grid>
  </section>
</template>
