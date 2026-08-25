<script setup lang="ts">
// AppSidebar.vue 负责应用主导航和本地数据边界提示，不承载具体书籍业务状态。
import { h, type Component } from 'vue'
import { NAlert, NLayoutSider, NMenu } from 'naive-ui'
import type { MenuOption } from 'naive-ui'
import {
  BarChart3,
  BookMarked,
  BookOpenText,
  Bookmark,
  Highlighter,
  Library,
  ListChecks,
  Settings,
} from 'lucide-vue-next'
import type { ViewKey } from '@/types/navigation'

defineProps<{
  activeView: ViewKey
}>()

const emit = defineEmits<{
  navigate: [view: ViewKey]
}>()

const navOptions: MenuOption[] = [
  {
    label: '书库',
    key: 'library',
    icon: renderMenuIcon(Library),
  },
  {
    label: '书库素材',
    key: 'libraryMaterials',
    icon: renderMenuIcon(Bookmark),
  },
  {
    type: 'group',
    label: '当前书',
    key: 'workspace',
    children: [
      { label: '阅读', key: 'reading', icon: renderMenuIcon(BookOpenText) },
      { label: '拆解', key: 'deconstructions', icon: renderMenuIcon(Highlighter) },
      { label: '素材', key: 'materials', icon: renderMenuIcon(Bookmark) },
      { label: '方案', key: 'plan', icon: renderMenuIcon(ListChecks) },
      { label: '分析', key: 'analysis', icon: renderMenuIcon(BarChart3) },
    ],
  },
  {
    label: '设置',
    key: 'settings',
    icon: renderMenuIcon(Settings),
  },
]

/**
 * 为 Naive UI 菜单渲染统一尺寸图标，保持导航视觉一致。
 */
function renderMenuIcon(icon: Component) {
  return () => h(icon, { size: 18, strokeWidth: 1.8 })
}
</script>

<template>
  <n-layout-sider
    bordered
    collapse-mode="width"
    :collapsed-width="72"
    :width="236"
    class="sidebar"
  >
    <div class="brand-block">
      <div class="brand-mark">
        <BookMarked :size="26" />
      </div>
      <div>
        <div class="brand-title">佳依拆书</div>
        <div class="brand-subtitle">本地小说拆解工作台</div>
      </div>
    </div>

    <n-menu
      :value="activeView"
      :options="navOptions"
      :root-indent="20"
      :indent="18"
      @update:value="(key) => emit('navigate', key as ViewKey)"
    />

    <div class="sidebar-footer">
      <n-alert type="info" :show-icon="false" class="storage-alert">
        <div class="storage-title">当前数据边界</div>
        <div class="storage-text">单本书正式数据保存在小说项目目录中。</div>
      </n-alert>
    </div>
  </n-layout-sider>
</template>
