/* store.test.mjs —— store 纯逻辑单测（零 DOM）：
 * snapshot 基线 / RFC6902 patch 增量（成功/失败回退信号）/ 订阅合帧 / 对象级 rv 缓存 /
 * 容器化访问器（objects={Kind:[…]} 与平铺两形态）。 */
import { test } from 'node:test';
import assert from 'node:assert/strict';
import { createStore, objKey } from '../src/core/store.js';

/** 等一拍 microtask（订阅合帧落地）。 */
const tick = () => new Promise((r) => queueMicrotask(() => queueMicrotask(r)));

const SNAP = {
  seq: 7,
  kinds: [{ name: 'Goal' }, { name: 'Rule' }],
  objects: {
    Goal: [
      { name: 'north-star', scope: 'global', resourceVersion: 3 },
      { name: 'meta', scope: 'core', resourceVersion: 1 },
    ],
    Rule: [{ name: 'rule_ux', scope: 'core_design', resourceVersion: 2 }],
  },
  events: [{ id: 'e1', type: 'DocApplied', level: 'info' }],
  drafts: [{ kind: 'Goal', scope: 'global', name: 'north-star', fieldManager: 'console:web' }],
  tasks: { tasks: [{ name: 'build', scope: 'global', status: { phase: 'running' } }] },
};

test('init 建基线 + get 点路径读', () => {
  const st = createStore();
  assert.equal(st.isReady, false);
  st.init(SNAP);
  assert.equal(st.isReady, true);
  assert.equal(st.get('seq'), 7);
  assert.equal(st.get('objects.Goal.0.name'), 'north-star');
  assert.equal(st.get('a.b.c'), undefined); // 缺路径=undefined 不抛
});

test('applyPatch 增量（add/replace/remove）+ seq 推进', () => {
  const st = createStore();
  st.init(SNAP);
  const ok = st.applyPatch([
    { op: 'add', path: '/events/-', value: { id: 'e2', type: 'Escalate' } },
    { op: 'replace', path: '/objects/Goal/0/resourceVersion', value: 4 },
    { op: 'remove', path: '/objects/Rule/0' },
  ], { seq: 8 });
  assert.equal(ok, true);
  assert.equal(st.get('objects.Goal.0.resourceVersion'), 4);
  assert.equal(st.get('events').length, 2);
  assert.equal(st.get('events.1.id'), 'e2');
  assert.equal(st.get('objects.Rule').length, 0);
  assert.equal(st.seq, 8);
});

test('applyPatch 失败（路径错位）=false 且状态保持原样（回退信号）', () => {
  const st = createStore();
  st.init(SNAP);
  const before = st.get('objects.Goal.0.name');
  const ok = st.applyPatch([{ op: 'replace', path: '/objects/Goal/99/name', value: 'x' }]);
  assert.equal(ok, false);
  assert.equal(st.get('objects.Goal.0.name'), before); // 不猜不补——保持原样
  assert.equal(st.seq, 7); // seq 不推进
});

test('空 patch=无变化成功语义', () => {
  const st = createStore();
  st.init(SNAP);
  assert.equal(st.applyPatch([], { seq: 99 }), true);
  assert.equal(st.seq, 7); // 空 patch 不推进 seq
});

test('订阅通知=批处理合帧（一拍多变更只触发一次）', async () => {
  const st = createStore();
  st.init(SNAP);
  let n = 0;
  const unsub = st.subscribe(() => { n += 1; });
  st.applyPatch([{ op: 'replace', path: '/seq', value: 9 }]);
  st.applyPatch([{ op: 'replace', path: '/seq', value: 10 }]);
  st.init({ seq: 11 });
  await tick();
  assert.equal(n, 1); // microtask 合帧
  unsub();
  st.applyPatch([{ op: 'replace', path: '/seq', value: 12 }]);
  await tick();
  assert.equal(n, 1); // 退订后不再通知
});

test('docCache 对象级 rv 语义（命中/过期/缺失）', () => {
  const st = createStore();
  st.init(SNAP);
  st.cacheDoc('Goal', 'global', 'north-star', 3, 'hello: world');
  assert.equal(st.cachedDoc('Goal', 'global', 'north-star', 3).content, 'hello: world');
  assert.equal(st.cachedDoc('Goal', 'global', 'north-star', 4), null); // rv 变=过期不返回
  assert.equal(st.cachedDoc('Goal', 'global', 'absent', 1), null); // 未缓存
  assert.equal(objKey('Goal', '', 'x'), 'Goal|global|x');
});

test('staleDocs：未缓存或 rv 不符的 refs（上限裁剪）', () => {
  const st = createStore();
  st.init(SNAP);
  const stale = st.staleDocs(24);
  assert.equal(stale.length, 3); // 三个对象均未缓存
  st.cacheDoc('Goal', 'global', 'north-star', 3, 'c1');
  st.cacheDoc('Rule', 'core_design', 'rule_ux', 99, 'c2'); // rv 故意不符
  const stale2 = st.staleDocs(24);
  assert.equal(stale2.length, 2);
  assert.ok(stale2.some((r) => r.name === 'rule_ux'));
  const capped = st.staleDocs(1);
  assert.equal(capped.length, 1);
});

test('objects() 容错：平铺形态与桶形态同归一化', () => {
  const st = createStore();
  st.init({ objects: [{ kind: 'Goal', scope: 's1', name: 'a', rv: 5 }] });
  assert.deepEqual(st.objects().map((o) => [o.kind, o.scope, o.name, o.rv]), [['Goal', 's1', 'a', 5]]);
  const st2 = createStore();
  st2.init({ objects: { Goal: [{ name: 'b' }] } }); // scope 缺省=global
  assert.deepEqual(st2.objects().map((o) => [o.kind, o.scope, o.name]), [['Goal', 'global', 'b']]);
});

test('访问器容错：tasks 桶形态 / 空快照不抛', () => {
  const st = createStore();
  st.init(SNAP);
  assert.equal(st.tasks().length, 1);
  assert.equal(st.tasks()[0].status.phase, 'running');
  assert.equal(st.drafts()[0].name, 'north-star');
  assert.equal(st.events().length, 1);
  const empty = createStore();
  empty.init({});
  assert.deepEqual(empty.tasks(), []);
  assert.deepEqual(empty.workflows(), []);
  assert.deepEqual(empty.sessions(), []);
  assert.deepEqual(empty.attention(), []);
  assert.equal(empty.inbox(), null);
});

test('setConn 广播连接态', async () => {
  const st = createStore();
  st.init(SNAP);
  const seen = [];
  const unsub = st.subscribe(() => seen.push(st.conn));
  st.setConn('degraded');
  await tick();
  assert.deepEqual(seen, ['degraded']);
  st.setConn('degraded'); // 同值不重复广播
  await tick();
  assert.equal(seen.length, 1);
  unsub();
});
