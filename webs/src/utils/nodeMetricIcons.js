export const NODE_TEST_STATUS = {
  UNTESTED: 'untested',
  SUCCESS: 'success',
  TIMEOUT: 'timeout',
  ERROR: 'error'
};

export const getSpeedIcon = (speed, speedStatus) => {
  if (speedStatus === NODE_TEST_STATUS.TIMEOUT) {
    return '⏱️';
  }
  if (speedStatus === NODE_TEST_STATUS.ERROR || speed === -1) {
    return '❌';
  }
  if (speedStatus === NODE_TEST_STATUS.UNTESTED || (!speedStatus && speed <= 0)) {
    return '⛔️';
  }
  if (speed >= 5) {
    return '🟢';
  }
  if (speed >= 1) {
    return '🟡';
  }
  return '🔴';
};

export const getDelayIcon = (delay, delayStatus) => {
  if (delayStatus === NODE_TEST_STATUS.TIMEOUT || delay === -1) {
    return '⏱️';
  }
  if (delayStatus === NODE_TEST_STATUS.ERROR) {
    return '❌';
  }
  if (delayStatus === NODE_TEST_STATUS.UNTESTED || (!delayStatus && delay <= 0)) {
    return '⛔️';
  }
  if (delay < 200) {
    return '🟢';
  }
  if (delay < 500) {
    return '🟡';
  }
  return '🔴';
};
