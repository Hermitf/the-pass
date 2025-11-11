# The Pass

一个使用 Go (Gin) + React 构建的多角色外卖/配送平台示例，实现用户/商户/骑手/员工账户体系、统一认证、短信验证码、扫码登录、基础限流与日发送统计。

> 当前处于“核心功能成型 + 扫码登录接入阶段”，适合作为中型服务的技术脚手架示例。

## ✨ 特性

- 统一账户注册与登录（用户 / 商户 / 骑手 / 员工）
- JWT 鉴权中间件（可扩展角色/权限）
- 短信验证码：限流（滑动窗口）+ 每日上限 + Redis Lua 原子脚本
- 扫码登录：移动端二次确认（Ticket 状态机 pending → scanned → confirmed/rejected）
- 可插拔日志接口（Logger）与哨兵错误 (ErrStoreFailure)
- 前端 React + Vite（登录页、仪表盘占位）
- 配置热加载（viper watch），预留多环境能力
- 结构清晰的服务 / 仓储 / 中间件分层

## 🧱 技术栈

| 层级 | 技术 |
| ---- | ---- |
| Backend | Go 1.24, Gin, GORM (PostgreSQL), Redis, Viper, JWT |
| Frontend | React, TypeScript, Vite |
| Auth | JWT (golang-jwt), 自定义中间件 |
| Cache / Store | Redis (go-redis v9), Lua scripts for atomic operations |
| Dev / Tools | Docker Compose（待完善）, Swagger 注释（待生成） |

## 📁 目录结构

```
backend/
	internal/
		auth_qr/        # 扫码登录票据 + Redis 存储 + Actions
		handler/        # Gin 处理器（注册/登录/SMS/档案）
		middleware/     # JWT 中间件
		service/        # 业务服务层 (User/Employee/...)
		repository/     # 仓储层 (UserRepository + WithTx)
		model/          # 数据模型
		config/         # 配置结构
	pkg/
		sms/            # 短信存储 + Provider + Service (Lua 优化)
		auth/           # JWT 封装
		crypto/         # 密码哈希（bcrypt）
		validator/      # 简易校验
frontend/
	src/              # React 应用源码
docker-compose.yaml # Redis / Sentinel 等初始编排
README.md
```

## 🧬 架构概览

- Handler → Service → Repository → DB/Redis
- SMS Service 编排顺序：限流 → 每日上限 → 生成验证码 → 存储 → Provider 发送
- 扫码登录：PC 创建票据 → 轮询 → 手机端扫码标记 scanned → 用户确认 → confirmed/rejected → PC 获取结果

（可选：后续可在文档加入流程图）

## ⚙️ 后端运行

```fish
# 进入后端目录
cd backend

# 设置环境变量（示例，支持 viper.AutomaticEnv）
set -x THE_PASS_REDIS_HOST 127.0.0.1
set -x THE_PASS_REDIS_PORT 6379
set -x THE_PASS_DB_HOST 127.0.0.1
# 其他数据库/Redis配置根据 config.yaml 与实际环境补充

# 编译
go build ./...

# 运行（示例入口：cmd/server/main.go）
go run ./cmd/server/main.go
```

## 🛠 配置与环境变量

- 使用 `config.yaml` + `viper.AutomaticEnv`
- 后续计划：多环境配置（`config.dev.yaml` / `config.prod.yaml`）+ flag 覆盖
- 热加载：文件变化后自动重新 Unmarshal（当前缺少失败回退策略）

示例关键字段：
- Database: host / port / username / password / dbName
- Redis: host / port / password / poolSize / minIdleConns
- SMSRuntimeConfig: Enabled / ExpireIn / RateMax / RateWindow / DailyMax / Template

## 📲 短信验证码模块 (pkg/sms)

- 存储：`RedisStore`（code / rate_z / daily），支持 key 前缀用于多环境区分
- 限流：Lua 脚本原子执行（删旧 + 插入 + 计数 + 过期）
- 每日计数：Lua INCR + TTL（自然日结束，跨天自动重置）
- 日志：可插拔 Logger（默认 StdLogger），手机号脱敏输出
- 错误：`ErrSendTooFrequent` / `ErrDailyLimitReached` / `ErrStoreFailure` 等

