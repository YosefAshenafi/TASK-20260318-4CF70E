<script setup lang="ts">
import { Delete, EditPen } from '@element-plus/icons-vue'
import { onMounted, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'

import { apiDelete, apiGet, apiPatch, apiPost, apiPutBytes } from '@/api/http'
import { useCreateScopeContext } from '@/composables/useDataScope'

type CandidateRow = {
  id: string
  name: string
  phoneMasked: string
  idNumberMasked: string
  email?: string
  phone?: string
  idNumber?: string
  experienceYears?: number
  educationLevel?: string
  skills: string[]
  tags: string[]
  customFields: Record<string, unknown>
  institutionId: string
  createdAt: string
  updatedAt: string
}
type ListResp = { items: CandidateRow[]; total: number; page: number; pageSize: number }
type PositionOption = { id: string; title: string }
type DuplicateGroup = { matchType: string; institutionId: string; candidateIds: string[] }
type MatchScore = {
  score: number
  breakdown: { skills: number; experience: number; education: number }
  reasons: string[]
}
type CustomFieldKV = { key: string; value: string }

const { requireContext } = useCreateScopeContext()
const loading = ref(false)
const rows = ref<CandidateRow[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)

const filterKeyword = ref('')
const filterSkills = ref('')
const filterEdu = ref('')
const filterMinExp = ref<number | undefined>()
const filterMaxExp = ref<number | undefined>()
const filterCreatedFrom = ref('')
const filterCreatedTo = ref('')
const filterUpdatedFrom = ref('')
const filterUpdatedTo = ref('')

const createVisible = ref(false)
const createSaving = ref(false)
const formName = ref('')
const formExp = ref<number | undefined>()
const formEdu = ref('')
const formSkills = ref('')
const formTags = ref('')

const editVisible = ref(false)
const editSaving = ref(false)
const editCandidateId = ref('')
const editName = ref('')
const editPhone = ref('')
const editIDNumber = ref('')
const editEmail = ref('')
const editExp = ref<number | undefined>()
const editEdu = ref('')
const editSkills = ref('')
const editTags = ref('')
const editCustomFields = ref<CustomFieldKV[]>([])

const importVisible = ref(false)
const importSaving = ref(false)
const importUploading = ref(false)
const importPreviewRows = ref<CustomFieldKV[]>([])
const importValidationErrors = ref<string[]>([])
const importBatchId = ref('')
const importFileInput = ref<File[]>([])

const duplicateLoading = ref(false)
const duplicates = ref<DuplicateGroup[]>([])

const matchVisible = ref(false)
const matching = ref(false)
const matchCandidateId = ref('')
const matchPositionId = ref('')
const positions = ref<PositionOption[]>([])
const matchResult = ref<MatchScore | null>(null)
const similarCandidates = ref<CandidateRow[]>([])
const similarPositions = ref<PositionOption[]>([])

function splitCSV(input: string): string[] {
  const seen = new Set<string>()
  return input
    .split(',')
    .map((x) => x.trim())
    .filter((x) => x.length > 0)
    .filter((x) => {
      if (seen.has(x)) return false
      seen.add(x)
      return true
    })
}
function customFieldMap(rows: CustomFieldKV[]): Record<string, unknown> {
  const out: Record<string, unknown> = {}
  for (const row of rows) {
    const k = row.key.trim()
    if (!k) continue
    out[k] = row.value
  }
  return out
}
function customFieldRows(input: Record<string, unknown> | undefined): CustomFieldKV[] {
  if (!input) return []
  return Object.keys(input)
    .sort()
    .map((k) => ({ key: k, value: String(input[k] ?? '') }))
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
    if (filterKeyword.value.trim()) q.set('keyword', filterKeyword.value.trim())
    if (filterSkills.value.trim()) q.set('skills', filterSkills.value.trim())
    if (filterEdu.value.trim()) q.set('educationLevel', filterEdu.value.trim())
    if (filterMinExp.value != null) q.set('minExperience', String(filterMinExp.value))
    if (filterMaxExp.value != null) q.set('maxExperience', String(filterMaxExp.value))
    if (filterCreatedFrom.value.trim()) q.set('createdFrom', filterCreatedFrom.value.trim())
    if (filterCreatedTo.value.trim()) q.set('createdTo', filterCreatedTo.value.trim())
    if (filterUpdatedFrom.value.trim()) q.set('updatedFrom', filterUpdatedFrom.value.trim())
    if (filterUpdatedTo.value.trim()) q.set('updatedTo', filterUpdatedTo.value.trim())
    const data = await apiGet<ListResp>(`/api/v1/recruitment/candidates?${q}`)
    rows.value = data.items
    total.value = data.total
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : 'Failed to load candidates')
  } finally {
    loading.value = false
  }
}
async function loadDuplicates() {
  duplicateLoading.value = true
  try {
    const data = await apiGet<{ items: DuplicateGroup[] }>('/api/v1/recruitment/candidates/duplicates')
    duplicates.value = data.items ?? []
  } catch {
    duplicates.value = []
  } finally {
    duplicateLoading.value = false
  }
}
async function loadPositions() {
  try {
    const data = await apiGet<{ items: PositionOption[] }>('/api/v1/recruitment/positions?page=1&pageSize=100')
    positions.value = data.items ?? []
  } catch {
    positions.value = []
  }
}

