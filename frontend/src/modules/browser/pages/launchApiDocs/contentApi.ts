export const DOC_API_PROFILES_LAUNCH = `# 实例与启动

## 实例接口

| 方法 | 路径 | 用途 |
|------|------|------|
| \`GET\` | \`/api/profiles\` | 列实例 |
| \`POST\` | \`/api/profiles\` | 创建实例 |
| \`GET\` | \`/api/profiles/{profileId}\` | 查单个实例 |
| \`PUT\` | \`/api/profiles/{profileId}\` | 更新实例 |
| \`DELETE\` | \`/api/profiles/{profileId}\` | 删除实例 |
| \`GET\` | \`/api/profiles/{profileId}/status\` | 查实例运行态 |
| \`POST\` | \`/api/profiles/{profileId}/stop\` | 停止实例 |

## 创建实例

\`\`\`bash
curl -X POST http://127.0.0.1:19876/api/profiles \\
  -H "Content-Type: application/json" \\
  -d '{
    "profile": {
      "profileName": "buyer-001",
      "proxyId": "proxy-us",
      "keywords": ["buyer-001"],
      "tags": ["电商"]
    },
    "launchCode": "BUYER_001"
  }'
\`\`\`

## 创建并立即启动

\`\`\`bash
curl -X POST http://127.0.0.1:19876/api/profiles \\
  -H "Content-Type: application/json" \\
  -d '{
    "profile": {
      "profileName": "buyer-002",
      "keywords": ["buyer-002"]
    },
    "autoLaunch": true,
    "start": {
      "skipDefaultStartUrls": true
    }
  }'
\`\`\`

## 创建实例（自定义代理配置）

\`\`\`bash
curl -X POST http://127.0.0.1:19876/api/profiles \\
  -H "Content-Type: application/json" \\
  -d '{
    "profile": {
      "profileName": "buyer-003",
      "proxyConfig": "http://127.0.0.1:18080",
      "keywords": ["buyer-003"]
    }
  }'
\`\`\`

## 查询 / 更新 / 删除

\`\`\`bash
curl http://127.0.0.1:19876/api/profiles
curl http://127.0.0.1:19876/api/profiles/550e8400-e29b-41d4-a716-446655440000
curl -X PUT http://127.0.0.1:19876/api/profiles/550e8400-e29b-41d4-a716-446655440000 -H "Content-Type: application/json" -d '{ ... }'
curl -X DELETE http://127.0.0.1:19876/api/profiles/550e8400-e29b-41d4-a716-446655440000
\`\`\`

## 启动接口

| 方法 | 路径 | 用途 |
|------|------|------|
| \`GET\` | \`/api/launch/{code}\` | 按唯一 Code 启动 |
| \`POST\` | \`/api/launch\` | 按 code / selector 参数化启动 |

### 按 Code 启动

\`\`\`bash
curl http://127.0.0.1:19876/api/launch/A3F9K2
\`\`\`

### 按 selector 启动

\`\`\`bash
curl -X POST http://127.0.0.1:19876/api/launch \\
  -H "Content-Type: application/json" \\
  -d '{
    "selector": {
      "keyword": "checkout",
      "tags": ["电商", "北美"],
      "groupId": "group-sales-us",
      "matchMode": "unique"
    },
    "skipDefaultStartUrls": true
  }'
\`\`\`

## 启动成功响应

\`\`\`json
{
  "ok": true,
  "profileId": "550e8400-e29b-41d4-a716-446655440000",
  "launchCode": "BUYER_001",
  "debugReady": true,
  "cdpUrl": "http://127.0.0.1:9333"
}
\`\`\`

## 单实例状态 / 停止

| 方法 | 路径 | 示例用途 |
|------|------|----------|
| \`GET\` | \`/api/profiles/{profileId}/status\` | 查实例是否运行、是否 ready |
| \`POST\` | \`/api/profiles/{profileId}/stop\` | 任务完成后精确停止 |

## 记住这几个规则

\`\`\`text
launchCode 冲突 -> 409
PUT 是整份更新
运行中的实例不能直接 DELETE
matchMode=all 只在 POST /api/launch 可用
proxyId 和 proxyConfig 同时传 -> 优先 proxyId
proxyId 无效 + proxyConfig 非空 -> 使用 proxyConfig
proxyId 无效 + proxyConfig 为空 -> 400
\`\`\`
`

export const DOC_API_RUNTIME = `# 运行态与接管

## 接口

| 方法 | 路径 | 用途 |
|------|------|------|
| \`POST\` | \`/api/runtime/session\` | 准备可接管会话 |
| \`POST\` | \`/api/runtime/status\` | 按 selector 查运行态 |
| \`POST\` | \`/api/runtime/stop\` | 按 selector 停止实例 |

## 查询当前活动实例

\`\`\`bash
\`\`\`

\`\`\`json
{
  "ok": true,
  "active": true,
  "profileId": "550e8400-e29b-41d4-a716-446655440000",
  "launchCode": "BUYER_001",
  "debugReady": true,
  "cdpUrl": "http://127.0.0.1:9333"
}
\`\`\`

## 准备可接管会话

\`\`\`bash
curl -X POST http://127.0.0.1:19876/api/runtime/session \\
  -H "Content-Type: application/json" \\
  -d '{
    "selector": {
      "code": "BUYER_001"
    },
    "timeoutMs": 45000,
    "skipDefaultStartUrls": true
  }'
\`\`\`

| 返回 | 含义 |
|------|------|
| \`200 + ready=true\` | 可以直接 attach |
| \`202 + ready=false\` | 已处理，但还没 ready |

## 按 selector 查状态

\`\`\`bash
curl -X POST http://127.0.0.1:19876/api/runtime/status \\
  -H "Content-Type: application/json" \\
  -d '{
    "selector": {
      "keyword": "shop",
      "matchMode": "first"
    }
  }'
\`\`\`

## 按 selector 停止

\`\`\`bash
curl -X POST http://127.0.0.1:19876/api/runtime/stop \\
  -H "Content-Type: application/json" \\
  -d '{
    "code": "BUYER_001"
  }'
\`\`\`

## 接管示例

\`\`\`javascript
import { chromium } from "playwright";

const res = await fetch("http://127.0.0.1:19876/api/runtime/session", {
  method: "POST",
  headers: { "Content-Type": "application/json" },
  body: JSON.stringify({
    selector: { code: "BUYER_001" },
    skipDefaultStartUrls: true
  })
});

const data = await res.json();
const browser = await chromium.connectOverCDP(data.cdpUrl);
\`\`\`

## 记住这几个规则

\`\`\`text
runtime/status 和 runtime/stop 不支持 matchMode=all
attach 前确认 debugReady=true 且 cdpUrl 非空
统一入口只指向一个活动实例
\`\`\`
`

