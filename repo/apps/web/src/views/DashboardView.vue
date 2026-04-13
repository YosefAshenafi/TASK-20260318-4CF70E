<script setup lang="ts">
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
</script>

<template>
  <el-card shadow="never">
    <h2>Welcome</h2>
    <p v-if="auth.username">Signed in as <strong>{{ auth.username }}</strong>.</p>
    <p class="muted">Use the sidebar to open modules. Menus follow RBAC once the API is connected.</p>
  </el-card>
</template>

<style scoped>
h2 {
  margin: 0 0 8px;
}
.muted {
  color: var(--el-text-color-secondary);
  margin-top: 8px;
}
</style>
