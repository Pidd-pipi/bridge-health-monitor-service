# BUG_REPRO

## Bug 是什么
审计与批量关闭（`backend/ops/ops_audit.go`、`ops_batch.go`）：`OpsAudit` 的读方法（`Since/Count/Latest/For`）用 `Lock` 配 `defer RUnlock`，调用即 panic；`BatchClose` 错误分支不释放租约（lease 泄漏），且关闭成功时写入的审计类型错误（`status_changed` 而非 `batch_closed`）。

## 如何触发
在埋错基线运行：

```bash
go test ./ops -run '^TestAuditSinceNoPanic$' -count=1
go test ./ops -run '^TestAuditCountNoPanic$' -count=1
go test ./ops -run '^TestAuditLatestNoPanic$' -count=1
go test ./ops -run '^TestAuditForNoPanic$' -count=1
go test ./ops -run '^TestBatchClosePartialFailureKeepsWorking$' -count=1
go test ./ops -run '^TestBatchCloseRecordsAudit$' -count=1
```

## 错误信息
审计读方法直接 panic：`sync: RUnlock of unlocked RWMutex`；批量关闭夹杂失败项后剩余有效项因租约耗尽超时失败；审计事件里找不到 `batch_closed` 类型。
