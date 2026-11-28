/* concept-visibility.test.mjs —— 概念可见性 V 批（V-01..V-22）全量机械回归锚（纯函数零 DOM）：
 *   V-01/16: tagRoute / docs/tag/:tag & tags 路由 / 三源去重 aggregateTags / tagGroups 真词表
 *   V-03:    depRoute / depChips（Task/Workflow 依赖深链）
 *   V-07:    syncBadge Running→wf/:name/r/:n 深链
 *   V-08:    REF_KINDS 补 Runbook/DriftReport/DocReview + renderMessage 死代码已删
 *   V-09:    PHASE_VIEW 补 pending 人话词形
 *   V-10:    sessBacklinksHtml（session→Task/Event 反向回链 + 空态）
 *   V-13/14: approvals/:focus 定位匹配 + prop/:name 路由
 *   V-17:    docHistoryRowsHtml 真账行 + api.docHistory/revContent 路径契约
 *   V-18:    renderAuditPayload 审核结果落面（XSS 转义+unbound+空态）+ audit 路由
 *   V-19/20: DOC_KINDS 含 BugDoc + templatesSectionHtml 真对象替换静态假卡
 *   V-21:    deliveriesPanelHtml 送达投影（accepted/sent/failed/attempts+event 深链）
 *   V-22:    i18n 幽灵词条「补标」已销 + 「未分组/代码审核」生效。 */

import { test } from 'node:test';
import assert from 'node:assert/strict';

import { createRouter, parseHash } from '../src/core/router.js';
import { createApi } from '../src/core/api.js';
import { setLocale, t as tr } from '../src/core/i18n.js';
import { tagRoute, syncBadge } from '../src/render/chips.js';
import * as refsMod from '../src/render/refs.js';
import { depRoute, depChips, phaseChip, groupOf, statusOf, taskPhaseOf } from '../src/screens/tasks.js';
import { tagGroups, tagVocabFromKinds, tagsOfDoc } from '../src/screens/objects.js';
import { frontmatterTags, tagsOfRef, aggregateTags, groupByType } from '../src/screens/tags.js';
import { sessBacklinksHtml } from '../src/screens/chat.js';
import { matchesApprovalFocus, verdictDocNames, verdictAnchorOf, apVerdictKeyOf } from '../src/screens/approvals.js';
import { docHistoryRowsHtml } from '../src/screens/editor.js';
import { renderAuditPayload } from '../src/screens/audit.js';
import { templatesSectionHtml } from '../src/screens/workflows.js';
import { DOC_KINDS, deliveriesPanelHtml } from '../src/screens/misc.js';

test('V-01/14/16/18 路由词形：docs/tag/:tag、tags、prop/:name、approvals/:focus、audit', () => {
  const router = createRouter();
  ['docs', 'docs/tag/:tag', 'tags', 'proposals', 'proposals/:name', 'prop/:name', 'approvals', 'approvals/:focus', 'audit']
    .forEach((p) => router.route(p, () => {}));

  const h1 = router.resolve(parseHash('#/docs/tag/console'));
  assert.equal(h1.r.pattern, 'docs/tag/:tag');
  assert.equal(h1.params.tag, 'console');

  const h2 = router.resolve(parseHash('#/tags'));
  assert.equal(h2.r.pattern, 'tags');

  const h3 = router.resolve(parseHash('#/prop/my-prop'));
  assert.equal(h3.r.pattern, 'prop/:name');
  assert.equal(h3.params.name, 'my-prop');

  const h4 = router.resolve(parseHash('#/approvals/Goal%2Fglobal%2Fnorth-star'));
  assert.equal(h4.r.pattern, 'approvals/:focus');
  assert.equal(h4.params.focus, 'Goal/global/north-star');

  const h5 = router.resolve(parseHash('#/audit'));
  assert.equal(h5.r.pattern, 'audit');

  assert.equal(tagRoute('domain:core'), 'docs/tag/domain%3Acore');
});

