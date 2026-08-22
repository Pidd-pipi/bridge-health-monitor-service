# BUG_REPRO

## Bug 是什么
HTTP 请求上下文处理（`backend/runtime.go`、`httpapi/handler.go`）：服务端不再给请求加超时（`deadlineMiddleware` 直接透传），`X-Request-Deadline-Ms` 期限头被忽略；handler 复用固定的 `staleRequestCtx` 而不是当前请求的 context，取消与超时不再传播；ctx 错误被映射成 500 而非 504。

## 如何触发
在埋错基线运行：

```bash
go test . -run '^TestSearchHonorsClientCancel$' -count=1
go test . -run '^TestFetchDetailHonorsCancel$' -count=1
go test . -run '^TestHeaderDeadlineEnforced$' -count=1
```

## 错误信息
已取消请求仍被处理返回 200（应 504）；`X-Request-Deadline-Ms: 0` 不生效返回 200（应 504）。
