/* inbox-count.test.mjs —— #20 收件箱计数与清单同源（welcome 大屏「52 待你处理」失真收口）：
 * ① inboxCountOf === pendingItems.length（同一过滤函数单点，raw inbox.items 不再进计数）；
 * ② scope 收窄同口径；③ 草稿不存在（apply/弃稿/现读 404）出队——计数与清单同步除名；
 * ④ 网络/5xx 不冒充死稿（诚实降级不误删）。 */
import { test } from 'node:test';
import assert from 'node:assert/strict';
import { pendingItems, draftKey, dequeueDeadDraft, isDeadDraftResult, pendingItemScope } from '../src/screens/approvals.js';
import { inboxCountOf } from '../src/screens/welcome.js';

/** 造 ctx：inbox 投影带 52 条陈旧草稿/归档行注水，真实待办只有 pendingItems 那几条。 */
function ctxOf(tag) {
  const drafts = [
    { kind: 'Workflow', scope: 'global', name: tag + '-live', fieldManager: 'agent:probe' },
    { kind: 'ServiceDesc', scope: 'svc-a', name: tag + '-scoped', fieldManager: 'agent:probe' },
  ];
  const runs = [
    { workflow: 'wf-a', run: 1, status: 'suspended', ownerScope: 'svc-a' },
    { workflow: 'wf-b', run: 2, phase: 'AwaitingApproval', ownerScope: 'svc-a' },
    { workflow: 'wf-c', run: 3, status: 'running', ownerScope: 'svc-a' }, // 非挂起=不在账
  ];
  const proposals = [
    { name: tag + '-prop-open', scope: 'svc-a', status: 'submitted' },
    { name: tag + '-prop-done', scope: 'svc-a', status: 'approved' }, // 已决=不在账
  ];
  const events = [
    { id: tag + '-ev-appr', type: 'ApprovalNeeded', involved: 'Goal/global/n', scope: 'svc-a' },
    { id: tag + '-ev-doc', type: 'DocApplied', involved: 'Goal/global/n' }, // 归档行=不在账
  ];
  const store = {
    drafts: () => drafts,
    runs: () => runs,
    proposals: () => proposals,
    events: () => events,
    inbox: () => ({ count: 52, items: Array.from({ length: 52 }, (_, i) => ({ digest: '陈旧行 ' + i, class: 'doc-review' })) }),
  };
  return { store };
}

test('收件箱计数 = 清单同一过滤函数的长度（52 条 raw 投影不再冒充待办）', () => {
  const ctx = ctxOf('t-parity');
  const list = pendingItems(ctx);
  assert.equal(list.length, 6, '2 稿 + 2 挂起 run + 1 待批提案 + 1 审批事件');
  assert.equal(inboxCountOf(ctx), list.length, 'welcome 计数与审批中心清单同源同长');
  assert.notEqual(ctx.store.inbox().count, list.length, 'raw inbox.count=52 明确不是待办数');
});

test('Service 视角 scope 收窄：计数与清单同口径', () => {
  const ctx = ctxOf('t-scope');
  const scoped = pendingItems(ctx).filter((it) => pendingItemScope(it) === 'svc-a');
  assert.equal(scoped.length, 5, 'global 稿被 svc-a 视角排除');
  assert.equal(inboxCountOf(ctx, 'svc-a'), scoped.length);
  assert.equal(inboxCountOf(ctx, 'nowhere'), 0, '无匹配 scope=零待办，不回退全量');
});

test('死稿出队：草稿不存在后清单与计数同步除名（不再挂死卡）', () => {
  const ctx = ctxOf('t-dead');
  const dead = { kind: 'Workflow', scope: 'global', name: 't-dead-live' };
  const before = pendingItems(ctx).length;
  assert.ok(pendingItems(ctx).some((it) => it.key === draftKey(dead)), '出队前死稿在账');

  dequeueDeadDraft(dead);
  assert.equal(pendingItems(ctx).length, before - 1, '清单除名');
  assert.equal(inboxCountOf(ctx), pendingItems(ctx).length, '计数同步除名（同一来源）');
  assert.ok(!pendingItems(ctx).some((it) => it.key === draftKey(dead)), '草稿不存在的条目不再出现');
});

test('isDeadDraftResult：只认 404 / ErrNoDraft 词形，网络与 5xx 不冒充死稿', () => {
  assert.equal(isDeadDraftResult({ ok: false, status: 404, error: 'draft: 草稿不存在' }), true);
  assert.equal(isDeadDraftResult({ ok: false, status: 0, error: 'draft: 草稿不存在' }), true);
  assert.equal(isDeadDraftResult({ ok: false, status: 500, error: 'internal error' }), false);
  assert.equal(isDeadDraftResult({ ok: false, status: 0, error: '请求超时' }), false);
  assert.equal(isDeadDraftResult({ ok: true, data: {} }), false);
});
