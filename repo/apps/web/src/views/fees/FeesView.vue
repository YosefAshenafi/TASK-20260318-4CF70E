<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'

import { apiGet, apiPatch, apiPost } from '@/api/http'
import { useCreateScopeContext } from '@/composables/useDataScope'
import { useAuthStore } from '@/stores/auth'

type FeeRow = {
  id: string
  institutionId: string
  departmentId?: string
  teamId?: string
  caseId?: string
  candidateId?: string
  feeType: string
  amount: number
  currency: string
  note?: string
  createdByUserId: string
  updatedByUserId?: string
  createdAt: string
  updatedAt: string
}

type ListResp = {
  items: FeeRow[]
  total: number
  page: number
  pageSize: number
}

const auth = useAuthStore()
const { requireContext } = useCreateScopeContext()
const canManage = () => auth.hasPermission('fees.manage')

const loading = ref(false)
const rows = ref<FeeRow[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)

const dialogVisible = ref(false)
const dialogSaving = ref(false)
const formFeeType = ref('case_handling')
const formAmount = ref<number | undefined>(undefined)
const formCurrency = ref('CNY')
const formCaseId = ref('')
const formCandidateId = ref('')
const formNote = ref('')

async function load() {
  loading.value = true
  try {
    const q = new URLSearchParams({
      page: String(page.value),
      pageSize: String(pageSize.value),
      sortBy: 'created_at',
      sortOrder: 'desc',
    })
    const data = await apiGet<ListResp>(`/api/v1/fees?${q}`)
    rows.value = data.items
    total.value = data.total
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : 'Failed to load fees')
  } finally {
    loading.value = false
  }
}

function onPageChange(p: number) {
  page.value = p
  load()
}

function openCreate() {
  formFeeType.value = 'case_handling'
  formAmount.value = undefined
  formCurrency.value = 'CNY'
  formCaseId.value = ''
  formCandidateId.value = ''
  formNote.value = ''
  dialogVisible.value = true
}

async function submitCreate() {
  if (!formFeeType.value.trim() || formAmount.value == null || formAmount.value <= 0) {
    ElMessage.warning('Fee type and positive amount are required.')
    return
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
    await apiPost<FeeRow>('/api/v1/fees', {
      institutionId: scope.institutionId,
      departmentId: scope.departmentId,
      teamId: scope.teamId,
      caseId: formCaseId.value.trim() || undefined,
      candidateId: formCandidateId.value.trim() || undefined,
      feeType: formFeeType.value.trim(),
      amount: formAmount.value,
      currency: formCurrency.value.trim() || 'CNY',
      note: formNote.value.trim() || undefined,
    })
    ElMessage.success('Fee record created.')
    dialogVisible.value = false
    await load()
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : 'Create failed')
  } finally {
    dialogSaving.value = false
  }
}

async function editRow(row: FeeRow) {
  if (!canManage()) return
  let amountValue: string
  let noteValue: string
  try {
    const a = await ElMessageBox.prompt('Amount', 'Edit fee amount', {
      inputValue: String(row.amount),
    })
    amountValue = a.value
    const n = await ElMessageBox.prompt('Note (optional)', 'Edit fee note', {
      inputValue: row.note ?? '',
      inputType: 'textarea',
    })
    noteValue = n.value
  } catch {
    return
  }
  const amount = Number(amountValue)
  if (!Number.isFinite(amount) || amount <= 0) {
    ElMessage.warning('Amount must be a positive number.')
    return
  }
  try {
    await apiPatch<FeeRow>(`/api/v1/fees/${row.id}`, {
      amount,
      note: noteValue.trim() || undefined,
    })
    ElMessage.success('Fee record updated.')
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
      <h2 class="rec-title">Fees</h2>
      <el-button v-if="canManage()" type="primary" round @click="openCreate">Add fee</el-button>
    </div>

    <el-card class="rec-card" shadow="never">
      <el-table v-loading="loading" :data="rows" stripe empty-text="No fees to show">
        <el-table-column prop="feeType" label="Type" width="160" />
        <el-table-column label="Amount" width="120">
          <template #default="{ row }">
            {{ row.currency }} {{ Number(row.amount).toFixed(2) }}
          </template>
        </el-table-column>
        <el-table-column prop="caseId" label="Case" width="220" show-overflow-tooltip>
          <template #default="{ row }">
            {{ row.caseId || '—' }}
          </template>
        </el-table-column>
        <el-table-column prop="candidateId" label="Candidate" width="220" show-overflow-tooltip>
          <template #default="{ row }">
            {{ row.candidateId || '—' }}
          </template>
        </el-table-column>
        <el-table-column prop="note" label="Note" min-width="240" show-overflow-tooltip>
          <template #default="{ row }">
            {{ row.note || '—' }}
          </template>
        </el-table-column>
        <el-table-column prop="updatedAt" label="Updated" width="180">
          <template #default="{ row }">
            {{ new Date(row.updatedAt).toLocaleString() }}
          </template>
        </el-table-column>
        <el-table-column v-if="canManage()" label="" width="90" fixed="right" align="center">
          <template #default="{ row }">
            <el-button link type="primary" @click="editRow(row)">Edit</el-button>
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

    <el-dialog v-model="dialogVisible" title="New fee" width="520px" destroy-on-close>
      <el-form label-position="top">
        <el-form-item label="Fee type" required>
          <el-input v-model="formFeeType" placeholder="e.g. case_handling" />
        </el-form-item>
        <el-form-item label="Amount" required>
          <el-input-number v-model="formAmount" :min="0.01" :precision="2" :step="10" style="width: 100%" />
        </el-form-item>
        <el-form-item label="Currency">
          <el-input v-model="formCurrency" placeholder="CNY" />
        </el-form-item>
        <el-form-item label="Case ID (optional)">
          <el-input v-model="formCaseId" />
        </el-form-item>
        <el-form-item label="Candidate ID (optional)">
          <el-input v-model="formCandidateId" />
        </el-form-item>
        <el-form-item label="Note">
          <el-input v-model="formNote" type="textarea" :rows="3" />
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
</style>