function onPageChange(p: number) {
  page.value = p
  load()
}
function applyFilters() {
  page.value = 1
  load()
}

function openCreate() {
  formName.value = ''
  formExp.value = undefined
  formEdu.value = ''
  formSkills.value = ''
  formTags.value = ''
  createVisible.value = true
}
async function submitCreate() {
  if (!formName.value.trim()) return ElMessage.warning('Name is required.')
  createSaving.value = true
  try {
    const scope = requireContext()
    await apiPost<CandidateRow>('/api/v1/recruitment/candidates', {
      name: formName.value.trim(),
      institutionId: scope.institutionId,
      departmentId: scope.departmentId,
      teamId: scope.teamId,
      experienceYears: formExp.value,
      educationLevel: formEdu.value || undefined,
      skills: splitCSV(formSkills.value),
      tags: splitCSV(formTags.value),
    })
    ElMessage.success('Candidate created. Duplicate phone/ID records are auto-merged.')
    createVisible.value = false
    await Promise.all([load(), loadDuplicates()])
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : 'Create failed')
  } finally {
    createSaving.value = false
  }
}

async function openEdit(row: CandidateRow) {
  editSaving.value = false
  try {
    const detail = await apiGet<CandidateRow>(`/api/v1/recruitment/candidates/${row.id}`)
    editCandidateId.value = detail.id
    editName.value = detail.name
    editPhone.value = detail.phone ?? ''
    editIDNumber.value = detail.idNumber ?? ''
    editEmail.value = detail.email ?? ''
    editExp.value = detail.experienceYears
    editEdu.value = detail.educationLevel ?? ''
    editSkills.value = (detail.skills ?? []).join(', ')
    editTags.value = (detail.tags ?? []).join(', ')
    editCustomFields.value = customFieldRows(detail.customFields)
    editVisible.value = true
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : 'Failed to load candidate details')
  }
}
async function submitEdit() {
  if (!editName.value.trim()) return ElMessage.warning('Name is required.')
  editSaving.value = true
  try {
    await apiPatch(`/api/v1/recruitment/candidates/${editCandidateId.value}`, {
      name: editName.value.trim(),
      phone: editPhone.value.trim(),
      idNumber: editIDNumber.value.trim(),
      email: editEmail.value.trim(),
      experienceYears: editExp.value,
      educationLevel: editEdu.value.trim() || undefined,
      skills: splitCSV(editSkills.value),
      tags: splitCSV(editTags.value),
      customFields: customFieldMap(editCustomFields.value),
    })
    ElMessage.success('Candidate updated.')
    editVisible.value = false
    await load()
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : 'Update failed')
  } finally {
    editSaving.value = false
  }
}

