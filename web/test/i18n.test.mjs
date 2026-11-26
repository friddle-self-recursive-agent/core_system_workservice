import { test } from 'node:test';
import assert from 'node:assert/strict';
import { t, getLocale, setLocale, onLocaleChange, dictOf } from '../src/core/i18n.js';

test('zh 恒等：源词形直返（测试与降级双语义）', () => {
  setLocale('zh');
  assert.equal(t('任务中心'), '任务中心');
  assert.equal(t('Not in dict'), 'Not in dict');
});

test('en 命中：词典现译', () => {
  setLocale('en');
  assert.equal(t('任务中心'), 'Task Center');
  assert.equal(t('实时（WS patch）'), 'Live (WS patch)');
  setLocale('zh');
});

test('en 缺译：诚实回退键本身（不冒充翻译）', () => {
  setLocale('en');
  assert.equal(t('___no_such_key___'), '___no_such_key___');
  setLocale('zh');
});

test('setLocale 未注册词形=忽略（不炸运行态）', () => {
  const before = getLocale();
  setLocale('fr');
  assert.equal(getLocale(), before);
});

test('onLocaleChange 订阅/退订', () => {
  let hits = 0;
  const off = onLocaleChange(() => { hits++; });
  setLocale('en');
  off();
  setLocale('zh');
  assert.equal(hits, 1);
});

test('EN 词典规模≥300（屏内键覆盖底线）', () => {
  assert.ok(dictOf('en').size >= 300, 'dict size=' + dictOf('en').size);
});
