# Backend

Go module: `example.com/bridge-health-monitor-service`.

```bash
go test ./...
go build ./...
PORT=8080 go run .
```

The entrypoint is `main.go`; it starts the HTTP service. Health check: `GET /healthz`.
