<script setup lang="ts">
import { CircleCheck, CircleClose, EditPen } from '@element-plus/icons-vue'
import { onMounted, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'

import { apiGet, apiPatch, apiPost } from '@/api/http'
import { DEV_INSTITUTION_ID } from '@/config/devSeed'

type PositionRow = {
  id: string
  title: string
  description?: string
  status: string
  institutionId: string
  createdAt: string
}

type ListResp = {
  items: PositionRow[]
  total: number
  page: number
  pageSize: number
}

const loading = ref(false)
const rows = ref<PositionRow[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)

const dialogVisible = ref(false)
const dialogSaving = ref(false)
const formTitle = ref('')
const formDesc = ref('')
const formStatus = ref('open')

async function load() {
  loading.value = true
  try {
    const q = new URLSearchParams({
      page: String(page.value),
      pageSize: String(pageSize.value),
      sortBy: 'created_at',
      sortOrder: 'desc',
    })
    const data = await apiGet<ListResp>(`/api/v1/recruitment/positions?${q}`)
    rows.value = data.items
    total.value = data.total
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : 'Failed to load positions')
  } finally {
    loading.value = false
  }
}

function onPageChange(p: number) {
  page.value = p
  load()
}

function openCreate() {
  formTitle.value = ''
  formDesc.value = ''
  formStatus.value = 'open'
  dialogVisible.value = true
}

async function submitCreate() {
  if (!formTitle.value.trim()) {
    ElMessage.warning('Title is required.')
    return
  }
  dialogSaving.value = true
  try {
    await apiPost<PositionRow>('/api/v1/recruitment/positions', {
      institutionId: DEV_INSTITUTION_ID,
      title: formTitle.value.trim(),
      description: formDesc.value.trim() || undefined,
      status: formStatus.value,
    })
    ElMessage.success('Position created.')
    dialogVisible.value = false
    await load()
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : 'Create failed')
  } finally {
    dialogSaving.value = false
  }
}

async function toggleStatus(row: PositionRow) {
  const next = row.status === 'open' ? 'closed' : 'open'
  try {
    await apiPatch<PositionRow>(`/api/v1/recruitment/positions/${row.id}`, { status: next })
    ElMessage.success('Status updated.')
    await load()
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : 'Update failed')
  }
}

async function editDescription(row: PositionRow) {
  let value: string
  try {
    const res = await ElMessageBox.prompt('Description', 'Edit position', {
      inputValue: row.description ?? '',
      inputType: 'textarea',
    })
    value = res.value
  } catch {
    return
  }
  try {
    await apiPatch<PositionRow>(`/api/v1/recruitment/positions/${row.id}`, {
      description: value.trim() || undefined,
    })
    ElMessage.success('Updated.')
    await load()
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : 'Update failed')
  }
}

onMounted(load)
</script>

<template>
  <div class="rec-page">
    <div class="rec-toolbar">
      <h2 class="rec-title">Positions</h2>
      <el-button type="primary" round @click="openCreate">Add position</el-button>
    </div>

    <el-card class="rec-card" shadow="never">
      <el-table v-loading="loading" :data="rows" stripe empty-text="No positions to show">
        <el-table-column prop="title" label="Title" min-width="200" />
        <el-table-column prop="status" label="Status" width="110">
          <template #default="{ row }">
            <el-tag :type="row.status === 'open' ? 'success' : 'info'" size="small">{{ row.status }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="Description" min-width="220" show-overflow-tooltip>
          <template #default="{ row }">
            {{ row.description || '—' }}
          </template>
        </el-table-column>
        <el-table-column prop="createdAt" label="Created" width="180">
          <template #default="{ row }">
            {{ new Date(row.createdAt).toLocaleString() }}
          </template>
        </el-table-column>
        <el-table-column label="Actions" width="120" fixed="right" align="center">
          <template #default="{ row }">
            <div class="action-icons">
              <el-tooltip content="Edit description" placement="top">
                <el-button
                  link
                  type="primary"
                  :icon="EditPen"
                  aria-label="Edit description"
                  @click="editDescription(row)"
                />
              </el-tooltip>
              <el-tooltip :content="row.status === 'open' ? 'Close position' : 'Reopen position'" placement="top">
                <el-button
                  v-if="row.status === 'open'"
                  link
                  type="warning"
                  :icon="CircleClose"
                  aria-label="Close position"
                  @click="toggleStatus(row)"
                />
                <el-button
                  v-else
                  link
                  type="success"
                  :icon="CircleCheck"
                  aria-label="Reopen position"
                  @click="toggleStatus(row)"
                />
              </el-tooltip>
            </div>
          </template>
        </el-table-column>
      </el-table>
      <div class="rec-pager">
        <el-pagination
          background
          layout="prev, pager, next, total"
          :total="total"
          :page-size="pageSize"
          :current-page="page"
          @current-change="onPageChange"
        />
      </div>
    </el-card>

    <el-dialog v-model="dialogVisible" title="New position" width="440px" destroy-on-close>
      <el-form label-position="top">
        <el-form-item label="Title" required>
          <el-input v-model="formTitle" placeholder="Job title" />
        </el-form-item>
        <el-form-item label="Description">
          <el-input v-model="formDesc" type="textarea" :rows="3" placeholder="Optional" />
        </el-form-item>
        <el-form-item label="Status">
          <el-select v-model="formStatus" style="width: 100%">
            <el-option label="Open" value="open" />
            <el-option label="Closed" value="closed" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">Cancel</el-button>
        <el-button type="primary" :loading="dialogSaving" @click="submitCreate">Create</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.rec-page {
  max-width: 1200px;
}
.rec-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  margin-bottom: 1rem;
}
.rec-title {
  margin: 0;
  font-size: 1.25rem;
  font-weight: 650;
  letter-spacing: -0.02em;
}
.rec-card {
  border-radius: 14px;
  border: 1px solid var(--el-border-color-lighter);
}
.rec-pager {
  display: flex;
  justify-content: flex-end;
  margin-top: 1rem;
}

.action-icons {
  display: inline-flex;
  align-items: center;
  gap: 2px;
}
</style>
