# Facade

> One Facade. Endless Faces.

[![Release](https://img.shields.io/github/v/release/zzabolazza/Facade?sort=semver)](https://github.com/zzabolazza/Facade/releases)
[![Platform](https://img.shields.io/badge/Platform-Windows%20%7C%20Linux%20%7C%20macOS-blue)](https://github.com/zzabolazza/Facade/releases)
[![Issues](https://img.shields.io/github/issues/zzabolazza/Facade)](https://github.com/zzabolazza/Facade/issues)

桌面端多账号浏览器隔离管理工具。为每个账号分配独立浏览器实例与代理出口，在一台设备上稳定运行多个隔离环境。

**适用场景：** 多账号运营 / 跨境电商 / 指纹浏览器测试 / 独立代理出口验证

---

## 功能概览

| 功能 | 说明 |
|------|------|
| **实例隔离** | 每个账号独立浏览器环境，Cookie、LocalStorage、IndexedDB 互不干扰 |
| **代理绑定** | 支持 `direct://` / `http://` / `https://` / `socks5://`，按实例分配独立出口 |
| **内核管理** | 维护多个 Chromium 内核版本，一键切换默认内核 |
| **插件管理** | 安装、导入、启停、删除，支持实例级插件限制 |
| **实例迁移** | ZIP 打包导出完整用户数据，跨设备导入为新实例 |
| **快捷启动** | `Ctrl + K` 按 Code / 名称 / 标签秒开目标实例 |
| **书签管理** | 统一维护书签，支持启动时自动打开并同步到已有实例 |
| **分组与标签** | 按业务场景组织实例，支持关键字检索 |
| **快照** | 创建 / 还原实例时间点快照 |
| **备份恢复** | 全量备份 / 还原，带进度追踪与数据合并 |
| **Launch API** | 本地 HTTP 接口，供外部系统调用浏览器实例 |
| **系统托盘** | 最小化到托盘，支持单实例保护 |

---

## 技术栈

| 层 | 技术 |
|----|------|
| 桌面框架 | [Wails v2](https://wails.io) — Go + Web |
| 后端 | Go 1.25 · SQLite (pure-Go) · WebSocket |
| 前端 | React 18 · TypeScript · [Tailwind CSS](https://tailwindcss.com) · [Vite 5](https://vitejs.dev) |
| 状态管理 | [Zustand](https://github.com/pmndrs/zustand) |
| 图标 | [Lucide React](https://lucide.dev) |

---

## 快速开始

### 下载安装

| 平台 | 方式 |
|------|------|
| **Windows** | 下载 `Facade-Setup-*.exe` 安装，或解压便携版运行 `facade.exe` |
| **Linux** | `sudo apt install ./facade_<version>_<arch>.deb`（推荐，自动安装依赖）；或解压 `.tar.gz` 后运行 `facade` |
| **macOS** | 解压 `Facade-*.app`，首次运行执行 `xattr -dr com.apple.quarantine Facade.app` |

前往 [Releases](https://github.com/zzabolazza/Facade/releases) 下载最新版本。

### 从源码运行

```bash
# 稳定模式：构建前端后启动
./dev.sh

# 热更新模式：Vite watch + Wails
./dev.sh live
```

### 准备浏览器内核

1. 下载 [fingerprint-chromium](https://github.com/adryfish/fingerprint-chromium/releases)
2. 解压到任意目录
3. 打开应用 → 内核管理 → 新增内核 → 选择目录

---

## 项目结构

```
├── main.go                    # 应用入口，Wails 引导
├── config.yaml                # 运行时配置
├── wails.json                 # Wails 项目配置
├── dev.sh                     # 开发启动脚本
│
├── backend/                   # Go 后端
│   ├── app*.go                # App 核心逻辑（实例、代理、备份、书签…）
│   ├── bootstrap.go           # 后端初始化
│   ├── internal/
│   │   ├── browser/           # 浏览器域：Profile CRUD、内核、分组
│   │   ├── launchcode/        # Launch API 服务端 + Swagger
│   │   ├── proxy/             # 代理解析、测速、健康检测
│   │   ├── config/            # 配置类型与默认值
│   │   ├── database/          # SQLite 数据层
│   │   ├── logger/            # 日志系统（多写入器、拦截器）
│   │   ├── backup/            # 备份规格与路径
│   │   ├── snapshot/          # 快照归档
│   │   └── tray/              # 系统托盘
│   └── cmd/profile-recover/   # 实例配置恢复工具
│
├── frontend/                  # React 前端
│   └── src/
│       ├── modules/
│       │   ├── browser/       # 实例、代理、内核、插件、书签等页面
│       │   └── settings/      # 系统设置
│       ├── shared/            # 布局、组件、工具
│       └── config/            # 路由、导航、项目配置
│
├── build/                     # 构建资源（图标、平台清单）
└── publish/                   # 发布脚本（Linux / macOS / Windows / NSIS）
```

---

## Launch API

本地 HTTP 服务，默认端口 `19876`，供外部系统调用浏览器实例。

```bash
# 按 Code 启动实例
curl -X POST http://localhost:19876/api/launch \
  -H "Content-Type: application/json" \
  -d '{"code": "my-instance"}'

# 按名称匹配
curl -X POST http://localhost:19876/api/launch \
  -H "Content-Type: application/json" \
  -d '{"key": "店铺A"}'
```

**功能：** 按 Code / 名称 / 标签 / 分组启动 · 实例状态查询 · 停止实例 · DevTools 端口等待

**认证：** 可选 API Key（配置文件 `launch_server.auth.api_key`）

**文档：** 启动应用后访问 `http://localhost:19876/swagger/`

---

## 数据目录

运行时数据写入用户目录，不污染项目树：

| 平台 | 路径 |
|------|------|
| macOS | `~/Library/Application Support/facade` |
| Linux | `$XDG_DATA_HOME/facade` 或 `~/.local/share/facade` |
| Windows | `%LOCALAPPDATA%\facade` |

其下包含 `config.yaml`、`proxies.yaml`、`data/app.db`、日志、扩展与各实例用户数据目录。

---

## 开发

### 环境要求

- Go 1.25+
- Node.js 18+
- [Wails CLI](https://wails.io/docs/gettingstarted/installation)
- Linux: `libgtk-3-dev libwebkit2gtk-4.0-dev`
- macOS: Xcode Command Line Tools
- Windows: WebView2（Win10/11 自带）

### 常用命令

```bash
# 开发
./dev.sh              # stable 模式
./dev.sh live         # 热更新

# 打包
bash publish/linux/publish-linux.sh --arch amd64
bash publish/mac/publish-mac.sh --arch arm64
bash publish/windows/publish-windows.sh --arch amd64

# 工具
wails generate module                                     # 生成前端绑定
go run ./backend/cmd/profile-recover --app-root '/path'   # 恢复实例配置
```

### 分支说明

| 分支 | 用途 |
|------|------|
| `master` | 开发基线，不包含用户数据 |
| `user_data` | 历史演示分支 |

---

## 贡献

欢迎通过 [Issue](https://github.com/zzabolazza/Facade/issues) 和 Pull Request 参与改进。

- Bug 反馈：附带版本号、系统版本、复现步骤
- 功能建议：说明业务场景与预期行为
- 较大改动建议先开 Issue 对齐

---

## License

暂未附带独立 LICENSE 文件，后续会补充。

---

*感谢 [fingerprint-chromium](https://github.com/adryfish/fingerprint-chromium) 项目的支持。*
