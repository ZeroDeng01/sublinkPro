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

function decorateTemplateAIError(error, data) {
  if (data && typeof data === 'object' && !Array.isArray(data)) {
    if (Object.prototype.hasOwnProperty.call(data, 'code')) {
      error.code = data.code;
    }
    if (Object.prototype.hasOwnProperty.call(data, 'message')) {
      error.message = data.message || error.message;
    }
    if (Object.prototype.hasOwnProperty.call(data, 'operationIndex')) {
      error.operationIndex = data.operationIndex;
    }
    if (Object.prototype.hasOwnProperty.call(data, 'validation')) {
      error.validation = data.validation;
    }
  }

  return error;
}

function dispatchTemplateAIStreamEvent(parsedEvent, handlers, setFinalPayload) {
  const { event, data } = parsedEvent;

  switch (event) {
    case 'template.edit.session.created':
      handlers.onSessionCreated?.(data);
      break;
    case 'template.edit.model.delta':
      handlers.onDelta?.(data);
      break;
    case 'template.edit.operations.ready':
      handlers.onOperationsReady?.(data);
      break;
    case 'template.edit.preview.validating':
      handlers.onPreviewValidating?.(data);
      break;
    case 'template.edit.preview.ready':
      setFinalPayload(data);
      handlers.onPreviewReady?.(data);
      break;
    case 'template.edit.warning':
      handlers.onWarning?.(data);
      break;
    case 'template.edit.completed':
      setFinalPayload(data);
      handlers.onCompleted?.(data);
      break;
    case 'template.edit.error': {
      handlers.onError?.(data);
      const message = data?.message || data?.msg || 'AI 生成失败';
      throw decorateTemplateAIError(createStreamRequestError(message, null, data), data);
    }
    default:
      break;
  }
}

export async function streamTemplateAIEditSession(data, handlers = {}) {
  const token = localStorage.getItem('accessToken');
  const response = await fetch('/api/v1/template/ai/edit-sessions/stream', {
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

export function getTemplateAIEditSession(sessionId) {
  return request({
    url: `/v1/template/ai/edit-sessions/${sessionId}`,
    method: 'get'
  });
}

export function acceptTemplateAIEditSession(sessionId, data = {}) {
  return request({
    url: `/v1/template/ai/edit-sessions/${sessionId}/accept`,
    method: 'post',
    data: data.currentText !== undefined ? { currentText: data.currentText } : {}
  });
}

export function discardTemplateAIEditSession(sessionId) {
  return request({
    url: `/v1/template/ai/edit-sessions/${sessionId}/discard`,
    method: 'post'
  });
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
