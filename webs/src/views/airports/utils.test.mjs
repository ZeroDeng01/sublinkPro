import assert from 'node:assert/strict';
import { test } from 'node:test';
import { getAirportTestState, parseAirportPriority } from './utils.js';

test('airport sort priority accepts signed integers and uses zero for a blank field', () => {
  assert.equal(parseAirportPriority(''), 0);
  assert.equal(parseAirportPriority('0'), 0);
  assert.equal(parseAirportPriority(' 10 '), 10);
  assert.equal(parseAirportPriority('-1'), -1);
  assert.equal(parseAirportPriority('-2147483648'), -2147483648);
  assert.equal(parseAirportPriority('2147483647'), 2147483647);
});

test('airport sort priority rejects invalid inputs and values outside the API range', () => {
  for (const value of ['1.5', '1e3', '-', 'abc', '-2147483649', '2147483648']) {
    assert.equal(parseAirportPriority(value), null, value);
  }
});

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
