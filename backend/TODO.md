# 后端任务列表 (全局视角)

## 阶段一：项目基础架构

### 模块：项目设置与配置
- [ ] 完善 `config.yaml` 结构，支持多环境（开发/测试/生产）。（当前仅单一配置文件，ENV 覆盖已通过 viper.AutomaticEnv 初步支持）
- [~] 检查并完善配置热加载逻辑 (`global/app.go`)。（已有 watchConfig，缺少：错误回退策略、多环境切换、动态连接资源重载）
- [ ] 添加对命令行参数的支持，以便覆盖配置文件中的部分选项。（尚未实现 flag 层）

### 模块：数据库与数据模型
- [x] 设计并确认所有核心模型（`user`, `merchant`, `rider`, `employee`）的数据库表结构。（模型文件与 AutoMigrate 已执行）
- [ ] 编写数据库迁移脚本（例如使用 `golang-migrate`，当前仅使用 `AutoMigrate`，缺少版本化回滚能力）。
- [~] 为所有模型实现基础的仓储层（Repository）方法（CRUD）。（User 仓库较完整，其它模型存在基础文件但仍需补充 CRUD 与查询测试）
- [~] 在 `database/manager.go` 中引入数据库事务管理机制。（User 侧通过 `repo.WithTx` 已使用事务封装，需抽象公共层与跨仓库事务示例）

### 模块：通用包 (pkg)
- [~] **pkg/crypto**: 完善密码加密/比对的逻辑，确保安全性。
	- 结构：已拆分为 `config.go` / `hash.go` / `verify.go` / `limiter.go` / `errors.go`，并补充流程性中文注释；对外 API 保持不变（向后兼容）。
	- 进度：已完成动态 bcrypt cost、pepper 支持（向后兼容验证）、可选尝试次数限制（ErrTooManyAttempts）。
	- 待办：补充单元测试；在 service 层接入 limiter 策略与结构化审计日志（替换 log.Printf 为可插拔 Logger）。
- [ ] **pkg/validator**: 添加更多自定义校验规则以满足业务需求。（目前仅基本手机号/邮箱校验）
- [~] **pkg/auth (JWT)**: 确认 JWT 的负载（Payload）与扩展能力（多算法、可扩展声明、多租户隔离）。
	- 现状：当前包含 userID + 角色字符串；缺少刷新策略、设备/租户字段与多算法支持。
	- 子任务（拆分与扩展）：
		- [ ] errors.go：错误哨兵分类（ErrTokenExpired/Invalid/Revoked/Algorithm/ClaimsInvalid 等）。
		- [ ] config.go：JWTConfig 扩展（Algorithm/Issuer/Audience/Leeway/MultiTenant），支持 env/viper 加载。
		- [ ] claims.go：基础 Claims + 可扩展自定义声明结构；提供 BuildUserClaims 工厂。
		- [ ] signer.go + algorithms.go：抽象 Signer 接口；实现 HS256/RS256/ECDSA。
		- [ ] service.go：IssueAccessToken / IssueRefreshToken / ParseAndValidate / RotateRefresh。
		- [ ] blacklist.go：TokenRevocationStore 接口 + Redis 实现（按 jti 存储，TTL=exp-now）。
		- [ ] refresh.go：RefreshToken 模型设计与旋转/失效机制（防重放）。
		- [ ] key_provider.go：多租户密钥提供者（按 tenantID 返回密钥/密钥对，带缓存与热更新）。
		- [ ] middleware.go：从 Header/Cookie 抽取令牌、Parse、注入上下文（用户/租户/角色）。
		- [ ] 结构化日志与 metrics：签发/验证失败原因、撤销命中、租户命中缓存等。
		- [ ] 单元测试与集成测试（覆盖 JTI 撤销、多算法、租户、过期/提前/受众校验）。
	- 演进顺序（建议）：
		1) 错误哨兵 + 验证函数加入 ctx 预留黑名单/审计
		2) 抽象 Signer（先保留 HS256 实现）
		3) 增加 Issuer/Audience/NotBefore/Subject 支持
		4) 引入 Refresh + Blacklist（若需要登出/撤销）
		5) 多租户 KeyProvider（有隔离需求时接入）

