# BUG_REPRO

## Bug 是什么
事件检索的过滤、分页与标签合并逻辑（`backend/ops/ops_filter.go`、`ops_query.go`）复用调用方切片底层数组或原地改写映射：`FilterRecords` 用 `items[:0]` 原地压缩，`Paginate` 直接切片返回（共享底层数组），`opsClonePage` 不拷贝，`MergeLabels` 原地改 base map。先过滤再翻页后，原列表与页面数据互相串场。

## 如何触发
在埋错基线运行：

```bash
go test ./ops -run '^TestSearchFilterKeepsSourceIntact$' -count=1
go test ./ops -run '^TestPageDoesNotShareBackingArray$' -count=1
go test ./ops -run '^TestPageCloneDetached$' -count=1
go test ./ops -run '^TestMergeLabelsDoesNotMutateBase$' -count=1
```

## 错误信息
断言失败：过滤后源切片内容被改写、分页返回的条目修改会透传到原始列表、克隆页与源共享底层数组、合并后 base 映射被污染。
