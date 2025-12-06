import request from "@/utils/request";

export interface APIKey {
  id: number;
  userID: number;
  username: string;
  createdAt: string;
  expiredAt: string | null;
  description: string;
}

export interface CreateAPIKeyParams {
  description: string;
  expiredAt?: string;
  username?: string;
}

/**
 * 获取用户API密钥列表
 */
export function getAPIKeys(userId: number) {
  return request({
    url: `/api/v1/apikey/get/${userId}`,
    method: "get",
  });
}

/**
 * 创建新的API密钥
 */
export function createAPIKey(data: CreateAPIKeyParams) {
  return request({
    url: "/api/v1/apikey/add",
    method: "post",
    data,
  });
}

/**
 * 删除API密钥
 */
export function deleteAPIKey(apiKeyId: number) {
  return request({
    url: `/api/v1/apikey/delete/${apiKeyId}`,
    method: "delete",
  });
}
