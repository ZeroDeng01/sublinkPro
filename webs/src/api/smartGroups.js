import request from './request';

export const getSmartGroups = () => request({ url: '/v1/smart-groups', method: 'get' });
export const createSmartGroup = (data) => request({ url: '/v1/smart-groups', method: 'post', data });
export const updateSmartGroup = (id, data) => request({ url: `/v1/smart-groups/${id}`, method: 'put', data });
export const deleteSmartGroup = (id) => request({ url: `/v1/smart-groups/${id}`, method: 'delete' });
export const getSmartGroupMembers = (id, page = 1) =>
  request({ url: `/v1/smart-groups/${id}/members`, method: 'get', params: { page, pageSize: 30 } });
export const checkSmartGroup = (id, profileId, nodeId = 0) =>
  request({ url: `/v1/smart-groups/${id}/check`, method: 'post', data: { profileId, nodeId } });
