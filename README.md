# Ant Browser

> 面向多账号隔离、代理绑定和本地环境管理的桌面浏览器工具（Windows / Linux / macOS unsigned）。

[![Release](https://img.shields.io/github/v/release/black-ant/Ant-Browser?sort=semver)](https://github.com/black-ant/Ant-Browser/releases)
[![Platform](https://img.shields.io/badge/Platform-Windows%20%7C%20Linux%20%7C%20macOS-blue)](https://github.com/black-ant/Ant-Browser/releases)
[![Issues](https://img.shields.io/github/issues/black-ant/Ant-Browser)](https://github.com/black-ant/Ant-Browser/issues)

## 推荐内核项目

Ant Browser 当前推荐配套使用的浏览器内核，来源于开源项目 [fingerprint-chromium](https://github.com/adryfish/fingerprint-chromium)。

如果你正在寻找可直接下载和维护的指纹内核版本，建议先查看它的 Releases 页面：

- <https://github.com/adryfish/fingerprint-chromium/releases>

这个项目为 Ant Browser 的内核准备提供了直接可用的基础来源，这里先对原项目做明确推荐与致谢。

Ant Browser 的目标很明确：在一台桌面设备上，帮助用户稳定管理多个彼此隔离的浏览器实例，并配合代理池、浏览器内核和快捷启动能力完成日常运营或测试工作。

## 目录

- [项目简介](#项目简介)
- [近期更新](#近期更新)
- [更新日志](CHANGELOG.md)
- [核心特性](#核心特性)
- [界面预览](#界面预览)
- [快速开始](#快速开始)
- [常用操作](#常用操作)
- [常见问题](#常见问题)
- [Roadmap](#roadmap)
- [贡献](#贡献)
- [支持与反馈](#支持与反馈)
- [License](#license)

## 项目简介

Ant Browser 适合以下场景：

- 多账号环境隔离
- 跨境电商与社媒账号运营
- 需要独立代理出口的本地测试
- 需要统一管理浏览器内核和实例配置的团队

这个项目当前提供的核心价值是：

- 给每个账号分配独立浏览器实例
- 给每个实例绑定独立代理
- 统一管理浏览器内核、标签、关键字和快捷打开码
- 在本地保存配置和运行数据，便于自主控制

## 近期更新

### 1.3.0 · 2026-06-23

- 插件管理：新增插件包管理能力，支持插件安装、导入、启停、删除、实例限制和单实例插件配置
- 代理优化：统一使用 Chromium 原生代理链接（`direct://` / `http://` / `https://` / `socks5://`），完善代理测速、IP 健康检测和代理异常处理
- 实例迁移：支持实例导入导出，可将实例配置和完整浏览器用户数据目录打包迁移到新环境
- 代理适配：实例导入时按代理名称匹配本地同名代理，匹配不到或同名不唯一时自动清空代理
- 界面优化：优化实例列表、关键字展示、操作菜单和导入导出入口，减少页面拥挤和无效信息

### 1.2.0 · 2026-05-09

- 重点升级接口调用：Launch API 补齐实例增删改查、按 code / selector 启动、runtime session / status / stop ，方便外部系统直接调用浏览器能力；实例 CDP 使用直连调试端口
- 增强代理池：优化原生代理链接的导入、编辑和预览能力，支持 HTTP(S) / SOCKS5 / 直连节点批量导入
- 优化代理检测：新增测速目标、IP 健康检测目标和检测超时配置
- 改进实例启动：代理异常时支持本次直连启动，不修改实例原有代理配置；默认代理池只保留直连节点
- 升级书签能力：新增 IP 检测站点默认书签，支持设置启动时自动打开，并可同步到已有未运行实例

### 1.1.0 · 2026-03-19

- 完善 Linux 支持：补齐 Linux 环境下的开发、打包、安装、启动与运行链路，并持续修复安装版启动与退出稳定性问题
- 补齐 macOS unsigned 内测构建链路：支持在原生 macOS 主机上打包 `.app` / `.zip`，并将用户状态目录放到 `~/Library/Application Support/ant-browser`
- 开发与安装版均将应用状态与实例数据写入用户 home 下的状态根目录，不再写入项目目录
- 新增 SOCKS 代理测试支持：SOCKS 代理能力已进入测试阶段，后续会继续验证稳定性与兼容性
- 实验性支持接口触发浏览器：支持通过接口启动浏览器实例，便于外部系统接入

完整历史版本记录见 [CHANGELOG.md](CHANGELOG.md)。

## 源码分支说明

- `master`：面向开发者的干净基线分支，不提交用户数据库、实例目录或其他运行时数据。首次启动时会在用户状态根目录自动初始化。
- `user_data`：历史演示分支；运行时数据现已统一落到 home 状态根，不再依赖仓库内 `data/`。

## 运行时数据目录

开发和打包运行都会把可写状态放到用户目录（不写项目树）：

| 平台 | 状态根目录 |
|------|------------|
| macOS | `~/Library/Application Support/ant-browser` |
| Linux | `$XDG_DATA_HOME/ant-browser` 或 `~/.local/share/ant-browser` |
| Windows | `%LOCALAPPDATA%\ant-browser` |

其下包含 `config.yaml`、`proxies.yaml`、`data/app.db`、日志、扩展与各实例用户数据目录。浏览器内核由「内核管理」自行注册，不打包、也不再使用仓库内 `chrome/` 占位目录。

## 核心特性

- 实例隔离管理：支持创建、编辑、启动、停止、重启、克隆和删除浏览器实例
- 代理池配置：支持统一维护原生代理链接，并将代理分配到具体实例
- 多协议支持：支持 `direct://` / `http://` / `https://` / `socks5://`
- 内核管理：支持维护多个 Chrome 内核版本，并设置默认内核
- 快捷启动：支持通过实例 Code 和 `Ctrl + K` 快速打开目标实例
- 标签与检索：支持按标签、关键字、状态、代理、内核、分组进行筛选
- 插件管理：支持插件安装、导入、启停、删除、实例限制和单实例插件配置
- 实例迁移：支持将实例配置和浏览器用户数据目录导出为 ZIP，并导入为新实例
- VPN / 代理检测：支持测速、IP 健康检测和代理异常处理
- 本地化存储：配置和实例数据保存在本地，适合长期使用和备份

## 界面预览

### 1. 实例列表

<img src="images/readme/002-实例列表.png" alt="实例列表" width="100%" />

对应功能点：

- 统一查看和管理所有浏览器实例
- 按状态、代理、内核、分组、关键字筛选实例
- 支持 `新建配置`、启动、停止、重启、配置、克隆、删除
- 给实例分配快捷打开码，后续可以直接快速启动

### 2. 代理配置

<img src="images/readme/003-设置代理池.png" alt="代理配置" width="100%" />

对应功能点：

- 统一管理代理节点
- 支持按协议、分组筛选代理
- 支持录入原生代理链接（`direct://` / `http://` / `https://` / `socks5://`）
- 支持查看延迟、IP 健康并挑选可用节点

代理规则：

- 指纹浏览器只使用 Chromium 原生代理链接，不再内置或管理 xray / sing-box / mihomo。
- 高级协议节点不可用于启动、测速与 IP 健康检测。

### 3. 代理生效验证

<img src="images/readme/004-自定义代理.png" alt="代理生效验证" width="100%" />

对应功能点：

- 启动实例后访问 IP 检测网站验证代理是否真正生效
- 检查 IP 地区、ASN、运营商和风险值等信息
- 用于确认当前实例是否已经走目标代理出口

## 快速开始

### 环境要求

- 操作系统：
  - Windows 10 / 11（64 位）
  - Linux（amd64 / arm64）
  - macOS（amd64 / arm64，当前为 unsigned 内测包）
- 建议内存：8 GB 及以上
- 建议磁盘空间：2 GB 以上

### 下载与运行

1. 前往 Releases 页面下载最新版本：<https://github.com/black-ant/Ant-Browser/releases>
2. 安装版直接运行 `AntBrowser-Setup-*.exe`
3. 便携版解压后运行 `ant-chrome.exe`
4. Linux：推荐用 `sudo apt install ./ant-browser_<version>_<arch>.deb` 安装（会自动拉取 `libwebkit2gtk-4.1-0` 等运行时依赖）；勿仅用 `dpkg -i`（不会从软件源安装依赖）。也可解压 `tar.gz` 后运行 `ant-chrome`，但需自行安装 WebKitGTK 运行时
5. macOS unsigned 包解压后运行 `AntBrowser-<version>-macos-<arch>.app`；如被 Gatekeeper 拦截，请对本机测试包执行 `xattr -dr com.apple.quarantine <app路径>` 后再打开

### 从源码运行

1. 开发默认使用 `master` 分支；该分支不带测试用户数据，适合作为日常开发基线。
2. 如需带测试库的演示环境，请切换到 `user_data` 分支。
3. 在 macOS / Linux 上执行 `./dev.sh`；默认是 `stable` 静态资源模式，热更新开发使用 `./dev.sh live`。

开发模式说明：

- `./dev.sh` / `./dev.sh stable`：先构建 `frontend/dist`，再以静态资源模式启动 Wails，不依赖外部 Vite dev server
- `./dev.sh live`：启动 Vite watcher，并通过 `-frontenddevserverurl` 接入桌面壳
- 如需为依赖下载配置代理，可在启动前设置 `DEV_PROXY_URL`、`DEV_NO_PROXY`、`DEV_GOPROXY`

开发入口（`./dev.sh`）仅提供 macOS / Linux shell 脚本。发布打包支持 macOS / Linux / Windows；Windows 打包脚本同样以 bash 编写，需在原生 Windows（Git Bash / MSYS）或 Windows CI runner 上执行。

### Linux 发布打包（源码）

Linux 发布脚本位于 `publish/linux/`。

```bash
bash publish/linux/publish-linux.sh --arch amd64
bash publish/linux/publish-linux.sh --arch arm64
```

详细说明见 [publish/linux/README.md](publish/linux/README.md)。

### macOS unsigned 发布打包（源码）

macOS 发布脚本位于 `publish/mac/`，必须在原生 macOS 主机上执行，且目标架构需与主机架构一致。

```bash
bash publish/mac/publish-mac.sh --arch amd64
bash publish/mac/publish-mac.sh --arch arm64
```

脚本会生成 unsigned `.app` 和 `.zip`，适合 PR 验证与内部测试。详细说明见 [publish/mac/README.md](publish/mac/README.md)。

### Windows 发布打包（源码）

Windows 发布脚本位于 `publish/windows/`，需在原生 Windows 主机（Git Bash / MSYS）上执行；安装包步骤依赖 `makensis`（`choco install nsis`）。

```bash
bash publish/windows/publish-windows.sh --arch amd64
bash publish/windows/publish-windows.sh --arch arm64
```

脚本会生成便携版 `.zip`，并在检测到 `makensis` 时额外生成 NSIS 安装包 `.exe`。详细说明见 [publish/windows/README.md](publish/windows/README.md)。

### 其他源码工具

从现有用户数据目录补回丢失的实例配置：

```bash
go run ./backend/cmd/profile-recover --app-root '/path/to/Ant Browser'
go run ./backend/cmd/profile-recover --app-root '/path/to/Ant Browser' --apply
```

生成 / 刷新 Wails 前端绑定：

```bash
wails generate module
```

### 准备浏览器内核

1. 打开应用，进入 `指纹浏览器 > 内核管理`
2. 点击页头链接前往 Releases 下载 fingerprint-chromium
3. 解压到任意目录后，点击 `新增内核` → `选择目录` 注册

确保所选目录下存在当前平台的浏览器可执行文件。

### 第一次使用建议流程

1. 在 `代理池配置` 中先导入或新增可用代理节点
2. 在 `实例列表` 中点击 `新建配置`
3. 选择实例名称、内核、代理、标签和需要的启动参数
4. 返回实例列表，点击启动按钮运行实例
5. 打开 IP 检测网站，确认代理结果是否符合预期

## 常用操作

| 目标 | 入口 | 说明 |
| --- | --- | --- |
| 新建浏览器实例 | `实例列表 > 新建配置` | 创建一个新的独立浏览器环境 |
| 配置代理池 | `代理池配置` | 维护代理节点并检查延迟、健康状态 |
| 绑定实例代理 | `实例编辑页` | 给指定实例分配目标代理节点 |
| 启动实例 | `实例列表` | 单击启动按钮即可运行目标实例 |
| 快速打开实例 | `Ctrl + K` | 可按 Code、实例名、标签、关键字快速检索 |
| 管理浏览器内核 | `内核管理` | 新增、编辑、删除和设置默认内核 |
| 验证代理结果 | 启动实例后访问 IP 检测网站 | 核对 IP、地区、ASN、风险值 |

## 常见问题

### 1. 应用无法启动怎么办？

先检查浏览器内核路径是否有效，并确认目标目录下存在 `chrome.exe`。

### 2. 实例启动了但代理没有生效怎么办？

先检查代理节点本身是否可用，再确认该实例已经正确绑定代理。建议启动后访问 IP 检测网站复核当前出口。

仅原生代理链接可用；vmess/vless/hysteria2 等高级协议不会被启动或测速。

### 3. 实例太多，怎么快速找到目标实例？

可以在 `实例列表` 中按状态、代理、内核、分组、关键字筛选，也可以通过 `Ctrl + K` 使用实例 Code 或名称快速启动。

### 4. 多个账号怎么避免串号？

建议采用一账号一实例、一实例一稳定代理的方式，不要混用浏览器环境，也不要频繁切换同一实例的出口 IP。

## Roadmap

- 持续补充使用文档和接口说明
- 增强实例模板、批量管理和检索体验

## 贡献

欢迎通过 Issue 和 Pull Request 参与改进。

- Bug 反馈：请附带版本号、系统版本、复现步骤和截图
- 功能建议：请说明业务场景、预期行为和现有问题
- 文档优化：欢迎直接提交 README、教程和截图说明相关改进

如果是较大改动，建议先开 Issue 对齐需求再提交 PR。

## 支持与反馈

- Releases：<https://github.com/black-ant/Ant-Browser/releases>
- Issues：<https://github.com/black-ant/Ant-Browser/issues>
- 感谢以下社区的支持：<https://linux.do/>

## License

当前仓库暂未附带独立的 `LICENSE` 文件，后续会补充。
