import request from "@/utils/request";
export function getSubs() {
  return request({
    url: "/api/v1/subcription/get",
    method: "get",
  }).then((response) => {
    // 确保每个订阅都有 SubLogs 数组
    if (response.data && Array.isArray(response.data)) {
      response.data.forEach((sub) => {
        // 如果 SubLogs 不存在或为空数组，添加一个默认空记录
        if (!sub.Nodes || sub.Nodes.length === 0) {
          sub.Nodes = [];
        }
      });
    }
    console.log("获取订阅列表", response.data);
    return response;
  });
}

export function AddSub(data: any) {
  return request({
    url: "/api/v1/subcription/add",
    method: "post",
    data,
    headers: {
      "Content-Type": "multipart/form-data",
    },
  });
}
export function DelSub(data: any) {
  return request({
    url: "/api/v1/subcription/delete",
    method: "delete",
    params: data,
  });
}

export function UpdateSub(data: any) {
  return request({
    url: "/api/v1/subcription/update",
    method: "post",
    data,
    headers: {
      "Content-Type": "multipart/form-data",
    },
  });
}

export function SortSub(data: any) {
  return request({
    url: "/api/v1/subcription/sort",
    method: "post",
    data,
    headers: {
      "Content-Type": "application/json",
    },
  });
}
