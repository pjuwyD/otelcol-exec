package execreceiver

import (
	"context"
	"os/exec"
	"time"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.uber.org/zap"
	"bytes"
	"go.opentelemetry.io/collector/pdata/pmetric"
)

type execreceiverReceiver struct {
	host         component.Host
	cancel       context.CancelFunc
	logger       *zap.Logger
	nextConsumer consumer.Metrics
	config       *Config
}

func (execRcvr *execreceiverReceiver) Start(ctx context.Context, host component.Host) error {
	execRcvr.host = host
	execRcvr.logger.Info("Starting execreceiverReceiver...")

	// Creating a background context and attaching cancel function
	ctx = context.Background()
	ctx, execRcvr.cancel = context.WithCancel(ctx)

	// Parse interval from configuration
	interval, err := time.ParseDuration(execRcvr.config.Interval)
	if err != nil {
		execRcvr.logger.Error("Invalid interval in config", zap.String("interval", execRcvr.config.Interval), zap.Error(err))
		return err
	}
	execRcvr.logger.Info("Interval parsed successfully", zap.Duration("interval", interval))

	// Start a goroutine to run the Python script at regular intervals
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		execRcvr.logger.Info("Started processing metrics...")

		for {
			select {
			case <-ticker.C:
				execRcvr.logger.Info("Tick received, running Python script...")
				output, err := runPythonScript(execRcvr.config.Script)
				if err != nil {
					execRcvr.logger.Error("Error running Python script", zap.Error(err))
					continue
				}

				execRcvr.logger.Info("Python script output received", zap.String("output", output))

				// Create the JSON unmarshaler
				metricsUnmarshaler := &pmetric.JSONUnmarshaler{}

				// Unmarshal the JSON into pmetric.Metrics
				metrics, err := metricsUnmarshaler.UnmarshalMetrics([]byte(output))
				if err != nil {
					execRcvr.logger.Error("Error unmarshaling metrics JSON", zap.Error(err))
					continue
				}

				// Check if the metrics object contains any data
				if metrics.ResourceMetrics().Len() == 0 {
					execRcvr.logger.Warn("Unmarshaled metrics is empty or nil")
					continue
				}

				// Send metrics to the next consumer
				err = execRcvr.nextConsumer.ConsumeMetrics(ctx, metrics)
				if err != nil {
					execRcvr.logger.Error("Error consuming metrics", zap.Error(err))
					continue
				}

				execRcvr.logger.Info("Metrics processed successfully")
			case <-ctx.Done():
				execRcvr.logger.Info("Receiver shutdown initiated")
				return
			}
		}
	}()

	execRcvr.logger.Info("execreceiverReceiver started successfully.")
	return nil
}


// runPythonScript runs the Python script and returns its output
func runPythonScript(path string) (string, error) {
	// Path to your Python script (adjust path as needed)
	cmd := exec.Command("/opt/axess/bin/python", path)

	// Capture stdout and stderr
	var out, stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	// Run the script
	err := cmd.Run()
	if err != nil {
		return stderr.String(), err
	}

	return out.String(), nil
}

func (execRcvr *execreceiverReceiver) Shutdown(ctx context.Context) error {
	// Log shutdown attempt
	execRcvr.logger.Info("Shutting down execreceiverReceiver...")
	if execRcvr.cancel != nil {
		execRcvr.cancel()
	}
	execRcvr.logger.Info("Receiver shutdown completed.")
	return nil
}
