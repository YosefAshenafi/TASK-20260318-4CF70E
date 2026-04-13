<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'

import { apiGet, apiPatch, apiPost } from '@/api/http'
import { DEV_DEPARTMENT_ID, DEV_INSTITUTION_ID, DEV_TEAM_ID } from '@/config/devSeed'

type UserRow = {
  id: string
  username: string
  displayName: string
  isActive: boolean
  roles: string[]
  createdAt: string
}

type UserDetail = UserRow & {
  roleIds: string[]
  scopeIds: string[]
}

type RoleRow = {
  id: string
  slug: string
  name: string
  description?: string
  createdAt: string
  updatedAt: string
}

type PermRow = {
  id: string
  code: string
  description?: string
  createdAt: string
}

type ScopeRow = {
  id: string
  scopeKey: string
  institutionId: string
  departmentId?: string
  teamId?: string
  createdAt: string
}

/** docs/design.md §1 + §8.1.1 — canonical product personas (seeded roles). */
const PRIMARY_ROLE_SLUGS = new Set([
  'business_specialist',
  'compliance_administrator',
  'recruitment_specialist',
  'system_admin',
])

function isPrimaryRole(slug: string): boolean {
  return PRIMARY_ROLE_SLUGS.has(slug)
}

const tab = ref<'users' | 'roles' | 'scopes'>('users')

const users = ref<UserRow[]>([])
const roles = ref<RoleRow[]>([])
const permissions = ref<PermRow[]>([])
const scopes = ref<ScopeRow[]>([])

const loadingUsers = ref(false)
const loadingRoles = ref(false)
const loadingScopes = ref(false)

const permDialogVisible = ref(false)
const permSaving = ref(false)
const editingRole = ref<RoleRow | null>(null)
const selectedPermIds = ref<string[]>([])

/** Add user */
const addUserVisible = ref(false)
const addUserSaving = ref(false)
const newUsername = ref('')
const newPassword = ref('')
const newDisplayName = ref('')
const newIsActive = ref(true)
const newRoleIds = ref<string[]>([])

/** Edit user */
const editUserVisible = ref(false)
const editUserSaving = ref(false)
const editingUserId = ref<string | null>(null)
const editDisplayName = ref('')
const editIsActive = ref(true)
const editPassword = ref('')
const editRoleIds = ref<string[]>([])

/** User data scopes */
const scopeDialogVisible = ref(false)
const scopeSaving = ref(false)
const scopeUserId = ref<string | null>(null)
const scopeUserLabel = ref('')
const selectedScopeIds = ref<string[]>([])

/** Create role */
const addRoleVisible = ref(false)
const addRoleSaving = ref(false)
const newRoleSlug = ref('')
const newRoleName = ref('')
const newRoleDescription = ref('')

/** Create data scope */
const addScopeVisible = ref(false)
const addScopeSaving = ref(false)
const newScopeKey = ref('')
const newScopeInstitutionId = ref(DEV_INSTITUTION_ID)
const newScopeDepartmentId = ref('')
const newScopeTeamId = ref('')

async function loadUsers() {
  loadingUsers.value = true
  try {
    const data = await apiGet<{ items: UserRow[] }>('/api/v1/users')
    users.value = data.items ?? []
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : 'Failed to load users')
  } finally {
    loadingUsers.value = false
  }
}

async function loadRoles() {
  loadingRoles.value = true
  try {
    const [r, p] = await Promise.all([
      apiGet<{ items: RoleRow[] }>('/api/v1/roles'),
      apiGet<{ items: PermRow[] }>('/api/v1/permissions'),
    ])
    roles.value = r.items ?? []
    permissions.value = p.items ?? []
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : 'Failed to load roles')
  } finally {
    loadingRoles.value = false
  }
}

async function loadScopes() {
  loadingScopes.value = true
  try {
    const data = await apiGet<{ items: ScopeRow[] }>('/api/v1/scopes')
    scopes.value = data.items ?? []
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : 'Failed to load scopes')
  } finally {
    loadingScopes.value = false
  }
}

