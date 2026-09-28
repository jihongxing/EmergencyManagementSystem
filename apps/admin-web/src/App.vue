<script setup lang="ts">
import { computed, ref } from 'vue'

const view = ref<'department' | 'purchase'>('department')
const title = computed(() => view.value === 'department' ? '部门管理' : '组织购买')
</script>

<template>
  <div class="shell">
    <header class="topbar">
      <strong>应急安全检查</strong>
      <span class="environment">开发环境</span>
    </header>
    <div class="workspace">
      <nav class="sidebar" aria-label="入口">
        <button :aria-current="view === 'department' ? 'page' : undefined" @click="view = 'department'">部门管理</button>
        <button :aria-current="view === 'purchase' ? 'page' : undefined" @click="view = 'purchase'">组织购买</button>
      </nav>
      <main>
        <h1>{{ title }}</h1>
        <p class="notice" role="status">身份认证与业务接口尚未接入。此页面不提供正式记录、行政确认或支付操作。</p>
      </main>
    </div>
  </div>
</template>
