/* sessionctx.test.mjs —— 会话上下文面板纯逻辑单测（G-14，零 DOM）：
 * 归一化（normalizeContext：载荷字段→视图模型缺省词形/缺席不冒充）、装配面收敛
 * （capTags：去重/每类≤5/总≤20/裸 tag 归 ''）、编辑回写 payload（buildSavePayload：
 * 预算域校验/乐观锁基准/fieldManager 词形）、权限降级渲染（renderSessionCtx：写面关断
 * 词形/注入词形转义/空态诚实词形/分片面缺席降级）。 */
import { test } from 'node:test';
import assert from 'node:assert/strict';
import { parseTagInput, capTags, normalizeContext, buildSavePayload, renderSessionCtx }
  from '../src/screens/sessionctx.js';

/* ---------- tag 输入解析 ---------- */

test('parseTagInput：逗号/空白/中文分隔符归一', () => {
  assert.deepEqual(parseTagInput('domain:console, mech:cache  lifecycle:recovery，d、e\nf'), [
    'domain:console', 'mech:cache', 'lifecycle:recovery', 'd', 'e', 'f',
  ]);
  assert.deepEqual(parseTagInput(''), []);
  assert.deepEqual(parseTagInput(null), []);
  assert.deepEqual(parseTagInput('  '), []);
});

/* ---------- ContextPolicy.packLimits 装配面收敛 ---------- */

test('capTags：保序去重 + 每 tagType ≤5 + 总 ≤20（超限进 dropped）', () => {
  const six = ['domain:a', 'domain:b', 'domain:c', 'domain:d', 'domain:e', 'domain:f'];
  const r1 = capTags(six);
  assert.equal(r1.kept.length, 5);
  assert.deepEqual(r1.dropped, ['domain:f']);
  assert.deepEqual(r1.kept, six.slice(0, 5));
  const many = [];
  for (let i = 0; i < 25; i++) many.push('t' + i + ':x');
  const r2 = capTags(many);
  assert.equal(r2.kept.length, 20);
  assert.equal(r2.dropped.length, 5);
  // 裸 tag 归 type=''（与后端 splitTagKey 同词形——'' 桶同样 ≤5）。
  const bare = ['a', 'b', 'c', 'd', 'e', 'f'];
  const r3 = capTags(bare);
  assert.deepEqual(r3.kept, bare.slice(0, 5));
  assert.deepEqual(r3.dropped, ['f']);
  // 去重：重复键不占额。
  const r4 = capTags(['domain:a', ' domain:a ', 'domain:b']);
  assert.deepEqual(r4.kept, ['domain:a', 'domain:b']);
  assert.deepEqual(r4.dropped, []);
});

/* ---------- GET 载荷归一化 ---------- */

test('normalizeContext：全载荷词形（pack/shards/override/writeGate）', () => {
  const m = normalizeContext({
    sessionId: 's1',
    session: { state: 'suspected-stall', phase: 'tool-call', lastAction: 'apply', heartbeats: 3, lastBeat: '2026-11-26T10:00:00Z' },
    pack: {
      takeBudget: 40000, takeBudgetDefault: 30000,
      tagSig: [{ tagType: 'domain', tag: 'notify', tagKey: 'domain:notify' }],
      tagSigSource: 'override',
      chunks: [{ kind: 'Rule', scope: 'deploy', name: 'other-doc', score: 71.3, why: 'tag 命中' }],
      tokens: null, degraded: true, degradedReason: '本代无三段式装配器',
    },
    shards: [{ tagKey: 'domain:notify', docs: 2, topDocs: ['Rule/deploy/other-doc'], stale: false }],
    shardsStats: { total: 1, stale: 0, rebuildThreshold: 75 },
    policy: { shardMaxTokens: 6000, minTakeBudget: 2000, maxTakeBudget: 60000, maxTagsPerType: 5, maxTagsTotal: 20 },
    override: { takeBudget: 40000, tagSig: ['domain:notify'], note: '试笔', by: 'console:web', revision: 2, updatedAt: '2026-11-26T11:00:00Z' },
    canRevert: true,
    writeGate: { open: true, hint: '--allow-apply' },
  });
  assert.equal(m.present, true);
  assert.equal(m.session.state, 'suspected-stall');
  assert.equal(m.takeBudget, 40000);
  assert.deepEqual(m.tags, ['domain:notify']);
  assert.equal(m.tagSigSource, 'override');
  assert.equal(m.chunks.length, 1);
  assert.equal(m.chunks[0].score, 71.3);
  assert.equal(m.shards[0].docs, 2);
  assert.equal(m.stats.threshold, 75);
  assert.equal(m.policy.maxTagsTotal, 20);
  assert.equal(m.override.revision, 2);
  assert.equal(m.canRevert, true);
  assert.equal(m.writeOpen, true);
  assert.equal(m.shardsFacePresent, true);
});

test('normalizeContext：缺席词形不冒充（null 面=空集+降级旗标）', () => {
  const m = normalizeContext({
    sessionId: 'ghost', session: null,
    pack: { takeBudget: 30000, takeBudgetDefault: 30000, tagSig: [], tagSigSource: 'none', chunks: [], tokens: null, degraded: true },
    shards: null, shardsStats: null,
    override: null, canRevert: false, writeGate: { open: false },
  });
  assert.equal(m.session, null);
  assert.deepEqual(m.tags, []);
  assert.deepEqual(m.chunks, []);
  assert.equal(m.stats, null);
  assert.equal(m.shardsFacePresent, false); // 分片面缺席旗标（渲染区显式降级词形）
  assert.equal(m.override, null);
  assert.equal(m.writeOpen, false);
  assert.equal(m.policy.maxTagsPerType, 5); // 缺省兜底（不冒充 0 上限）
  assert.equal(m.degraded, true);
  // 非对象入参=全缺省（不抛错）。
  const empty = normalizeContext(null);
  assert.equal(empty.present, false);
  assert.deepEqual(empty.chunks, []);
});

