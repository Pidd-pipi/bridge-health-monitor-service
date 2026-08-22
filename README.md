# 桥梁健康监测服务（bridge-health-monitor-service）

标准库实现的桥梁巡检与结构监测事件处置服务，默认监听 `8080`，可用 `PORT` 环境变量覆盖。
服务包含两个业务域：

1. **桥梁台账**：`GET /api/v1/bridges` 列出桥梁，`POST /api/v1/bridges/{id}/condition`
   提交现场评估（带版本校验与状态流转约束），另提供 `GET /api/v1/bridges/{id}`、
   `GET /api/v1/bridges/{id}/history`、`GET /api/v1/bridges/report`、`GET /api/v1/bridges/export`。
2. **监测事件处置**：`POST /api/v1/events` 登记监测事件，`GET /api/v1/events` 检索，
   `POST /api/v1/events/{id}/transition` 状态流转（queued/active/reviewing/paused/closed，乐观并发），
   以及 `audit`、`policy`、`deadline`、`report`、`export`、`overdue`、`batch-close` 等接口。

桥梁状态值包括 `monitored`、`watch`、`restricted`、`cleared`，状态迁移受业务规则约束。
监测事件带优先级（low/normal/high/critical）与标签（site/operator/evidence/reviewed），
critical 事件关闭前必须完成复核。根路径为内置前端页面。

## 目录结构

```text
.
├── backend/
│   ├── main.go / runtime.go        # 启动、中间件（超时、请求 ID、恢复）
│   ├── httpapi/                    # HTTP 路由与处理器（bridges + events）
│   ├── bridge/                     # 桥梁台账服务（状态流转、报表、导出）
│   ├── ops/                        # 监测事件域（存储、状态机、审计、策略、期限、批量）
│   ├── store/ validation/ domain/  # 桥梁数据、校验、领域模型
│   ├── web/                        # 内置前端页面（index.html / app.js）
│   └── config/ health/             # 配置与健康检查
├── runtime_smoke.json              # 启动契约（服务模式 + 健康检查）
└── README.md
```

## 运行与测试

```bash
cd backend
go build ./...
go test ./...
PORT=8080 go run .
```

健康检查：`GET /healthz` 返回 `{"service":"bridge-health-monitor","status":"ok"}`。

## 环境变量

| 变量 | 说明 | 默认 |
|---|---|---|
| `PORT` | 监听端口 | `8080` |

## API 示例

```bash
curl -s localhost:8080/api/v1/bridges
curl -s -X POST localhost:8080/api/v1/bridges/br-201/condition -d '{"condition":"watch","expected_revision":1}'
curl -s -X POST localhost:8080/api/v1/events -d '{"subject":"支座锈蚀复核","owner":"赵工","priority":"high","labels":{"site":"东江大桥","operator":"赵工","evidence":"ph-9"}}'
curl -s localhost:8080/api/v1/events?priority=critical
curl -s -X POST localhost:8080/api/v1/events/evt-1001/review -d '{"revision":3}'
curl -s localhost:8080/api/v1/events/report?days=7
```

请求可携带 `X-Request-ID`（透传/生成）、`X-Operator`（操作人）与 `X-Request-Deadline-Ms`（覆盖单请求超时）。
