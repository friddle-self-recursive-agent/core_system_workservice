// home.test.mjs —— 首页=对话（2026-11-27 DeepSeek chat 模式）机械验收：
// ① 大 composer 词形：真实示例 chips/图片附加/圆形发送；无旧版 session select/方块发送/假 toggle；
// ② 最近会话：真实清单排序/上限/空清单不造假行/会话路由编码；
// ③ 发送路由：sessionRoute 单段编码。
import { test } from 'node:test'
import assert from 'node:assert/strict'
import { heroComposerHtml, recentSessionsHtml, sessionRoute, submitHome } from '../src/screens/home.js'

test('heroComposerHtml：DeepSeek 版大 composer（示例 chips/图片/圆形发送；无旧 select）', () => {
  const html = heroComposerHtml()
  assert.match(html, /class="home-box"/)
  assert.match(html, /data-home-text/)
  assert.match(html, /data-home-send/)
  assert.match(html, /data-home-img/, '图片附加在册')
  assert.match(html, /data-home-sample=/, '真实示例 chips 在册')
  assert.equal((html.match(/data-home-sample=/g) || []).length, 3, '三枚真实示例')
  assert.doesNotMatch(html, /cc-sess|root 会话（默认）/, '旧版 session select 退场（宿主发 id）')
  assert.doesNotMatch(html, /DeepThink|Search/, '不摆 DeepSeek 能力位假 toggle（本系统无对应真通路）')
})

test('recentSessionsHtml：清单倒序/上限 6/空清单不渲染/标题转义/会话路由', () => {
  assert.equal(recentSessionsHtml([]), '', '空清单=空串（不造假行）')
  assert.equal(recentSessionsHtml(null), '')
  const items = Array.from({ length: 9 }, (_, i) => ({
    sessionId: 's-' + i, title: i === 8 ? '<img src=x onerror=alert(1)>' : '会话 ' + i, updatedAt: i * 1000,
  }))
  const html = recentSessionsHtml(items)
  assert.equal((html.match(/data-go="session\//g) || []).length, 6, '只取最近 6 条')
  assert.ok(html.indexOf('s-8') < html.indexOf('s-7'), 'updatedAt 倒序（最新在前）')
  assert.doesNotMatch(html, /<img src=x/, '标题必转义（XSS 红线）')
  assert.match(html, /&lt;img/, '转义后文本在册')
  assert.match(html, /全部会话/, '全部会话入口在册')
})

test('sessionRoute：单段编码（#/session/<id>）', () => {
  assert.equal(sessionRoute('abc'), 'session/abc')
  assert.equal(sessionRoute('a/b c'), 'session/a%2Fb%20c')
})

test('submitHome：create→send(queue)→进入 #/session/<id>（宿主发 id，零手填）', async () => {
  const calls = []
  const notes = []
  const el = {
    querySelector(sel) {
      if (sel === '[data-home-text]') return { value: '帮我看一下文档差距' }
      if (sel === '[data-home-send]') return { disabled: false }
      if (sel === '[data-home-note]') return { set textContent(v) { notes.push(String(v)) }, get textContent() { return notes[notes.length - 1] || '' } }
      return null
    },
  }
  globalThis.window = { __CC: { chatFaces: {
    create: async () => { calls.push(['create']); return { status: 200, data: { sessionId: 's-new' } } },
    send: async (sid, text) => { calls.push(['send', sid, text]); return { status: 200 } },
  } } }
  let nav = ''
  try {
    const ok = await submitHome(el, { go: (p) => { nav = p } })
    assert.equal(ok, true)
    assert.deepEqual(calls, [['create'], ['send', 's-new', '帮我看一下文档差距']], 'create 先行、send 落同一会话（默认 queue）')
    assert.equal(nav, 'session/s-new', '发送成功=进入真实会话屏')
    assert.ok(notes.some((n) => n.includes('正在发送')), '过程词形在册')
  } finally {
    delete globalThis.window
  }
})

test('submitHome：通路缺席=显式缺席词形，不冒充已发送', async () => {
  const notes = []
  const el = {
    querySelector(sel) {
      if (sel === '[data-home-text]') return { value: 'hello' }
      if (sel === '[data-home-send]') return { disabled: false }
      if (sel === '[data-home-note]') return { set textContent(v) { notes.push(String(v)) }, get textContent() { return notes[notes.length - 1] || '' } }
      return null
    },
  }
  let navigated = false
  try {
    const ok = await submitHome(el, { go: () => { navigated = true } })
    assert.equal(ok, false)
    assert.equal(navigated, false)
    assert.ok(notes.some((n) => n.includes('对话通路未注入')), '显式缺席词形')
  } finally {
    delete globalThis.window
  }
})
