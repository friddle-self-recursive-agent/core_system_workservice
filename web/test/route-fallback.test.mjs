/* route-fallback.test.mjs —— #22（R-UX-3 同族）路由兜底回归：
 * ① #/chat 裸路径重定向默认 agent 对话；② 需参数裸路径报「缺参数+可点出口」；
 * ③ 真未知路径仍报未知路径（不冒充）。纯 router 单测——假 location + 假 view 零 DOM。 */
import { test } from 'node:test';
import assert from 'node:assert/strict';
import { createRouter } from '../src/core/router.js';

function fakeEnv(hash) {
  globalThis.location = { hash };
  return { view: { innerHTML: '' } };
}

test('#/chat 裸路径 → 默认 agent 对话（不再未知路径死胡同）', () => {
  const { view } = fakeEnv('#/chat');
  const hits = [];
  const router = createRouter({ defaultRoute: 'home' });
  router
    .route('home', () => hits.push('home'))
    .route('chat', () => router.go('chat/core-headless'))
    .route('chat/:agent', (el, p) => hits.push('chat/' + p.agent));
  router.mount(view);
  assert.equal(location.hash, '#/chat/core-headless', '重定向到可分享的默认 agent 深链');
  assert.doesNotMatch(view.innerHTML, /未知路径/);
  router.render(); // 浏览器由 hashchange 推进；node 无 window，手动再渲染一帧
  assert.deepEqual(hits, ['chat/core-headless'], '落到对话屏而非错误屏');
});

test('需参数裸路径 → 缺参数+可点出口（不甩参数拼写）', () => {
  const cases = [
    ['#/chat', 'agent', 'chat/core-headless'],
    ['#/task', 'id', 'tasks'],
    ['#/doc', 'name', 'docs'],
    ['#/wf', 'name', 'workflows'],
  ];
  for (const [hash, param, go] of cases) {
    const { view } = fakeEnv(hash);
    const router = createRouter({ defaultRoute: 'home' });
    router
      .route('home', () => {})
      .route('task/:id', () => {})
      .route('chat/:agent', () => {})
      .route('doc/:name', () => {})
      .route('wf/:name', () => {});
    router.mount(view);
    assert.match(view.innerHTML, /缺参数/, hash);
    assert.ok(view.innerHTML.includes('<code>' + param + '</code>'), hash + ' 报出缺失参数名');
    assert.ok(view.innerHTML.includes('data-go="' + go + '"'), hash + ' 给出该族可点出口');
    assert.ok(view.innerHTML.includes('data-go="home"'), hash + ' 保留回首页');
    assert.doesNotMatch(view.innerHTML, /未知路径/, hash);
  }
});

test('真未知路径仍报未知路径（不冒充缺参数）', () => {
  const { view } = fakeEnv('#/nope/nothing');
  const router = createRouter();
  router.route('home', () => {});
  router.mount(view);
  assert.match(view.innerHTML, /未知路径（nope\/nothing）/);
  assert.doesNotMatch(view.innerHTML, /缺参数/);
  assert.match(view.innerHTML, /data-go="home"/);
});
