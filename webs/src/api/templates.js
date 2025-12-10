import request from './request';

// 获取模板列表（支持分页参数）
// params: { page, pageSize }
// 带page/pageSize时返回 { items, total, page, pageSize, totalPages }
export function getTemplates(params = {}) {
  return request({
    url: '/v1/template/get',
    method: 'get',
    params
  });
}

// 添加模板
export function addTemplate(data) {
  const formData = new FormData();
  Object.keys(data).forEach((key) => {
    if (data[key] !== undefined && data[key] !== null) {
      formData.append(key, data[key]);
    }
  });
  return request({
    url: '/v1/template/add',
    method: 'post',
    data: formData,
    headers: {
      'Content-Type': 'multipart/form-data'
    }
  });
}

// 更新模板
export function updateTemplate(data) {
  const formData = new FormData();
  Object.keys(data).forEach((key) => {
    if (data[key] !== undefined && data[key] !== null) {
      formData.append(key, data[key]);
    }
  });
  return request({
    url: '/v1/template/update',
    method: 'post',
    data: formData,
    headers: {
      'Content-Type': 'multipart/form-data'
    }
  });
}

// 删除模板
export function deleteTemplate(data) {
  const formData = new FormData();
  Object.keys(data).forEach((key) => {
    if (data[key] !== undefined && data[key] !== null) {
      formData.append(key, data[key]);
    }
  });
  return request({
    url: '/v1/template/delete',
    method: 'post',
    data: formData,
    headers: {
      'Content-Type': 'multipart/form-data'
    }
  });
}

// 获取 ACL4SSR 规则预设列表
export function getACL4SSRPresets() {
  return request({
    url: '/v1/template/presets',
    method: 'get'
  });
}

// 转换规则
export function convertRules(data) {
  return request({
    url: '/v1/template/convert',
    method: 'post',
    data
  });
}
