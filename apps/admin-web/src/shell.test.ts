import { describe, expect, it } from 'vitest'
import { createSSRApp } from 'vue'
import { renderToString } from 'vue/server-renderer'
import App from './App.vue'

describe('engineering shell', () => {
  it('renders the unauthenticated shell without operational forms', async () => {
    const html = await renderToString(createSSRApp(App))
    expect(html).toContain('身份认证与业务接口尚未接入')
    expect(html).not.toContain('<form')
  })
})
