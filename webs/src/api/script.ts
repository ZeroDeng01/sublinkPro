import request from "@/utils/request";

export function getScriptList() {
  return request({
    url: "/api/v1/script/list",
    method: "get",
  });
}

export function addScript(data: any) {
  return request({
    url: "/api/v1/script/add",
    method: "post",
    data,
  });
}

export function updateScript(data: any) {
  return request({
    url: "/api/v1/script/update",
    method: "post",
    data,
  });
}

export function deleteScript(data: any) {
  return request({
    url: "/api/v1/script/delete",
    method: "delete",
    data,
  });
}
