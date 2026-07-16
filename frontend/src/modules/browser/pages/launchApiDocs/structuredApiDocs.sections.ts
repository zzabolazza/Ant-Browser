import type { StructuredApiSectionDoc } from './structuredApiDocs.types'

export const STRUCTURED_API_SECTION_DOCS: StructuredApiSectionDoc[] = [
  {
    id: 'api-profiles-launch',
    title: '实例与启动',
    intro: '这页只做实例管理和启动能力的总览。先通过实例接口完成配置，再按 launchCode 或 selector 启动实例。',
    highlights: [
      '/api/profiles 负责实例增删改查。',
      '/api/launch 负责启动。',
      '详情只从表格里的“查看详情”进入。',
    ],
  },
  {
    id: 'api-runtime',
    title: '运行态与 CDP',
    intro: '这页只列运行态控制。启动后用返回的实例 cdpUrl 直连浏览器调试端口，Launch Server 不再转发 CDP。',
    highlights: [
      'runtime/session 用于等待 debugReady。',
      'runtime/status / runtime/stop 都按 selector 工作。',
      'attach 时使用响应里的直连 cdpUrl。',
    ],
  },
]
