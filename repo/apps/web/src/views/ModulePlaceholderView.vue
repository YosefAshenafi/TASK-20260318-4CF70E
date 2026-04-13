<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'

const route = useRoute()
const title = computed(() => (route.meta.title as string) || 'Module')
</script>

<template>
  <el-card class="panel" shadow="never">
    <h2>{{ title }}</h2>
    <p class="muted">Placeholder screen — implementation follows in domain milestones.</p>
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
  line-height: 1.5;
}
</style>
