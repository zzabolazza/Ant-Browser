import type { StructuredApiEndpointDoc } from './structuredApiDocs.types'

export const RUNTIME_API_ENDPOINT_DOCS: StructuredApiEndpointDoc[] = [
  {
    id: 'api-runtime-session-detail',
    parentId: 'api-runtime',
    label: '准备可接管会话',
    method: 'POST',
    path: '/api/runtime/session',
    purpose: '准备一个可 attach 的运行时会话。',
    description: '按 selector 命中实例，必要时自动启动，并在给定超时时间内等待 debugReady=true。就绪后用返回的 cdpUrl 直连该实例调试端口。',
    fields: [
      { name: 'selector', type: 'object', required: false, location: 'Body', description: '目标实例选择条件；新接入推荐使用。' },
      { name: 'code', type: 'string', required: false, location: 'Body', description: '兼容写法：等价于 selector.code。' },
      { name: 'profileId', type: 'string', required: false, location: 'Body', description: '兼容写法：等价于 selector.profileId。' },
      { name: 'profileName', type: 'string', required: false, location: 'Body', description: '兼容写法：等价于 selector.profileName。' },
      { name: 'keyword / keywords', type: 'string / string[]', required: false, location: 'Body', description: '兼容写法：等价于 selector.keyword / selector.keywords。' },
      { name: 'tag / tags', type: 'string / string[]', required: false, location: 'Body', description: '兼容写法：等价于 selector.tag / selector.tags。' },
      { name: 'groupId', type: 'string', required: false, location: 'Body', description: '兼容写法：等价于 selector.groupId。' },
      { name: 'matchMode', type: 'unique | first', required: false, location: 'Body', description: '兼容写法：等价于 selector.matchMode。' },
      { name: 'timeoutMs', type: 'integer', required: false, location: 'Body', description: '等待 debugReady 的超时时间。' },
      { name: 'startUrls', type: 'string[]', required: false, location: 'Body', description: '本次启动时额外打开的网址。' },
      { name: 'skipDefaultStartUrls', type: 'boolean', required: false, location: 'Body', description: '是否跳过实例默认启动 URL。' },
      { name: 'launchArgs', type: 'string[]', required: false, location: 'Body', description: '本次启动时临时附加的启动参数。' },
    ],
    requestExample: {
      language: 'bash',
      code: ({ launchBaseUrl, authHeader }) => `curl -X POST ${launchBaseUrl}/api/runtime/session \\
  -H "Content-Type: application/json" \\
  -H "${authHeader}: <your-api-key>" \\
  -d '{
    "selector": { "code": "BUYER_001" },
    "timeoutMs": 45000,
    "skipDefaultStartUrls": true
  }'`,
    },
    responseExample: {
      language: 'json',
      code: () => `{
  "ok": true,
  "ready": true,
  "waitTimedOut": false,
  "retryable": false,
  "profileId": "550e8400-e29b-41d4-a716-446655440000",
  "launchCode": "BUYER_001",
  "running": true,
  "debugReady": true,
  "cdpPort": 9333,
  "cdpUrl": "http://127.0.0.1:9333",
  "timeoutMs": 45000
}`,
    },
    responseCodes: [
      { code: '200', description: '实例已 ready，可直接 attach。' },
      { code: '202', description: '实例已处理但暂未 ready，可稍后重试。' },
      { code: '400', description: 'selector 缺失或 matchMode 非法。' },
      { code: '404', description: '目标实例不存在。' },
    ],
    notes: [
      'selector 为空且没有任何兼容顶层选择字段时返回 400。',
      '200 表示 ready，用 cdpUrl 直连即可。',
      '202 表示未 ready，需要重试。',
    ],
  },
  {
    id: 'api-runtime-status-detail',
    parentId: 'api-runtime',
    label: '按 selector 查状态',
    method: 'POST',
    path: '/api/runtime/status',
    purpose: '按 selector 查询实例当前运行态。',
    description: '不启动新实例，不等待 ready，只看当前 selector 命中的实例状态。',
    fields: [
      { name: 'selector', type: 'object', required: false, location: 'Body', description: '目标实例选择条件；新接入推荐使用。' },
      { name: 'code / profileId / profileName', type: 'string', required: false, location: 'Body', description: '兼容顶层选择字段。' },
      { name: 'keyword / keywords', type: 'string / string[]', required: false, location: 'Body', description: '兼容顶层选择字段。' },
      { name: 'tag / tags / groupId', type: 'string / string[]', required: false, location: 'Body', description: '兼容顶层选择字段。' },
      { name: 'matchMode', type: 'unique | first', required: false, location: 'Body', description: '运行态控制不支持 all。' },
    ],
    requestExample: {
      language: 'bash',
      code: ({ launchBaseUrl, authHeader }) => `curl -X POST ${launchBaseUrl}/api/runtime/status \\
  -H "Content-Type: application/json" \\
  -H "${authHeader}: <your-api-key>" \\
  -d '{
    "selector": { "keyword": "checkout", "matchMode": "first" }
  }'`,
    },
    responseExample: {
      language: 'json',
      code: () => `{
  "ok": true,
  "profileId": "550e8400-e29b-41d4-a716-446655440000",
  "launchCode": "BUYER_001",
  "running": true,
  "debugReady": true,
  "cdpPort": 9333,
  "cdpUrl": "http://127.0.0.1:9333"
}`,
    },
    responseCodes: [
      { code: '200', description: '返回运行态。' },
      { code: '400', description: 'selector 缺失或 matchMode 非法。' },
      { code: '404', description: '目标实例不存在。' },
    ],
    notes: [
      '不会启动实例。',
      'selector 为空且没有任何兼容顶层选择字段时返回 400。',
      'debugReady=true 时 cdpUrl 为该实例直连地址。',
    ],
  },
  {
    id: 'api-runtime-stop-detail',
    parentId: 'api-runtime',
    label: '按 selector 停止',
    method: 'POST',
    path: '/api/runtime/stop',
    purpose: '按 selector 停止实例。',
    description: '和 runtime/status 一样使用 selector，但动作改为停止实例，适合编排侧做统一回收。',
    fields: [
      { name: 'selector', type: 'object', required: false, location: 'Body', description: '目标实例选择条件；新接入推荐使用。' },
      { name: 'code / profileId / profileName', type: 'string', required: false, location: 'Body', description: '兼容顶层选择字段。' },
      { name: 'keyword / keywords', type: 'string / string[]', required: false, location: 'Body', description: '兼容顶层选择字段。' },
      { name: 'tag / tags / groupId', type: 'string / string[]', required: false, location: 'Body', description: '兼容顶层选择字段。' },
      { name: 'matchMode', type: 'unique | first', required: false, location: 'Body', description: '运行态控制不支持 all。' },
    ],
    requestExample: {
      language: 'bash',
      code: ({ launchBaseUrl, authHeader }) => `curl -X POST ${launchBaseUrl}/api/runtime/stop \\
  -H "Content-Type: application/json" \\
  -H "${authHeader}: <your-api-key>" \\
  -d '{
    "selector": { "code": "BUYER_001" }
  }'`,
    },
    responseExample: {
      language: 'json',
      code: () => `{
  "ok": true,
  "stopped": true,
  "profileId": "550e8400-e29b-41d4-a716-446655440000",
  "launchCode": "BUYER_001",
  "running": false,
  "debugReady": false,
  "cdpPort": 0,
  "cdpUrl": ""
}`,
    },
    responseCodes: [
      { code: '200', description: '停止成功。' },
      { code: '400', description: 'selector 缺失或 matchMode 非法。' },
      { code: '404', description: '目标实例不存在。' },
    ],
    notes: [
      'selector 为空且没有任何兼容顶层选择字段时返回 400。',
      '不支持 matchMode=all。',
    ],
  },
]
