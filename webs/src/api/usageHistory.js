import request from './request';

// 浏览器时区偏移（分钟，与 Date.getTimezoneOffset() 一致），
// 让后端按客户端本地自然日划分「今日」和按天统计。
const clientTimezoneOffset = () => new Date().getTimezoneOffset();

// 批量获取机场用量迷你图（列表卡片内联展示）
// @param {number[]} ids - 机场ID列表
// @param {object} params - { days, granularity }
export function getAirportUsageSparklines(ids, params = {}) {
  const normalizedIds = (Array.isArray(ids) ? ids : [ids]).filter((id) => Number(id) > 0);
  return request({
    url: '/v1/airports/usage-sparklines',
    method: 'get',
    params: { ids: normalizedIds.join(','), tzOffset: clientTimezoneOffset(), ...params }
  });
}

// 获取按天统计的流量消耗；不传 airportId 时为全部机场汇总
// @param {object} params - { days, airportId }
export function getAirportUsageDaily(params = {}) {
  return request({
    url: '/v1/airports/usage-daily',
    method: 'get',
    params: { tzOffset: clientTimezoneOffset(), ...params }
  });
}

// 获取单个机场的用量历史（图表数据）
export function getAirportUsageHistory(id, params = {}) {
  return request({
    url: `/v1/airports/${id}/usage-history`,
    method: 'get',
    params
  });
}

// 获取用量历史保留配置
export function getUsageHistoryConfig() {
  return request({
    url: '/v1/settings/usage-history',
    method: 'get'
  });
}

// 更新用量历史保留配置
export function updateUsageHistoryConfig(data) {
  return request({
    url: '/v1/settings/usage-history',
    method: 'post',
    data
  });
}
