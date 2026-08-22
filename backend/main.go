package main

import (
	"example.com/bridge-health-monitor-service/bridge"
	"example.com/bridge-health-monitor-service/config"
	"example.com/bridge-health-monitor-service/health"
	"example.com/bridge-health-monitor-service/httpapi"
	"example.com/bridge-health-monitor-service/ops"
	"example.com/bridge-health-monitor-service/store"
	"example.com/bridge-health-monitor-service/web"
	"log"
	"net/http"
)

func main() {
	c := config.Load()
	m := http.NewServeMux()
	m.HandleFunc("/healthz", health.Handler)
	m.Handle("/api/v1/", httpapi.New(bridge.New(store.New()), ops.NewService(ops.SeedRecords())))
	m.HandleFunc("/", web.Handler)
	log.Printf("bridge monitor listening on %s", c.Address())
	log.Fatal(serveAddress(c.Address(), m))
}
