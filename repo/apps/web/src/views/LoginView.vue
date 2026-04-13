<script setup lang="ts">
import { ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const route = useRoute()
const auth = useAuthStore()

const username = ref('demo')
const password = ref('')

function submit() {
  if (!username.value.trim()) {
    return
  }
  auth.login(username.value.trim())
  const redirect = (route.query.redirect as string) || '/dashboard'
  router.replace(redirect)
}
</script>

<template>
  <div class="login-page">
    <el-card class="login-card" shadow="hover">
      <template #header>
        <span>PharmaOps — Sign in</span>
      </template>
      <el-form label-position="top" @submit.prevent="submit">
        <el-form-item label="Username">
          <el-input v-model="username" autocomplete="username" />
        </el-form-item>
        <el-form-item label="Password">
          <el-input v-model="password" type="password" autocomplete="current-password" show-password />
        </el-form-item>
        <el-button type="primary" class="login-btn" native-type="submit">Continue</el-button>
      </el-form>
      <p class="hint">UI scaffold: any password works; session wiring comes with the API.</p>
    </el-card>
  </div>
</template>

<style scoped>
.login-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 24px;
  background: linear-gradient(160deg, var(--el-fill-color-light), var(--el-bg-color));
}
.login-card {
  width: min(420px, 100%);
}
.login-btn {
  width: 100%;
  margin-top: 8px;
}
.hint {
  margin-top: 12px;
  font-size: 12px;
  color: var(--el-text-color-secondary);
}
</style>
