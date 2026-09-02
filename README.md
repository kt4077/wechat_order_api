# 服务端 service_api（Golang + Gin + MySQL）

活动报名工具的接口服务，为小程序端与 PC 管理后台提供统一 API。

## 技术栈

| 组件 | 说明 |
|---|---|
| Gin | HTTP 框架，路由分组 + 中间件 |
| GORM | ORM，自动迁移、软删除、事务与行锁 |
| MySQL 8 | 业务数据存储 |
| PowerWeChat | 微信小程序能力（登录、订阅消息、手机号解析） |
| 七牛云 SDK | 图片对象存储（可切换本地磁盘） |
| excelize | 报名数据 Excel 导出 |
| zap | 结构化日志 |
| viper | 多环境配置 |

## 目录结构

```
service_api/
├── main.go                 # 入口：配置加载、数据库初始化、优雅退出
├── config/                 # 配置文件与加载逻辑
├── internal/
│   ├── router/             # 路由分组（公共 / 小程序端 / 管理端）
│   ├── middleware/         # 跨域、限流、鉴权
│   ├── controller/         # 参数接收、校验、结果返回
│   ├── service/            # 业务逻辑（报名、活动、入驻、统计、导出）
│   ├── model/              # 数据实体与数据库连接
│   └── dto/                # 请求 / 响应数据结构
├── pkg/
│   ├── crypto/             # AES 加密、脱敏、bcrypt
│   ├── jwtx/               # 用户端 / 管理端令牌
│   ├── wechat/             # PowerWeChat 封装
│   ├── upload/             # 七牛云 / 本地存储
│   ├── errcode/            # 统一错误码
│   ├── response/           # 统一响应
│   ├── logger/             # 日志
│   └── validate/           # 手机号、身份证、分页
└── runtime/                # 日志、缓存、上传（gitignore）
```

## 本地运行

```bash
# 1. 创建数据库
mysql -u root -p -e "CREATE DATABASE activity DEFAULT CHARACTER SET utf8mb4;"

# 2. 调整配置
vim config/config.yaml

# 3. 启动
go mod download
go run . -c config/config.yaml

# 4. 验证
curl http://127.0.0.1:8080/health
```

首次启动会自动建表、创建超级管理员（`super_admin` 配置节）与系统配置默认项。

## 编译

```bash
# 当前平台
go build -o bin/activity .

# Linux 生产包
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o bin/activity .
```

## 配置说明

见 `config/config.yaml`，关键项：

- `database.auto_migrate`：首次启动自动建表，生产稳定后建议关闭
- `oss.driver`：`qiniu` 使用七牛云，`local` 落盘到 `oss.local.root`
- `wechat.cache.driver`：多实例部署改为 `redis`
- `wechat.tmpl_*`：订阅消息模板 ID，留空则跳过微信推送（站内消息仍生成）
- `ratelimit`：单 IP 令牌桶限流，`enable: false` 可关闭

## 设计要点

- **防超卖**：`SubmitSignup` 在事务内对活动行加排他锁，校验名额后再写入，保证 1000+ 并发不超卖
- **权限实时生效**：活动发布权限以数据库 `role` + `merchant_flag` 实时校验，令牌内角色仅作展示，入驻通过或权限回收后立即生效
- **软删除**：所有业务表采用 `delete_time` 软删除，保留审计数据
- **隐私加密**：手机号、身份证使用 AES-CBC 加密入库，对外按需脱敏
- **操作留痕**：后台审核、权限变更、删除等操作统一写入 `operation_log`