async function confirmDelete(row: CandidateRow) {
  try {
    await ElMessageBox.confirm(`Remove candidate “${row.name}”?`, 'Confirm', { type: 'warning' })
  } catch {
    return
  }
  try {
    await apiDelete(`/api/v1/recruitment/candidates/${row.id}`)
    ElMessage.success('Candidate removed.')
    await load()
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : 'Delete failed')
  }
}

function openImport() {
  importBatchId.value = ''
  importValidationErrors.value = []
  importPreviewRows.value = []
  importFileInput.value = []
  importVisible.value = true
}

function onResumePick(ev: Event) {
  const input = ev.target as HTMLInputElement
  importFileInput.value = Array.from(input.files ?? [])
}

function guessMime(file: File): string {
  if (file.type && file.type !== 'application/octet-stream') return file.type
  const n = file.name.toLowerCase()
  if (n.endsWith('.pdf')) return 'application/pdf'
  if (n.endsWith('.txt')) return 'text/plain'
  if (n.endsWith('.docx')) return 'application/vnd.openxmlformats-officedocument.wordprocessingml.document'
  if (n.endsWith('.doc')) return 'application/msword'
  return 'text/plain'
}

async function sha256Hex(buf: ArrayBuffer): Promise<string> {
  const hash = await crypto.subtle.digest('SHA-256', buf)
  return [...new Uint8Array(hash)].map((b) => b.toString(16).padStart(2, '0')).join('')
}

async function uploadResumeFile(file: File): Promise<string> {
  const chunkSize = 1024 * 1024
  const init = await apiPost<{ uploadId: string; totalChunks: number }>('/api/v1/files/uploads/init', {
    fileName: file.name,
    size: file.size,
    mimeType: guessMime(file),
    chunkSize,
  })
  const buf = await file.arrayBuffer()
  for (let i = 0; i < init.totalChunks; i++) {
    const start = i * chunkSize
    const end = Math.min(start + chunkSize, buf.byteLength)
    await apiPutBytes(`/api/v1/files/uploads/${init.uploadId}/chunks/${i}`, buf.slice(start, end))
  }
  const done = await apiPost<{ fileId: string }>(`/api/v1/files/uploads/${init.uploadId}/complete`, {
    sha256: await sha256Hex(buf),
  })
  return done.fileId
}

async function buildImportPreview() {
  if (!importFileInput.value.length) return ElMessage.warning('Select resume files first.')
  importUploading.value = true
  importValidationErrors.value = []
  try {
    const scope = requireContext()
    const fileIds: string[] = []
    for (const f of importFileInput.value) {
      fileIds.push(await uploadResumeFile(f))
    }
    const batch = await apiPost<{ id: string; validationReport?: { rows?: Array<Record<string, unknown>>; errors?: Array<{ rowIndex: number; message: string }>; warnings?: Array<{ rowIndex: number; message: string }> } }>(
      '/api/v1/recruitment/candidates/imports',
      {
        institutionId: scope.institutionId,
        resumeFileIds: fileIds,
      },
    )
    importBatchId.value = batch.id
    const rows = batch.validationReport?.rows ?? []
    importPreviewRows.value = rows.map((r) => ({ key: String(r.name ?? '(no name)'), value: String(r.email ?? r.phone ?? r.idNumber ?? 'no contact extracted') }))
    importValidationErrors.value = [
      ...(batch.validationReport?.errors ?? []).map((e) => `Row ${e.rowIndex + 1}: ${e.message}`),
      ...(batch.validationReport?.warnings ?? []).map((e) => `Row ${e.rowIndex + 1} warning: ${e.message}`),
    ]
    ElMessage.success('Preview generated from uploaded resumes.')
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : 'Failed to build import preview')
  } finally {
    importUploading.value = false
  }
}

async function commitImport() {
  if (!importBatchId.value) return ElMessage.warning('Generate preview first.')
  importSaving.value = true
  try {
    await apiPost(`/api/v1/recruitment/candidates/imports/${importBatchId.value}/commit`, {})
    ElMessage.success('Import committed. Duplicate rows are auto-merged by phone/ID.')
    importVisible.value = false
    await Promise.all([load(), loadDuplicates()])
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : 'Import failed')
  } finally {
    importSaving.value = false
  }
}