function openAddUser() {
  newUsername.value = ''
  newPassword.value = ''
  newDisplayName.value = ''
  newIsActive.value = true
  newRoleIds.value = []
  addUserVisible.value = true
}

async function saveAddUser() {
  const u = newUsername.value.trim()
  const p = newPassword.value
  const d = newDisplayName.value.trim()
  if (!u || !d) {
    ElMessage.warning('Username and display name are required')
    return
  }
  if (p.length < 8) {
    ElMessage.warning('Password must be at least 8 characters')
    return
  }
  addUserSaving.value = true
  try {
    await apiPost<UserDetail>('/api/v1/users', {
      username: u,
      password: p,
      displayName: d,
      isActive: newIsActive.value,
      roleIds: newRoleIds.value,
    })
    ElMessage.success('User created.')
    addUserVisible.value = false
    await loadUsers()
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : 'Create failed')
  } finally {
    addUserSaving.value = false
  }
}

async function openEditUser(row: UserRow) {
  editingUserId.value = row.id
  editPassword.value = ''
  editUserVisible.value = true
  try {
    const detail = await apiGet<UserDetail>(`/api/v1/users/${row.id}`)
    editDisplayName.value = detail.displayName
    editIsActive.value = detail.isActive
    editRoleIds.value = [...(detail.roleIds ?? [])]
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : 'Failed to load user')
    editUserVisible.value = false
  }
}

async function saveEditUser() {
  const id = editingUserId.value
  if (!id) return
  const d = editDisplayName.value.trim()
  if (!d) {
    ElMessage.warning('Display name is required')
    return
  }
  if (editPassword.value && editPassword.value.length < 8) {
    ElMessage.warning('New password must be at least 8 characters')
    return
  }
  const body: Record<string, unknown> = {
    displayName: d,
    isActive: editIsActive.value,
    roleIds: editRoleIds.value,
  }
  if (editPassword.value.trim()) {
    body.password = editPassword.value
  }
  editUserSaving.value = true
  try {
    await apiPatch<UserDetail>(`/api/v1/users/${id}`, body)
    ElMessage.success('User updated.')
    editUserVisible.value = false
    await loadUsers()
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : 'Update failed')
  } finally {
    editUserSaving.value = false
  }
}

async function openScopeDialog(row: UserRow) {
  scopeUserId.value = row.id
  scopeUserLabel.value = row.username
  scopeDialogVisible.value = true
  try {
    const detail = await apiGet<UserDetail>(`/api/v1/users/${row.id}`)
    selectedScopeIds.value = [...(detail.scopeIds ?? [])]
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : 'Failed to load user')
    selectedScopeIds.value = []
  }
}

async function saveUserScopes() {
  const id = scopeUserId.value
  if (!id) return
  scopeSaving.value = true
  try {
    await apiPost(`/api/v1/users/${id}/scopes`, { scopeIds: selectedScopeIds.value })
    ElMessage.success('Data scopes updated.')
    scopeDialogVisible.value = false
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : 'Save failed')
  } finally {
    scopeSaving.value = false
  }
}

async function openPermDialog(row: RoleRow) {
  editingRole.value = row
  permDialogVisible.value = true
  try {
    const detail = await apiGet<{ permissionIds: string[] }>(`/api/v1/roles/${row.id}`)
    selectedPermIds.value = detail.permissionIds ?? []
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : 'Failed to load role')
    selectedPermIds.value = []
  }
}

async function savePermissions() {
  if (!editingRole.value) {
    return
  }
  permSaving.value = true
  try {
    await apiPost(`/api/v1/roles/${editingRole.value.id}/permissions`, {
      permissionIds: selectedPermIds.value,
    })
    ElMessage.success('Permissions updated.')
    permDialogVisible.value = false
    await loadRoles()
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : 'Save failed')
  } finally {
    permSaving.value = false
  }
}

function openAddRole() {
  newRoleSlug.value = ''
  newRoleName.value = ''
  newRoleDescription.value = ''
  addRoleVisible.value = true
}

