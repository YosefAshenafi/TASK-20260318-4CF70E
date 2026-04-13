<script setup lang="ts">
import { EditPen, View } from '@element-plus/icons-vue'
import { onMounted, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'

import { apiGet, apiPatch, apiPost } from '@/api/http'
import { DEV_ADMIN_USER_ID, DEV_INSTITUTION_ID } from '@/config/devSeed'
import { useAuthStore } from '@/stores/auth'

type CaseRow = {
  id: string
  caseNumber: string
  caseType: string
  title: string
  description: string
  status: string
  assigneeId?: string
  reportedAt: string
  createdAt: string
}

type ProcessingRow = {
  id: string
  stepCode: string
  actorUserId: string
  note?: string
  createdAt: string
}

type TransitionRow = {
  id: string
  fromStatus: string
  toStatus: string
  actorUserId: string
  createdAt: string
}

type ListResp = {
  items: CaseRow[]
  total: number
  page: number
  pageSize: number
}

const auth = useAuthStore()
const canManage = () => auth.hasPermission('cases.manage')

const loading = ref(false)
const rows = ref<CaseRow[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const searchQ = ref('')
const statusFilter = ref('')

const detailVisible = ref(false)
const detailLoading = ref(false)
const activeCase = ref<CaseRow | null>(null)
const processing = ref<ProcessingRow[]>([])
const transitions = ref<TransitionRow[]>([])

const createVisible = ref(false)
const createSaving = ref(false)
const formType = ref('quality')
const formTitle = ref('')
const formDesc = ref('')
const formReported = ref('')

const statusOptions = [
  { value: '', label: 'All statuses' },
  { value: 'submitted', label: 'Submitted' },
  { value: 'assigned', label: 'Assigned' },
  { value: 'in_progress', label: 'In progress' },
  { value: 'pending_review', label: 'Pending review' },
  { value: 'closed', label: 'Closed' },
]

const nextStatus = ref('')

async function load() {
  loading.value = true
  try {
    const q = new URLSearchParams({
      page: String(page.value),
      pageSize: String(pageSize.value),
      sortBy: 'created_at',
      sortOrder: 'desc',
    })
    if (searchQ.value.trim()) {
      q.set('q', searchQ.value.trim())
    }
    if (statusFilter.value) {
      q.set('status', statusFilter.value)
    }
    const data = await apiGet<ListResp>(`/api/v1/cases?${q}`)
    rows.value = data.items
    total.value = data.total
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : 'Failed to load cases')
  } finally {
    loading.value = false
  }
}

function onPageChange(p: number) {
  page.value = p
  load()
}

function openCreate() {
  const t = new Date()
  t.setMinutes(t.getMinutes() - t.getTimezoneOffset())
  formType.value = 'quality'
  formTitle.value = ''
  formDesc.value = ''
  formReported.value = t.toISOString().slice(0, 16)
  createVisible.value = true
}

async function submitCreate() {
  if (!formTitle.value.trim() || !formDesc.value.trim()) {
    ElMessage.warning('Title and description are required.')
    return
  }
  const reported = new Date(formReported.value)
  if (Number.isNaN(reported.getTime())) {
    ElMessage.warning('Invalid reported time.')
    return
  }
  createSaving.value = true
  try {
    await apiPost<CaseRow>('/api/v1/cases', {
      institutionId: DEV_INSTITUTION_ID,
      caseType: formType.value.trim(),
      title: formTitle.value.trim(),
      description: formDesc.value.trim(),
      reportedAt: reported.toISOString(),
    })
    ElMessage.success('Case created.')
    createVisible.value = false
    await load()
  } catch (e) {
    const msg = e instanceof Error ? e.message : 'Create failed'
    if (msg.includes('duplicate') || msg.includes('409')) {
      ElMessage.warning('Duplicate submission blocked (same details within 5 minutes).')
    } else {
      ElMessage.error(msg)
    }
  } finally {
    createSaving.value = false
  }
}

async function openDetail(row: CaseRow) {
  nextStatus.value = ''
  activeCase.value = row
  detailVisible.value = true
  detailLoading.value = true
  processing.value = []
  transitions.value = []
  try {
    const [c, pr, tr] = await Promise.all([
      apiGet<CaseRow>(`/api/v1/cases/${row.id}`),
      apiGet<{ items: ProcessingRow[] }>(`/api/v1/cases/${row.id}/processing-records`),
      apiGet<{ items: TransitionRow[] }>(`/api/v1/cases/${row.id}/status-transitions`),
    ])
    activeCase.value = c
    processing.value = pr.items ?? []
    transitions.value = tr.items ?? []
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : 'Failed to load case')
  } finally {
    detailLoading.value = false
  }
}

