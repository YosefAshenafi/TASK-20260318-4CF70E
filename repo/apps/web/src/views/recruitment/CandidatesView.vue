<script setup lang="ts">
import { Delete, EditPen } from '@element-plus/icons-vue'
import { onMounted, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'

import { apiDelete, apiGet, apiPatch, apiPost } from '@/api/http'
import { DEV_INSTITUTION_ID } from '@/config/devSeed'

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

const loading = ref(false)
const rows = ref<CandidateRow[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)

const dialogVisible = ref(false)
const dialogSaving = ref(false)
const formName = ref('')
const formExp = ref<number | undefined>(undefined)
const formEdu = ref('')
const formSkills = ref('')

async function load() {
  loading.value = true
  try {
    const q = new URLSearchParams({
      page: String(page.value),
      pageSize: String(pageSize.value),
      sortBy: 'created_at',
      sortOrder: 'desc',
    })
    const data = await apiGet<ListResp>(`/api/v1/recruitment/candidates?${q}`)
    rows.value = data.items
    total.value = data.total
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : 'Failed to load candidates')
  } finally {
    loading.value = false
  }
}

function onPageChange(p: number) {
  page.value = p
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
    const skills = formSkills.value
      .split(',')
      .map((s) => s.trim())
      .filter(Boolean)
    await apiPost<CandidateRow>('/api/v1/recruitment/candidates', {
      name: formName.value.trim(),
      institutionId: DEV_INSTITUTION_ID,
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

onMounted(load)
</script>

<template>
  <div class="rec-page">
    <div class="rec-toolbar">
      <h2 class="rec-title">Candidates</h2>
      <el-button type="primary" round @click="openCreate">Add candidate</el-button>
    </div>

    <el-card class="rec-card" shadow="never">
      <el-table v-loading="loading" :data="rows" stripe empty-text="No candidates in scope">
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
.tag-pill {
  margin-right: 4px;
}

.action-icons {
  display: inline-flex;
  align-items: center;
  gap: 2px;
}
</style>