async function saveAddRole() {
  const slug = newRoleSlug.value.trim().toLowerCase()
  const name = newRoleName.value.trim()
  if (!slug || !name) {
    ElMessage.warning('Slug and name are required')
    return
  }
  addRoleSaving.value = true
  try {
    await apiPost<RoleRow>('/api/v1/roles', {
      slug,
      name,
      description: newRoleDescription.value.trim() || undefined,
    })
    ElMessage.success('Role created.')
    addRoleVisible.value = false
    await loadRoles()
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : 'Create failed')
  } finally {
    addRoleSaving.value = false
  }
}

function openAddScope() {
  newScopeKey.value = ''
  newScopeInstitutionId.value = DEV_INSTITUTION_ID
  newScopeDepartmentId.value = ''
  newScopeTeamId.value = ''
  addScopeVisible.value = true
}

async function saveAddScope() {
  const key = newScopeKey.value.trim()
  const inst = newScopeInstitutionId.value.trim()
  if (!key || inst.length !== 36) {
    ElMessage.warning('Scope key and institution id (UUID) are required')
    return
  }
  const body: Record<string, unknown> = {
    scopeKey: key,
    institutionId: inst,
  }
  const d = newScopeDepartmentId.value.trim()
  const t = newScopeTeamId.value.trim()
  if (d) body.departmentId = d
  if (t) body.teamId = t
  addScopeSaving.value = true
  try {
    await apiPost<ScopeRow>('/api/v1/scopes', body)
    ElMessage.success('Data scope created.')
    addScopeVisible.value = false
    await loadScopes()
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : 'Create failed')
  } finally {
    addScopeSaving.value = false
  }
}

async function editRoleName(row: RoleRow) {
  let name: string
  let desc: string
  try {
    const r1 = await ElMessageBox.prompt('Display name', 'Edit role', { inputValue: row.name })
    name = r1.value
    const r2 = await ElMessageBox.prompt('Description (optional)', 'Edit role', {
      inputValue: row.description ?? '',
      inputType: 'textarea',
    })
    desc = r2.value
  } catch {
    return
  }
  try {
    await apiPatch(`/api/v1/roles/${row.id}`, {
      name: name.trim(),
      description: desc.trim() || null,
    })
    ElMessage.success('Role updated.')
    await loadRoles()
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : 'Update failed')
  }
}

onMounted(async () => {
  await loadUsers()
  await loadRoles()
  await loadScopes()
})
</script>

