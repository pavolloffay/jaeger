package metric

import (
	"github.com/stretchr/testify/require"
	jMetrics "github.com/uber/jaeger-lib/metrics"
	"go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/exporters/metric/stdout"
	"go.uber.org/zap"
	"os"
	"testing"
	"time"
)

func TestA(t *testing.T) {
	pusher, err := stdout.InstallNewPipeline(stdout.Config{Writer:os.Stdout})
	require.NoError(t, err)
	defer pusher.Stop()

	f := New("", global.MeterProvider(), zap.NewNop())

	opts := jMetrics.Options{
		Name: "some metric",
		Tags: map[string]string{"re": "ro"},
		Help: "This helps a lot",
	}
	c := f.Counter(opts)
	c.Inc(15)
	c.Inc(22)
	time.Sleep(10*time.Second)
}
