import request from "@/utils/request";
import type { PluginInfo, PluginConfigRequest } from "./types";

// API响应类型
interface ApiResponse<T> {
  code: string;
  msg: string;
  data: T;
}

/**
 * 获取所有插件列表
 * @returns 插件列表
 */
export const getPluginsApi = () => {
  return request<ApiResponse<PluginInfo[]>>({
    url: "/api/v1/plugins/get",
    method: "get",
  });
};

/**
 * 启用插件
 * @param name 插件名称
 * @returns 操作结果
 */
export const enablePluginApi = (name: string) => {
  return request<ApiResponse<null>>({
    url: `/api/v1/plugins/enable/${name}`,
    method: "post",
  });
};

/**
 * 禁用插件
 * @param name 插件名称
 * @returns 操作结果
 */
export const disablePluginApi = (name: string) => {
  return request<ApiResponse<null>>({
    url: `/api/v1/plugins/disable/${name}`,
    method: "post",
  });
};

/**
 * 获取插件配置
 * @param name 插件名称
 * @returns 插件配置
 */
export const getPluginConfigApi = (name: string) => {
  return request<ApiResponse<Record<string, any>>>({
    url: `/api/v1/plugins/config/${name}`,
    method: "get",
  });
};

/**
 * 更新插件配置
 * @param data 包含插件名称和配置的对象
 * @returns 操作结果
 */
export const updatePluginConfigApi = (data: PluginConfigRequest) => {
  return request<ApiResponse<null>>({
    url: `/api/v1/plugins/config`,
    method: "put",
    data: data,
  });
};

/**
 * 重新加载所有插件
 * @returns 操作结果
 */
export const reloadPluginsApi = () => {
  return request<ApiResponse<null>>({
    url: "/api/v1/plugins/reload",
    method: "post",
  });
};
