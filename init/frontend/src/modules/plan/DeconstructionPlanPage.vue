<script setup lang="ts">
// DeconstructionPlanPage.vue 负责当前书拆解模板列表、类型维度编辑和当前模板切换。
import { computed, onMounted, ref, watch } from 'vue'
import { NAlert, NButton, NCard, NInput, NList, NListItem, NSpace, NSpin, NTag, NThing } from 'naive-ui'
import { CheckCircle, Copy, Plus, Save, Trash2 } from 'lucide-vue-next'
import {
  OpenDeconstructionPlanWorkspace,
  OpenDeconstructionTemplate,
  SaveDeconstructionTemplate,
  SetCurrentDeconstructionTemplate,
} from '@wails/go/main/App'
import type { library } from '@wails/go/models'
import type { BookProject } from '@/types/library'
import { getErrorMessage } from '@/utils/format'

const props = defineProps<{
  project: BookProject | null
}>()

const workspace = ref<library.DeconstructionPlanWorkspace | null>(null)
const draftPlan = ref<library.DeconstructionPlan | null>(null)
const selectedTemplateFile = ref('')
const activeTypeId = ref('')
const saveAsName = ref('')
const loading = ref(false)
const saving = ref(false)
const pageError = ref('')
const pageNotice = ref('')

const activeBookPath = computed(() => props.project?.book.bookPath ?? '')

const activeType = computed(() => {
  return draftPlan.value?.types.find((item) => item.id === activeTypeId.value) ?? draftPlan.value?.types[0] ?? null
})

/**
 * 读取当前书方案工作区，确保页面只编辑该书项目目录中的模板。
 */
async function loadWorkspace() {
  if (!activeBookPath.value) {
    workspace.value = null
    draftPlan.value = null
    return
  }
  loading.value = true
  pageError.value = ''
  pageNotice.value = ''
  try {
    const loaded = await OpenDeconstructionPlanWorkspace(activeBookPath.value)
    await applyWorkspace(loaded)
  } catch (error) {
    pageError.value = getErrorMessage(error)
  } finally {
    loading.value = false
  }
}

async function applyWorkspace(nextWorkspace: library.DeconstructionPlanWorkspace, preferredTemplateFile = '') {
  workspace.value = nextWorkspace
  draftPlan.value = clonePlan(nextWorkspace.currentPlan)
  const nextSelectedFile =
    preferredTemplateFile || nextWorkspace.currentTemplateFileName || selectedTemplateFile.value || nextWorkspace.templates[0]?.fileName || ''
  if (nextSelectedFile) {
    selectedTemplateFile.value = nextSelectedFile
    draftPlan.value = clonePlan(await OpenDeconstructionTemplate(activeBookPath.value, nextSelectedFile))
  }
  activeTypeId.value = draftPlan.value.types[0]?.id ?? ''
}

async function selectTemplate(fileName: string) {
  if (!activeBookPath.value) {
    return
  }
  selectedTemplateFile.value = fileName
  loading.value = true
  pageError.value = ''
  pageNotice.value = ''
  try {
    draftPlan.value = clonePlan(await OpenDeconstructionTemplate(activeBookPath.value, fileName))
    activeTypeId.value = draftPlan.value.types[0]?.id ?? ''
  } catch (error) {
    pageError.value = getErrorMessage(error)
  } finally {
    loading.value = false
  }
}

async function saveTemplate(overwrite: boolean) {
  if (!activeBookPath.value || !draftPlan.value) {
    return
  }
  const templateName = overwrite ? draftPlan.value.templateName : saveAsName.value
  saving.value = true
  pageError.value = ''
  pageNotice.value = ''
  try {
    const loaded = await SaveDeconstructionTemplate({
      bookPath: activeBookPath.value,
      templateName,
      fileName: overwrite ? selectedTemplateFile.value : '',
      plan: draftPlan.value,
      overwrite,
    } as library.SaveDeconstructionTemplateRequest)
    const savedFileName = overwrite
      ? selectedTemplateFile.value
      : (loaded.templates.find((template) => template.name === templateName.trim())?.fileName ?? loaded.templates[0]?.fileName ?? '')
    await applyWorkspace(loaded, savedFileName)
    pageNotice.value = overwrite ? '模板已保存。' : '已另存为新模板。'
    if (!overwrite) {
      saveAsName.value = ''
    }
  } catch (error) {
    pageError.value = getErrorMessage(error)
  } finally {
    saving.value = false
  }
}

async function setSelectedTemplateAsCurrent() {
  if (!activeBookPath.value || !selectedTemplateFile.value) {
    return
  }
  saving.value = true
  pageError.value = ''
  pageNotice.value = ''
  try {
    const loaded = await SetCurrentDeconstructionTemplate({
      bookPath: activeBookPath.value,
      fileName: selectedTemplateFile.value,
    })
    await applyWorkspace(loaded, selectedTemplateFile.value)
    pageNotice.value = '已设为当前模板，阅读拆解会使用这份内容。'
  } catch (error) {
    pageError.value = getErrorMessage(error)
  } finally {
    saving.value = false
  }
}

function addType() {
  if (!draftPlan.value) {
    return
  }
  const order = draftPlan.value.types.length + 1
  const type = {
    id: `type-${Date.now()}`,
    name: `拆解类型 ${order}`,
    order,
    dimensions: [{ id: `dimension-${Date.now()}`, name: '通用分析', order: 1 }],
    dimension: '',
    fields: [],
  } as unknown as library.DeconstructionType
  draftPlan.value.types.push(type)
  activeTypeId.value = type.id
}

