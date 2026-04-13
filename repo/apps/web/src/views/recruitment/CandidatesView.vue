<script setup lang="ts">
import { Delete, EditPen } from '@element-plus/icons-vue'
import { onMounted, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'

import { apiDelete, apiGet, apiPatch, apiPost } from '@/api/http'
import { useCreateScopeContext } from '@/composables/useDataScope'

type CandidateRow = {
  id: string
  name: string
  phoneMasked: string
  experienceYears?: number
  educationLevel?: string
  skills: string[]
  tags: string[]
  institutionId: string
  createdAt: string
}

type ListResp = {
  items: CandidateRow[]
  total: number
  page: number
  pageSize: number
}

type PositionOption = { id: string; title: string }
type DuplicateGroup = {
  matchType: string
  institutionId: string
  candidateIds: string[]
}
type MatchScore = {
  score: number
  breakdown: { skills: number; experience: number; education: number }
  reasons: string[]
}

const { requireContext } = useCreateScopeContext()

const loading = ref(false)
const rows = ref<CandidateRow[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)

const filterKeyword = ref('')
const filterSkills = ref('')
const filterEdu = ref('')
const filterMinExp = ref<number | undefined>(undefined)
const filterMaxExp = ref<number | undefined>(undefined)

const dialogVisible = ref(false)
const dialogSaving = ref(false)
const formName = ref('')
const formExp = ref<number | undefined>(undefined)
const formEdu = ref('')
const formSkills = ref('')

const importVisible = ref(false)
const importSaving = ref(false)
const importRowsText = ref('[\n  {"name":"Demo Candidate","skills":["GMP"],"educationLevel":"Bachelor","experienceYears":2}\n]')

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
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : 'Failed to load duplicates')
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
  dialogVisible.value = true
}

async function submitCreate() {
  if (!formName.value.trim()) {
    ElMessage.warning('Name is required.')
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
    const skills = formSkills.value
      .split(',')
      .map((s) => s.trim())
      .filter(Boolean)
    await apiPost<CandidateRow>('/api/v1/recruitment/candidates', {
      name: formName.value.trim(),
      institutionId: scope.institutionId,
      departmentId: scope.departmentId,
      teamId: scope.teamId,
      experienceYears: formExp.value,
      educationLevel: formEdu.value || undefined,
      skills,
    })
    ElMessage.success('Candidate created.')
    dialogVisible.value = false
    await load()
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : 'Create failed')
  } finally {
    dialogSaving.value = false
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

async function quickEdit(row: CandidateRow) {
  let value: string
  try {
    const res = await ElMessageBox.prompt('Update name', 'Edit', {
      inputValue: row.name,
      inputPattern: /.+/,
      inputErrorMessage: 'Name required',
    })
    value = res.value
  } catch {
    return
  }
  try {
    await apiPatch<CandidateRow>(`/api/v1/recruitment/candidates/${row.id}`, { name: value.trim() })
    ElMessage.success('Updated.')
    await load()
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : 'Update failed')
  }
}

function openImport() {
  importVisible.value = true
}

async function commitImport() {
  let rows: Array<Record<string, unknown>>
  try {
    const parsed = JSON.parse(importRowsText.value)
    if (!Array.isArray(parsed) || parsed.length === 0) {
      ElMessage.warning('Import payload must be a non-empty JSON array.')
      return
    }
    rows = parsed as Array<Record<string, unknown>>
  } catch {
    ElMessage.warning('Import payload must be valid JSON.')
    return
  }

  importSaving.value = true
  try {
    let scope
    try {
      scope = requireContext()
    } catch (err) {
      ElMessage.error(err instanceof Error ? err.message : 'No data scope')
      return
    }
    const batch = await apiPost<{ id: string }>('/api/v1/recruitment/candidates/imports', {
      institutionId: scope.institutionId,
      rows,
    })
    await apiPost(`/api/v1/recruitment/candidates/imports/${batch.id}/commit`, {})
    ElMessage.success('Import committed.')
    importVisible.value = false
    await load()
    await loadDuplicates()
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : 'Import failed')
  } finally {
    importSaving.value = false
  }
}

