const defaultConfig = {
  enabled: false,
  listenAddress: '127.0.0.1',
  port: 1080,
  username: '',
  hasPassword: false,
  maskedPassword: '',
  nodeId: 0,
  selection: 'best',
  requireAuth: true,
  maxAttempts: 1,
  dialTimeoutSeconds: 30,
  failureCooldownSeconds: 0,
  specificFallback: false,
  maxConnections: 256,
  maxConnectionsPerClient: 32,
  idleTimeoutSeconds: 600,
  maxConnectionDurationSeconds: 0,
  stickySessionEnabled: false,
  stickySessionMode: 'client_ip',
  stickySessionTtlSeconds: 1800,
  healthCheckEnabled: false,
  healthCheckIntervalSeconds: 60,
  healthCheckTimeoutSeconds: 5,
  candidateGroups: [],
  candidateSources: [],
  candidateProtocols: [],
  candidateCountries: []
};

export function settingsDraft(settings) {
  const config = Object.fromEntries(Object.keys(defaultConfig).map((key) => [key, settings[key] ?? defaultConfig[key]]));
  return { ...config, password: '', clearPassword: false };
}

export function settingsPayload(form) {
  return {
    enabled: Boolean(form.enabled),
    listenAddress: form.listenAddress.trim(),
    port: Number(form.port) || 1080,
    username: form.username.trim(),
    password: form.password,
    clearPassword: Boolean(form.clearPassword),
    nodeId: Number(form.nodeId) || 0,
    selection: form.selection,
    requireAuth: Boolean(form.requireAuth),
    maxAttempts: Number(form.maxAttempts) || 1,
    dialTimeoutSeconds: Number(form.dialTimeoutSeconds) || 30,
    failureCooldownSeconds: Number(form.failureCooldownSeconds) || 0,
    specificFallback: Boolean(form.specificFallback),
    maxConnections: Number(form.maxConnections) || 256,
    maxConnectionsPerClient: Number(form.maxConnectionsPerClient) || 32,
    idleTimeoutSeconds: Number(form.idleTimeoutSeconds) || 0,
    maxConnectionDurationSeconds: Number(form.maxConnectionDurationSeconds) || 0,
    stickySessionEnabled: Boolean(form.stickySessionEnabled),
    stickySessionMode: form.stickySessionMode,
    stickySessionTtlSeconds: Number(form.stickySessionTtlSeconds) || 1800,
    healthCheckEnabled: Boolean(form.healthCheckEnabled),
    healthCheckIntervalSeconds: Number(form.healthCheckIntervalSeconds) || 60,
    healthCheckTimeoutSeconds: Number(form.healthCheckTimeoutSeconds) || 5,
    candidateGroups: Array.isArray(form.candidateGroups) ? form.candidateGroups : [],
    candidateSources: Array.isArray(form.candidateSources) ? form.candidateSources : [],
    candidateProtocols: Array.isArray(form.candidateProtocols) ? form.candidateProtocols : [],
    candidateCountries: Array.isArray(form.candidateCountries) ? form.candidateCountries : []
  };
}

export const numericFields = {
  maxConnections: { tab: 'listeners', section: 'limits', min: 1, max: 10000 },
  maxConnectionsPerClient: { tab: 'listeners', section: 'limits', min: 1, max: 10000 },
  idleTimeoutSeconds: { tab: 'listeners', section: 'limits', min: 0, max: 86400 },
  maxConnectionDurationSeconds: { tab: 'listeners', section: 'limits', min: 0, max: 604800 },
  maxAttempts: { tab: 'routing', section: 'advancedRouting', min: 1, max: 5 },
  dialTimeoutSeconds: { tab: 'routing', section: 'advancedRouting', min: 1, max: 120 },
  failureCooldownSeconds: { tab: 'routing', section: 'advancedRouting', min: 0, max: 3600 },
  stickySessionTtlSeconds: { tab: 'routing', section: 'advancedRouting', min: 60, max: 604800 },
  healthCheckIntervalSeconds: { tab: 'monitor', section: 'probeSettings', min: 10, max: 3600 },
  healthCheckTimeoutSeconds: { tab: 'monitor', section: 'probeSettings', min: 1, max: 30 }
};

export function requiresSharedCredentials(settings, listeners = []) {
  return settings.requireAuth || listeners.some((listener) => listener.enabled && listener.requireAuth === true);
}

export function validateSettings(form, listeners) {
  const specific = form.selection === 'specific';
  if (specific && Number(form.nodeId) <= 0) {
    return { tab: 'routing', field: 'nodeId', message: 'form.nodeRequired' };
  }
  if (form.enabled && requiresSharedCredentials(form, listeners)) {
    if (!form.username.trim()) return { tab: 'auth', field: 'username', message: 'page.validation.username' };
    if (form.clearPassword || (!form.password && !form.hasPassword)) {
      return { tab: 'auth', field: form.clearPassword ? 'clearPassword' : 'password', message: 'page.validation.password' };
    }
  }
  for (const field of ['username', 'password']) {
    if (new TextEncoder().encode(form[field]).length > 255) {
      return { tab: 'auth', field, message: 'page.validation.credentialLength' };
    }
  }
  if (form.stickySessionEnabled && form.stickySessionMode === 'username' && !form.requireAuth) {
    return { tab: 'auth', field: 'requireAuth', message: 'form.stickySessionUsernameAuthRequired' };
  }
  for (const [field, { tab, section, min, max }] of Object.entries(numericFields)) {
    const value = Number(form[field]);
    if (form[field] === '' || !Number.isInteger(value) || value < min || value > max) {
      return { tab, section, field, message: 'page.validation.range', params: { min, max } };
    }
  }
  if (Number(form.maxConnectionsPerClient) > Number(form.maxConnections)) {
    return { tab: 'listeners', section: 'limits', field: 'maxConnectionsPerClient', message: 'page.validation.clientLimit' };
  }
  if (specific && form.specificFallback && Number(form.maxAttempts) < 2) {
    return { tab: 'routing', section: 'advancedRouting', field: 'maxAttempts', message: 'page.validation.fallbackAttempts' };
  }
  return null;
}

export function gatewayInputId(field) {
  return `proxy-gateway-${field}`;
}
