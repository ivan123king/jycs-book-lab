<script setup lang="ts">
// SettingsPage.vue 承载应用级偏好原型，不保存单本书正式业务数据。
import { NButton, NCard, NForm, NFormItem, NList, NListItem, NSelect, NSpace, NSwitch, NThing } from 'naive-ui'
import { CircleAlert, Database } from 'lucide-vue-next'
</script>

<template>
  <section class="page-section">
    <div class="section-heading">
      <div>
        <h1>全局设置</h1>
        <p>只管理应用级偏好，不保存某一本书的正式业务数据。</p>
      </div>
    </div>

    <div class="settings-grid">
      <n-card :bordered="false" title="界面偏好">
        <n-form label-placement="left" label-width="110">
          <n-form-item label="阅读字号">
            <n-select
              value="18"
              :options="[
                { label: '16px', value: '16' },
                { label: '18px', value: '18' },
                { label: '20px', value: '20' },
              ]"
            />
          </n-form-item>
          <n-form-item label="紧凑布局">
            <n-switch />
          </n-form-item>
          <n-form-item label="最近项目数">
            <n-select
              value="10"
              :options="[
                { label: '5', value: '5' },
                { label: '10', value: '10' },
                { label: '20', value: '20' },
              ]"
            />
          </n-form-item>
        </n-form>
      </n-card>

      <n-card :bordered="false" title="本地数据">
        <n-list>
          <n-list-item>
            <n-thing title="项目目录" description="单本书正式数据随项目目录迁移" />
          </n-list-item>
          <n-list-item>
            <n-thing title="缓存索引" description=".cache/index.sqlite 可删除并重建" />
          </n-list-item>
          <n-list-item>
            <n-space>
              <n-button>
                <template #icon>
                  <Database :size="16" />
                </template>
                清理缓存
              </n-button>
              <n-button>
                <template #icon>
                  <CircleAlert :size="16" />
                </template>
                检查项目
              </n-button>
            </n-space>
          </n-list-item>
        </n-list>
      </n-card>
    </div>
  </section>
</template>