async function mergeGroup(group: DuplicateGroup) {
  if (!group.candidateIds || group.candidateIds.length < 2) return
  const ids = [...group.candidateIds]
  const baseCandidateId = ids[ids.length - 1]
  const sourceCandidateIds = ids.slice(0, -1)
  try {
    await ElMessageBox.confirm(
      `Merge ${sourceCandidateIds.length} duplicate record(s) into base ${baseCandidateId.slice(0, 8)}…?`,
      'Merge duplicates',
      { type: 'warning' },
    )
  } catch {
    return
  }
  try {
    await apiPost('/api/v1/recruitment/candidates/merge', {
      baseCandidateId,
      sourceCandidateIds,
      strategy: 'latest_wins_fill_missing',
    })
    ElMessage.success('Merge completed.')
    await load()
    await loadDuplicates()
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : 'Merge failed')
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
  if (!matchCandidateId.value || !matchPositionId.value) {
    ElMessage.warning('Select both candidate and position.')
    return
  }
  matching.value = true
  try {
    const res = await apiPost<MatchScore>('/api/v1/recruitment/match/candidate-to-position', {
      candidateId: matchCandidateId.value,
      positionId: matchPositionId.value,
    })
    matchResult.value = res
    const [sc, sp] = await Promise.all([
      apiGet<{ items: CandidateRow[] }>(
        `/api/v1/recruitment/recommendations/similar-candidates/${matchCandidateId.value}?limit=5`,
      ),
      apiGet<{ items: PositionOption[] }>(
        `/api/v1/recruitment/recommendations/similar-positions/${matchPositionId.value}?limit=5`,
      ),
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
        <el-form-item label="Search">
          <el-input v-model="filterKeyword" placeholder="Name keyword" clearable @clear="load" />
        </el-form-item>
        <el-form-item label="Skills">
          <el-input v-model="filterSkills" placeholder="GMP, QA" clearable @clear="load" />
        </el-form-item>
        <el-form-item label="Education">
          <el-input v-model="filterEdu" placeholder="e.g. Bachelor" clearable @clear="load" />
        </el-form-item>
        <el-form-item label="Exp min">
          <el-input-number v-model="filterMinExp" :min="0" :max="60" controls-position="right" placeholder="0" />
        </el-form-item>
        <el-form-item label="Exp max">
          <el-input-number v-model="filterMaxExp" :min="0" :max="60" controls-position="right" placeholder="60" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="applyFilters">Filter</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card class="rec-card" shadow="never">
      <el-table v-loading="loading" :data="rows" stripe empty-text="No candidates to show">
        <el-table-column prop="name" label="Name" min-width="160" />
        <el-table-column label="Experience" width="110">
          <template #default="{ row }">
            {{ row.experienceYears != null ? `${row.experienceYears} yrs` : '—' }}
          </template>
        </el-table-column>
        <el-table-column prop="educationLevel" label="Education" width="120" />
        <el-table-column label="Skills" min-width="160">
          <template #default="{ row }">
            {{ row.skills?.length ? row.skills.join(', ') : '—' }}
          </template>
        </el-table-column>
        <el-table-column label="Tags" width="120">
          <template #default="{ row }">
            <el-tag v-for="t in row.tags" :key="t" size="small" class="tag-pill">{{ t }}</el-tag>
            <span v-if="!row.tags?.length">—</span>
          </template>
        </el-table-column>
        <el-table-column prop="createdAt" label="Created" width="180">
          <template #default="{ row }">
            {{ new Date(row.createdAt).toLocaleString() }}
          </template>
        </el-table-column>
        <el-table-column label="Actions" width="104" fixed="right" align="center">
          <template #default="{ row }">
            <div class="action-icons">
              <el-tooltip content="Match" placement="top">
                <el-button link type="success" aria-label="Match" @click="openMatch(row)">M</el-button>
              </el-tooltip>
              <el-tooltip content="Rename" placement="top">
                <el-button
                  link
                  type="primary"
                  :icon="EditPen"
                  aria-label="Rename"
                  @click="quickEdit(row)"
                />
              </el-tooltip>
              <el-tooltip content="Remove" placement="top">
                <el-button
                  link
                  type="danger"
                  :icon="Delete"
                  aria-label="Remove"
                  @click="confirmDelete(row)"
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

    <el-card class="rec-card" shadow="never">
      <template #header>
        <div class="card-head">
          <span>Duplicate candidates</span>
          <el-button link type="primary" @click="loadDuplicates">Refresh</el-button>
        </div>
      </template>
      <el-table v-loading="duplicateLoading" :data="duplicates" stripe empty-text="No duplicates found">
        <el-table-column prop="matchType" label="Type" width="120" />
        <el-table-column label="Candidate IDs" min-width="420" show-overflow-tooltip>
          <template #default="{ row }">
            {{ row.candidateIds.join(', ') }}
          </template>
        </el-table-column>
        <el-table-column label="" width="120" align="center">
          <template #default="{ row }">
            <el-button size="small" type="warning" @click="mergeGroup(row)">Merge</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="dialogVisible" title="New candidate" width="420px" destroy-on-close>
      <el-form label-position="top">
        <el-form-item label="Name" required>
          <el-input v-model="formName" placeholder="Full name" />
        </el-form-item>
        <el-form-item label="Experience (years)">
          <el-input-number v-model="formExp" :min="0" :max="60" controls-position="right" />
        </el-form-item>
        <el-form-item label="Education level">
          <el-input v-model="formEdu" placeholder="e.g. Bachelor" />
        </el-form-item>
        <el-form-item label="Skills (comma-separated)">
          <el-input v-model="formSkills" placeholder="GMP, QA" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">Cancel</el-button>
        <el-button type="primary" :loading="dialogSaving" @click="submitCreate">Create</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="importVisible" title="Bulk import candidates" width="720px" destroy-on-close>
      <p class="muted">Paste JSON array rows (name, skills, educationLevel, experienceYears, etc.).</p>
      <el-input v-model="importRowsText" type="textarea" :rows="14" />
      <template #footer>
        <el-button @click="importVisible = false">Cancel</el-button>
        <el-button type="primary" :loading="importSaving" @click="commitImport">Create & commit</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="matchVisible" title="Match and recommendations" width="760px" destroy-on-close>
      <el-form label-position="top">
        <el-form-item label="Candidate ID">
          <el-input v-model="matchCandidateId" readonly />
        </el-form-item>
        <el-form-item label="Position">
          <el-select v-model="matchPositionId" placeholder="Select position" style="width: 100%">
            <el-option v-for="p in positions" :key="p.id" :label="p.title" :value="p.id" />
          </el-select>
        </el-form-item>
      </el-form>
      <el-button type="primary" :loading="matching" @click="runMatch">Run match</el-button>

      <div v-if="matchResult" class="match-panel">
        <p><strong>Score:</strong> {{ matchResult.score }} / 100</p>
        <p>
          <strong>Breakdown:</strong> skills {{ matchResult.breakdown.skills }}, experience
          {{ matchResult.breakdown.experience }}, education {{ matchResult.breakdown.education }}
        </p>
        <ul>
          <li v-for="r in matchResult.reasons" :key="r">{{ r }}</li>
        </ul>
        <p><strong>Similar candidates:</strong> {{ similarCandidates.map((c) => c.name).join(', ') || '—' }}</p>
        <p><strong>Similar positions:</strong> {{ similarPositions.map((p) => p.title).join(', ') || '—' }}</p>
      </div>
      <template #footer>
        <el-button @click="matchVisible = false">Close</el-button>
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
.tool-actions {
  display: flex;
  gap: 0.5rem;
}
.rec-title {
  margin: 0;
  font-size: 1.25rem;
  font-weight: 650;
  letter-spacing: -0.02em;
}
.rec-filters {
  border-radius: 14px;
  border: 1px solid var(--el-border-color-lighter);
  margin-bottom: 1rem;
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
.tag-pill {
  margin-right: 4px;
}

.action-icons {
  display: inline-flex;
  align-items: center;
  gap: 2px;
}
.card-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.match-panel {
  margin-top: 1rem;
  border-top: 1px solid var(--el-border-color-lighter);
  padding-top: 0.75rem;
}
.muted {
  color: var(--el-text-color-secondary);
}
</style>
