<script setup lang="ts">
import { useRouter } from 'vue-router'

const router = useRouter()
</script>

<template>
  <div class="wrap">
    <el-result icon="warning" title="Page not found" sub-title="The link may be wrong or the page was removed.">
      <template #extra>
        <el-button type="primary" @click="router.push('/dashboard')">Go to dashboard</el-button>
      </template>
    </el-result>
  </div>
</template>

<style scoped>
.wrap {
  min-height: 60vh;
  display: flex;
  align-items: center;
  justify-content: center;
}
</style>
