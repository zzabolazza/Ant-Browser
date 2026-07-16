export const DOC_TUTORIAL = `# 使用教程

## 只在应用内使用

\`\`\`text
内核管理 -> 下载 / 识别内核
代理池配置 -> 导入 / 添加代理
插件包管理 -> 安装 / 导入插件
实例列表 -> 新建实例 -> 选择内核 / 代理 -> 启动
实例列表 -> 导入实例 / 选中实例 -> 导出
\`\`\`

## 最小上手顺序

1. 打开 \`指纹浏览器 > 内核管理\`
2. 准备一个可用内核
3. 打开 \`指纹浏览器 > 代理池配置\`，按需导入代理
4. 需要固定插件时，打开 \`指纹浏览器 > 插件包管理\` 安装或导入插件
5. 打开 \`指纹浏览器 > 实例列表\`，新建实例并按需配置插件
6. 保存后直接启动

## 实例迁移

\`\`\`text
导出：实例列表 -> 勾选实例 -> 导出
单个导出：实例列表 -> 更多 -> 导出
导入：实例列表 -> 导入实例 -> 选择 ZIP
\`\`\`

- 导出前先停止实例，避免浏览器数据文件被占用
- 导出包包含实例配置和完整浏览器用户数据目录
- 导入时会生成新实例，不覆盖当前已有实例
- 代理只按名称适配本地同名代理，未匹配时自动清空代理
`

export const DOC_OPERATION_FLOW = `# 操作流程

## 首次配置

\`\`\`text
1. 内核管理：准备 Chrome 内核
2. 代理池配置：导入或录入代理
3. 插件包管理：安装插件，按需限制实例
4. 实例列表：创建实例并选择内核、代理、分组
5. 启动实例
6. 迁移实例：停止实例后导出 ZIP，新环境中导入为新实例
\`\`\`

## 插件接入流程

\`\`\`text
插件包管理 -> 输入商店链接 / 插件 ID -> 安装
插件包管理 -> 限制实例 -> 按分组选择实例 -> 保存
实例列表 -> 单个实例 -> 插件 -> 开启单独配置（可选）
\`\`\`

## 分组限制怎么用

- 在实例列表或分组管理里先维护实例分组
- 在插件包管理里点插件卡片的 \`限制实例\`
- 弹窗会按分组展示实例，并显示每组已选数量
- 点 \`选择本组\` 可以快速让整组实例加载该插件
- 点 \`取消本组\` 可以快速让整组实例排除此插件

## 启动前检查

\`\`\`text
1. 实例有可用内核
2. 代理可连通，或确认不需要代理
3. 必要插件已安装，并且该实例在插件限制范围内
4. 启动后状态显示为已就绪
\`\`\`
`


export const DOC_CORE_INTRO = `# 内核介绍

## 推荐目录

\`\`\`text
chrome/
  chrome-<version>/
    chrome.exe
    ...
\`\`\`

## 两种准备方式

\`\`\`text
方式 A：应用内下载
指纹浏览器 -> 内核管理 -> 下载内核

方式 B：手动下载
下载 ZIP -> 解压到 chrome/<version>/ -> 回到内核管理确认识别
\`\`\`

## 下载渠道

- [fingerprint-chromium](https://github.com/adryfish/fingerprint-chromium)
- [Releases](https://github.com/adryfish/fingerprint-chromium/releases)

## 在应用里怎么用

\`\`\`text
内核管理 -> 设默认内核
实例编辑页 -> 选择内核
\`\`\`

## 全局设置

\`\`\`text
内核管理 -> 全局设置 -> 编辑
\`\`\`

- 用户数据根目录决定实例数据保存位置
- 默认启动参数会追加到实例启动参数里
- 默认指纹参数会追加到指纹内核启动参数里
- 默认启动页面为空时，不会额外打开页面
- 轻启动模式用于减少启动阶段的额外等待

## 自检

\`\`\`text
1. 目录下能看到 chrome.exe
2. 内核管理页能识别到该目录
3. 实例绑定的是正确版本
\`\`\`
`

