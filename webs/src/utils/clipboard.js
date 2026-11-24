export async function copyText(text) {
  if (typeof text !== 'string') return false;

  if (window.isSecureContext && navigator.clipboard?.writeText) {
    try {
      await navigator.clipboard.writeText(text);
      return true;
    } catch {
      // The browser can reject clipboard writes even in a secure context.
    }
  }

  const previousFocus = document.activeElement;
  // Keep the selection inside a dialog so its focus trap does not intercept it.
  const container = previousFocus?.closest('[role="dialog"], [role="alertdialog"]') ?? document.body;
  const input = document.createElement('textarea');
  input.value = text;
  input.readOnly = true;
  input.tabIndex = -1;
  Object.assign(input.style, { position: 'fixed', top: '0', left: '0', opacity: '0', pointerEvents: 'none' });

  try {
    container.appendChild(input);
    input.focus({ preventScroll: true });
    if (document.activeElement !== input) return false;
    input.select();
    return document.execCommand('copy');
  } catch {
    return false;
  } finally {
    input.remove();
    if (document.activeElement !== previousFocus) previousFocus?.focus?.({ preventScroll: true });
  }
}