/* ---------- 编辑回写 payload ---------- */

test('buildSavePayload：预算域校验 + 乐观锁基准 + 收敛词形', () => {
  const model = normalizeContext({
    policy: { minTakeBudget: 2000, maxTakeBudget: 60000, maxTagsPerType: 5, maxTagsTotal: 20 },
    override: { revision: 7, takeBudget: 40000 },
  });
  const good = buildSavePayload(model, { takeBudget: '45000', tagList: 'domain:notify, domain:notify mech:cache bare', note: '试笔' });
  assert.equal(good.ok, true);
  assert.equal(good.payload.takeBudget, 45000); // 字符串数字归一
  assert.deepEqual(good.payload.tagSig, ['domain:notify', 'mech:cache', 'bare']);
  assert.equal(good.payload.baseRevision, 7); // 乐观锁基准=现行 override revision
  assert.equal(good.payload.by, 'console:web');
  assert.deepEqual(good.dropped, []);
  // 无覆盖=基准 0。
  const fresh = buildSavePayload(normalizeContext({ policy: { minTakeBudget: 2000, maxTakeBudget: 60000 } }), { takeBudget: 30000, tagList: '' });
  assert.equal(fresh.payload.baseRevision, 0);
  assert.deepEqual(fresh.payload.tagSig, []);
  // 预算越界/非数=ok:false 携词形（不静默收敛——诚实拒绝）。
  for (const bad of ['abc', '1999', '60001', '', null, 'NaN']) {
    const r = buildSavePayload(model, { takeBudget: bad, tagList: '' });
    assert.equal(r.ok, false, 'bad=' + bad);
    assert.ok(r.error.includes('2000'));
  }
  // 超限 tag 进 dropped（后端同一 packLimits——前端先收敛+回显）。
  const six = buildSavePayload(model, { takeBudget: 30000, tagList: 'd:1,d:2,d:3,d:4,d:5,d:6' });
  assert.equal(six.ok, true);
  assert.equal(six.payload.tagSig.length, 5);
  assert.deepEqual(six.dropped, ['d:6']);
  // note 截断 512（写面同上限——双闸一致）。
  const long = buildSavePayload(model, { takeBudget: 30000, tagList: '', note: 'x'.repeat(600) });
  assert.equal(long.payload.note.length, 512);
});

/* ---------- 渲染（权限降级/转义/空态） ---------- */

test('renderSessionCtx：写面关断=诚实禁用词形（不冒充可编辑）', () => {
  const m = normalizeContext({ sessionId: 's1', writeGate: { open: false }, pack: { degraded: true, tagSig: [], chunks: [] }, shards: null });
  const html = renderSessionCtx(m);
  assert.ok(html.includes('写面关断'));
  assert.ok(html.includes('--allow-apply'));
  assert.ok(!html.includes('data-ctx-save')); // 关断面零保存入口
  assert.ok(html.includes('分片缓存面缺席')); // shards:null 显式降级（不冒充空缓存）
});

test('renderSessionCtx：动态文本全量转义（注入词形不透出）', () => {
  const m = normalizeContext({
    sessionId: 'x',
    pack: {
      degraded: true, tagSigSource: 'override',
      tagSig: [{ tagType: 'domain', tag: '<img src=x onerror=alert(1)>', tagKey: 'domain:<img src=x onerror=alert(1)>' }],
      chunks: [{ kind: 'Rule', scope: 's', name: '<script>alert(1)</script>', score: 50, why: '"quoted" & <angly>' }],
    },
    shards: [{ tagKey: '"><svg onload=alert(1)>', docs: 1, topDocs: [], stale: true }],
    override: { revision: 1, takeBudget: 30000, note: '<b>not-html</b>', by: 'a"b', updatedAt: '2026-11-26T00:00:00Z' },
    writeGate: { open: true }, canRevert: true,
  });
  const html = renderSessionCtx(m);
  assert.ok(!html.includes('<img src=x'));
  assert.ok(!html.includes('<script>'));
  assert.ok(!html.includes('<svg onload'));
  assert.ok(html.includes('&lt;script&gt;'));
  assert.ok(html.includes('&quot;quoted&quot;'));
  assert.ok(html.includes('&lt;b&gt;not-html&lt;/b&gt;')); // 备注同样过 esc（非 md 通路）
});

test('renderSessionCtx：空态与降级词形（tagSig 未装配/chunks 空/会话不在账）', () => {
  const m = normalizeContext({
    sessionId: 'ghost', session: null,
    pack: { takeBudget: 30000, takeBudgetDefault: 30000, tagSig: [], tagSigSource: 'none', chunks: [], degraded: true, degradedReason: '本代无三段式装配器' },
    shards: [], shardsStats: { total: 0, stale: 0, rebuildThreshold: 75 },
  });
  const html = renderSessionCtx(m);
  assert.ok(html.includes('会话不在活性账')); // 不冒充在册
  assert.ok(html.includes('无索引候选'));
  assert.ok(html.includes('无生效 tag 分片'));
  assert.ok(html.includes('索引候选投影')); // degraded 恒真词形
  assert.ok(html.includes('本代无三段式装配器'));
  assert.ok(html.includes('无覆盖')); // override 缺席词形
  assert.ok(html.includes('disabled')); // 撤销钮禁用（栈空）
});