function removeType(typeId: string) {
  if (!draftPlan.value || draftPlan.value.types.length <= 1) {
    pageError.value = '至少需要保留一个拆解类型。'
    return
  }
  draftPlan.value.types = draftPlan.value.types
    .filter((item) => item.id !== typeId)
    .map((item, index) => ({ ...item, order: index + 1 }))
  activeTypeId.value = draftPlan.value.types[0]?.id ?? ''
}

function addDimension() {
  if (!activeType.value) {
    return
  }
  const order = activeType.value.dimensions.length + 1
  activeType.value.dimensions.push({
    id: `dimension-${Date.now()}`,
    name: `拆解维度 ${order}`,
    order,
  })
}

function removeDimension(dimensionId: string) {
  if (!activeType.value || activeType.value.dimensions.length <= 1) {
    pageError.value = '每个拆解类型至少需要保留一个维度。'
    return
  }
  activeType.value.dimensions = activeType.value.dimensions
    .filter((item) => item.id !== dimensionId)
    .map((item, index) => ({ ...item, order: index + 1 }))
}

function clonePlan(plan: library.DeconstructionPlan) {
  return JSON.parse(JSON.stringify(plan)) as library.DeconstructionPlan
}

watch(
  () => props.project?.book.bookPath,
  () => {
    void loadWorkspace()
  },
)

onMounted(loadWorkspace)
</script>

<template>
  <section class="page-section">
    <div class="section-heading">
      <div>
        <h1>拆解方案</h1>
        <p>按当前书管理拆解模板，阅读中选文后会按当前模板选择类型和维度。</p>
      </div>
      <n-space>
        <n-button
          type="primary"
          :disabled="!selectedTemplateFile || selectedTemplateFile === workspace?.currentTemplateFileName"
          :loading="saving"
          @click="setSelectedTemplateAsCurrent"
        >
          <template #icon>
            <CheckCircle :size="16" />
          </template>
          设为当前模板
        </n-button>
      </n-space>
    </div>

    <n-alert v-if="!activeBookPath" type="info" class="page-alert">
      请先从书库打开一本小说。拆解方案和模板都会保存在该小说项目目录中。
    </n-alert>

    <n-spin v-else :show="loading">
      <n-alert v-if="pageError" type="error" class="page-alert">{{ pageError }}</n-alert>
      <n-alert v-else-if="pageNotice" type="success" class="page-alert">{{ pageNotice }}</n-alert>

      <div class="plan-layout">
        <n-card :bordered="false" title="模板列表" class="dimension-card">
          <n-list>
            <n-list-item v-for="template in workspace?.templates ?? []" :key="template.fileName">
              <button
                class="template-list-item"
                :class="{
                  active: workspace?.currentTemplateFileName === template.fileName,
                  selected: selectedTemplateFile === template.fileName,
                }"
                @click="selectTemplate(template.fileName)"
              >
                <span>
                  {{ template.name }}
                  <n-tag v-if="workspace?.currentTemplateFileName === template.fileName" size="small" type="success">
                    当前方案
                  </n-tag>
                </span>
                <small>{{ template.fileName }}</small>
              </button>
            </n-list-item>
          </n-list>
        </n-card>

        <n-card :bordered="false" title="模板内容" class="fields-card">
          <template v-if="draftPlan">
            <div class="template-name-row">
              <n-input v-model:value="draftPlan.templateName" placeholder="模板名称，例如：人物感情分析模板" />
              <n-button :loading="saving" @click="saveTemplate(true)">
                <template #icon>
                  <Save :size="16" />
                </template>
                保存模板
              </n-button>
            </div>

            <div class="save-as-row">
              <n-input v-model:value="saveAsName" placeholder="另存为模板名称" />
              <n-button type="primary" secondary :loading="saving" @click="saveTemplate(false)">
                <template #icon>
                  <Copy :size="16" />
                </template>
                另存为模板
              </n-button>
            </div>

            <n-alert type="warning" class="page-alert">
              修改模板不会改写已经保存的拆解卡片；历史卡片会保留创建时的类型、维度和看法。
            </n-alert>

            <div class="plan-editor-grid">
              <div class="type-list-panel">
                <div class="panel-title-row">
                  <strong>类型</strong>
                  <n-button size="small" @click="addType">
                    <template #icon>
                      <Plus :size="14" />
                    </template>
                    新增
                  </n-button>
                </div>
                <button
                  v-for="type in draftPlan.types"
                  :key="type.id"
                  class="type-list-item"
                  :class="{ active: activeTypeId === type.id }"
                  @click="activeTypeId = type.id"
                >
                  {{ type.name }}
                </button>
              </div>

              <div v-if="activeType" class="type-editor-panel">
                <n-thing title="类型名称" description="例如：表现手法、人物感情、结构设计">
                  <template #header-extra>
                    <n-button size="small" tertiary type="error" @click="removeType(activeType.id)">
                      <template #icon>
                        <Trash2 :size="14" />
                      </template>
                      删除类型
                    </n-button>
                  </template>
                  <n-input v-model:value="activeType.name" class="type-name-input" />
                </n-thing>

                <div class="panel-title-row dimension-heading">
                  <strong>维度</strong>
                  <n-button size="small" @click="addDimension">
                    <template #icon>
                      <Plus :size="14" />
                    </template>
                    新增维度
                  </n-button>
                </div>
                <div class="dimension-editor-list">
                  <div v-for="dimension in activeType.dimensions" :key="dimension.id" class="dimension-editor-row">
                    <n-input v-model:value="dimension.name" />
                    <n-tag :bordered="false">第 {{ dimension.order }} 项</n-tag>
                    <n-button quaternary type="error" @click="removeDimension(dimension.id)">
                      <template #icon>
                        <Trash2 :size="15" />
                      </template>
                    </n-button>
                  </div>
                </div>
              </div>
            </div>
          </template>
        </n-card>
      </div>
    </n-spin>
  </section>
</template>
