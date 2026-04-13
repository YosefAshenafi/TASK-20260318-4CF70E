<script setup lang="ts">
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
</script>

<template>
  <el-card class="panel" shadow="never">
    <h2>Welcome</h2>
    <p v-if="auth.username">Signed in as <strong>{{ auth.username }}</strong>.</p>
    <p class="muted">Use the sidebar to open modules.</p>
  </el-card>
</template>

<style scoped>
.panel {
  border-radius: 14px;
  border: 1px solid var(--el-border-color-lighter);
  box-shadow: var(--el-box-shadow-light);
}

.panel :deep(.el-card__body) {
  padding: 1.5rem 1.75rem;
}

h2 {
  margin: 0 0 8px;
  font-size: 1.25rem;
  font-weight: 650;
  letter-spacing: -0.02em;
}

.muted {
  color: var(--el-text-color-secondary);
  margin-top: 8px;
  line-height: 1.5;
}
</style>
