# BUG_REPRO

## Bug 是什么
桥梁存储与服务层零值路径（`backend/store/store.go`、`bridge/bridge.go`）：`Store.Get/UpdateCondition` 对缺失 id 返回零值结构与 nil 错误，`Service.History` 跳过存在性检查，缺失桥返回 200 空数据；`Service.RiskLevel` 用 `levels[risk/40]` 索引，风险分过大时越界 panic。

## 如何触发
在埋错基线运行：

```bash
go test ./httpapi -run '^TestFetchMissingBridgeRejects$' -count=1
go test ./httpapi -run '^TestUpdateMissingBridgeRejects$' -count=1
go test ./bridge -run '^TestBridgeRiskLevelBoundariesSafe$' -count=1
go test ./bridge -run '^TestBridgeHistoryMissing$' -count=1
go test ./store -run '^TestStoreUpdateMissingReturnsNotFound$' -count=1
```

## 错误信息
不存在的桥 `GET/POST` 返回 200 与空数据（应 404）；`RiskLevel(120)` 直接 panic：

```
panic: runtime error: index out of range [3] with length 3 [recovered]

goroutine 3 [running]:
example.com/bridge-health-monitor-service/bridge.(*Service).RiskLevel(...)
    .../bridge/bridge.go:49
```

不存在桥的历史记录不报错返回空列表（应 404）。
