<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'

const route = useRoute()
const title = computed(() => (route.meta.title as string) || 'Module')
</script>

<template>
  <el-card shadow="never">
    <h2>{{ title }}</h2>
    <p class="muted">Placeholder screen — implementation follows in domain milestones.</p>
  </el-card>
</template>

<style scoped>
h2 {
  margin: 0 0 8px;
}
.muted {
  color: var(--el-text-color-secondary);
}
</style>
