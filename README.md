# arkSign — 森空岛自动签到

[![Go Version](https://img.shields.io/badge/Go-1.22-00ADD8?logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-GPLv3-blue.svg)](LICENSE)
[![Platform](https://img.shields.io/badge/Platform-Linux%20%7C%20macOS%20%7C%20Windows-brightgreen)](#下载)

基于 Go 语言编写的轻量化命令行签到工具，支持**明日方舟**与**明日方舟：终末地**每日自动签到。

---

## ✨ 特性

- **多游戏支持** — 同时支持《明日方舟》与《明日方舟：终末地》，自动识别账号绑定角色
- **定时任务** — 内置 cron 调度器，指定时间自动签到，无需外部 crontab
- **多账号** — 支持添加多个账号，逐一签到
- **加密存储** — 账号信息 AES-256-GCM 加密本地保存，无泄露风险
- **单文件部署** — 编译后为独立二进制文件，零依赖，开箱即用

---

## 📦 下载

从 [Releases](../../releases) 页面下载对应平台的二进制文件：

| 平台 | 架构 | 文件名 |
|---|---|---|
| Linux | x86_64 | `arkSign_linux_amd64` |
| Linux | ARM64 | `arkSign_linux_arm64` |
| macOS | x86_64 | `arkSign_darwin_amd64` |
| macOS | ARM64 | `arkSign_darwin_arm64` |
| Windows | x86_64 | `arkSign_windows_amd64.exe` |

---

## 🛠️ 编译安装

```bash
# 克隆仓库
git clone https://github.com/<your-repo>/arkSign.git
cd arkSign

# 安装依赖
go mod tidy

# 编译当前平台（开发用）
make dev

# 编译所有平台
make all

# 查看所有 make 目标
make help
```

---

## 🚀 快速开始

### 1. 添加账号

```bash
./arkSign -a
```

按提示输入手机号和密码，密码输入时回显 `*` 掩码。

### 2. 运行签到

```bash
# 立即执行一次
./arkSign -o

# 定时执行（每日 04:30）
./arkSign -t 04:30

# 查看版本
./arkSign -v
```

---

## 📖 命令行参数

| 参数 | 说明 | 示例 |
|---|---|---|
| `-a` | 交互式添加/更新账号 | `./arkSign -a` |
| `-o` | 立即执行一次签到后退出 | `./arkSign -o` |
| `-t HH:MM` | 每日定时签到时刻（默认 4:30） | `./arkSign -t 08:00` |
| `-v` | 显示版本信息 | `./arkSign -v` |

---

## 📁 数据存储

账号信息加密存储在 `configs/accounts.json`，使用 **AES-256-GCM** 认证加密。文件权限设为 `0600`（仅所有者可读写）。

> ⚠️ 请勿将 `configs/` 目录下的文件上传至公开平台或分享给他人。

---

## 🔧 开发

```bash
make fmt       # 格式化代码
make vet       # 静态检查
make lint      # 代码规范检查（需安装 golangci-lint）
make test      # 运行测试（含竞态检测）
make cover     # 查看测试覆盖率
make tidy      # 清理依赖
```

---

## ⚠️ 免责声明

本项目仅供学习交流使用，请勿用于任何非法用途。使用本工具产生的任何后果由用户自行承担。
