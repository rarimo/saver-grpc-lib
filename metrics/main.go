package metrics

import (
	"net/http"
	"net/http/pprof"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	WebsocketAvailable    = 0
	WebsocketDisconnected = 1
)

var (
	WebsocketMetric = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "rpc_websocket_status",
	})
)

func RunProfiling() {
	go profiling()
}

func profiling() {
	r := http.NewServeMux()
	r.HandleFunc("/debug/pprof/", pprof.Index)
	r.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	r.HandleFunc("/debug/pprof/profile", pprof.Profile)
	r.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	r.HandleFunc("/debug/pprof/trace", pprof.Trace)
	r.Handle("/metrics", promhttp.Handler())

	if err := http.ListenAndServe(":8080", r); err != nil {
		panic(err)
	}
}
