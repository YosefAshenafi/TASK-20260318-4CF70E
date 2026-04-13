<script setup lang="ts">
import { EditPen, Search } from '@element-plus/icons-vue'
import { onMounted, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'

import { apiGet, apiPatch, apiPost } from '@/api/http'
import { useCreateScopeContext } from '@/composables/useDataScope'
import { humanizeTechnicalLabel } from '@/utils/display'
import { useAuthStore } from '@/stores/auth'

type RestrictionRow = {
  id: string
  institutionId: string
  clientId: string
  medicationId?: string
  rule: Record<string, unknown>
  isActive: boolean
  createdAt: string
}

type ViolationRow = {
  id: string
  clientId: string
  medicationId?: string
  details?: Record<string, unknown>
  createdAt: string
}

type ListResp = {
  items: RestrictionRow[]
  total: number
  page: number
  pageSize: number
}

type ViolationListResp = {
  items: ViolationRow[]
  total: number
  page: number
  pageSize: number
}

const auth = useAuthStore()
const { requireContext } = useCreateScopeContext()
const canManage = () => auth.hasPermission('compliance.manage')

const loading = ref(false)
const rows = ref<RestrictionRow[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)

const violLoading = ref(false)
const violRows = ref<ViolationRow[]>([])
const violTotal = ref(0)
const violPage = ref(1)

const dialogVisible = ref(false)
const dialogSaving = ref(false)
const formClient = ref('')
const formMed = ref('')
const formRx = ref(false)
const formFreq = ref<number | undefined>(undefined)
const formActive = ref(true)

const checkVisible = ref(false)
const checkSaving = ref(false)
const checkClient = ref('')
const checkMed = ref('')
const checkControlled = ref(true)
const checkRxId = ref('')
const checkAt = ref('')

function ruleSummary(rule: Record<string, unknown>): string {
  const parts: string[] = []
  if (rule.requiresPrescription) {
    parts.push('Rx required')
  }
  if (typeof rule.frequencyDays === 'number' && rule.frequencyDays > 0) {
    parts.push(`${rule.frequencyDays}d frequency`)
  }
  return parts.length ? parts.join(' · ') : '—'
}

/** Human-readable violation line for operators (no raw JSON in the grid). */
function violationReasonText(details: Record<string, unknown> | undefined): string {
  if (!details || typeof details !== 'object') {
    return '—'
  }
  const reasons = details.reasons
  if (Array.isArray(reasons) && reasons.length > 0) {
    const lines = reasons.filter((r): r is string => typeof r === 'string' && r.trim() !== '')
    if (lines.length) {
      return lines.join('; ')
    }
  }
  if (typeof details.message === 'string' && details.message.trim() !== '') {
    return details.message.trim()
  }
  const raw = details.reason
  if (typeof raw === 'string' && raw.trim() !== '') {
    const r = raw.trim()
    // Sentence-style reasons (e.g. demo seed / legacy rows) — show as-is.
    if (/[\s]/.test(r) || r === r.toLowerCase()) {
      return r
    }
    switch (r) {
      case 'PRESCRIPTION_REQUIRED':
        return 'Prescription attachment required for controlled medication'
      case 'FREQUENCY':
        return 'Purchase blocked by frequency restriction (within the allowed window)'
      case 'RESTRICTION_VIOLATION':
        return 'Purchase blocked by restriction rule'
      default:
        return r.includes('_') || r === r.toUpperCase()
          ? humanizeTechnicalLabel(r.toLowerCase())
          : r
    }
  }
  return '—'
}

async function load() {
  loading.value = true
  try {
    const q = new URLSearchParams({
      page: String(page.value),
      pageSize: String(pageSize.value),
      sortBy: 'created_at',
      sortOrder: 'desc',
    })
    const data = await apiGet<ListResp>(`/api/v1/compliance/restrictions?${q}`)
    rows.value = data.items
    total.value = data.total
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : 'Failed to load restrictions')
  } finally {
    loading.value = false
  }
}

async function loadViolations() {
  violLoading.value = true
  try {
    const q = new URLSearchParams({
      page: String(violPage.value),
      pageSize: '10',
      sortBy: 'created_at',
      sortOrder: 'desc',
    })
    const data = await apiGet<ViolationListResp>(`/api/v1/compliance/restrictions/violations?${q}`)
    violRows.value = data.items
    violTotal.value = data.total
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : 'Failed to load violations')
  } finally {
    violLoading.value = false
  }
}

function onPageChange(p: number) {
  page.value = p
  load()
}

function onViolPageChange(p: number) {
  violPage.value = p
  loadViolations()
}

function openCreate() {
  formClient.value = ''
  formMed.value = ''
  formRx.value = true
  formFreq.value = 7
  formActive.value = true
  dialogVisible.value = true
}

