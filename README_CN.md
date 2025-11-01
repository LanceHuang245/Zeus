[English](README.md) | 简体中文

# Zeus - Zephyr 后端服务

Zeus 是 Zephyr 项目的高性能后端服务，基于 Gin 框架开发，提供天气数据 API 服务。项目支持多种天气数据源，包括和风天气(QWeather)和 OpenMeteo，具备完整的缓存机制和 TLS 安全支持;\
后续将让Zephyr允许使用您通过构建的Zeus服务器来获取数据。

## 主要特性

- **多数据源支持**: 集成 QWeather 和 OpenMeteo 天气数据源
- **高性能**: 基于 Gin 框架，支持高并发请求
- **安全可靠**: 支持 TLS 加密传输
- **智能缓存**: Redis 缓存机制，提升响应速度
- **城市搜索**: 支持全球城市搜索和地理位置查询
- **天气预警**: 实时天气预警信息推送
- **监控就绪**: 内置健康检查接口

## 项目架构

```
Zeus/
├── api_group/          # API 路由组
├── config/            # 配置管理
├── models/            # 数据模型
├── qweather/          # QWeather 数据源
├── openmeteo/         # OpenMeteo 数据源
├── openstreetmap/     # OpenStreetMap 集成
├── utils/             # 工具函数
├── cert/              # TLS 证书
└── bin/               # 构建输出
```

## 快速开始

### 环境要求

- Go 1.25+
- Redis 8.0+
- TLS 证书 (生产环境)

### 安装步骤

1. **克隆项目**
   ```bash
   git clone <repository-url>
   cd Zeus
   ```

2. **安装依赖**
   ```bash
   make deps
   ```

3. **配置环境变量**
   ```bash
   cp .env.example .env
   # 编辑 .env 文件，填入你的配置信息
   ```

4. **启动服务**
   ```bash
   make run
   ```

服务启动后，访问 `https://localhost:3899/api/v1/healthcheck` 验证服务是否正常运行。

## 配置说明

### 环境变量配置

| 变量名 | 说明 | 默认值 |
|--------|------|--------|
| `REDIS_ADDR` | Redis 地址 | `127.0.0.1:6379` |
| `REDIS_PASSWORD` | Redis 密码 | 空 |
| `REDIS_DB` | Redis 数据库 | `0` |
| `CACHE_TTL_MINUTES` | 缓存TTL(分钟) | `30` |
| `QWEATHER_PROJECT_ID` | QWeather 项目ID | - |
| `QWEATHER_KEY_ID` | QWeather Key ID | - |
| `QWEATHER_PRIVATE_KEY` | QWeather 私钥 | - |
| `QWEATHER_URL` | QWeather API地址 | `https://devapi.qweather.com/v7` |
| `SERVER_PORT` | 服务端口 | `:3899` |
| `ENABLE_TLS` | 启用TLS | `true` |
| `CERT_FILE` | TLS证书路径 | `./cert/zephyr.crt` |
| `KEY_FILE` | TLS私钥路径 | `./cert/zephyr.key` |

## 开发指南

### 构建项目

```bash
# 本地构建
make build

# 构建 Linux 版本
make build-linux

# 构建 Windows 版本
make build-windows

# 构建所有平台
make build-all
```

### 代码规范

```bash
# 格式化代码
make fmt

# 运行测试
make test

# 清理构建文件
make clean
```

## 贡献

1. Fork 项目
2. 创建功能分支
3. 提交变更
4. 推送到分支
5. 创建 Pull Request

## 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

---

⭐ 如果这个项目对你有帮助，请给个 Star 支持一下！