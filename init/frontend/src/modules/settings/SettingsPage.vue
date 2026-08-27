<script setup lang="ts">
// SettingsPage.vue 承载应用级偏好读写，不保存或覆盖单本书正式业务数据。
import { onMounted, ref } from 'vue'
import { NAlert, NButton, NCard, NForm, NFormItem, NInputNumber, NList, NListItem, NSpace, NThing, useMessage } from 'naive-ui'
import { Type } from 'lucide-vue-next'
import { GetAppSettings, SaveAppSettings } from '@wails/go/main/App'
import type { library } from '@wails/go/models'
import { getErrorMessage } from '@/utils/format'

const minReaderFontSize = 14
const maxReaderFontSize = 28
const fallbackReaderFontSize = 18

const message = useMessage()
const loading = ref(false)
const saving = ref(false)
const settingsError = ref('')
const settings = ref<library.AppSettings | null>(null)
const defaultReaderFontSize = ref(fallbackReaderFontSize)

/**
 * 限制全局默认字号范围，保持它与阅读页单书字号控件的可选范围一致。
 */
function normalizeReaderFontSize(value?: number | null) {
  if (!value) {
    return fallbackReaderFontSize
  }
  return Math.min(maxReaderFontSize, Math.max(minReaderFontSize, Math.round(value)))
}

/**
 * 设置页打开时读取应用系统数据目录中的全局偏好。
 */
async function loadSettings() {
  loading.value = true
  settingsError.value = ''
  try {
    const loaded = await GetAppSettings()
    settings.value = loaded
    defaultReaderFontSize.value = normalizeReaderFontSize(loaded.defaultReaderFontSize)
  } catch (error) {
    settingsError.value = getErrorMessage(error)
  } finally {
    loading.value = false
  }
}

/**
 * 保存新导入小说使用的默认阅读字号，不改写任何已存在的单书 reading-state.json。
 */
async function saveSettings() {
  saving.value = true
  settingsError.value = ''
  try {
    const saved = await SaveAppSettings({
      defaultReaderFontSize: normalizeReaderFontSize(defaultReaderFontSize.value),
    })
    settings.value = saved
    defaultReaderFontSize.value = normalizeReaderFontSize(saved.defaultReaderFontSize)
    message.success('全局字号已保存')
  } catch (error) {
    settingsError.value = getErrorMessage(error)
  } finally {
    saving.value = false
  }
}

onMounted(loadSettings)
</script>

<template>
  <section class="page-section">
    <div class="section-heading">
      <div>
        <h1>全局设置</h1>
        <p>只管理应用级偏好，不保存某一本书的正式业务数据。</p>
      </div>
    </div>

    <n-alert v-if="settingsError" type="error" class="page-alert">
      {{ settingsError }}
    </n-alert>

    <div class="settings-grid">
      <n-card :bordered="false" title="阅读偏好">
        <n-form label-placement="left" label-width="120">
          <n-form-item label="新书默认字号">
            <n-input-number
              v-model:value="defaultReaderFontSize"
              :disabled="loading"
              :min="minReaderFontSize"
              :max="maxReaderFontSize"
              :step="1"
              :show-button="true"
            >
              <template #prefix>
                <Type :size="16" />
              </template>
            </n-input-number>
          </n-form-item>
        </n-form>

        <n-space justify="space-between" align="center" class="settings-actions">
          <span class="settings-hint">已导入书籍继续使用各自阅读页字号。</span>
          <n-button type="primary" :loading="saving" :disabled="loading" @click="saveSettings">
            保存设置
          </n-button>
        </n-space>
      </n-card>

      <n-card :bordered="false" title="本地数据">
        <n-list>
          <n-list-item>
            <n-thing title="书库目录" :description="settings?.libraryPath || '尚未设置书库目录'" />
          </n-list-item>
          <n-list-item>
            <n-thing title="最近打开" :description="settings?.lastOpenedBookPath || '暂无最近打开小说'" />
          </n-list-item>
          <n-list-item>
            <n-thing title="项目目录" description="单本书正式数据随项目目录迁移" />
          </n-list-item>
        </n-list>
      </n-card>
    </div>
  </section>
</template>

<style scoped>
.settings-actions {
  margin-top: 8px;
}

.settings-hint {
  color: #667085;
  font-size: 13px;
}
</style>
