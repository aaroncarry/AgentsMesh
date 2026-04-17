package otel

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
)

var (
	PodActiveCount    metric.Int64UpDownCounter = noop.Int64UpDownCounter{}
	RunnerConnected   metric.Int64UpDownCounter = noop.Int64UpDownCounter{}
	GRPCMessagesRecv  metric.Int64Counter       = noop.Int64Counter{}
	PodCreateDuration metric.Float64Histogram   = noop.Float64Histogram{}
)

func InitMetrics() {
	m := otel.Meter("agentsmesh-backend")
	PodActiveCount, _ = m.Int64UpDownCounter("agentsmesh.backend.pod.active")
	RunnerConnected, _ = m.Int64UpDownCounter("agentsmesh.backend.runner.connected")
	GRPCMessagesRecv, _ = m.Int64Counter("agentsmesh.backend.grpc.messages.received")
	PodCreateDuration, _ = m.Float64Histogram("agentsmesh.backend.pod.create.duration",
		metric.WithUnit("ms"))
}
