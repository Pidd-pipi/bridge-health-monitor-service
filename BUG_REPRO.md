# BUG_REPRO

## Bug 是什么
监测事件复核状态机（`backend/ops/ops_state.go`、`ops_model.go`、`ops_review.go`）：`reviewing` 中间态没进 `opsStatusValid`，转换表缺 `active→reviewing` 与 `reviewing→closed` 边；`Review` 复核成功却把状态回写成 `active`；`Terminal()` 与 `opsStatusTerminal` 把 `reviewing` 当成终态。复核流程 400 失败、状态不落库、列表查不到 reviewing。

## 如何触发
在埋错基线运行：

```bash
go test ./httpapi -run '^TestReviewLifecycleFlow$' -count=1
go test ./httpapi -run '^TestReviewStatusPersisted$' -count=1
go test ./ops -run '^TestIntermediateStateKnown$' -count=1
go test ./ops -run '^TestReviewingNonTerminal$' -count=1
go test ./ops -run '^TestModelTerminal$' -count=1
```

## 错误信息
复核接口返回 400：`operations status transition is not allowed: active to reviewing`；`reviewing should be a valid status` / `reviewing should not be terminal` 断言失败。
