import { describe, expect, it } from 'vitest'
import { createSSRApp } from 'vue'
import { renderToString } from 'vue/server-renderer'
import App from './App.vue'

describe('engineering shell', () => {
  it('renders the identity entry without operational forms during bootstrap', async () => {
    const html = await renderToString(createSSRApp(App))
    expect(html).toContain('正在读取身份')
    expect(html).not.toContain('部门管理')
    expect(html).not.toContain('组织购买')
  })
})
