import assert from 'node:assert/strict';
import test from 'node:test';
import { buildShareUpdateData, buildShareBatchUpdates, resolveSelectedShares, runShareBatchUpdates } from './shareBatchUpdate.js';

test('editing a device limit omits the original token while explicit token edits are preserved', () => {
  const share = { id: 1, token: 'old-token' };
  assert.deepEqual(buildShareUpdateData({ token: 'old-token', max_devices: 2 }, share, 3), { id: 1, subscription_id: 3, max_devices: 2 });
  assert.equal(buildShareUpdateData({ token: 'custom-token' }, share, 3).token, 'custom-token');
});

test('keeping a batch setting unchanged sends no update; explicit zero only changes the device limit', () => {
  for (const [mode, data] of [
    ['devices', {}],
    ['expire', { expireType: 'keep' }],
    ['enabled', { action: 'keep' }]
  ]) {
    assert.deepEqual(buildShareBatchUpdates(mode, data, [1], []), []);
  }
  assert.deepEqual(buildShareBatchUpdates('devices', { maxDevices: 0, token: 'stale', enabled: true }, [1, 2], []), [
    { ids: [1, 2], updates: { max_devices: 0 } }
  ]);
  assert.deepEqual(buildShareBatchUpdates('expire', { expireType: 0, expireDays: 30, expireAt: 'stale' }, [1], []), [
    { ids: [1], updates: { expire_type: 0 } }
  ]);
});

test('a partial toggle failure preserves item errors and still updates the other group', async () => {
  const batches = buildShareBatchUpdates(
    'enabled',
    { action: 'toggle' },
    [1, 2, 3],
    [
      { id: 1, enabled: false },
      { id: 2, enabled: false },
      { id: 3, enabled: true }
    ]
  );
  const calls = [];
  const results = await runShareBatchUpdates(batches, async (ids, updates) => {
    calls.push({ ids, updates });
    if (updates.enabled) {
      const error = new Error('partial failure');
      error.response = {
        data: {
          results: [
            { id: 1, success: true },
            { id: 2, success: false, error: 'share unavailable' }
          ]
        }
      };
      throw error;
    }
    return { results: [{ id: 3, success: true }] };
  });
  assert.deepEqual(calls, [
    { ids: [1, 2], updates: { enabled: true } },
    { ids: [3], updates: { enabled: false } }
  ]);
  assert.deepEqual(results, [
    { id: 1, success: true },
    { id: 2, success: false, error: 'share unavailable' },
    { id: 3, success: true }
  ]);
});

test('legacy responses and request failures retain per-item outcomes', async () => {
  const results = await runShareBatchUpdates(
    [
      { ids: [1], updates: { enabled: true } },
      { ids: [2], updates: { enabled: false } }
    ],
    async (ids) => {
      if (ids[0] === 2) throw new Error('network failed');
      return { code: 200 };
    }
  );
  assert.deepEqual(results, [
    { id: 1, success: true },
    { id: 2, success: false, error: 'network failed' }
  ]);
});

test('failed shares from later pages remain available for retry after refreshing the first page', () => {
  const retained = [
    { id: 1, name: 'first page', enabled: false },
    { id: 51, name: 'later page', enabled: false }
  ];
  const refreshed = [{ id: 1, name: 'updated first page', enabled: true }];
  const selected = resolveSelectedShares([1, 51], refreshed, retained);
  assert.deepEqual(selected, [refreshed[0], retained[1]]);
  assert.deepEqual(resolveSelectedShares([51], refreshed, retained), [retained[1]]);
  assert.deepEqual(buildShareBatchUpdates('enabled', { action: 'toggle' }, [1, 51], selected), [
    { ids: [51], updates: { enabled: true } },
    { ids: [1], updates: { enabled: false } }
  ]);
});
