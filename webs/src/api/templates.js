import request from './request';

function parseStreamEventData(data) {
  if (!data) {
    return null;
  }

  try {
    return JSON.parse(data);
  } catch {
    return data;
  }
}

function parseStreamEventBlock(block) {
  const lines = block.split(/\r?\n/);
  let event = 'message';
  const dataLines = [];

  lines.forEach((line) => {
    if (!line || line.startsWith(':')) {
      return;
    }

    const separatorIndex = line.indexOf(':');
    const field = separatorIndex === -1 ? line : line.slice(0, separatorIndex);
    const value = separatorIndex === -1 ? '' : line.slice(separatorIndex + 1).trimStart();

    if (field === 'event') {
      event = value || 'message';
    }

    if (field === 'data') {
      dataLines.push(value);
    }
  });

  return {
    event,
    data: parseStreamEventData(dataLines.join('\n'))
  };
}

function createStreamRequestError(message, response, data) {
  const error = new Error(message);
  error.response = response
    ? {
        status: response.status,
        data
      }
    : undefined;
  error.data = data;
  return error;
}

async function buildStreamResponseError(response) {
  const contentType = response.headers.get('content-type') || '';
  let data = null;

  if (contentType.includes('application/json')) {
    data = await response.json().catch(() => null);
  } else {
    const text = await response.text().catch(() => '');
    data = parseStreamEventData(text) || text;
  }

  const message = data?.message || data?.msg || `AI 生成失败 (${response.status})`;
  return createStreamRequestError(message, response, data);
}

function dispatchTemplateAIStreamEvent(parsedEvent, handlers, setFinalPayload) {
  const { event, data } = parsedEvent;

  switch (event) {
    case 'response.created':
      handlers.onStart?.(data);
      break;
    case 'response.output_text.delta':
      handlers.onDelta?.(data);
      break;
    case 'response.completed':
      handlers.onComplete?.(data);
      break;
    case 'response.failed':
      handlers.onError?.(data);
      throw createStreamRequestError(data?.message || data?.error?.message || 'AI 生成失败', null, data);
    case 'template.final':
      setFinalPayload(data);
      handlers.onFinal?.(data);
      break;
    case 'error': {
      handlers.onError?.(data);
      const message = data?.message || data?.msg || (typeof data === 'string' ? data : 'AI 生成失败');
      throw createStreamRequestError(message, null, data);
    }
    default:
      break;
  }
}

export async function generateTemplateAICandidateStream(data, handlers = {}) {
  const token = localStorage.getItem('accessToken');
  const response = await fetch('/api/v1/template/ai/generate-stream', {
    method: 'POST',
    headers: {
      Accept: 'text/event-stream',
      'Content-Type': 'application/json',
      ...(token ? { Authorization: token } : {})
    },
    body: JSON.stringify(data),
    signal: handlers.signal
  });

  if (response.status === 401) {
    localStorage.removeItem('accessToken');
    window.location.href = '/login';
  }

  if (!response.ok) {
    throw await buildStreamResponseError(response);
  }

  if (!response.body) {
    throw new Error('AI 流式响应不可用');
  }

  const reader = response.body.getReader();
  const decoder = new TextDecoder();
  let buffer = '';
  let finalPayload = null;

  try {
    while (true) {
      const { value, done } = await reader.read();

      if (done) {
        break;
      }

      buffer += decoder.decode(value, { stream: true });
      const blocks = buffer.split(/\r?\n\r?\n/);
      buffer = blocks.pop() || '';

      blocks.forEach((block) => {
        if (!block.trim()) {
          return;
        }

        dispatchTemplateAIStreamEvent(parseStreamEventBlock(block), handlers, (payload) => {
          finalPayload = payload;
        });
      });
    }

    buffer += decoder.decode();

    if (buffer.trim()) {
      dispatchTemplateAIStreamEvent(parseStreamEventBlock(buffer), handlers, (payload) => {
        finalPayload = payload;
      });
    }
  } finally {
    reader.releaseLock();
  }

  if (!finalPayload) {
    throw new Error('AI 生成未返回最终结果');
  }

  return finalPayload;
}

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

export function getTemplateUsage(params) {
  return request({
    url: '/v1/template/usage',
    method: 'get',
    params
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

export function generateTemplateAICandidate(data) {
  return request({
    url: '/v1/template/ai/generate',
    method: 'post',
    data
  });
}

export function validateTemplateAICandidate(data) {
  return request({
    url: '/v1/template/ai/validate',
    method: 'post',
    data
  });
}

export function applyTemplateAICandidate(data) {
  return request({
    url: '/v1/template/ai/apply',
    method: 'post',
    data
  });
}
