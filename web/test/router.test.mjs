/* router.test.mjs —— 路由词形解析单测（纯函数零 DOM）：
 * hash 切段/解码、模式匹配参数、路由解析首中。词形兼容旧 console 全量路由。 */
import { test } from 'node:test';
import assert from 'node:assert/strict';
import { parseHash, matchSegs, createRouter } from '../src/core/router.js';

test('parseHash：空/裸词/深链/查询串剥除', () => {
  assert.deepEqual(parseHash(''), []);
  assert.deepEqual(parseHash('#/'), []);
  assert.deepEqual(parseHash('#/obj'), ['obj']);
  assert.deepEqual(parseHash('#/wf/core/r/3'), ['wf', 'core', 'r', '3']);
  assert.deepEqual(parseHash('#/docs?q=x'), ['docs']);
  assert.deepEqual(parseHash(''), []);
});

test('parseHash：%2F 整段单次编码解码（task id / designdoc path 词形）', () => {
  assert.deepEqual(parseHash('#/task/global%2Fbuild'), ['task', 'global/build']);
  assert.deepEqual(
    parseHash('#/designdoc/core_design/rule%2Frule_ux.md'),
    ['designdoc', 'core_design', 'rule/rule_ux.md'],
  );
});

test('parseHash：非法编码段原样保留不抛', () => {
  assert.deepEqual(parseHash('#/obj/%ZZ'), ['obj', '%ZZ']);
});

test('matchSegs：参数段/字面段/长度不齐', () => {
  assert.deepEqual(
    matchSegs('doc/:kind/:scope/:name', ['doc', 'Goal', 'global', 'north-star']),
    { kind: 'Goal', scope: 'global', name: 'north-star' },
  );
  assert.deepEqual(matchSegs('wf/:name/r/:n', ['wf', 'core', 'r', '12']), { name: 'core', n: '12' });
  assert.equal(matchSegs('obj', ['obj', 'Goal']), null); // 长度不齐
  assert.equal(matchSegs('events', ['event']), null); // 字面不匹配
  assert.deepEqual(matchSegs('obj', ['obj']), {});
});

test('createRouter.resolve：按注册序首中', () => {
  const router = createRouter();
  const hits = [];
  router
    .route('doc/:name', (el, p) => hits.push(['short', p]))
    .route('doc/:kind/:scope/:name', (el, p) => hits.push(['long', p]));
  const hit = router.resolve(['doc', 'Goal', 'global', 'north-star']);
  assert.ok(hit, '命中长词形');
  assert.deepEqual(hit.params, { kind: 'Goal', scope: 'global', name: 'north-star' });
  assert.equal(router.resolve(['doc']), null); // 无匹配
});

test('createRouter：mount 前不触 DOM/location（node 可构造）', () => {
  const router = createRouter();
  router.route('obj', () => {});
  assert.equal(router.render(), null); // 无容器=安全 no-op
  assert.deepEqual(router.current.segs, []);
});

test('词形全量兼容抽样（旧 console 路由表）', () => {
  const router = createRouter();
  const table = [
    'home', 'welcome', 'obj', 'obj/:kind', 'obj/:kind/:scope/:name', 'events', 'event/:id',
    'tasks', 'task/:id', 'report', 'doc/:name', 'doc/:kind/:scope/:name',
    'chat/:agent', 'session/:id', 'approvals', 'inbox', 'proposals',
    'workflows', 'wf/:name', 'wf/:name/r/:n', 'tree', 'docs',
    'designdoc', 'designdoc/:repo/:path', 'connectors', 's/:name',
  ];
  table.forEach((p) => router.route(p, () => {}));
  const cases = [
    [['obj', 'Goal'], 'obj/:kind'],
    [['obj', 'Goal', 'core', 'meta'], 'obj/:kind/:scope/:name'],
    [['event', 'abc'], 'event/:id'],
    [['task', 'global/build'], 'task/:id'],
    [['chat', 'root-agent'], 'chat/:agent'],
    [['session', 's-1'], 'session/:id'],
    [['wf', 'core', 'r', '7'], 'wf/:name/r/:n'],
    [['designdoc', 'core_design', 'x/y.md'], 'designdoc/:repo/:path'],
    [['s', 'core'], 's/:name'],
  ];
  cases.forEach(([segs, want]) => {
    const hit = router.resolve(segs);
    assert.ok(hit, segs.join('/'));
    assert.equal(hit.r.pattern, want, segs.join('/'));
  });
});