test('V-03 依赖深链：depRoute/depChips 覆盖 Task/Workflow 大小写与简写，未知词形不编造', () => {
  assert.equal(depRoute('Task/self-recursive-core/t1'), 'task/self-recursive-core%2Ft1');
  assert.equal(depRoute('task/global/upgrade-base'), 'task/global%2Fupgrade-base');
  assert.equal(depRoute('Workflow/global/dev-delivery'), 'wf/dev-delivery');
  assert.equal(depRoute('workflow/global/core-workflow'), 'wf/core-workflow');
  assert.equal(depRoute('global/t2', 'task'), 'task/global%2Ft2');
  assert.equal(depRoute('core-workflow', 'workflow'), 'wf/core-workflow');
  assert.equal(depRoute('??? invalid ref ???'), '');
  const html = depChips(['Task/global/a', 'unknown thing']);
  assert.ok(html.includes('data-go="task/global%2Fa"'));
  assert.ok(!html.includes('data-go="unknown'));
});

test('V-07 syncBadge：活跃 run 返回可点 wf/:name/r/:n；终态/空=已同步（R-UX-3 单一语义词形）', () => {
  const running = syncBadge([{ workflow: 'dev-delivery', run: 7, status: 'running', involved: 'Goal/global/g1' }], 'g1');
  assert.ok(running.includes('data-go="wf/dev-delivery/r/7"'));
  assert.ok(running.includes('同步中'));
  const done = syncBadge([{ workflow: 'dev-delivery', run: 7, status: 'done', involved: 'Goal/global/g1' }], 'g1');
  assert.ok(done.includes('已同步'));
  assert.ok(!done.includes('data-go='));
});

test('V-08 refs.js：REF_KINDS 含 Runbook/DriftReport/DocReview 且死代码 renderMessage 已删', () => {
  assert.equal(refsMod.renderMessage, undefined, '死代码 renderMessage 必须已清理');
  const html = refsMod.linkifyRefs('见 Runbook/global/rb1 与 DriftReport/core/dr1 及 DocReview/core/rev1');
  assert.ok(html.includes('data-go="obj/Runbook/global/rb1"'));
  assert.ok(html.includes('data-go="obj/DriftReport/core/dr1"'));
  assert.ok(html.includes('data-go="obj/DocReview/core/rev1"'));
});

test('V-09 PHASE_VIEW：pending 人话词形=等待拉起 且归排队组', () => {
  const html = phaseChip({ status: { phase: 'pending' } });
  assert.ok(html.includes('等待拉起'));
  assert.equal(groupOf({ status: { phase: 'pending' } }), 'queued');
});

test('F-05 任务扁平字段：statusOf 兼容 /v1/tasks 简表 t.phase 顶层词形', () => {
  const flat = { scope: 'self-recursive-core', name: 'console-audit-20261126', phase: 'running', lastAction: '拉起中（spawn）', attempts: 2 };
  assert.equal(statusOf(flat).phase, 'running');
  assert.equal(statusOf(flat).lastAction, '拉起中（spawn）');
  assert.equal(groupOf(flat), 'doing');
  assert.ok(phaseChip(flat).includes('正在做'));
  // 混合词形：t.status 存在但缺 phase 键 → 顶层 phase 兜底
  const hybrid = { name: 'x', status: { lastAction: '写面' }, phase: 'verifying' };
  assert.equal(statusOf(hybrid).phase, 'verifying');
  assert.equal(statusOf(hybrid).lastAction, '写面');
  // 空守卫：null/undefined 不抛
  assert.equal(statusOf(null).phase, undefined);
  assert.equal(groupOf([]), 'queued');
});

