export interface NavItem {
  name: string
  path: string
  icon: string
}

export const navigationConfig: NavItem[] = [
  { name: '实例列表', path: '/browser/list', icon: 'Monitor' },
  { name: '内核管理', path: '/browser/cores', icon: 'Cpu' },
  { name: '插件管理', path: '/browser/extensions', icon: 'Puzzle' },
  { name: '代理配置', path: '/browser/proxy-pool', icon: 'Globe' },
  { name: '默认书签', path: '/browser/bookmarks', icon: 'Bookmark' },
  { name: '标签管理', path: '/browser/tags', icon: 'Tag' },
  { name: '系统设置', path: '/settings', icon: 'Settings' },
  { name: '日志查看', path: '/browser/logs', icon: 'FileText' },
]
