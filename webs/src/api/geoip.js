import request from './request';

// 获取 GeoIP 配置
export const getGeoIPConfig = () => {
  return request({
    url: '/v1/geoip/config',
    method: 'get'
  });
};

// 保存 GeoIP 配置
export const saveGeoIPConfig = (data) => {
  return request({
    url: '/v1/geoip/config',
    method: 'put',
    data
  });
};

// 获取 GeoIP 状态
export const getGeoIPStatus = () => {
  return request({
    url: '/v1/geoip/status',
    method: 'get'
  });
};

// 下载 GeoIP 数据库
export const downloadGeoIP = () => {
  return request({
    url: '/v1/geoip/download',
    method: 'post'
  });
};

// 停止 GeoIP 下载
export const stopGeoIPDownload = () => {
  return request({
    url: '/v1/geoip/stop',
    method: 'post'
  });
};
