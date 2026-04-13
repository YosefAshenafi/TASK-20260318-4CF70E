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
    <div class="login-brand" aria-hidden="true">
      <div class="login-brand-inner">
        <p class="login-kicker">PharmaOps</p>
        <h1 class="login-headline">Talent &amp; compliance operations</h1>
        <p class="login-sub">
          Role-aware access, auditable workflows, and secure handling of sensitive data — built for
          intranet deployment.
        </p>
      </div>
    </div>

    <div class="login-panel">
      <el-card class="login-card" shadow="never">
        <div class="login-card-header">
          <h2 class="login-title">Sign in</h2>
          <p class="login-lede">Use your organization account to continue.</p>
        </div>
        <el-form label-position="top" size="large" class="login-form" @submit.prevent="submit">
          <el-form-item label="Username">
            <el-input v-model="username" autocomplete="username" placeholder="Enter username" />
          </el-form-item>
          <el-form-item label="Password">
            <el-input
              v-model="password"
              type="password"
              autocomplete="current-password"
              show-password
              placeholder="Enter password"
            />
          </el-form-item>
          <el-button type="primary" class="login-btn" native-type="submit" round>Continue</el-button>
        </el-form>
        <p class="hint">
          Scaffold mode: password is not validated yet. Server-side session and bcrypt apply with the
          API.
        </p>
      </el-card>
    </div>
  </div>
</template>

<style scoped>
.login-page {
  min-height: 100vh;
  width: 100%;
  display: flex;
  flex-direction: column;
}

@media (min-width: 900px) {
  .login-page {
    flex-direction: row;
  }
}

.login-brand {
  flex: 1;
  min-height: 220px;
  padding: clamp(2rem, 5vw, 3.5rem);
  display: flex;
  align-items: flex-end;
  background:
    radial-gradient(120% 80% at 80% 0%, rgba(94, 234, 212, 0.35) 0%, transparent 55%),
    radial-gradient(90% 60% at 10% 100%, rgba(15, 118, 110, 0.5) 0%, transparent 50%),
    linear-gradient(145deg, #042f2e 0%, #0f766e 42%, #115e59 100%);
  color: #ecfdf5;
}

@media (min-width: 900px) {
  .login-brand {
    align-items: center;
    min-height: 100vh;
  }
}

.login-brand-inner {
  max-width: 28rem;
}

.login-kicker {
  margin: 0 0 0.5rem;
  font-size: 0.8125rem;
  font-weight: 600;
  letter-spacing: 0.12em;
  text-transform: uppercase;
  opacity: 0.85;
}

.login-headline {
  margin: 0 0 0.75rem;
  font-size: clamp(1.5rem, 3vw, 2rem);
  font-weight: 650;
  letter-spacing: -0.02em;
  line-height: 1.2;
  color: #f0fdfa;
}

.login-sub {
  margin: 0;
  font-size: 0.9375rem;
  line-height: 1.55;
  opacity: 0.88;
}

.login-panel {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: clamp(1.5rem, 4vw, 2.5rem);
  background: linear-gradient(180deg, #f8fafc 0%, #f1f5f9 100%);
}

.login-card {
  width: min(100%, 26rem);
  border-radius: 16px !important;
  border: 1px solid var(--el-border-color-lighter);
  box-shadow: var(--el-box-shadow-lighter) !important;
}

.login-card :deep(.el-card__body) {
  padding: clamp(1.5rem, 4vw, 2rem);
}

.login-card-header {
  margin-bottom: 1.25rem;
}

.login-title {
  margin: 0 0 0.35rem;
  font-size: 1.375rem;
  font-weight: 650;
  letter-spacing: -0.02em;
  color: var(--el-text-color-primary);
}

.login-lede {
  margin: 0;
  font-size: 0.875rem;
  color: var(--el-text-color-secondary);
  line-height: 1.45;
}

.login-form :deep(.el-form-item__label) {
  font-weight: 500;
}

.login-form :deep(.el-input__wrapper) {
  border-radius: 10px;
  box-shadow: 0 0 0 1px var(--el-border-color) inset;
  transition: box-shadow 0.2s ease;
}

.login-form :deep(.el-input__wrapper:hover),
.login-form :deep(.el-input__wrapper.is-focus) {
  box-shadow: 0 0 0 1px var(--el-color-primary-light-5) inset;
}

.login-btn {
  width: 100%;
  margin-top: 0.25rem;
  height: 44px;
  font-weight: 600;
}

.hint {
  margin: 1.25rem 0 0;
  padding-top: 1rem;
  border-top: 1px solid var(--el-border-color-lighter);
  font-size: 0.75rem;
  line-height: 1.45;
  color: var(--el-text-color-secondary);
}
</style>
