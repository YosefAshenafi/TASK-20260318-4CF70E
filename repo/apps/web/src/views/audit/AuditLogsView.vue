<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'

import { apiGet, apiPost } from '@/api/http'
import { useAuthStore } from '@/stores/auth'
import { humanizeTechnicalLabel } from '@/utils/display'

type AuditRow = {
  id: string
  module: string
  operation: string
  operatorId: string
  requestSource?: string
  targetType: string
  targetId: string
  before?: Record<string, unknown>
  after?: Record<string, unknown>
  createdAt: string
}

type ListResp = {
  items: AuditRow[]
  total: number
  page: number
  pageSize: number
}

const auth = useAuthStore()
const canExport = () => auth.hasPermission('audit.view')

const loading = ref(false)
const rows = ref<AuditRow[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const moduleFilter = ref('')
const targetTypeFilter = ref('')

async function load() {
  loading.value = true
  try {
    const q = new URLSearchParams({
      page: String(page.value),
      pageSize: String(pageSize.value),
      sortBy: 'created_at',
      sortOrder: 'desc',
    })
    if (moduleFilter.value.trim()) {
      q.set('module', moduleFilter.value.trim())
    }
    if (targetTypeFilter.value.trim()) {
      q.set('targetType', targetTypeFilter.value.trim())
    }
    const data = await apiGet<ListResp>(`/api/v1/audit/logs?${q}`)
    rows.value = data.items
    total.value = data.total
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : 'Failed to load audit logs')
  } finally {
    loading.value = false
  }
}

function onPageChange(p: number) {
  page.value = p
  load()
}

async function requestExport() {
  try {
    await ElMessageBox.confirm(
      'Start an export using the filters above. Large exports may take a while and can finish in the background.',
      'Export audit log',
      { type: 'info' },
    )
  } catch {
    return
  }
  try {
    const res = await apiPost<{ id: string; status: string; createdAt: string }>('/api/v1/audit/logs/export', {
      module: moduleFilter.value.trim() || undefined,
      targetType: targetTypeFilter.value.trim() || undefined,
    })
    ElMessage.success(`Export requested. Reference: ${res.id.slice(0, 8)}…`)
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : 'Export failed')
  }
}

/**
 * Human-readable audit payload (field-level before/after when present).
 * Surfaces free-text fields like `note` as plain content, not JSON quotes.
 */
function humanizeKey(key: string): string {
  const s = key.replace(/_/g, ' ')
  return s.replace(/\b\w/g, (c) => c.toUpperCase())
}

function formatValue(v: unknown): string {
  if (v === null || v === undefined) return '—'
  if (typeof v === 'string') return v
  if (typeof v === 'number' || typeof v === 'boolean') return String(v)
  try {
    return JSON.stringify(v)
  } catch {
    return String(v)
  }
}

function formatPayloadForDisplay(obj: Record<string, unknown> | undefined): string {
  if (!obj || Object.keys(obj).length === 0) {
    return '—'
  }
  const rest: Record<string, unknown> = { ...obj }
  const lines: string[] = []

  if (typeof rest.note === 'string') {
    const note = rest.note.trim()
    if (note.length > 0) {
      lines.push(note)
    }
    delete rest.note
  }

  const otherKeys = Object.keys(rest).sort()
  for (let i = 0; i < otherKeys.length; i++) {
    const k = otherKeys[i]
    if (i === 0 && lines.length > 0) {
      lines.push('')
    }
    lines.push(`${humanizeKey(k)}: ${formatValue(rest[k])}`)
  }

  return lines.length > 0 ? lines.join('\n') : '—'
}

onMounted(load)
</script>

<template>
  <div class="rec-page">
    <div class="rec-toolbar">
      <h2 class="rec-title">Audit logs</h2>
      <el-button v-if="canExport()" round type="primary" @click="requestExport">Request export</el-button>
    </div>

    <el-card class="rec-card" shadow="never">
      <div class="audit-filters">
        <el-input v-model="moduleFilter" clearable placeholder="Area" style="width: 140px" @clear="load" />
        <el-input v-model="targetTypeFilter" clearable placeholder="What changed" style="width: 160px" @clear="load" />
        <el-button @click="load">Apply filters</el-button>
      </div>
      <el-table v-loading="loading" :data="rows" stripe empty-text="No audit entries">
        <el-table-column prop="createdAt" label="Time" width="180">
          <template #default="{ row }">
            {{ new Date(row.createdAt).toLocaleString() }}
          </template>
        </el-table-column>
        <el-table-column prop="module" label="Area" width="130">
          <template #default="{ row }">
            {{ humanizeTechnicalLabel(row.module) }}
          </template>
        </el-table-column>
        <el-table-column prop="operation" label="Operation" width="180" show-overflow-tooltip>
          <template #default="{ row }">
            {{ humanizeTechnicalLabel(row.operation) }}
          </template>
        </el-table-column>
        <el-table-column prop="operatorId" label="Operator" width="200">
          <template #default="{ row }">
            <span class="mono">{{ row.operatorId }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="targetType" label="Target" width="140">
          <template #default="{ row }">
            {{ humanizeTechnicalLabel(row.targetType) }}
          </template>
        </el-table-column>
        <el-table-column prop="targetId" label="Record id" min-width="200">
          <template #default="{ row }">
            <span class="mono">{{ row.targetId }}</span>
          </template>
        </el-table-column>
        <el-table-column label="Before" min-width="200">
          <template #default="{ row }">
            <div class="audit-payload">{{ formatPayloadForDisplay(row.before) }}</div>
          </template>
        </el-table-column>
        <el-table-column label="After" min-width="200">
          <template #default="{ row }">
            <div class="audit-payload">{{ formatPayloadForDisplay(row.after) }}</div>
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
.audit-filters {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  margin-bottom: 1rem;
  align-items: center;
}
.mono {
  font-family: ui-monospace, monospace;
  font-size: 12px;
}
.audit-payload {
  white-space: pre-wrap;
  word-break: break-word;
  font-size: 12px;
  line-height: 1.45;
  max-height: 160px;
  overflow-y: auto;
  color: var(--el-text-color-regular);
}
</style>