test('V-10 sessBacklinksHtml：按 agentSession 反向回链 Task/Event + 无关联诚实空态', () => {
  const store = {
    tasks: () => [{ scope: 'self-recursive-core', name: 't1', spec: { title: '概念整改' }, status: { phase: 'running', agentSession: 'sess-42' } }],
    events: () => [{ id: 99, type: 'TaskStateChanged', agentSession: 'sess-42' }],
  };
  const hit = sessBacklinksHtml(store, 'sess-42');
  assert.ok(hit.includes('data-go="task/self-recursive-core%2Ft1"'));
  assert.ok(hit.includes('data-go="event/99"'));
  const miss = sessBacklinksHtml(store, 'sess-none');
  assert.ok(miss.includes('无在册关联'));
  assert.ok(!miss.includes('data-go='));
});

test('V-13/V-14 matchesApprovalFocus：draft 对象键/name 与 ApprovalNeeded 事件精准定位', () => {
  const draftItem = { kind: 'draft', draft: { kind: 'Goal', scope: 'global', name: 'north-star' } };
  assert.equal(matchesApprovalFocus(draftItem, 'Goal/global/north-star'), true);
  assert.equal(matchesApprovalFocus(draftItem, 'north-star'), true);
  assert.equal(matchesApprovalFocus(draftItem, 'other-doc'), false);

  const evItem = { kind: 'event', ev: { involved: 'Proposal/global/p-1', dedupKey: 'proposal:p-1' } };
  assert.equal(matchesApprovalFocus(evItem, 'p-1'), true);
  assert.equal(matchesApprovalFocus(evItem, 'p-2'), false);
});

test('V-16 tags 三源严格去重聚合 + frontmatter 双形态 + tagGroups 真词表', () => {
  assert.deepEqual(frontmatterTags('---\ntags: [console, "ux"]\n---\n'), ['console', 'ux']);
  assert.deepEqual(frontmatterTags('---\ntags:\n  - alpha\n  - beta\n---\n'), ['alpha', 'beta']);
  assert.deepEqual(tagsOfDoc('---\nname: t1\nmetadata:\n  tags: [console, ux]\n---\n# Heading\nbody\n'), ['console', 'ux']);
  assert.deepEqual(tagsOfDoc('---\ntags: [{tag: core, tag_type: domain}]\n---\n# Body\n'), ['domain:core']);
  assert.deepEqual(tagsOfRef({ raw: { spec: { metadata: { tags: ['domain:core'] } } } }), ['domain:core']);

  // 同一对象 Task/s/t1 在三源均出现 tag "console" → 计数必须=1（严格按对象键去重）
  const indexDocs = [
    { kind: 'Task', scope: 's', name: 't1', tags: ['console', 'domain:core'] },
    { kind: 'Goal', scope: 's', name: 'g1', tags: ['console'] },
  ];
  const objs = [
    { kind: 'Task', scope: 's', name: 't1', rv: 1, raw: { metadata: { tags: ['console'] } } },
    { kind: 'Rule', scope: 's', name: 'r1', rv: 1 },
  ];
  const cached = (k, s, n) => (k === 'Task' ? { content: '---\ntags: [console]\n---' } : { content: '---\ntags: [ux]\n---' });
  const vocab = tagVocabFromKinds([{ spec: { metadata: { tagTypes: { domain: '领域' } } } }]);
  const agg = aggregateTags(indexDocs, objs, cached, vocab);
  const byTag = new Map(agg.map((x) => [x.tag, x]));
  assert.equal(byTag.get('console').n, 2, 'Task/s/t1 三源命中只计 1 次 + Goal/s/g1 = 2');
  assert.equal(byTag.get('domain:core').type, '领域');
  assert.equal(byTag.get('ux').n, 1);

  const grouped = groupByType(agg);
  assert.equal(grouped[0].type, '领域');
  assert.equal(grouped[grouped.length - 1].type, '未分组');

  const tg = tagGroups(['console', 'domain:core'], vocab);
  assert.equal(tg[0].type, '领域');
  assert.equal(tg[1].type, '未分组');
});

