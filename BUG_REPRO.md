# BUG_REPRO

## Bug 是什么
`backend/ops` 的时钟、期限与报表存在零值路径缺陷：`OpsClock.Now` 未判空直接调用 nil 函数 panic；`OpsService.Deadline` 对不可解析时间戳返回零值且不报错；`Overdue` 忽略 Deadline 错误把坏时间戳事件判为逾期；`Report` 在空库时不初始化 `ByStatus/ByPriority`，返回 null map。

## 如何触发
在埋错基线运行：

```bash
go test ./ops -run '^TestClockNilNowFuncSafe$' -count=1
go test ./ops -run '^TestDeadlineInvalidStampErrors$' -count=1
go test ./ops -run '^TestOverdueSkipsUnparseableStamp$' -count=1
go test ./ops -run '^TestReportEmptyStoreMapsInitialized$' -count=1
```

## 错误信息
`TestClockNilNowFuncSafe` 直接 panic：

```
panic: runtime error: invalid memory address or nil pointer dereference [recovered]
[signal SIGSEGV: segmentation violation code=0x2 addr=0x0 pc=0x10296c8a8]

goroutine 35 [running]:
example.com/bridge-health-monitor-service/ops.OpsClock.Now(...)
    .../ops/ops_clock.go:12 +0x18
```

坏时间戳期限接口不报错、逾期列表把坏数据算进去、空库报表 `ByStatus/ByPriority` 为 null。
