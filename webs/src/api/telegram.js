import request from './request';

// 获取 Telegram 配置
export const getTelegramConfig = () => {
  return request.get('/v1/settings/telegram');
};

// 保存 Telegram 配置
export const saveTelegramConfig = (data) => {
  return request.post('/v1/settings/telegram', data);
};

// 测试 Telegram 连接
export const testTelegramConnection = (data) => {
  return request.post('/v1/settings/telegram/test', data);
};

// 获取 Telegram 状态
export const getTelegramStatus = () => {
  return request.get('/v1/settings/telegram/status');
};

// 重新连接 Telegram
export const reconnectTelegram = () => {
  return request.post('/v1/settings/telegram/reconnect');
};
