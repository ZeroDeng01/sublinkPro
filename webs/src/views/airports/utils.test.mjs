import assert from 'node:assert/strict';
import { test } from 'node:test';
import { getAirportTestState } from './utils.js';

test('airport checks distinguish recorded failures from never tested nodes', () => {
  assert.equal(getAirportTestState({ hasTestedNodes: true, delayPassCount: 0, speedPassCount: 0 }, 2), 'tested');
  assert.equal(getAirportTestState({ hasTestedNodes: false, delayPassCount: 0, speedPassCount: 0 }, 2), 'untested');
  assert.equal(getAirportTestState(undefined, 2), 'untested');
});

test('zero nodes have a separate empty state', () => {
  assert.equal(getAirportTestState({ hasTestedNodes: false, delayPassCount: 0, speedPassCount: 0 }, 0), 'empty');
});

test('latency only, partial success and older successful responses retain pass counts', () => {
  assert.equal(getAirportTestState({ hasTestedNodes: true, delayPassCount: 1, speedPassCount: 0 }, 4), 'tested');
  assert.equal(getAirportTestState({ delayPassCount: 42, speedPassCount: 0 }, 45), 'tested');
  assert.equal(getAirportTestState({ delayPassCount: 0, speedPassCount: 1 }, 2), 'tested');
});
