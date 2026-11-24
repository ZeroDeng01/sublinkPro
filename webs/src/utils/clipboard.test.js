import assert from 'node:assert/strict';
import { test } from 'node:test';
import { copyText } from './clipboard.js';

function mockBrowser(t, { isSecureContext = false, writeText, inDialog = false, execCommand = () => true } = {}) {
  const dialogSelector = '[role="dialog"], [role="alertdialog"]';
  const body = { appendChild: t.mock.fn() };
  const dialog = { appendChild: t.mock.fn() };
  const previousFocus = {
    closest: t.mock.fn(() => (inDialog ? dialog : null)),
    focus: t.mock.fn(() => {
      document.activeElement = previousFocus;
    })
  };
  const input = {
    style: {},
    focus: () => {
      document.activeElement = input;
    },
    select: t.mock.fn(),
    remove: t.mock.fn(() => {
      document.activeElement = body;
    })
  };

  globalThis.window = { isSecureContext };
  globalThis.document = { activeElement: previousFocus, body, createElement: t.mock.fn(() => input), execCommand: t.mock.fn(execCommand) };
  t.mock.getter(globalThis, 'navigator', () => ({ clipboard: writeText ? { writeText } : undefined }));
  t.after(() => {
    delete globalThis.window;
    delete globalThis.document;
  });

  return { body, dialog, dialogSelector, input, previousFocus };
}

test('uses the Clipboard API in a secure context', async (t) => {
  const writeText = t.mock.fn(async () => {});
  mockBrowser(t, { isSecureContext: true, writeText });

  assert.equal(await copyText('proxy://subscription'), true);
  assert.deepEqual(writeText.mock.calls[0].arguments, ['proxy://subscription']);
  assert.equal(document.createElement.mock.callCount(), 0);
});

test('copies synchronously inside the active dialog over HTTP and restores focus', async (t) => {
  const writeText = t.mock.fn();
  const { body, dialog, dialogSelector, input, previousFocus } = mockBrowser(t, { writeText, inDialog: true });
  const text = '第一行\nsecond line';

  const result = copyText(text);
  assert.equal(document.execCommand.mock.callCount(), 1);
  assert.equal(await result, true);
  assert.equal(writeText.mock.callCount(), 0);
  assert.deepEqual(previousFocus.closest.mock.calls[0].arguments, [dialogSelector]);
  assert.deepEqual(dialog.appendChild.mock.calls[0].arguments, [input]);
  assert.equal(body.appendChild.mock.callCount(), 0);
  assert.equal(input.value, text);
  assert.equal(input.select.mock.callCount(), 1);
  assert.equal(input.remove.mock.callCount(), 1);
  assert.equal(document.activeElement, previousFocus);
});

test('falls back when the Clipboard API rejects or is unavailable in a secure context', async (t) => {
  await t.test('write is rejected', async (t) => {
    const writeText = t.mock.fn(async () => {
      throw new Error('Clipboard write denied');
    });
    const { body, input } = mockBrowser(t, { isSecureContext: true, writeText });

    assert.equal(await copyText('127.0.0.1'), true);
    assert.equal(writeText.mock.callCount(), 1);
    assert.deepEqual(body.appendChild.mock.calls[0].arguments, [input]);
  });

  await t.test('API is unavailable', async (t) => {
    mockBrowser(t, { isSecureContext: true });

    assert.equal(await copyText('127.0.0.1'), true);
    assert.equal(document.execCommand.mock.callCount(), 1);
  });
});

test('reports failure and clears the temporary input if the fallback is denied', async (t) => {
  const { input, previousFocus } = mockBrowser(t, { execCommand: () => false });

  assert.equal(await copyText('test'), false);
  assert.equal(input.remove.mock.callCount(), 1);
  assert.equal(document.activeElement, previousFocus);
});

test('reports failure and clears the temporary input if the fallback throws', async (t) => {
  const { input, previousFocus } = mockBrowser(t, {
    execCommand: () => {
      throw new Error('Copy unsupported');
    }
  });

  assert.equal(await copyText('test'), false);
  assert.equal(input.remove.mock.callCount(), 1);
  assert.equal(document.activeElement, previousFocus);
});
