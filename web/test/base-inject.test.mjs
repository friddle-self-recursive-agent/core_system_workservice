import { test } from 'node:test';
import assert from 'node:assert/strict';
import { createApi } from '../src/core/api.js';
import { wsUrlOf } from '../src/core/ws.js';

function withFetchCapture() {
  const seen = [];
  const orig = globalThis.fetch;
  globalThis.fetch = async (url, opts) => {
    seen.push({ url: String(url), opts: opts || {} });
    return { ok: true, status: 200, text: async () => '{}' };
  };
  return { seen, restore: () => { globalThis.fetch = orig; } };
}

test('createApi 直连缺省：读 v1 词位+写 v1/apply·drafts', async () => {
  const fc = withFetchCapture();
  try {
    const api = createApi();
    await api.snapshot();
    await api.apply({ a: 1 });
    await api.saveDraft({ b: 2 });
    assert.deepEqual(fc.seen.map((s) => s.url), ['/v1/snapshot', '/v1/apply', '/v1/drafts']);
    assert.equal(fc.seen[1].opts.method, 'POST');
  } finally { fc.restore(); }
});

test('createApi 代理基路径：读走 api 前缀+写走宿主注入词位（BASES 注入契约）', async () => {
  const fc = withFetchCapture();
  try {
    const api = createApi('/core-console/api', { apply: '/core-console/apply', drafts: '/core-console/drafts' });
    await api.snapshot();
    await api.events(5);
    await api.apply({});
    await api.saveDraft({});
    assert.deepEqual(fc.seen.map((s) => s.url),
      ['/core-console/api/v1/snapshot', '/core-console/api/v1/events?limit=5', '/core-console/apply', '/core-console/drafts']);
    assert.equal(fc.seen[2].opts.headers['Content-Type'], 'application/json');
  } finally { fc.restore(); }
});

test('wsUrlOf 自定义词位：node 无 location=相对路径形态', () => {
  assert.equal(wsUrlOf(null, 'core-console/ws'), '/core-console/ws');
  assert.equal(wsUrlOf(null), '/v1/stream');
});
