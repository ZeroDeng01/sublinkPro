import request from './request';

// ========== 标签管理 ==========

// 获取标签列表
export const getTags = () => request.get('/v1/tags/list');

// 添加标签
export const addTag = (data) => request.post('/v1/tags/add', data);

// 更新标签
export const updateTag = (data) => request.post('/v1/tags/update', data);

// 删除标签
export const deleteTag = (name) => request.delete(`/v1/tags/delete?name=${encodeURIComponent(name)}`);

// 获取标签组列表
export const getTagGroups = () => request.get('/v1/tags/groups');

// ========== 标签规则管理 ==========

// 获取规则列表
export const getTagRules = () => request.get('/v1/tags/rules');

// 添加规则
export const addTagRule = (data) => request.post('/v1/tags/rules/add', data);

// 更新规则
export const updateTagRule = (data) => request.post('/v1/tags/rules/update', data);

// 删除规则
export const deleteTagRule = (id) => request.delete(`/v1/tags/rules/delete?id=${id}`);

// 手动触发规则
export const triggerTagRule = (id) => request.post(`/v1/tags/rules/trigger?id=${id}`);

// ========== 节点标签操作 ==========

// 给节点添加标签
export const addNodeTag = (data) => request.post('/v1/tags/node/add', data);

// 从节点移除标签
export const removeNodeTag = (data) => request.post('/v1/tags/node/remove', data);

// 批量给节点添加标签
export const batchAddNodeTag = (data) => request.post('/v1/tags/node/batch-add', data);

// 批量设置节点标签（覆盖模式）
export const batchSetNodeTags = (data) => request.post('/v1/tags/node/batch-set', data);

// 批量从节点移除指定标签
export const batchRemoveNodeTags = (data) => request.post('/v1/tags/node/batch-remove', data);

// 获取节点的标签
export const getNodeTags = (nodeId) => request.get(`/v1/tags/node/tags?nodeId=${nodeId}`);