test('V-17 editor 历史真账：docHistoryRowsHtml 渲染 + createApi 端点路径', async () => {
  const html = docHistoryRowsHtml([{ rev: 5, at: '2026-11-26T14:59:26Z', fieldManager: 'task-scheduler', bytes: 3280 }]);
  assert.ok(html.includes('r5'));
  assert.ok(html.includes('task-scheduler'));
  assert.ok(html.includes('3280 B'));
  assert.ok(html.includes('data-rev-peek="5"'));
  assert.ok(docHistoryRowsHtml([]).includes('doc_history 暂无历史翻页行'));

  const seen = [];
  const origFetch = globalThis.fetch;
  globalThis.fetch = async (url) => {
    seen.push(String(url));
    return { ok: true, status: 200, text: async () => '{}' };
  };
  try {
    const api = createApi('/core-console/api');
    await api.docHistory('Task', 'self-recursive-core', 't1');
    await api.revContent('Task', 'self-recursive-core', 't1', 3);
    await api.codeAuditFindings();
    await api.deliveries();
    assert.equal(seen[0], '/core-console/api/v1/doc-history/Task/t1?scope=self-recursive-core');
    assert.equal(seen[1], '/core-console/api/v1/rev-content/Task/t1?scope=self-recursive-core&rv=3');
    assert.equal(seen[2], '/core-console/api/v1/code-audit/findings');
    assert.equal(seen[3], '/core-console/api/v1/notifications/deliveries');
  } finally {
    globalThis.fetch = origFetch;
  }
});

test('V-18 audit.js：renderAuditPayload 真实 findings/unbound/skipped + XSS 转义 + 诚实空态', () => {
  assert.ok(renderAuditPayload(null).includes('未装配'));
  assert.ok(renderAuditPayload({ count: 0, audits: [] }).includes('暂无代码审核记录'));
  const html = renderAuditPayload({
    source: 'connectops-code-audit',
    count: 1,
    audits: [{
      auditId: 'audit-1',
      ok: false,
      ranAt: '2026-11-26T15:00:00Z',
      tool: { name: 'connectops-code-audit', version: 'v0.1.0' },
      findings: [{ rule: 'R-PUB-1', severity: 'error', file: '<script>x</script>', line: 12, message: 'hit', checkRef: 'lint/personal-info.sh' }],
      unbound: [{ rule: 'R-X', checkRef: 'missing.sh', reason: '未注册' }],
      skipped: [{ rule: 'R-ENG-8', mode: 'agent', reason: 'agent 路由' }],
    }],
  });
  assert.ok(html.includes('audit-1'));
  assert.ok(html.includes('R-PUB-1'));
  assert.ok(html.includes('unbound'));
  assert.ok(html.includes('R-ENG-8'));
  assert.ok(!html.includes('<script>x</script>'), '动态字段必须转义');
});

test('V-19/V-20 BugDoc 入 DOC_KINDS + Template 静态假卡退役为真对象列表', () => {
  assert.ok(DOC_KINDS.includes('BugDoc'));
  const emptyCtx = { store: { refs: () => [], runs: () => [] } };
  const emptyHtml = templatesSectionHtml(emptyCtx);
  assert.ok(emptyHtml.includes('暂无流程模板')); // R-UX-3 黑话清零：假卡退役词形改人话
  assert.ok(!emptyHtml.includes('实例化'), '不得残留 disabled 假实例化按钮');

  const realCtx = {
    store: {
      refs: (k) => k === 'Template'
        ? [{ kind: 'Template', scope: 'global', name: 'dev-tpl', rv: 2 }]
        : (k === 'Workflow' ? [{ kind: 'Workflow', scope: 'global', name: 'bug-lifecycle', rv: 1 }] : []),
      runs: () => [{ workflow: 'bug-lifecycle', run: 1, status: 'done' }],
    },
  };
  const realHtml = templatesSectionHtml(realCtx);
  assert.ok(realHtml.includes('dev-tpl'));
  assert.ok(realHtml.includes('data-go="obj/Template/global/dev-tpl"'));
  assert.ok(realHtml.includes('data-go="wf/bug-lifecycle"'));
});