function openMatch(row: CandidateRow) {
  matchCandidateId.value = row.id
  matchPositionId.value = ''
  matchResult.value = null
  similarCandidates.value = []
  similarPositions.value = []
  matchVisible.value = true
}
async function runMatch() {
  if (!matchCandidateId.value || !matchPositionId.value) return ElMessage.warning('Select both candidate and position.')
  matching.value = true
  try {
    const res = await apiPost<MatchScore>('/api/v1/recruitment/match/candidate-to-position', {
      candidateId: matchCandidateId.value,
      positionId: matchPositionId.value,
    })
    matchResult.value = res
    const [sc, sp] = await Promise.all([
      apiGet<{ items: CandidateRow[] }>(`/api/v1/recruitment/recommendations/similar-candidates/${matchCandidateId.value}?limit=5`),
      apiGet<{ items: PositionOption[] }>(`/api/v1/recruitment/recommendations/similar-positions/${matchPositionId.value}?limit=5`),
    ])
    similarCandidates.value = sc.items ?? []
    similarPositions.value = sp.items ?? []
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : 'Match failed')
  } finally {
    matching.value = false
  }
}

onMounted(async () => {
  await Promise.all([load(), loadDuplicates(), loadPositions()])
})
</script>

<template>
  <div class="rec-page">
    <div class="rec-toolbar">
      <h2 class="rec-title">Candidates</h2>
      <div class="tool-actions">
        <el-button round @click="openImport">Bulk import</el-button>
        <el-button type="primary" round @click="openCreate">Add candidate</el-button>
      </div>
    </div>

    <el-card class="rec-filters" shadow="never">
      <el-form :inline="true" @submit.prevent="load">
        <el-form-item label="Search"><el-input v-model="filterKeyword" placeholder="Name keyword" clearable /></el-form-item>
        <el-form-item label="Skills"><el-input v-model="filterSkills" placeholder="GMP, QA" clearable /></el-form-item>
        <el-form-item label="Education"><el-input v-model="filterEdu" placeholder="e.g. Bachelor" clearable /></el-form-item>
        <el-form-item label="Exp min"><el-input-number v-model="filterMinExp" :min="0" :max="60" controls-position="right" /></el-form-item>
        <el-form-item label="Exp max"><el-input-number v-model="filterMaxExp" :min="0" :max="60" controls-position="right" /></el-form-item>
        <el-form-item label="Created from"><el-input v-model="filterCreatedFrom" placeholder="RFC3339" /></el-form-item>
        <el-form-item label="Created to"><el-input v-model="filterCreatedTo" placeholder="RFC3339" /></el-form-item>
        <el-form-item label="Updated from"><el-input v-model="filterUpdatedFrom" placeholder="RFC3339" /></el-form-item>
        <el-form-item label="Updated to"><el-input v-model="filterUpdatedTo" placeholder="RFC3339" /></el-form-item>
        <el-form-item><el-button type="primary" @click="applyFilters">Filter</el-button></el-form-item>
      </el-form>
    </el-card>

    <el-card class="rec-card" shadow="never">
      <el-table v-loading="loading" :data="rows" stripe empty-text="No candidates to show">
        <el-table-column prop="name" label="Name" min-width="160" />
        <el-table-column prop="phoneMasked" label="Phone" width="130" />
        <el-table-column prop="idNumberMasked" label="ID" width="130" />
        <el-table-column label="Experience" width="110"><template #default="{ row }">{{ row.experienceYears != null ? `${row.experienceYears} yrs` : '—' }}</template></el-table-column>
        <el-table-column prop="educationLevel" label="Education" width="120" />
        <el-table-column label="Skills" min-width="160"><template #default="{ row }">{{ row.skills?.length ? row.skills.join(', ') : '—' }}</template></el-table-column>
        <el-table-column label="Tags" width="140"><template #default="{ row }"><el-tag v-for="t in row.tags" :key="t" size="small" class="tag-pill">{{ t }}</el-tag><span v-if="!row.tags?.length">—</span></template></el-table-column>
        <el-table-column prop="createdAt" label="Created" width="180"><template #default="{ row }">{{ new Date(row.createdAt).toLocaleString() }}</template></el-table-column>
        <el-table-column label="Actions" width="104" fixed="right" align="center">
          <template #default="{ row }">
            <div class="action-icons">
              <el-button link type="success" aria-label="Match" @click="openMatch(row)">M</el-button>
              <el-button link type="primary" :icon="EditPen" aria-label="Edit" @click="openEdit(row)" />
              <el-button link type="danger" :icon="Delete" aria-label="Remove" @click="confirmDelete(row)" />
            </div>
          </template>
        </el-table-column>
      </el-table>
      <div class="rec-pager"><el-pagination background layout="prev, pager, next, total" :total="total" :page-size="pageSize" :current-page="page" @current-change="onPageChange" /></div>
    </el-card>

    <el-card class="rec-card" shadow="never">
      <template #header><div class="card-head"><span>Duplicate candidates (auto-merged on create/import)</span><el-button link type="primary" @click="loadDuplicates">Refresh</el-button></div></template>
      <el-table v-loading="duplicateLoading" :data="duplicates" stripe empty-text="No duplicates found">
        <el-table-column prop="matchType" label="Type" width="120" />
        <el-table-column label="Candidate IDs" min-width="420" show-overflow-tooltip><template #default="{ row }">{{ row.candidateIds.join(', ') }}</template></el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="createVisible" title="New candidate" width="440px" destroy-on-close>
      <el-form label-position="top">
        <el-form-item label="Name" required><el-input v-model="formName" placeholder="Full name" /></el-form-item>
        <el-form-item label="Experience (years)"><el-input-number v-model="formExp" :min="0" :max="60" controls-position="right" /></el-form-item>
        <el-form-item label="Education level"><el-input v-model="formEdu" placeholder="e.g. Bachelor" /></el-form-item>
        <el-form-item label="Skills (comma-separated)"><el-input v-model="formSkills" placeholder="GMP, QA" /></el-form-item>
        <el-form-item label="Tags (comma-separated)"><el-input v-model="formTags" placeholder="priority, referral" /></el-form-item>
      </el-form>
      <template #footer><el-button @click="createVisible = false">Cancel</el-button><el-button type="primary" :loading="createSaving" @click="submitCreate">Create</el-button></template>
    </el-dialog>

    <el-dialog v-model="editVisible" title="Edit candidate" width="560px" destroy-on-close>
      <el-form label-position="top">
        <el-form-item label="Name" required><el-input v-model="editName" /></el-form-item>
        <el-form-item label="Phone"><el-input v-model="editPhone" placeholder="Requires recruitment.view_pii to reveal existing value" /></el-form-item>
        <el-form-item label="ID number"><el-input v-model="editIDNumber" /></el-form-item>
        <el-form-item label="Email"><el-input v-model="editEmail" /></el-form-item>
        <el-form-item label="Experience years"><el-input-number v-model="editExp" :min="0" :max="60" controls-position="right" /></el-form-item>
        <el-form-item label="Education level"><el-input v-model="editEdu" /></el-form-item>
        <el-form-item label="Skills (comma-separated)"><el-input v-model="editSkills" /></el-form-item>
        <el-form-item label="Tags (comma-separated)"><el-input v-model="editTags" /></el-form-item>
        <el-form-item label="Custom fields">
          <div class="kv-wrap">
            <div v-for="(row, idx) in editCustomFields" :key="idx" class="kv-row">
              <el-input v-model="row.key" placeholder="key" />
              <el-input v-model="row.value" placeholder="value" />
              <el-button link type="danger" @click="editCustomFields.splice(idx, 1)">Remove</el-button>
            </div>
            <el-button link type="primary" @click="editCustomFields.push({ key: '', value: '' })">+ Add field</el-button>
          </div>
        </el-form-item>
      </el-form>
      <template #footer><el-button @click="editVisible = false">Cancel</el-button><el-button type="primary" :loading="editSaving" @click="submitEdit">Save</el-button></template>
    </el-dialog>

    <el-dialog v-model="importVisible" title="Bulk import resumes" width="760px" destroy-on-close>
      <p class="muted">Upload resume files, review extracted candidate previews, then commit import.</p>
      <input type="file" multiple accept=".pdf,.txt,.doc,.docx" @change="onResumePick" />
      <div class="tool-actions" style="margin-top: 8px">
        <el-button :loading="importUploading" @click="buildImportPreview">Upload + preview extraction</el-button>
      </div>
      <el-alert v-if="importValidationErrors.length" title="Validation feedback" type="warning" :closable="false" style="margin-top: 10px">
        <template #default>
          <ul><li v-for="e in importValidationErrors" :key="e">{{ e }}</li></ul>
        </template>
      </el-alert>
      <el-table v-if="importPreviewRows.length" :data="importPreviewRows" size="small" style="margin-top: 10px">
        <el-table-column prop="key" label="Candidate" />
        <el-table-column prop="value" label="Extracted contact" />
      </el-table>
      <template #footer><el-button @click="importVisible = false">Cancel</el-button><el-button type="primary" :loading="importSaving" @click="commitImport">Commit import</el-button></template>
    </el-dialog>

    <el-dialog v-model="matchVisible" title="Match and recommendations" width="760px" destroy-on-close>
      <el-form label-position="top">
        <el-form-item label="Candidate ID"><el-input v-model="matchCandidateId" readonly /></el-form-item>
        <el-form-item label="Position"><el-select v-model="matchPositionId" placeholder="Select position" style="width: 100%"><el-option v-for="p in positions" :key="p.id" :label="p.title" :value="p.id" /></el-select></el-form-item>
      </el-form>
      <el-button type="primary" :loading="matching" @click="runMatch">Run match</el-button>
      <div v-if="matchResult" class="match-panel">
        <p><strong>Score:</strong> {{ matchResult.score }} / 100</p>
        <p><strong>Breakdown:</strong> skills {{ matchResult.breakdown.skills }}, experience {{ matchResult.breakdown.experience }}, education {{ matchResult.breakdown.education }}</p>
        <ul><li v-for="r in matchResult.reasons" :key="r">{{ r }}</li></ul>
        <p><strong>Similar candidates:</strong> {{ similarCandidates.map((c) => c.name).join(', ') || '—' }}</p>
        <p><strong>Similar positions:</strong> {{ similarPositions.map((p) => p.title).join(', ') || '—' }}</p>
      </div>
      <template #footer><el-button @click="matchVisible = false">Close</el-button></template>
    </el-dialog>
  </div>
</template>

<style scoped>
.rec-page { max-width: 1200px; }
.rec-toolbar { display: flex; align-items: center; justify-content: space-between; gap: 1rem; margin-bottom: 1rem; }
.tool-actions { display: flex; gap: 0.5rem; }
.rec-title { margin: 0; font-size: 1.25rem; font-weight: 650; letter-spacing: -0.02em; }
.rec-filters { border-radius: 14px; border: 1px solid var(--el-border-color-lighter); margin-bottom: 1rem; }
.rec-card { border-radius: 14px; border: 1px solid var(--el-border-color-lighter); margin-bottom: 1rem; }
.rec-pager { display: flex; justify-content: flex-end; margin-top: 1rem; }
.tag-pill { margin-right: 4px; }
.action-icons { display: inline-flex; align-items: center; gap: 2px; }
.card-head { display: flex; align-items: center; justify-content: space-between; }
.match-panel { margin-top: 1rem; border-top: 1px solid var(--el-border-color-lighter); padding-top: 0.75rem; }
.muted { color: var(--el-text-color-secondary); }
.kv-wrap { width: 100%; display: flex; flex-direction: column; gap: 8px; }
.kv-row { display: grid; grid-template-columns: 1fr 1fr auto; gap: 8px; }
</style>