async function assignToAdmin() {
  const row = activeCase.value
  if (!row) {
    return
  }
  try {
    await ElMessageBox.confirm('Assign this case to the administrator account?', 'Assign', { type: 'info' })
  } catch {
    return
  }
  try {
    const c = await apiPost<CaseRow>(`/api/v1/cases/${row.id}/assign`, {
      assigneeUserId: DEV_ADMIN_USER_ID,
    })
    activeCase.value = c
    ElMessage.success('Assigned.')
    await load()
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : 'Assign failed')
  }
}

async function addProcessingNote() {
  const row = activeCase.value
  if (!row) {
    return
  }
  let note: string
  try {
    const res = await ElMessageBox.prompt('Note (optional)', 'Processing step', {
      confirmButtonText: 'Add',
      inputPlaceholder: 'e.g. Reviewed batch record',
    })
    note = res.value
  } catch {
    return
  }
  try {
    await apiPost(`/api/v1/cases/${row.id}/processing-records`, {
      stepCode: 'note',
      note: note.trim() || undefined,
    })
    ElMessage.success('Recorded.')
    await openDetail(row)
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : 'Failed')
  }
}

async function applyTransition() {
  const row = activeCase.value
  const toStatus = nextStatus.value
  if (!row || !toStatus) {
    return
  }
  try {
    await ElMessageBox.confirm(`Move status to “${toStatus}”?`, 'Status', { type: 'warning' })
  } catch {
    return
  }
  try {
    const c = await apiPost<CaseRow>(`/api/v1/cases/${row.id}/status-transitions`, { toStatus })
    activeCase.value = c
    nextStatus.value = ''
    ElMessage.success('Status updated.')
    await load()
    await openDetail(c)
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : 'Transition failed')
  }
}

