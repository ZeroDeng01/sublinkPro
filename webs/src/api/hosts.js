import request from './request';

// ========== Host 管理 ==========

// 获取 Host 列表
export const getHosts = () => request.get('/v1/hosts/list');

// 添加 Host
export const addHost = (data) => request.post('/v1/hosts/add', data);

// 更新 Host
export const updateHost = (data) => request.post('/v1/hosts/update', data);

// 删除 Host
export const deleteHost = (id) => request.delete(`/v1/hosts/delete?id=${id}`);

// 批量删除 Host
export const batchDeleteHosts = (ids) => request.delete('/v1/hosts/batch-delete', { data: { ids } });

// 导出 Host 为文本
export const exportHosts = () => request.get('/v1/hosts/export');

// 从文本同步 Host
export const syncHosts = (text) => request.post('/v1/hosts/sync', { text });

// ========== Host 模块设置 ==========

// 获取模块设置
export const getHostSettings = () => request.get('/v1/hosts/settings');

// 更新模块设置
export const updateHostSettings = (data) => request.post('/v1/hosts/settings', data);

// 设置 Host 固定状态
export const pinHost = (id, pinned) => request.post('/v1/hosts/pin', { id, pinned });