## 阶段二：核心功能开发

### 模块：用户与认证 (Auth)
- [x] 实现用户注册业务逻辑 (`service/user_service.go`) 及 API 接口 (`handler/auth_handler.go`)。（含短信验证码校验与事务写入）
- [x] 实现用户登录（手机/密码 + 预留 sms/oauth）业务逻辑及 API 接口。（`LoginUser` 已支持 password / sms，oauth 类型占位返回错误）
- [~] 实现 JWT 认证中间件 (`middleware/jwt_auth.go`)，保护需要授权的路由。（基本验证完成；待补角色/权限校验与租户扩展）
- [~] 支持扫码登录并在移动端二次确认流程。（`internal/auth_qr` 已实现 Ticket + Store + Actions；缺少 HTTP Handler + 前端轮询 + 移动端确认/拒绝接口）
- [ ] **(可选)** 实现 JWT 刷新（Refresh Token）机制。（未开始）
- [ ] **(可选)** 实现用户登出功能（例如：基于 Redis 的 Token 黑名单）。

### 模块：短信服务 (SMS)
- [x] **(已完成)** 基于 Redis 实现验证码存储、限流及日上限统计。（含 ctx 版本接口与基础单测）
- [~] 为 Redis 操作添加更详细的错误处理和日志记录。（已加入可插拔 Logger 与 ErrStoreFailure；后续统一结构化/脱敏与错误分级）
- [ ] 评估当前的 Redis 键命名约定是否能适应大规模数据集。（需列键模式 & 预估数量级/分片策略）
- [x] 优化高并发场景下的 Redis Pipeline 操作。（已改为 Lua 脚本原子化合并操作）
- [x] 确保每日计数的递增操作具有原子性。（Lua 中设置 TTL 与 INCR 一次完成）

### 模块：核心业务 (商户/骑手/员工)
- [ ] **Merchant**: 核心业务逻辑与服务未实现（仅模型 + 注册占位）。
- [ ] **Rider**: 核心业务逻辑与服务未实现。
- [ ] **Employee**: 核心业务逻辑与服务未实现（仅注册和添加员工接口）。
- [ ] 为以上各模块设计并实现对应的 API 接口 (`handler`)。（部分注册/添加员工接口存在，需补 CRUD / 状态流转）

## 阶段三：测试与部署

### 模块：测试
- [ ] 为所有 `service` 层的方法编写单元测试。（当前仅 sms store 测试）
- [ ] 为所有 API `handler` 编写集成测试。（未开始）
- [~] **(SMS)** 扩展单元测试，覆盖 Redis 存储方法的边缘情况。（基础路径已测；缺失异常/并发/TTL 竞态用例）
- [ ] **(SMS)** 针对高并发场景添加集成测试。（未开始）

### 模块：部署
- [ ] 编写 `Dockerfile` 以便容器化部署。（尚未存在 backend 独立 Dockerfile）
- [ ] 完善 `docker-compose.yaml`，用于本地开发环境的快速启动。（已有基础 Redis/Sentinel 配置，缺少应用服务编排与环境变量）
- [ ] 编写 CI/CD 脚本（例如使用 GitHub Actions）。（未开始）

---
标记说明：
- [x] 已完成
- [~] 部分完成（核心可用，仍有增强项）
- [ ] 未开始 / 待实现

后续优先级建议：
1) 登录增强（在 service 接入密码尝试限制 + 审计日志）& 扫码登录 HTTP 接口 (P1)
2) 配置多环境 & JWT 刷新/Signer 抽象初步 (P2)
3) 键命名评估与大规模分片策略设计 (P2)
4) Service/Handler 测试矩阵 (P2)
5) 角色/权限 & 文档/部署 (P3)
