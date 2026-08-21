<script setup lang="ts">
// ReadingWorkbenchPage.vue 承载当前书阅读、编辑和选文拆解原型，后续只通过后端服务保存正式数据。
import { ref } from 'vue'
import {
  NAlert,
  NBadge,
  NButton,
  NButtonGroup,
  NForm,
  NFormItem,
  NInput,
  NProgress,
  NScrollbar,
  NSelect,
  NSpace,
  NTag,
} from 'naive-ui'
import { Copy, Highlighter, Pencil, Save } from 'lucide-vue-next'
import type { DemoChapter } from '@/types/reading'

defineProps<{
  chapters: DemoChapter[]
}>()

const readerMode = ref<'read' | 'edit'>('read')
const selectedChapterKey = ref('chapter-03')
const selectedDeconstructionType = ref('scene')
</script>

<template>
  <section class="reading-page">
    <aside class="chapter-panel">
      <div class="panel-title">章节目录</div>
      <n-scrollbar style="max-height: calc(100vh - 180px)">
        <button
          v-for="chapter in chapters"
          :key="chapter.key"
          class="chapter-item"
          :class="{ active: selectedChapterKey === chapter.key }"
          @click="selectedChapterKey = chapter.key"
        >
          <span>{{ chapter.title }}</span>
          <small>{{ chapter.count }} 张卡片</small>
          <n-progress
            type="line"
            :percentage="chapter.progress"
            :show-indicator="false"
            :height="4"
          />
        </button>
      </n-scrollbar>
    </aside>

    <main class="reader-pane">
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
          <n-tag type="info">第三章 雨夜码头</n-tag>
        </n-space>
        <n-space>
          <n-button quaternary circle>
            <template #icon>
              <Copy :size="16" />
            </template>
          </n-button>
          <n-button quaternary circle>
            <template #icon>
              <Pencil :size="16" />
            </template>
          </n-button>
        </n-space>
      </div>

      <article v-if="readerMode === 'read'" class="novel-text">
        <p>雨在码头上铺开，像一层被反复揉皱的灰布。旧仓库檐下的灯忽明忽暗，照见铁轨尽头慢慢停下来的货车。</p>
        <p>
          沈知遥站在阴影里，手指压着信封的边角。她听见身后有人喊她的名字，却没有立刻回头。
          <span class="selected-source">那声音比三年前低了一点，像被潮气磨过。</span>
        </p>
        <p>渡轮鸣笛时，周围所有人都短暂地看向江面。只有她仍看着仓库的玻璃窗，看里面倒映出的那个人影。</p>
        <div class="floating-toolbar">
          <n-button size="small" type="primary">
            <template #icon>
              <Highlighter :size="14" />
            </template>
            创建拆解
          </n-button>
          <n-button size="small">标记</n-button>
          <n-button size="small">复制</n-button>
        </div>
      </article>

      <div v-else class="edit-pane">
        <n-form label-placement="top">
          <n-form-item label="章节标题">
            <n-input class="edit-title-input" value="第三章 雨夜码头" />
          </n-form-item>
          <n-form-item label="章节正文">
            <n-input
              type="textarea"
              value="雨在码头上铺开，像一层被反复揉皱的灰布。旧仓库檐下的灯忽明忽暗，照见铁轨尽头慢慢停下来的货车。

沈知遥站在阴影里，手指压着信封的边角。她听见身后有人喊她的名字，却没有立刻回头。那声音比三年前低了一点，像被潮气磨过。

渡轮鸣笛时，周围所有人都短暂地看向江面。只有她仍看着仓库的玻璃窗，看里面倒映出的那个人影。"
              :autosize="{ minRows: 15, maxRows: 24 }"
            />
          </n-form-item>
        </n-form>
        <n-alert type="info" :show-icon="false">
          章节标题和正文修改放在阅读页编辑模式中，导入阶段只展示识别结果。
        </n-alert>
        <n-space justify="end" class="card-actions">
          <n-button>放弃修改</n-button>
          <n-button type="primary">
            <template #icon>
              <Save :size="16" />
            </template>
            保存章节
          </n-button>
        </n-space>
      </div>
    </main>

    <aside class="workbench-panel">
      <div class="panel-title-row">
        <div>
          <div class="panel-title">拆解工作台</div>
          <div class="panel-subtitle">当前选文 · 可创建多张拆解</div>
        </div>
        <n-badge :value="3" type="success" />
      </div>
      <n-alert type="default" :show-icon="false" class="source-box">
        那声音比三年前低了一点，像被潮气磨过。
      </n-alert>
      <n-form label-placement="top">
        <n-form-item label="拆解类型">
          <n-select
            v-model:value="selectedDeconstructionType"
            :options="[
              { label: '场景拆解', value: 'scene' },
              { label: '对白拆解', value: 'dialogue' },
              { label: '人物弧光', value: 'character' },
            ]"
          />
        </n-form-item>
        <n-form-item label="场景目的">
          <n-input value="让重逢带有迟疑和对峙感" />
        </n-form-item>
        <n-form-item label="情绪变化">
          <n-input
            type="textarea"
            value="从确认对方出现，到压住回应，再到被环境迫使面对。"
            :autosize="{ minRows: 4, maxRows: 5 }"
          />
        </n-form-item>
        <n-form-item label="标签">
          <n-select
            multiple
            :value="['氛围', '重逢']"
            :options="[
              { label: '氛围', value: '氛围' },
              { label: '重逢', value: '重逢' },
              { label: '悬念', value: '悬念' },
            ]"
          />
        </n-form-item>
      </n-form>
      <n-space justify="end">
        <n-button>继续建卡</n-button>
        <n-button type="primary">保存拆解</n-button>
      </n-space>
    </aside>
  </section>
</template>
