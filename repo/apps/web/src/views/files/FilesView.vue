<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'

import { apiGet, apiPost, apiPutBytes } from '@/api/http'
import { DEV_CASE_ID } from '@/config/devSeed'
import { useAuthStore } from '@/stores/auth'

type FileRow = {
  id: string
  sha256: string
  sizeBytes: number
  mimeType?: string
  createdAt: string
}

type ListResp = {
  items: FileRow[]
  total: number
  page: number
  pageSize: number
}

type InitResp = {
  uploadId: string
  totalChunks: number
  expiresAt: string
}

type CompleteResp = {
  fileId: string
  sha256: string
  deduplicated: boolean
}

const auth = useAuthStore()
const canManage = () => auth.hasPermission('files.manage')

const loading = ref(false)
const rows = ref<FileRow[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)

const uploadBusy = ref(false)
const uploadProgress = ref('')
const selectedFile = ref<File | null>(null)

const linkFileId = ref('')
const linkRefId = ref(DEV_CASE_ID)

const chunkSizeBytes = 1024 * 1024

function guessMime(file: File): string {
  if (file.type && file.type !== 'application/octet-stream') {
    return file.type
  }
  const n = file.name.toLowerCase()
  if (n.endsWith('.pdf')) return 'application/pdf'
  if (n.endsWith('.txt')) return 'text/plain'
  if (n.endsWith('.png')) return 'image/png'
  if (n.endsWith('.jpg') || n.endsWith('.jpeg')) return 'image/jpeg'
  if (n.endsWith('.webp')) return 'image/webp'
  if (n.endsWith('.docx')) return 'application/vnd.openxmlformats-officedocument.wordprocessingml.document'
  if (n.endsWith('.doc')) return 'application/msword'
  return 'text/plain'
}

async function sha256Hex(buf: ArrayBuffer): Promise<string> {
  const hash = await crypto.subtle.digest('SHA-256', buf)
  return [...new Uint8Array(hash)].map((b) => b.toString(16).padStart(2, '0')).join('')
}

async function load() {
  loading.value = true
  try {
    const q = new URLSearchParams({
      page: String(page.value),
      pageSize: String(pageSize.value),
    })
    const data = await apiGet<ListResp>(`/api/v1/files?${q}`)
    rows.value = data.items
    total.value = data.total
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : 'Failed to load files')
  } finally {
    loading.value = false
  }
}

function onPageChange(p: number) {
  page.value = p
  load()
}

function onFilePick(e: Event) {
  const input = e.target as HTMLInputElement
  selectedFile.value = input.files?.[0] ?? null
}

async function runUpload() {
  const file = selectedFile.value
  if (!file) {
    ElMessage.warning('Choose a file first')
    return
  }
  const effectiveMime = guessMime(file)

  uploadBusy.value = true
  uploadProgress.value = 'Initializing…'
  try {
    const init = await apiPost<InitResp>('/api/v1/files/uploads/init', {
      fileName: file.name,
      size: file.size,
      mimeType: effectiveMime,
      chunkSize: chunkSizeBytes,
    })
    const buf = await file.arrayBuffer()
    const totalChunks = init.totalChunks
    for (let i = 0; i < totalChunks; i++) {
      const start = i * chunkSizeBytes
      const end = Math.min(start + chunkSizeBytes, buf.byteLength)
      const slice = buf.slice(start, end)
      uploadProgress.value = `Uploading part ${i + 1} of ${totalChunks}`
      await apiPutBytes(`/api/v1/files/uploads/${init.uploadId}/chunks/${i}`, slice)
    }
    const hash = await sha256Hex(buf)
    uploadProgress.value = 'Completing…'
    const done = await apiPost<CompleteResp>(`/api/v1/files/uploads/${init.uploadId}/complete`, {
      sha256: hash,
    })
    ElMessage.success(
      done.deduplicated ? `Deduplicated — file id ${done.fileId}` : `Uploaded — file id ${done.fileId}`,
    )
    selectedFile.value = null
    await load()
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : 'Upload failed')
  } finally {
    uploadBusy.value = false
    uploadProgress.value = ''
  }
}

