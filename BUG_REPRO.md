# BUG_REPRO

## Bug 是什么
策略接口的错误链与状态码映射（`backend/ops/ops_policy.go`、`ops_errors.go`、`ops_http.go`）：`EvaluatePolicy` 用 `%v` 包装丢失 sentinel，`opsCode` 改用直接相等判断导致 `errors.Is` 失效，`StatusForError` 把策略拒绝映射成 500；`RequireReview` 未做任何检查，critical 事件不满足复核标签也能直接关闭。

## 如何触发
在埋错基线运行：

```bash
go test ./httpapi -run '^TestPolicyUnknownEventReturns404$' -count=1
go test ./httpapi -run '^TestPolicyMissingLabelsReportsUnsatisfied$' -count=1
go test ./httpapi -run '^TestTransitionCriticalWithoutReviewRejected$' -count=1
```

## 错误信息
查询不存在事件的策略接口返回 500（应 404）；缺少复核标签的事件策略返回 500（应 200 且 `satisfied:false`）；critical 事件未复核直接关闭返回 200（应 403）。
