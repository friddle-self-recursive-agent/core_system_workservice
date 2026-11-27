import test from 'node:test';
import assert from 'node:assert/strict';
import { normalizeSessionConfig, deriveStageState } from '../src/screens/session-stages.js';

test('normalizeSessionConfig：缺席字段归一 scheme 缺省 ∧ persistent 词形', () => {
  const m = normalizeSessionConfig(null);
  assert.equal(m.backgroundMaxChars, 5000);
  assert.equal(m.blockAttribution, true);
  assert.equal(m.metaAutoUpdate, true);
  assert.deepEqual(m.metaScope, ['ref', 'tag', 'importance', 'metainfo']);
  assert.equal(m.provider, 'openai-default');
  assert.equal(m.model, 'reasoning-tier');
  const m2 = normalizeSessionConfig({ config: { backgroundMaxChars: 3000, blockAttribution: false, metaScope: ['tag'] }, persistent: false });
  assert.equal(m2.backgroundMaxChars, 3000);
  assert.equal(m2.blockAttribution, false);
  assert.equal(m2.persistent, false);
  assert.deepEqual(m2.metaScope, ['tag']);
});

test('deriveStageState：ctx 缺席=第一段待 ∧ running=第二段进行中 ∧ 回整账=第三段完成', () => {
  const idle = deriveStageState({ ctxReady: false, running: false, writebacks: [] });
  assert.equal(idle[0].done, false);
  assert.equal(idle[0].active, true);
  assert.equal(idle[1].active, false);
  assert.equal(idle[2].done, false);
  const run = deriveStageState({ ctxReady: true, running: true, writebacks: [] });
  assert.equal(run[0].done, true);
  assert.equal(run[1].active, true);
  assert.equal(run[2].done, false);
  const done = deriveStageState({ ctxReady: true, running: false, writebacks: [{ resource: 'r/p.md', importanceDelta: 5, tagsAdded: ['a'] }] });
  assert.equal(done[1].done, true);
  assert.equal(done[2].done, true);
  assert.ok(String(done[2].detail).includes('r/p.md'));
  assert.equal(done[3].wb.resource, 'r/p.md');
});