test('V-21 deliveriesPanelHtml：accepted/sent/failed/attempts 与 event 深链', () => {
  assert.ok(deliveriesPanelHtml(null).includes('未就绪'));
  const html = deliveriesPanelHtml({
    count: 1, accepted: 1, sent: 0, failed: 0,
    deliveries: [{ status: 'accepted', type: 'ApprovalNeeded', channel: '*', attempts: 1, eventId: 'evt-1', dedupKey: 'k1' }],
  });
  assert.ok(html.includes('accepted'));
  assert.ok(html.includes('attempts: 1'));
  assert.ok(html.includes('data-go="event/evt-1"'));
});

test('V-22 i18n：幽灵词条「补标」已销，「未分组/代码审核」在册', () => {
  setLocale('en');
  try {
    assert.equal(tr('补标'), '补标', '幽灵词条补标应已删除（回退原词）');
    assert.equal(tr('未分组'), 'Ungrouped');
    assert.equal(tr('代码审核'), 'Code Audit');
  } finally {
    setLocale('zh');
  }
});

test('taskPhaseOf：TaskStateChanged dedupKey 词尾=相位词；非 task: 前缀=空（不冒充相位词）', () => {
  assert.equal(taskPhaseOf({ dedupKey: 'task:self-recursive-core/align-docs-reality-20261127:running' }), 'running');
  assert.equal(taskPhaseOf({ dedupKey: 'task:t/a:done' }), 'done');
  assert.equal(taskPhaseOf({ dedupKey: 'service-session:wf:1:n' }), '');
  assert.equal(taskPhaseOf({}), '');
  assert.equal(taskPhaseOf(null), '');
});

test('verdictDocNames：per-run 名优先+旧名回退；run 缺席=仅旧名（稿批准卡路径）', () => {
  assert.deepEqual(
    verdictDocNames({ kind: 'Workflow', name: 'gh-issue-trigger', run: 83 }),
    ['docdiff-workflow-gh-issue-trigger-run83', 'docdiff-workflow-gh-issue-trigger'],
  );
  assert.deepEqual(
    verdictDocNames({ kind: 'Workflow', name: 'core-system-developer-loop' }),
    ['docdiff-workflow-core-system-developer-loop'],
  );
  assert.deepEqual(
    verdictDocNames({ kind: 'Workflow', name: 'gh_issue.trigger', run: 0 }),
    ['docdiff-workflow-gh-issue-trigger'],
  );
});

test('verdictAnchorOf：involved 三段键=实际挂起对象（非卡头编排器名）；缺席回退卡头', () => {
  assert.deepEqual(
    verdictAnchorOf({ workflow: 'core-workflow', owner: 'global', run: 83, involved: 'Workflow/global/gh-issue-trigger' }),
    { kind: 'Workflow', scope: 'global', name: 'gh-issue-trigger' },
  );
  assert.deepEqual(
    verdictAnchorOf({ workflow: 'wf-x', owner: 'global', involved: 'Workflow/global/wf-x,extra' }),
    { kind: 'Workflow', scope: 'global', name: 'wf-x' },
  );
  assert.deepEqual(
    verdictAnchorOf({ workflow: 'wf-x', owner: 'global' }),
    { kind: 'Workflow', scope: 'global', name: 'wf-x' },
  );
});

test('apVerdictKeyOf：per-run 唯一（同对象多 run 不撞首卡）；稿卡无 run=退化为对象键', () => {
  const a = verdictAnchorOf({ involved: 'Workflow/global/gh-issue-trigger' });
  assert.equal(apVerdictKeyOf({ run: 83 }, a), 'Workflow_global_gh_issue_trigger-run83');
  assert.equal(apVerdictKeyOf({ run: 84 }, a), 'Workflow_global_gh_issue_trigger-run84');
  assert.equal(apVerdictKeyOf({ run: null }, a), 'Workflow_global_gh_issue_trigger');
});
