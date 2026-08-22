# BUG_REPRO

## Bug 是什么
桥梁状态迁移（`backend/domain/models.go`、`validation/validation.go`）：`conditionTransitions` 转换表缺失 `monitored→watch`、`restricted→cleared` 两条合法边，却混入了非法的 `restricted→monitored`；`validation.Transition` 额外拒绝同状态更新。合法迁移报错、非法迁移放行、重复提交被拒。

## 如何触发
在埋错基线运行：

```bash
go test ./httpapi -run '^TestBridgeWatchStateUpdate$' -count=1
go test ./httpapi -run '^TestBridgeClearAfterRestrict$' -count=1
go test ./httpapi -run '^TestBridgeIllegalConditionReject$' -count=1
go test ./httpapi -run '^TestBridgeSameConditionAllow$' -count=1
```

## 错误信息
`monitored→watch`、`restricted→cleared` 返回 400（应 200）；`restricted→monitored` 返回 200（应 400）；同状态更新 `watch→watch` 返回 400（应 200）。
