// proposals-empty-status.test.mjs —— R-ENG-9 数据零编造（issue #19）：
// 提案卡空 status 不得兜底成编造业务词形「待定」，须显式「未声明」；
// 空字段（summary/status/proposer）整项省略标签，不再输出无信息量的 "k: " 噪音。
import { test } from 'node:test'
import assert from 'node:assert/strict'
import { render } from '../src/screens/misc.js'
import { setLocale } from '../src/core/i18n.js'

function proposalsHtml(rows, params = {}) {
  const el = {
    innerHTML: '',
    querySelectorAll() { return [] },
    querySelector() { return null },
  }
  setLocale('zh')
  render(el, { route: 'proposals', params, store: { proposals: () => rows } })
  return el.innerHTML
}

test('空 status=显式「未声明」，不再有「待定」编造兜底', () => {
  const html = proposalsHtml([{ name: 'align-docs-reality-20261127', scope: 'global', summary: '', status: '', proposer: '' }])
  assert.match(html, /未声明/, '缺席词形在场')
  assert.doesNotMatch(html, /待定/, 'R-ENG-9：空态不得冒充业务状态')
})

test('空字段整项省略标签：summary/status/proposer 不渲染，非空字段保留', () => {
  const html = proposalsHtml([{ name: 'align-docs-reality-20261127', scope: 'global', summary: '', status: '', proposer: '' }])
  assert.doesNotMatch(html, /summary:/, '空 summary 无标签')
  assert.doesNotMatch(html, /status:/, '空 status 无标签')
  assert.doesNotMatch(html, /proposer:/, '空 proposer 无标签')
  assert.match(html, /scope: global/, '非空 scope 保留')
  assert.match(html, /name: align-docs-reality-20261127/, '非空 name 保留')
})

test('真实值仍按原样呈现：status chip 与非空要素字段不被误删', () => {
  const html = proposalsHtml([{ name: 'p-live', title: '真提案', status: 'approved', summary: '已对齐', proposer: 'alice', scope: 'global' }])
  assert.match(html, /approved/, '真实 status 作 chip')
  assert.doesNotMatch(html, /未声明/, '有真实状态不显示缺席词')
  assert.match(html, /summary: 已对齐/)
  assert.match(html, /status: approved/)
  assert.match(html, /proposer: alice/)
})

test('纯空白字段等同缺席（trim 判据），展开详情对空 summary 显式缺席', () => {
  const html = proposalsHtml([{ name: 'p-blank', scope: 'global', status: '   ', summary: '  ', proposer: undefined }], { name: 'p-blank' })
  assert.match(html, /未声明/, '空白 status=缺席')
  assert.doesNotMatch(html, /status:/, '空白 status 不出标签')
  assert.doesNotMatch(html, /summary:/, '空白 summary 不出标签')
  assert.match(html, /summary 缺席——详情见对象工作台/, '展开详情显式缺席词形（非编造）')
})

test('全字段缺席=要素字段缺席（显式空态），不造占位字段', () => {
  const html = proposalsHtml([{}])
  assert.match(html, /要素字段缺席/)
  assert.match(html, /未命名提案/)
  assert.doesNotMatch(html, /summary:|status:|proposer:/)
  assert.doesNotMatch(html, /待定/)
})