async function submitCreate() {
  if (!formClient.value.trim()) {
    ElMessage.warning('Client ID is required.')
    return
  }
  const rule: Record<string, unknown> = {}
  if (formRx.value) {
    rule.requiresPrescription = true
  }
  if (formFreq.value != null && formFreq.value > 0) {
    rule.frequencyDays = formFreq.value
  }
  dialogSaving.value = true
  try {
    let scope
    try {
      scope = requireContext()
    } catch (err) {
      ElMessage.error(err instanceof Error ? err.message : 'No data scope')
      return
    }
    await apiPost<RestrictionRow>('/api/v1/compliance/restrictions', {
      institutionId: scope.institutionId,
      clientId: formClient.value.trim(),
      medicationId: formMed.value.trim() || undefined,
      rule,
      isActive: formActive.value,
    })
    ElMessage.success('Restriction created.')
    dialogVisible.value = false
    await load()
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : 'Create failed')
  } finally {
    dialogSaving.value = false
  }
}

async function toggleActive(row: RestrictionRow) {
  try {
    await apiPatch<RestrictionRow>(`/api/v1/compliance/restrictions/${row.id}`, {
      isActive: !row.isActive,
    })
    ElMessage.success('Updated.')
    await load()
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : 'Update failed')
  }
}

async function editRule(row: RestrictionRow) {
  const cur = { ...(row.rule || {}) }
  let freqStr: string
  try {
    const res = await ElMessageBox.prompt('Frequency window (days, 0 to disable)', 'Edit rule', {
      inputValue: String(typeof cur.frequencyDays === 'number' ? cur.frequencyDays : ''),
    })
    freqStr = res.value
  } catch {
    return
  }
  let freq = 0
  if (freqStr.trim() !== '') {
    const n = parseInt(freqStr.trim(), 10)
    if (!Number.isNaN(n)) {
      freq = n
    }
  }
  const rule: Record<string, unknown> = { ...cur }
  if (freq > 0) {
    rule.frequencyDays = freq
  } else {
    delete rule.frequencyDays
  }
  try {
    await apiPatch<RestrictionRow>(`/api/v1/compliance/restrictions/${row.id}`, { rule })
    ElMessage.success('Rule updated.')
    await load()
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : 'Update failed')
  }
}

function openCheck() {
  const now = new Date()
  checkClient.value = 'client-demo-1'
  checkMed.value = 'med-controlled-1'
  checkControlled.value = true
  checkRxId.value = ''
  checkAt.value = now.toISOString().slice(0, 16)
  checkVisible.value = true
}

async function submitCheck() {
  if (!checkClient.value.trim() || !checkMed.value.trim()) {
    ElMessage.warning('Client and medication are required.')
    return
  }
  let purchaseAt = new Date(checkAt.value)
  if (Number.isNaN(purchaseAt.getTime())) {
    purchaseAt = new Date()
  }
  checkSaving.value = true
  try {
    let scope
    try {
      scope = requireContext()
    } catch (err) {
      ElMessage.error(err instanceof Error ? err.message : 'No data scope')
      return
    }
    const res = await apiPost<{ allowed: boolean; reasons: string[] }>('/api/v1/compliance/restrictions/check-purchase', {
      institutionId: scope.institutionId,
      clientId: checkClient.value.trim(),
      medicationId: checkMed.value.trim(),
      isControlled: checkControlled.value,
      prescriptionAttachmentId: checkRxId.value.trim() || undefined,
      purchaseAt: purchaseAt.toISOString(),
    })
    if (res.allowed) {
      ElMessage.success('Purchase allowed.')
    } else {
      ElMessage.warning(res.reasons?.length ? res.reasons.join(', ') : 'Blocked')
    }
    checkVisible.value = false
    await loadViolations()
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : 'Check failed')
  } finally {
    checkSaving.value = false
  }
}

async function runExpirationJob() {
  try {
    await ElMessageBox.confirm(
      'Run the daily expiration check for the qualifications you can manage?',
      'Expiration check',
      { type: 'info' },
    )
  } catch {
    return
  }
  try {
    const res = await apiPost<{ deactivated: number }>('/api/v1/compliance/jobs/qualifications/run', {})
    ElMessage.success(`Deactivated ${res.deactivated} profile(s).`)
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : 'Job failed')
  }
}

onMounted(async () => {
  await load()
  await loadViolations()
})
</script>

