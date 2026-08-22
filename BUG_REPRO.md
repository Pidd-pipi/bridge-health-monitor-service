# BUG_REPRO

## Bug 是什么
监测事件库（`backend/ops`）的读路径不加锁且直接返回内部引用：`OpsStore.Get/List` 返回了 map 里的原始值，`OpsService.Snapshot` 复用包级共享 map。并发读写事件库时产生 data race，读取方拿到的事件与快照会被后续写入污染，列表会缺条目或串数据。

## 如何触发
在埋错基线运行：

```bash
go test -race ./ops -run '^TestLedgerConcurrentReadWriteStable$' -count=1
go test ./ops -run '^TestLedgerGetCopyIsolated$' -count=1
go test ./ops -run '^TestLedgerListCopiesIsolated$' -count=1
go test ./ops -run '^TestLedgerUpdateIsolated$' -count=1
go test ./ops -run '^TestLedgerSnapshotFreshPerCall$' -count=1
```

## 错误信息
`go test -race` 报 DATA RACE（concurrent map read and map write）；克隆/快照用例失败：修改返回值的 labels 后再次读取仍能看到修改（内部引用逃逸），快照 map 被上一次调用污染。
