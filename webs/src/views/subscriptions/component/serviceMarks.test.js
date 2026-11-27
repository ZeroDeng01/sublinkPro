import assert from 'node:assert/strict';
import { test } from 'node:test';
import {
  parseNodeRenameRule,
  previewNodeRenameItems,
  previewSubscriptionNodeName,
  updateServiceMarkAffixes
} from './nodeRenameRuleUtils.js';
import { formatServiceMarks, renderServiceMarks, serviceMarksError } from './serviceMarks.js';

const marks = {
  prefix: ' [',
  suffix: ']',
  services: [
    { provider: 'youtube_premium', label: 'YT', acceptedStatuses: ['available'] },
    { provider: 'openai', label: 'A/"未送中(字面量)" $Name', acceptedStatuses: ['available', 'partial'], region: true }
  ]
};

test('the builder keeps an escaped block as one reorderable item', () => {
  const block = formatServiceMarks(marks);
  const rule = `$Name${block}$DuplicateIndex`;
  const items = parseNodeRenameRule(rule);
  assert.deepEqual(
    items.map(({ type }) => type),
    ['variable', 'serviceMarks', 'variable']
  );
  assert.deepEqual(items[1].marks, marks);
  assert.equal(items.map(({ value }) => value).join(''), rule);
  assert.equal(previewSubscriptionNodeName(rule), 'HK-Node-Remark [YT · A/"未送中(字面量)" $Name-US]1');
});

test('legacy variable arguments and literal text do not open a mark block', () => {
  const tag = `$TagGroup(team${formatServiceMarks({ services: [marks.services[0]] })}`;
  const rule = `$Name${tag}$ServiceMarks(v1)${formatServiceMarks(marks)}`;
  const items = parseNodeRenameRule(rule);
  assert.equal(items[1].value, tag);
  assert.equal(items.filter((item) => item.type === 'serviceMarks').length, 1);
  assert.equal(items.map(({ value }) => value).join(''), rule);
});

test('both editors show the same legacy and mark values', () => {
  const rule = `$Name $UnlockRegion ${formatServiceMarks(marks)} $UnlockStatus/$UnlockLabel/$Unlock`;
  assert.equal(previewNodeRenameItems(parseNodeRenameRule(rule)), previewSubscriptionNodeName(rule));
  assert.equal(previewSubscriptionNodeName(rule), 'HK-Node-Remark US  [YT · A/"未送中(字面量)" $Name-US] available/解锁/Netflix-US');
});

test('moving or deleting a neighbor updates only default mark affixes', () => {
  const block = formatServiceMarks({ ...marks, prefix: ' · ', suffix: '' });
  const items = parseNodeRenameRule(`$Name${block}`);
  const moved = updateServiceMarkAffixes(items, [items[1], items[0]]);
  assert.equal(moved[0].marks.prefix, '');
  assert.equal(moved[0].marks.suffix, ' · ');
  assert.equal(updateServiceMarkAffixes(moved, [moved[0]])[0].marks.suffix, '');

  const configured = parseNodeRenameRule(`$Name${formatServiceMarks(marks)}`);
  const movedConfigured = updateServiceMarkAffixes(configured, [configured[1], configured[0]]);
  assert.equal(movedConfigured[0].value, configured[1].value);
});

test('preview marks use service order and omit the wrapper if none match', () => {
  const results = new Map([
    ['openai', { status: 'partial', region: 'US' }],
    ['youtube_premium', { status: 'available', region: 'JP' }]
  ]);
  assert.equal(renderServiceMarks(marks, results), ' [YT · A/"未送中(字面量)" $Name-US]');
  results.delete('youtube_premium');
  assert.equal(renderServiceMarks(marks, results), ' [A/"未送中(字面量)" $Name-US]');
  results.set('openai', { status: 'reachable' });
  assert.equal(renderServiceMarks(marks, results), '');
});

test('a status is required and config delimiters are rejected in labels', () => {
  const statuses = ['available', 'partial'];
  assert.equal(serviceMarksError(marks, statuses), null);
  assert.equal(serviceMarksError({ services: [{ provider: 'openai', label: 'GPT', acceptedStatuses: [] }] }, statuses).reason, 'statuses');
  assert.equal(
    serviceMarksError({ services: [{ provider: 'openai', label: 'GPT=x', acceptedStatuses: statuses }] }, statuses).reason,
    'text'
  );
});
