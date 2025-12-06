import request from './request';

// 获取 Access Key 列表
export function getAccessKeys(userId) {
    return request({
        url: `/v1/accesskey/get/${userId}`,
        method: 'get'
    });
}

// 创建 Access Key
export function createAccessKey(data) {
    return request({
        url: '/v1/accesskey/add',
        method: 'post',
        data
    });
}

// 删除 Access Key
export function deleteAccessKey(id) {
    return request({
        url: `/v1/accesskey/delete/${id}`,
        method: 'delete'
    });
}
