import request from './request';

export function getSocks5Settings() {
  return request({
    url: '/v1/proxy-gateway/settings',
    method: 'get'
  });
}

export function updateSocks5Settings(data) {
  return request({
    url: '/v1/proxy-gateway/settings',
    method: 'post',
    data
  });
}

export function stopSocks5() {
  return request({
    url: '/v1/proxy-gateway/stop',
    method: 'post'
  });
}

export function getSocks5Status(signal) {
  return request({ url: '/v1/proxy-gateway/status', method: 'get', signal });
}

export function getSocks5Listeners() {
  return request({ url: '/v1/proxy-gateway/listeners', method: 'get' });
}

export function createSocks5Listener(data) {
  return request({ url: '/v1/proxy-gateway/listeners', method: 'post', data });
}

export function updateSocks5Listener(id, data) {
  return request({ url: `/v1/proxy-gateway/listeners/${encodeURIComponent(id)}`, method: 'put', data });
}

export function deleteSocks5Listener(id) {
  return request({ url: `/v1/proxy-gateway/listeners/${encodeURIComponent(id)}`, method: 'delete' });
}

export function getSocks5Accounts() {
  return request({ url: '/v1/proxy-gateway/accounts', method: 'get' });
}

export function createSocks5Account(data) {
  return request({ url: '/v1/proxy-gateway/accounts', method: 'post', data });
}

export function updateSocks5Account(id, data) {
  return request({ url: `/v1/proxy-gateway/accounts/${encodeURIComponent(id)}`, method: 'put', data });
}

export function deleteSocks5Account(id) {
  return request({ url: `/v1/proxy-gateway/accounts/${encodeURIComponent(id)}`, method: 'delete' });
}

export function getSocks5RoutingProfiles() {
  return request({ url: '/v1/proxy-gateway/routing-profiles', method: 'get' });
}

export function createSocks5RoutingProfile(data) {
  return request({ url: '/v1/proxy-gateway/routing-profiles', method: 'post', data });
}

export function updateSocks5RoutingProfile(id, data) {
  return request({ url: `/v1/proxy-gateway/routing-profiles/${encodeURIComponent(id)}`, method: 'put', data });
}

export function deleteSocks5RoutingProfile(id) {
  return request({ url: `/v1/proxy-gateway/routing-profiles/${encodeURIComponent(id)}`, method: 'delete' });
}

export function getSocks5RoutingSnapshot(params, signal) {
  return request({ url: '/v1/proxy-gateway/node-stats', method: 'get', params, signal });
}

export function resetSocks5RuntimeStats() {
  return request({ url: '/v1/proxy-gateway/node-stats/reset', method: 'post' });
}

export function resetProxyGatewayStats() {
  return request({ url: '/v1/proxy-gateway/stats/reset', method: 'post' });
}

export function getSocks5Connections(signal) {
  return request({ url: '/v1/proxy-gateway/connections', method: 'get', signal });
}

export function closeSocks5Connection(id) {
  return request({ url: `/v1/proxy-gateway/connections/${encodeURIComponent(id)}`, method: 'delete' });
}

export function closeAllSocks5Connections() {
  return request({ url: '/v1/proxy-gateway/connections', method: 'delete' });
}

export function probeSocks5Health(params) {
  return request({ url: '/v1/proxy-gateway/health/probe', method: 'post', params });
}

export function startProxyGateway() {
  return request({ url: '/v1/proxy-gateway/start', method: 'post' });
}
