package gogrinder

import(
	"testing"
	//dto "github.com/prometheus/client_model/go"
	//"github.com/prometheus/client_golang/prometheus"
)


func TestMetricsReporterElapsed(t *testing.T) {

	// enter a measurement
	meta := Meta{"elapsed": 33333333, "teststep": "01_01_teststep"}
	mr := NewMetricsReporter()
	mr.Update(meta)

	//metric, _ := mr.elapsed.GetMetricWith(prometheus.Labels{"teststep": "01_01_teststep"})
	// most likely .Metric() -> Write(&dto.Metric)

}