<template>
  <div class="rec-page">
    <div class="rec-toolbar">
      <h2 class="rec-title">Purchase restrictions</h2>
      <div class="rec-actions">
        <el-button v-if="canManage()" round @click="runExpirationJob">Run qualification expiry job</el-button>
        <el-button v-if="canManage()" type="primary" round :icon="Search" @click="openCheck">Check purchase</el-button>
        <el-button v-if="canManage()" type="primary" round @click="openCreate">Add restriction</el-button>
      </div>
    </div>

    <el-card class="rec-card" shadow="never">
      <el-table v-loading="loading" :data="rows" stripe empty-text="No restrictions to show">
        <el-table-column label="Client" min-width="160">
          <template #default="{ row }">
            <span>{{ row.clientId }}</span>
            <div v-if="row.medicationId" class="qual-sub">Med: {{ row.medicationId }}</div>
            <div v-else class="qual-sub">All medications</div>
          </template>
        </el-table-column>
        <el-table-column label="Rule" min-width="200">
          <template #default="{ row }">
            {{ ruleSummary(row.rule) }}
          </template>
        </el-table-column>
        <el-table-column prop="isActive" label="Active" width="100">
          <template #default="{ row }">
            <el-switch
              v-if="canManage()"
              :model-value="row.isActive"
              @change="() => {
                void toggleActive(row)
              }"
            />
            <el-tag v-else :type="row.isActive ? 'success' : 'info'" size="small">
              {{ row.isActive ? 'yes' : 'no' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column v-if="canManage()" label="Actions" width="88" fixed="right" align="center">
          <template #default="{ row }">
            <el-tooltip content="Edit rule" placement="top">
              <el-button
                link
                type="primary"
                :icon="EditPen"
                aria-label="Edit rule"
                @click="editRule(row)"
              />
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

    <h3 class="rec-sub">Violation history</h3>
    <el-card class="rec-card" shadow="never">
      <el-table v-loading="violLoading" :data="violRows" stripe empty-text="No violations recorded">
        <el-table-column prop="clientId" label="Client" width="140" />
        <el-table-column prop="medicationId" label="Medication" width="160">
          <template #default="{ row }">
            {{ row.medicationId || '—' }}
          </template>
        </el-table-column>
        <el-table-column label="Reason" min-width="280" show-overflow-tooltip>
          <template #default="{ row }">
            {{ violationReasonText(row.details) }}
          </template>
        </el-table-column>
        <el-table-column prop="createdAt" label="Recorded" width="180">
          <template #default="{ row }">
            {{ new Date(row.createdAt).toLocaleString() }}
          </template>
        </el-table-column>
      </el-table>
      <div class="rec-pager">
        <el-pagination
          background
          layout="prev, pager, next, total"
          :total="violTotal"
          :page-size="10"
          :current-page="violPage"
          @current-change="onViolPageChange"
        />
      </div>
    </el-card>

    <el-dialog v-model="dialogVisible" title="New restriction" width="440px" destroy-on-close>
      <el-form label-position="top">
        <el-form-item label="Client ID" required>
          <el-input v-model="formClient" placeholder="e.g. client-demo-1" />
        </el-form-item>
        <el-form-item label="Medication ID (optional)">
          <el-input v-model="formMed" placeholder="Leave empty for all meds" />
        </el-form-item>
        <el-form-item label="Requires prescription (controlled)">
          <el-switch v-model="formRx" />
        </el-form-item>
        <el-form-item label="Frequency (days)">
          <el-input-number v-model="formFreq" :min="0" :max="365" controls-position="right" style="width: 100%" />
        </el-form-item>
        <el-form-item label="Active">
          <el-switch v-model="formActive" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">Cancel</el-button>
        <el-button type="primary" :loading="dialogSaving" @click="submitCreate">Create</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="checkVisible" title="Check purchase" width="520px" destroy-on-close>
      <el-form label-position="top">
        <el-form-item label="Client ID" required>
          <el-input v-model="checkClient" />
        </el-form-item>
        <el-form-item label="Medication ID" required>
          <el-input v-model="checkMed" />
        </el-form-item>
        <el-form-item label="Controlled medication">
          <el-switch v-model="checkControlled" />
        </el-form-item>
        <el-form-item label="Prescription attachment ID (optional)">
          <el-input v-model="checkRxId" placeholder="Prescription attachment id if required" />
        </el-form-item>
        <el-form-item label="Purchase time (local)">
          <el-input v-model="checkAt" type="datetime-local" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="checkVisible = false">Cancel</el-button>
        <el-button type="primary" :loading="checkSaving" @click="submitCheck">Run check</el-button>
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
  flex-wrap: wrap;
}
.rec-actions {
  display: flex;
  gap: 0.5rem;
  flex-wrap: wrap;
}
.rec-title {
  margin: 0;
  font-size: 1.25rem;
  font-weight: 650;
  letter-spacing: -0.02em;
}
.rec-sub {
  margin: 1.5rem 0 0.75rem;
  font-size: 1rem;
  font-weight: 650;
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
.qual-sub {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  margin-top: 2px;
}
</style>
