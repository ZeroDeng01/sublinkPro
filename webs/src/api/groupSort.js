import request from './request';

// 获取所有分组概要信息
export function getGroupSortGroups() {
  return request({ url: '/v1/group-sort/groups', method: 'get' });
}

// 获取分组详情
export function getGroupSortDetail(group) {
  return request({ url: '/v1/group-sort/detail', method: 'get', params: { group } });
}

// 保存分组内机场排序
export function saveGroupAirportSort(data) {
  return request({ url: '/v1/group-sort/save', method: 'post', data });
}