async function linkToCase() {
  const fid = linkFileId.value.trim()
  const rid = linkRefId.value.trim()
  if (!fid || !rid) {
    ElMessage.warning('Enter the file and case identifiers')
    return
  }
  try {
    await apiPost(`/api/v1/files/${fid}/link`, { refType: 'case', refId: rid })
    ElMessage.success('Linked file to case')
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : 'Link failed')
  }
}

function apiBase(): string {
  return import.meta.env.VITE_API_BASE_URL?.replace(/\/$/, '') ?? ''
}

async function downloadFile(row: FileRow) {
  const token = sessionStorage.getItem('pharmaops_session_token')
  const res = await fetch(`${apiBase()}/api/v1/files/${row.id}/download`, {
    headers: token ? { Authorization: `Bearer ${token}` } : {},
  })
  if (!res.ok) {
    const t = await res.text()
    let msg = ''
    try {
      const j = JSON.parse(t) as { message?: string }
      if (j.message) msg = j.message
    } catch {
      /* ignore */
    }
    if (!msg) {
      if (res.status === 401 || res.status === 403) {
        msg = 'You do not have permission to download this file.'
      } else if (res.status === 404) {
        msg = 'This file is no longer available.'
      } else {
        msg = 'Download failed. Please try again.'
      }
    }
    ElMessage.error(msg)
    return
  }
  const blob = await res.blob()
  const ext =
    row.mimeType === 'application/pdf'
      ? '.pdf'
      : row.mimeType?.startsWith('image/')
        ? '.bin'
        : '.bin'
  const name = `file-${row.id.slice(0, 8)}${ext}`
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = name
  a.click()
  URL.revokeObjectURL(url)
}

onMounted(() => {
  load()
})
</script>

<template>
  <div class="page">
    <header class="page-head">
      <h1>Files</h1>
      <p class="muted">Stored files and uploads you can resume if the connection drops.</p>
    </header>

    <el-card v-if="canManage()" class="panel" shadow="never">
      <template #header>
        <span>Upload</span>
      </template>
      <div class="upload-row">
        <input type="file" accept=".pdf,.txt,.png,.jpg,.jpeg,.webp,.doc,.docx" @change="onFilePick" />
        <el-button type="primary" :loading="uploadBusy" :disabled="!selectedFile" @click="runUpload">
          Start upload
        </el-button>
        <span v-if="uploadProgress" class="muted">{{ uploadProgress }}</span>
      </div>
      <p class="muted small">Allowed types include PDF, images, plain text, and Word. Maximum size about 100 MB.</p>
    </el-card>

    <el-card v-if="canManage()" class="panel" shadow="never">
      <template #header>
        <span>Link to case</span>
      </template>
      <div class="link-grid">
        <el-input v-model="linkFileId" placeholder="File identifier" clearable />
        <el-input v-model="linkRefId" placeholder="Case identifier" clearable />
        <el-button type="primary" @click="linkToCase">Link</el-button>
      </div>
      <p class="muted small">Only cases you can access can be linked. A sample case identifier may appear when demo data is loaded.</p>
    </el-card>

    <el-card class="panel" shadow="never">
      <template #header>
        <span>Stored files</span>
      </template>
      <el-table v-loading="loading" :data="rows" stripe style="width: 100%">
        <el-table-column prop="id" label="Id" min-width="280" show-overflow-tooltip />
        <el-table-column prop="mimeType" label="MIME" width="160" />
        <el-table-column prop="sizeBytes" label="Size" width="120" />
        <el-table-column prop="createdAt" label="Created" width="220" />
        <el-table-column label="Actions" width="140" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="downloadFile(row)">Download</el-button>
          </template>
        </el-table-column>
      </el-table>
      <div class="pager">
        <el-pagination
          background
          layout="prev, pager, next"
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
.page {
  max-width: 1100px;
}
.page-head h1 {
  margin: 0 0 0.25rem;
  font-size: 1.5rem;
}
.muted {
  color: var(--el-text-color-secondary);
  font-size: 0.9rem;
}
.small {
  font-size: 0.85rem;
  margin-top: 0.5rem;
}
.panel {
  margin-bottom: 1rem;
}
.upload-row {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.75rem;
}
.link-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
  align-items: center;
}
.link-grid .el-input {
  flex: 1;
  min-width: 200px;
}
.pager {
  margin-top: 1rem;
  display: flex;
  justify-content: flex-end;
}
</style>
