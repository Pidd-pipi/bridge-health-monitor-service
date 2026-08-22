# BUG_REPRO

## Bug 是什么
桥梁报表与导出（`backend/bridge/report.go`、`export.go`）：`Report` 用直接相等判断包裹 `store.ErrNotFound`，`errors.Is` 失效导致缺失 id 被当作致命错误返回 500；重复 id 未去重被重复计数；`ExportCSV` 用直接相等判断并静默跳过缺失 id，导出少行。

## 如何触发
在埋错基线运行：

```bash
go test ./httpapi -run '^TestSummaryCountsMissing$' -count=1
go test ./httpapi -run '^TestSummaryDeduplication$' -count=1
go test ./httpapi -run '^TestExportHasMissingRow$' -count=1
```

## 错误信息
报表接口传 `ids=br-201,nope` 返回 500（应 200 且 `missing:1`）；`ids=br-201,br-201` 报表 `total:2`（应 1）；导出 `ids=br-201,nope` 缺少 `nope,,,missing,0,0` 行。
