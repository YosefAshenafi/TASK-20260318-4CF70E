<script setup lang="ts">
import { CircleCheck, CircleClose, EditPen } from '@element-plus/icons-vue'
import { onMounted, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'

import { apiGet, apiPatch, apiPost } from '@/api/http'
import { DEV_INSTITUTION_ID } from '@/config/devSeed'
import { useAuthStore } from '@/stores/auth'

type QualRow = {
  id: string
  institutionId: string
  clientId: string
  displayName: string
  status: string
  expiresOn?: string
  createdAt: string
  updatedAt: string
}

type ListResp = {
  items: QualRow[]
  total: number
  page: number
  pageSize: number
}

const auth = useAuthStore()
const canManage = () => auth.hasPermission('compliance.manage')

const loading = ref(false)
const rows = ref<QualRow[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)

const expiringLoading = ref(false)
const expiringSoon = ref<QualRow[]>([])

const dialogVisible = ref(false)
const dialogSaving = ref(false)
const formClient = ref('')
const formName = ref('')
const formExpires = ref('')

async function loadExpiring() {
  expiringLoading.value = true
  try {
    const data = await apiGet<{ items: QualRow[] }>('/api/v1/compliance/qualifications/expiring?days=30')
    expiringSoon.value = data.items ?? []
  } catch {
    expiringSoon.value = []
  } finally {
    expiringLoading.value = false
  }
}

async function load() {
  loading.value = true
  try {
    const q = new URLSearchParams({
      page: String(page.value),
      pageSize: String(pageSize.value),
      sortBy: 'expires_on',
      sortOrder: 'asc',
    })
    const data = await apiGet<ListResp>(`/api/v1/compliance/qualifications?${q}`)
    rows.value = data.items
    total.value = data.total
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : 'Failed to load qualifications')
  } finally {
    loading.value = false
  }
}

function onPageChange(p: number) {
  page.value = p
  load()
}

function openCreate() {
  formClient.value = ''
  formName.value = ''
  formExpires.value = ''
  dialogVisible.value = true
}

async function submitCreate() {
  if (!formClient.value.trim() || !formName.value.trim()) {
    ElMessage.warning('Client ID and display name are required.')
    return
  }
  dialogSaving.value = true
  try {
    await apiPost<QualRow>('/api/v1/compliance/qualifications', {
      institutionId: DEV_INSTITUTION_ID,
      clientId: formClient.value.trim(),
      displayName: formName.value.trim(),
      expiresOn: formExpires.value || undefined,
    })
    ElMessage.success('Qualification created.')
    dialogVisible.value = false
    await load()
    await loadExpiring()
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : 'Create failed')
  } finally {
    dialogSaving.value = false
  }
}

async function editRow(row: QualRow) {
  let name: string
  let exp: string
  try {
    const res = await ElMessageBox.prompt('Display name', 'Edit qualification', {
      inputValue: row.displayName,
    })
    name = res.value
    const res2 = await ElMessageBox.prompt('Expires on (YYYY-MM-DD, empty to clear)', 'Expiration', {
      inputValue: row.expiresOn ?? '',
    })
    exp = res2.value
  } catch {
    return
  }
  try {
    await apiPatch<QualRow>(`/api/v1/compliance/qualifications/${row.id}`, {
      displayName: name.trim(),
      expiresOn: exp.trim() === '' ? '' : exp.trim(),
    })
    ElMessage.success('Updated.')
    await load()
    await loadExpiring()
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : 'Update failed')
  }
}

async function deactivate(row: QualRow) {
  try {
    await ElMessageBox.confirm(`Deactivate “${row.displayName}”?`, 'Confirm', { type: 'warning' })
  } catch {
    return
  }
  try {
    await apiPost<QualRow>(`/api/v1/compliance/qualifications/${row.id}/deactivate`, {})
    ElMessage.success('Deactivated.')
    await load()
    await loadExpiring()
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : 'Update failed')
  }
}

async function activate(row: QualRow) {
  try {
    await apiPost<QualRow>(`/api/v1/compliance/qualifications/${row.id}/activate`, {})
    ElMessage.success('Activated.')
    await load()
    await loadExpiring()
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : 'Update failed')
  }
}

function expiringIds(): Set<string> {
  return new Set(expiringSoon.value.map((r) => r.id))
}

onMounted(async () => {
  await loadExpiring()
  await load()
})
</script>

<template>
  <div class="rec-page">
    <div class="rec-toolbar">
      <h2 class="rec-title">Qualifications</h2>
      <el-button v-if="canManage()" type="primary" round @click="openCreate">Add qualification</el-button>
    </div>

    <el-alert
      v-if="expiringSoon.length > 0"
      v-loading="expiringLoading"
      class="rec-alert"
      type="warning"
      :closable="false"
      show-icon
      :title="`${expiringSoon.length} qualification(s) expiring within 30 days`"
    />

    <el-card class="rec-card" shadow="never">
      <el-table v-loading="loading" :data="rows" stripe empty-text="No qualifications to show">
        <el-table-column prop="displayName" label="Client / name" min-width="220">
          <template #default="{ row }">
            <div class="qual-name">{{ row.displayName }}</div>
            <div class="qual-sub">{{ row.clientId }}</div>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="Status" width="110">
          <template #default="{ row }">
            <el-tag :type="row.status === 'active' ? 'success' : 'info'" size="small">{{ row.status }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="expiresOn" label="Expires" width="120">
          <template #default="{ row }">
            <span v-if="row.expiresOn">{{ row.expiresOn }}</span>
            <span v-else>—</span>
          </template>
        </el-table-column>
        <el-table-column label="" width="100">
          <template #default="{ row }">
            <el-tag v-if="expiringIds().has(row.id)" type="warning" size="small">Soon</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="updatedAt" label="Updated" width="180">
          <template #default="{ row }">
            {{ new Date(row.updatedAt || row.createdAt).toLocaleString() }}
          </template>
        </el-table-column>
        <el-table-column v-if="canManage()" label="Actions" width="120" fixed="right" align="center">
          <template #default="{ row }">
            <div class="action-icons">
              <el-tooltip content="Edit" placement="top">
                <el-button
                  link
                  type="primary"
                  :icon="EditPen"
                  aria-label="Edit qualification"
                  @click="editRow(row)"
                />
              </el-tooltip>
              <el-tooltip v-if="row.status === 'active'" content="Deactivate" placement="top">
                <el-button
                  link
                  type="warning"
                  :icon="CircleClose"
                  aria-label="Deactivate"
                  @click="deactivate(row)"
                />
              </el-tooltip>
              <el-tooltip v-else content="Activate" placement="top">
                <el-button
                  link
                  type="success"
                  :icon="CircleCheck"
                  aria-label="Activate"
                  @click="activate(row)"
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

    <el-dialog v-model="dialogVisible" title="New qualification" width="440px" destroy-on-close>
      <el-form label-position="top">
        <el-form-item label="Client ID" required>
          <el-input v-model="formClient" placeholder="e.g. client-demo-1" />
        </el-form-item>
        <el-form-item label="Display name" required>
          <el-input v-model="formName" placeholder="Shown in lists" />
        </el-form-item>
        <el-form-item label="Expires on">
          <el-input v-model="formExpires" placeholder="YYYY-MM-DD (optional)" />
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
.rec-alert {
  margin-bottom: 1rem;
  border-radius: 12px;
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
.qual-name {
  font-weight: 600;
}
.qual-sub {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  margin-top: 2px;
}
.action-icons {
  display: inline-flex;
  align-items: center;
  gap: 2px;
}
</style>
