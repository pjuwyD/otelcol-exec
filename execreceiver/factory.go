package execreceiver

import (
	"context"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver"
)

var (
	typeStr = component.MustNewType("execreceiver")
)

const (
	defaultInterval = 1 * time.Minute
)

func createDefaultConfig() component.Config {
	return &Config{
		Interval: string(defaultInterval),
	}
}

func createMetricsReceiver(_ context.Context, params receiver.Settings, baseCfg component.Config, consumer consumer.Metrics) (receiver.Metrics, error) {
	logger := params.Logger
	execreceiverCfg := baseCfg.(*Config)

	metricRcvr := &execreceiverReceiver{
		logger:       logger,
		nextConsumer: consumer,
		config:       execreceiverCfg,
	}

	return metricRcvr, nil
}

// NewFactory creates a factory for execreceiver receiver.
func NewFactory() receiver.Factory {
	return receiver.NewFactory(
		typeStr,
		createDefaultConfig,
		receiver.WithMetrics(createMetricsReceiver, component.StabilityLevelAlpha))
}