<template>
  <div class="rec-page">
    <div class="rec-toolbar">
      <h2 class="rec-title">Roles &amp; access</h2>
    </div>

    <el-tabs v-model="tab" class="rbac-tabs">
      <el-tab-pane label="Users" name="users">
        <el-card class="rec-card" shadow="never">
          <div class="users-toolbar">
            <el-button type="primary" @click="openAddUser">Add user</el-button>
            <span class="muted small">Create accounts and assign roles (design §10 / api-spec AccessController).</span>
          </div>
          <el-table v-loading="loadingUsers" :data="users" stripe empty-text="No users">
            <el-table-column prop="username" label="Username" width="140" />
            <el-table-column prop="displayName" label="Display name" width="180" />
            <el-table-column label="Roles" min-width="200">
              <template #default="{ row }">
                <el-tag v-for="r in row.roles" :key="r" size="small" style="margin: 2px">{{ r }}</el-tag>
                <span v-if="!row.roles?.length" class="muted">—</span>
              </template>
            </el-table-column>
            <el-table-column prop="isActive" label="Active" width="90">
              <template #default="{ row }">
                <el-tag :type="row.isActive ? 'success' : 'info'" size="small">
                  {{ row.isActive ? 'yes' : 'no' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="Actions" width="200" fixed="right">
              <template #default="{ row }">
                <el-button link type="primary" size="small" @click="openEditUser(row)">Edit</el-button>
                <el-button link type="primary" size="small" @click="openScopeDialog(row)">Data scopes</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-tab-pane>

      <el-tab-pane label="Roles" name="roles">
        <el-card class="rec-card" shadow="never">
          <div class="users-toolbar">
            <el-button type="primary" @click="openAddRole">Add role</el-button>
            <span class="muted small">
              Lists all rows in <code class="mono">roles</code> (design 8.1). Four primary personas —
              Business Specialist, Compliance Administrator, Recruitment Specialist, System Administrator — are
              seeded and shown with a Primary badge. Extra roles extend the model as needed.
            </span>
          </div>
          <el-table v-loading="loadingRoles" :data="roles" stripe empty-text="No roles">
            <el-table-column label="Primary" width="96">
              <template #default="{ row }">
                <el-tag v-if="isPrimaryRole(row.slug)" type="success" size="small" effect="plain">
                  Primary
                </el-tag>
                <span v-else class="muted">—</span>
              </template>
            </el-table-column>
            <el-table-column prop="slug" label="Slug" width="200" show-overflow-tooltip />
            <el-table-column prop="name" label="Name" min-width="160" />
            <el-table-column prop="description" label="Description" min-width="200" show-overflow-tooltip />
            <el-table-column label="Actions" width="200" fixed="right">
              <template #default="{ row }">
                <el-button link type="primary" size="small" @click="editRoleName(row)">Edit</el-button>
                <el-button link type="primary" size="small" @click="openPermDialog(row)">Permissions</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-tab-pane>

      <el-tab-pane label="Data scopes" name="scopes">
        <el-card class="rec-card" shadow="never">
          <div class="users-toolbar">
            <el-button type="primary" @click="openAddScope">Add data scope</el-button>
            <span class="muted small">
              Institution-wide, or narrow by department / team (design 10.2). Dev seed ids —
              <code class="mono">dept</code> {{ DEV_DEPARTMENT_ID.slice(0, 8) }}…,
              <code class="mono">team</code> {{ DEV_TEAM_ID.slice(0, 8) }}…
            </span>
          </div>
          <el-table v-loading="loadingScopes" :data="scopes" stripe empty-text="No scopes">
            <el-table-column prop="scopeKey" label="Key" min-width="180" />
            <el-table-column prop="institutionId" label="Institution" width="200">
              <template #default="{ row }">
                <span class="mono">{{ row.institutionId }}</span>
              </template>
            </el-table-column>
            <el-table-column prop="departmentId" label="Dept" width="120" />
            <el-table-column prop="teamId" label="Team" width="120" />
          </el-table>
        </el-card>
      </el-tab-pane>
    </el-tabs>

    <el-dialog v-model="addUserVisible" title="Add user" width="480px" destroy-on-close>
      <el-form label-position="top">
        <el-form-item label="Username" required>
          <el-input v-model="newUsername" autocomplete="off" />
        </el-form-item>
        <el-form-item label="Password (min 8 characters)" required>
          <el-input v-model="newPassword" type="password" show-password autocomplete="new-password" />
        </el-form-item>
        <el-form-item label="Display name" required>
          <el-input v-model="newDisplayName" autocomplete="off" />
        </el-form-item>
        <el-form-item label="Active">
          <el-switch v-model="newIsActive" />
        </el-form-item>
        <el-form-item label="Roles">
          <el-select v-model="newRoleIds" multiple filterable placeholder="Roles" style="width: 100%">
            <el-option v-for="r in roles" :key="r.id" :label="`${r.slug} — ${r.name}`" :value="r.id" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="addUserVisible = false">Cancel</el-button>
        <el-button type="primary" :loading="addUserSaving" @click="saveAddUser">Create</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="editUserVisible" title="Edit user" width="480px" destroy-on-close>
      <el-form label-position="top">
        <el-form-item label="Display name" required>
          <el-input v-model="editDisplayName" autocomplete="off" />
        </el-form-item>
        <el-form-item label="New password (leave blank to keep)">
          <el-input v-model="editPassword" type="password" show-password autocomplete="new-password" />
        </el-form-item>
        <el-form-item label="Active">
          <el-switch v-model="editIsActive" />
        </el-form-item>
        <el-form-item label="Roles">
          <el-select v-model="editRoleIds" multiple filterable placeholder="Roles" style="width: 100%">
            <el-option v-for="r in roles" :key="r.id" :label="`${r.slug} — ${r.name}`" :value="r.id" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="editUserVisible = false">Cancel</el-button>
        <el-button type="primary" :loading="editUserSaving" @click="saveEditUser">Save</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="scopeDialogVisible" :title="`Data scopes — ${scopeUserLabel}`" width="520px" destroy-on-close>
      <p class="muted small">Assigns institution / department / team visibility per design §10.2.</p>
      <el-select
        v-model="selectedScopeIds"
        multiple
        filterable
        placeholder="Data scopes"
        style="width: 100%"
        collapse-tags
        collapse-tags-tooltip
      >
        <el-option
          v-for="s in scopes"
          :key="s.id"
          :label="`${s.scopeKey} (${s.institutionId.slice(0, 8)}…)`"
          :value="s.id"
        />
      </el-select>
      <template #footer>
        <el-button @click="scopeDialogVisible = false">Cancel</el-button>
        <el-button type="primary" :loading="scopeSaving" @click="saveUserScopes">Save</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="addRoleVisible" title="Add role" width="480px" destroy-on-close>
      <el-form label-position="top">
        <el-form-item label="Slug" required>
          <el-input v-model="newRoleSlug" placeholder="e.g. compliance_specialist" autocomplete="off" />
        </el-form-item>
        <el-form-item label="Display name" required>
          <el-input v-model="newRoleName" autocomplete="off" />
        </el-form-item>
        <el-form-item label="Description">
          <el-input v-model="newRoleDescription" type="textarea" :rows="2" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="addRoleVisible = false">Cancel</el-button>
        <el-button type="primary" :loading="addRoleSaving" @click="saveAddRole">Create</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="addScopeVisible" title="Add data scope" width="520px" destroy-on-close>
      <el-form label-position="top">
        <el-form-item label="Scope key" required>
          <el-input v-model="newScopeKey" placeholder="e.g. inst:acme-dept-pharmacy" autocomplete="off" />
        </el-form-item>
        <el-form-item label="Institution id" required>
          <el-input v-model="newScopeInstitutionId" class="mono" />
        </el-form-item>
        <el-form-item label="Department id (optional)">
          <el-input v-model="newScopeDepartmentId" :placeholder="DEV_DEPARTMENT_ID" class="mono" clearable />
        </el-form-item>
        <el-form-item label="Team id (optional; must match department)">
          <el-input v-model="newScopeTeamId" :placeholder="DEV_TEAM_ID" class="mono" clearable />
        </el-form-item>
      </el-form>
      <p class="muted small">Leave department and team empty for institution-wide scope.</p>
      <template #footer>
        <el-button @click="addScopeVisible = false">Cancel</el-button>
        <el-button type="primary" :loading="addScopeSaving" @click="saveAddScope">Create</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="permDialogVisible" :title="`Permissions — ${editingRole?.slug ?? ''}`" width="520px" destroy-on-close>
      <el-select
        v-model="selectedPermIds"
        multiple
        filterable
        placeholder="Select permissions"
        style="width: 100%"
        collapse-tags
        collapse-tags-tooltip
      >
        <el-option v-for="p in permissions" :key="p.id" :label="p.code" :value="p.id" />
      </el-select>
      <template #footer>
        <el-button @click="permDialogVisible = false">Cancel</el-button>
        <el-button type="primary" :loading="permSaving" @click="savePermissions">Save</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.rec-page {
  max-width: 1200px;
}
.rec-toolbar {
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
.rbac-tabs :deep(.el-tabs__header) {
  margin-bottom: 1rem;
}
.users-toolbar {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.75rem;
  margin-bottom: 1rem;
}
.muted {
  color: var(--el-text-color-secondary);
  font-size: 12px;
}
.small {
  font-size: 12px;
  line-height: 1.4;
}
.mono {
  font-family: ui-monospace, monospace;
  font-size: 12px;
}
</style>