async function editCase() {
  const row = activeCase.value
  if (!row) {
    return
  }
  let title: string
  let desc: string
  try {
    const r1 = await ElMessageBox.prompt('Title', 'Edit case', { inputValue: row.title })
    title = r1.value
    const r2 = await ElMessageBox.prompt('Description', 'Edit case', {
      inputValue: row.description,
      inputType: 'textarea',
    })
    desc = r2.value
  } catch {
    return
  }
  try {
    const c = await apiPatch<CaseRow>(`/api/v1/cases/${row.id}`, {
      title: title.trim(),
      description: desc.trim(),
    })
    activeCase.value = c
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
      <h2 class="rec-title">Cases</h2>
      <el-button v-if="canManage()" type="primary" round @click="openCreate">New case</el-button>
    </div>

    <el-card class="rec-card" shadow="never">
      <div class="case-filters">
        <el-input
          v-model="searchQ"
          clearable
          placeholder="Search title, number, type…"
          style="max-width: 280px"
          @clear="load"
          @keyup.enter="load"
        />
        <el-select v-model="statusFilter" placeholder="Status" clearable style="width: 180px" @change="load">
          <el-option v-for="o in statusOptions" :key="o.value || 'all'" :label="o.label" :value="o.value" />
        </el-select>
        <el-button @click="load">Search</el-button>
      </div>
      <el-table v-loading="loading" :data="rows" stripe empty-text="No cases to show">
        <el-table-column prop="caseNumber" label="Case #" width="200" show-overflow-tooltip />
        <el-table-column prop="title" label="Title" min-width="200" show-overflow-tooltip />
        <el-table-column prop="caseType" label="Type" width="120" />
        <el-table-column prop="status" label="Status" width="130">
          <template #default="{ row }">
            <el-tag size="small">{{ row.status }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="reportedAt" label="Reported" width="180">
          <template #default="{ row }">
            {{ new Date(row.reportedAt).toLocaleString() }}
          </template>
        </el-table-column>
        <el-table-column label="" width="88" fixed="right" align="center">
          <template #default="{ row }">
            <el-tooltip content="Open details" placement="top">
              <el-button link type="primary" :icon="View" aria-label="Open case" @click="openDetail(row)" />
            </el-tooltip>
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

    <el-dialog v-model="createVisible" title="New case" width="480px" destroy-on-close>
      <el-form label-position="top">
        <el-form-item label="Type" required>
          <el-input v-model="formType" placeholder="e.g. quality, safety" />
        </el-form-item>
        <el-form-item label="Title" required>
          <el-input v-model="formTitle" />
        </el-form-item>
        <el-form-item label="Description" required>
          <el-input v-model="formDesc" type="textarea" :rows="4" />
        </el-form-item>
        <el-form-item label="Reported at (local)">
          <el-input v-model="formReported" type="datetime-local" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createVisible = false">Cancel</el-button>
        <el-button type="primary" :loading="createSaving" @click="submitCreate">Create</el-button>
      </template>
    </el-dialog>

    <el-drawer v-model="detailVisible" :title="activeCase?.caseNumber ?? 'Case'" size="520px" destroy-on-close>
      <div v-if="activeCase" v-loading="detailLoading" class="detail-body">
        <p class="detail-title">{{ activeCase.title }}</p>
        <el-descriptions :column="1" border size="small">
          <el-descriptions-item label="Status">
            <el-tag size="small">{{ activeCase.status }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="Type">{{ activeCase.caseType }}</el-descriptions-item>
          <el-descriptions-item label="Reported">{{ new Date(activeCase.reportedAt).toLocaleString() }}</el-descriptions-item>
          <el-descriptions-item label="Description">
            <span class="detail-desc">{{ activeCase.description }}</span>
          </el-descriptions-item>
        </el-descriptions>

        <div v-if="canManage()" class="detail-actions">
          <el-button size="small" :icon="EditPen" @click="editCase">Edit title/description</el-button>
          <el-button size="small" @click="assignToAdmin">Assign to admin (dev)</el-button>
        </div>
        <div v-if="canManage()" class="detail-actions">
          <span class="muted">Next status</span>
          <el-select v-model="nextStatus" placeholder="Select transition" clearable style="width: 200px" size="small">
            <el-option label="assigned" value="assigned" />
            <el-option label="in_progress" value="in_progress" />
            <el-option label="pending_review" value="pending_review" />
            <el-option label="closed" value="closed" />
          </el-select>
          <el-button size="small" type="primary" :disabled="!nextStatus" @click="applyTransition">Apply</el-button>
        </div>
        <div v-if="canManage()" class="detail-actions">
          <el-button size="small" @click="addProcessingNote">Add processing note</el-button>
        </div>

        <h4 class="detail-h">Timeline</h4>
        <el-timeline>
          <el-timeline-item
            v-for="p in processing"
            :key="p.id"
            :timestamp="new Date(p.createdAt).toLocaleString()"
          >
            <strong>{{ p.stepCode }}</strong>
            <span v-if="p.note" class="muted"> — {{ p.note }}</span>
          </el-timeline-item>
        </el-timeline>

        <h4 class="detail-h">Status history</h4>
        <el-timeline>
          <el-timeline-item
            v-for="t in transitions"
            :key="t.id"
            :timestamp="new Date(t.createdAt).toLocaleString()"
          >
            {{ t.fromStatus }} → {{ t.toStatus }}
          </el-timeline-item>
        </el-timeline>
      </div>
    </el-drawer>
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
.case-filters {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  margin-bottom: 1rem;
  align-items: center;
}
.detail-body {
  min-height: 120px;
}
.detail-title {
  font-weight: 650;
  margin: 0 0 0.75rem;
}
.detail-desc {
  white-space: pre-wrap;
}
.detail-actions {
  margin-top: 0.75rem;
  display: flex;
  flex-wrap: wrap;
  gap: 0.35rem;
  align-items: center;
}
.detail-h {
  margin: 1.25rem 0 0.5rem;
  font-size: 0.95rem;
  font-weight: 650;
}
.muted {
  color: var(--el-text-color-secondary);
  font-size: 12px;
}
</style>