export const DOC_PROXY_INTRO = `# 代理介绍

## 直接录入示例

\`\`\`text
<proxy-url>
<proxy-url>
<proxy-url>
\`\`\`

## Clash YAML 导入示例

\`\`\`yaml
proxies:
  - name: hk-vless
    type: vless
    server: example.com
    port: 443
    uuid: your-uuid
    tls: true
    servername: example.com
\`\`\`

## 批量 DNS 示例

\`\`\`yaml
dns:
  enable: true
  nameserver:
    - 119.29.29.29
    - 223.5.5.5
\`\`\`

## 应用内路径

\`\`\`text
代理池配置 -> 导入 Clash YAML / 录入 HTTP(S) / SOCKS5
实例编辑页 -> 选择代理池节点
\`\`\`

## 在实例接口里绑定代理

\`\`\`json
{
  "profile": {
    "profileName": "buyer-001",
    "proxyId": "proxy-us",
    "keywords": ["buyer-001"]
  },
  "launchCode": "BUYER_001"
}
\`\`\`

不需要代理时，\`proxyId\` 留空即可。
\`proxyId\` 与 \`proxyConfig\` 同时传时，优先使用 \`proxyId\` 对应的代理池节点。
如果 \`proxyId\` 无效但传了 \`proxyConfig\`，会自动改为使用该 \`proxyConfig\`。
如果 \`proxyId\` 无效且 \`proxyConfig\` 也为空，请求会直接报错。
`

export const DOC_EXTENSION_INTRO = `# 插件包管理

## 入口

\`\`\`text
指纹浏览器 -> 插件包管理
\`\`\`

## 安装方式

- 输入 Chrome Web Store 链接或 32 位插件 ID 后安装
- 本地已有 \`.crx\` / \`.zip\` 时，用手动安装导入文件
- 已解压的插件目录可直接导入目录
- 安装完成后，插件默认按全局启用状态参与实例启动

## 限制实例

\`\`\`text
插件包管理 -> 插件卡片 -> 限制实例
\`\`\`

- 勾选的实例会加载该插件
- 未勾选的实例会排除该插件
- 弹窗按实例分组展示，未设置分组的实例进入 \`未分组\`
- 每个分组支持 \`选择本组\` 和 \`取消本组\`
- 分组标题会显示该组 \`已选 / 总数\`

## 单实例插件配置

\`\`\`text
实例列表 -> 实例行 -> 插件
\`\`\`

- 默认情况下，实例继承全局已启用插件和插件限制规则
- 打开单独配置后，该实例只加载弹窗内勾选的插件
- 单独配置适合少量特殊实例，不建议替代分组限制

## 推荐用法

\`\`\`text
同一批业务实例 -> 放入同一分组
同一批业务需要的插件 -> 在限制实例里选择对应分组
个别实例例外 -> 再用单实例插件配置微调
\`\`\`
`

export const DOC_API_OVERVIEW = `# 接口总览

## 基础地址

\`\`\`text
http://127.0.0.1:19876
\`\`\`

## 认证示例

\`\`\`bash
curl -H "X-Ant-Api-Key: <your-api-key>" http://127.0.0.1:19876/api/health
\`\`\`

没开认证时，去掉这个请求头即可。

## 全部接口

| 分类 | 方法 | 路径 |
|------|------|------|
| 健康检查 | \`GET\` | \`/api/health\` |
| 实例管理 | \`GET\` | \`/api/profiles\` |
| 实例管理 | \`POST\` | \`/api/profiles\` |
| 实例管理 | \`GET\` | \`/api/profiles/{profileId}\` |
| 实例管理 | \`PUT\` | \`/api/profiles/{profileId}\` |
| 实例管理 | \`DELETE\` | \`/api/profiles/{profileId}\` |
| 实例管理 | \`GET\` | \`/api/profiles/{profileId}/status\` |
| 实例管理 | \`POST\` | \`/api/profiles/{profileId}/stop\` |
| 启动 | \`GET\` | \`/api/launch/{code}\` |
| 启动 | \`POST\` | \`/api/launch\` |
| 运行态 | \`POST\` | \`/api/runtime/session\` |
| 运行态 | \`POST\` | \`/api/runtime/status\` |
| 运行态 | \`POST\` | \`/api/runtime/stop\` |
| 调用日志 | \`GET\` | \`/api/launch/logs\` |

## selector 最小写法

\`\`\`json
{
  "selector": {
    "code": "BUYER_001",
    "matchMode": "unique"
  }
}
\`\`\`

也可以把 \`code / profileId / profileName / keyword / tags / groupId / matchMode\` 放在请求体顶层；新接入建议统一放进 \`selector\`。

## 怎么选接口

| 场景 | 用哪个 |
|------|--------|
| 已知唯一 Code | \`GET /api/launch/{code}\` |
| 需要参数 / selector | \`POST /api/launch\` |
| 需要 ready 后再接管 | \`POST /api/runtime/session\` |
| 只有 selector，想查状态 | \`POST /api/runtime/status\` |
| 只有 selector，想停止 | \`POST /api/runtime/stop\` |
`
