// repo-link.test.mjs —— repoUrl/repoLinkChip 映射锚（N-02：代码 ref 直达仓库地址）。
// 内部仓家族（裸名/pkg://）与对外 github 词形双路；不匹配=空串不编造外链。
import { test } from 'node:test';
import assert from 'node:assert/strict';

globalThis.window = { __CC: {} };
const { repoUrl, repoLinkChip } = await import('../src/render/chips.js');

test('repoUrl：pkg:// 词形 → blob/main 路径', () => {
  assert.equal(repoUrl('pkg://common_connector/notify'),
    'https://github.com/friddle-self-recursive-agent/common_connector/blob/main/notify');
  assert.equal(repoUrl('pkg://org/repo/src/x.go'),
    'https://github.com/org/repo/blob/main/src/x.go');
});

test('repoUrl：git:// 与裸仓名家族 → 默认 org 仓', () => {
  assert.equal(repoUrl('git://friddle-self-recursive-agent/core_system_workservice'),
    'https://github.com/friddle-self-recursive-agent/core_system_workservice');
  assert.equal(repoUrl('core_design/goal.md'),
    'https://github.com/friddle-self-recursive-agent/core_design/blob/main/goal.md');
  assert.equal(repoUrl('core_system_dsh_plugin'),
    'https://github.com/friddle-self-recursive-agent/core_system_dsh_plugin');
});

test('repoUrl：github.com 词形透传 https 化（对外仓可用）', () => {
  assert.equal(repoUrl('github.com/deepseek-ai/deepseek-harness'),
    'https://github.com/deepseek-ai/deepseek-harness');
  assert.equal(repoUrl('github.com/org/repo/docs/a.md#L3'),
    'https://github.com/org/repo/blob/main/docs/a.md#L3');
});

test('repoUrl：不匹配=空串（不编造外链）', () => {
  assert.equal(repoUrl(''), '');
  assert.equal(repoUrl('random-word'), '');
  assert.equal(repoUrl('not-a-repo/path/x.md'), '');
});

test('repoLinkChip：命中=锚 href 外链；未命中=空串', () => {
  const hit = repoLinkChip('core_design', 'core_design');
  assert.match(hit, /href="https:\/\/github\.com\/friddle-self-recursive-agent\/core_design"/);
  assert.equal(repoLinkChip('random-word'), '');
});
