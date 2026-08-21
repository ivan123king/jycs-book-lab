<script setup lang="ts">
// DeconstructionPlanPage.vue 展示当前书独立拆解方案原型，不把字段写死到 App 壳层。
import { NAlert, NButton, NCard, NDataTable, NList, NListItem, NSpace, NTabPane, NTabs, NThing } from 'naive-ui'
import type { DemoPlanField } from '@/types/plan'

defineProps<{
  fields: DemoPlanField[]
}>()

const planColumns = [
  { title: '维度', key: 'dimension' },
  { title: '拆解类型', key: 'type' },
  { title: '字段名称', key: 'field' },
  { title: '字段类型', key: 'fieldType' },
  { title: '必填', key: 'required' },
  { title: '状态', key: 'visible' },
]
</script>

<template>
  <section class="page-section">
    <div class="section-heading">
      <div>
        <h1>拆解方案</h1>
        <p>当前书独立配置拆解维度、类型和字段。</p>
      </div>
      <n-space>
        <n-button>从模板创建</n-button>
        <n-button type="primary">另存为模板</n-button>
      </n-space>
    </div>

    <div class="plan-layout">
      <n-card :bordered="false" title="维度与类型" class="dimension-card">
        <n-tabs type="line" animated>
          <n-tab-pane name="scene" tab="场景">
            <n-list>
              <n-list-item>
                <n-thing title="场景拆解" description="场景目的、冲突、情绪变化、环境调度" />
              </n-list-item>
              <n-list-item>
                <n-thing title="氛围拆解" description="感官描写、节奏、意象组合" />
              </n-list-item>
            </n-list>
          </n-tab-pane>
          <n-tab-pane name="character" tab="人物">
            <n-list>
              <n-list-item>
                <n-thing title="人物弧光" description="欲望、阻力、选择、变化" />
              </n-list-item>
              <n-list-item>
                <n-thing title="对白拆解" description="表层台词、潜台词、关系变化" />
              </n-list-item>
            </n-list>
          </n-tab-pane>
          <n-tab-pane name="structure" tab="结构">
            <n-list>
              <n-list-item>
                <n-thing title="伏笔回收" description="埋设位置、回收章节、情绪收益" />
              </n-list-item>
            </n-list>
          </n-tab-pane>
        </n-tabs>
      </n-card>

      <n-card :bordered="false" title="字段配置" class="fields-card">
        <n-alert type="warning" class="page-alert">
          字段调整需要保留旧卡片数据，并在真实实现中提供迁移提示。
        </n-alert>
        <n-data-table :columns="planColumns" :data="fields" :bordered="false" />
      </n-card>
    </div>
  </section>
</template>
