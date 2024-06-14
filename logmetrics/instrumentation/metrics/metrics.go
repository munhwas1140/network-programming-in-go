package metrics

import (
	"flag"

	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/prometheus"
	prom "github.com/prometheus/client_golang/prometheus"
)

var (
	Namespace = flag.String("namespace", "web", "metrics namespace")
	Subsystem = flag.String("subsystem", "server1", "metrics subsystem")

	// 카운터
	// 사용자 요청 카운트, 에러카운트, 혹은 완료된 카운트 등
	// 단순히 증가하는 값들을 추적하기 위해 사용된다.
	Requests metrics.Counter = prometheus.NewCounterFrom(
		prom.CounterOpts{
			Namespace: *Namespace,
			Subsystem: *Subsystem,
			Name:      "request_count",
			Help:      "Total requests",
		},
		[]string{},
	)

	WriteErrors metrics.Counter = prometheus.NewCounterFrom(
		prom.CounterOpts{
			Namespace: *Namespace,
			Subsystem: *Subsystem,
			Name:      "write_errors_count",
			Help:      "Total write errors",
		},
		[]string{},
	)

	// 게이지
	// 현재 메모리 사용량이나 현재 처리중인 요청의 개수,
	// 큐의 크기, 팬의 속도 처럼 증가하거나 감소하는 값을 추적할 수 있다.
	// 게이지는 카운터와 달리 분단 연결 개수나 초당 메가바이트 전송 등
	// 비율 계산을 지원하지 않는다.
	OpenConnections metrics.Gauge = prometheus.NewGaugeFrom(
		prom.GaugeOpts{
			Namespace: *Namespace,
			Subsystem: *Subsystem,
			Name:      "open_connections",
			Help:      "Current open connections",
		},
		[]string{},
	)

	// 히스토그램
	// 값을 미리 정의된 버킷에 배치한다.
	// 각 버킷은 값 범위와 연결되며 최댓값의 이름을 따서 명명된다.
	// 값이 관측되면 히스토그램은 값의 범위에 들어맞는 가장 작은 버킷의 최댓값을 증가시킨다.
	// 이러한 방식으로 각 버킷에 대한 관측 값의 빈도를 추적한다.
	RequestDuration metrics.Histogram = prometheus.NewHistogramFrom(
		prom.HistogramOpts{
			Namespace: *Namespace,
			Subsystem: *Subsystem,
			Buckets: []float64{
				0.0000001, 0.0000002, 0.0000003, 0.0000004, 0.0000005,
				0.000001, 0.0000025, 0.000005, 0.0000075, 0.00001,
				0.0001, 0.001, 0.01,
			},
			Name: "request_duration_historgram_seconds",
			Help: "Total duration of all requests",
		},
		[]string{},
	)

	// Summary
	// 몇 가지 다른 점이 있는 히스토그램이다.
	// 히스토그램은 미리 정의된 버킷이 필요하지만 요약의 경우 스스로 버킷을 계산한다.
	// 메트릭스 서버는 히스토그램 정보를 기준으로 평균이나 퍼센티지를 계산하지만,
	// 서비스는 요약 정보를 기준으로 평균이나 퍼센티지를 계산한다.
	// 따라서 메트릭스 서버의 여러 서비스에 걸쳐 히스토그램은 집계할 수 있지만
	// 요약은 집계할 수 없다.
// 	RequestDuration metrics.Histogram = prometheus.NewSummaryFrom(
// 		prom.SummaryOpts{
// 			Namespace: *Namespace,
// 			Subsystem: *Subsystem,
// 			Name:      "request_duration_summary_seconds",
// 			Help:      "Total duration of all requests",
// 		},
// 		[]string{},
// 	)
)
