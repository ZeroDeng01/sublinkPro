import request from "@/utils/request";
export function getNodes() {
  return request({
    url: "/api/v1/nodes/get",
    method: "get",
  });
}

export function AddNodes(data: any) {
  return request({
    url: "/api/v1/nodes/add",
    method: "post",
    data,
    headers: {
      "Content-Type": "multipart/form-data",
    },
  });
}
export function UpdateNode(data: any) {
  return request({
    url: "/api/v1/nodes/update",
    method: "post",
    data,
    headers: {
      "Content-Type": "multipart/form-data",
    },
  });
}
export function DelNode(data: any) {
  return request({
    url: "/api/v1/nodes/delete",
    method: "delete",
    params: data,
  });
}

export function getSpeedTestConfig() {
  return request({
    url: "/api/v1/nodes/speed-test/config",
    method: "get",
  });
}

export function updateSpeedTestConfig(data: any) {
  return request({
    url: "/api/v1/nodes/speed-test/config",
    method: "post",
    data,
  });
}

export function runSpeedTest(ids?: number[]) {
  return request({
    url: "/api/v1/nodes/speed-test/run",
    method: "post",
    data: { ids },
  });
}