调用示例：

```go
store := sms.NewRedisStore(redisClient)
provider := sms.NewMockProvider()
svc := sms.NewService(store, provider, sms.SMSRuntimeConfig{
	Enabled: true,
	ExpireIn: 5*time.Minute,
	RateMax: 1,
	RateWindow: time.Minute,
	DailyMax: 5,
	Template: "您的验证码是 %s",
})
err := svc.SendCode(ctx, "13800000000")
```

## 🔳 扫码登录流程 (internal/auth_qr)

状态机：
- pending: 初始票据（PC 生成）
- scanned: 手机端已扫码
- confirmed: 手机端授权成功（绑定用户）
- rejected: 手机端拒绝

关键方法：
- `CreateTicket` / `GetTicket` / `UpdateTicket`
- 高阶动作：`MarkScanned` / `Confirm` / `Reject`

待接入：HTTP 接口 + 前端轮询 + 移动端确认页面

## 🔐 认证与授权

- 登录支持：password / sms（oauth 预留）
- JWT 负载：userID + 角色（后续扩展设备/租户/刷新策略）
- 中间件：`middleware/jwt_auth.go` 提取 token → 校验 → 设置 `userID` 到 Gin Context
- 后续增强：角色/权限矩阵、失败次数限制、设备指纹

## 💻 前端运行

```fish
cd frontend
pnpm install
pnpm dev
# 默认端口 http://localhost:5173
```

TODO：
- 登录表单对接真实 API 与错误提示
- 扫码登录页面（展示二维码 + 轮询状态）
- Token 过期统一处理（拦截 401 → 清除 localStorage → 跳转 /login）

## 🧪 测试

现状：
- `pkg/sms/store_redis_test.go` 基础路径覆盖（miniredis）

计划：
- 添加 Lua 行为边界测试（频率窗口刚满、每日计数跨天）
- Service 层与 Handler 集成测试（Gin + httptest）
- auth_qr 状态流转单测与并发安全测试

运行：

```fish
go test ./... -count=1
```

## 🚀 部署

当前：
- `docker-compose.yaml` 包含 Redis/Sentinel（待补 Postgres + backend 服务）

计划：
- 添加 backend Dockerfile（多阶段构建：编译 + 最小运行镜像）
- CI (GitHub Actions)：lint + test + build + push image
- 环境分离：dev / staging / prod compose 叠加文件

## 🗺 Roadmap

| 阶段 | 项目 | 状态 |
| ---- | ---- | ---- |
| Auth | 登录增强（设备/失败次数限制） | TODO |
| QR Login | HTTP 接口 + 前端轮询 | 部分完成 |
| Config | 多环境配置/flag 支持 | TODO |
| JWT | 刷新/登出黑名单 | TODO |
| SMS | 结构化日志 + 分片键评估 | 进行中 |
| Tests | Service / Handler / QR 状态 | TODO |

完整进度见 `backend/TODO.md`

## 🤝 Contributing

1. Fork & Clone
2. 建议分支命名：`feat/qr-handler` / `fix/sms-ttl`
3. 提交前运行：`go test ./...`（后续将加入 golangci-lint）
4. 提交 PR 附带变更说明与动机

## 📄 License

本项目使用 Apache 2.0 License（见 `LICENSE`）。

## 📎 附录

### Redis 键命名规范

- `{prefix}:code:{phone}` 验证码
- `{prefix}:rate_z:{phone}` 频率窗口（ZSET）
- `{prefix}:daily:{YYYYMMDD}:{phone}` 当日计数

> 说明：`prefix` 默认 `sms`，建议按环境设定如 `dev:sms` / `prod:sms`；可进一步采用分片策略 `{prefix}:code:{shard}:{phone}`（`shard = hash(phone)%N`）改善集群热点。

### 性能说明

- 频率限制与每日计数采用 Lua 脚本，减少 3~5 次往返 → 1 次
- 后续可引入分片键与批量裁剪提升高并发性能

### Swagger

- 已在 Handler 注释中添加 swagger tag，需接入 swag CLI 生成 `/docs`（可选后续）
