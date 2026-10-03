<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { createAuth, workspaceFor } from './auth'

const auth = createAuth()
const loginId = ref('')
const password = ref('')
const error = ref('')
const workspace = computed(() => workspaceFor(auth.member.value))

onMounted(() => auth.bootstrap())

async function submit() {
  error.value = ''
  try {
    await auth.login(loginId.value, password.value)
  } catch {
    error.value = '登录失败或当前账号不可用'
  }
}
</script>

<template>
  <div class="shell">
    <header class="topbar"><strong>应急安全检查</strong><span class="environment">开发环境</span></header>
    <main v-if="auth.state.value === 'loading'" class="center">正在读取身份...</main>
    <section v-else-if="auth.state.value !== 'authenticated'" class="login-panel">
      <h1>登录</h1>
      <p v-if="auth.state.value === 'forbidden'" class="notice" role="alert">当前账号无可用工作区或已被停用。</p>
      <form @submit.prevent="submit">
        <label>登录标识<input v-model="loginId" autocomplete="username" required /></label>
        <label>密码<input v-model="password" type="password" autocomplete="current-password" required /></label>
        <p v-if="error" class="error" role="alert">{{ error }}</p>
        <button class="primary" type="submit">进入系统</button>
      </form>
    </section>
    <div v-else class="workspace">
      <nav class="sidebar" aria-label="工作区">
        <strong>{{ workspace === 'enterprise' ? '企业工作区' : '行政工作区' }}</strong>
        <span v-if="workspace === 'enterprise'">组织购买（占位）</span>
        <span v-else>部门监管入口</span>
        <button @click="auth.logout">退出登录</button>
      </nav>
      <main>
        <h1>{{ workspace === 'enterprise' ? '企业工作区' : '行政工作区' }}</h1>
        <p class="notice">当前入口由服务端身份决定。业务记录和正式行政操作将在后续任务实现。</p>
      </main>
    </div>
  </div>
</template>
