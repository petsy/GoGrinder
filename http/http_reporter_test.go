package http

import (
	"net/http"
	"testing"

	"github.com/finklabs/GoGrinder"
	time "github.com/finklabs/ttime"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

// TODO: Pending prometheus/client_golang#58
// read metric helpers needs rework once testability is improved!
func readSummaryVec(m *prometheus.SummaryVec, l prometheus.Labels) []*dto.Quantile {
	pb := &dto.Metric{}
	s := m.With(l)
	s.Write(pb)
	return pb.GetSummary().GetQuantile()
}

func readCounterVec(m *prometheus.CounterVec, l prometheus.Labels) float64 {
	pb := &dto.Metric{}
	c := m.With(l)
	c.Write(pb)
	return pb.GetCounter().GetValue()
}

func TestHttpMetricUpdate(t *testing.T) {
	hmr := NewHttpMetricReporter()

	// add datapoint
	hm := HttpMetric{gogrinder.Meta{"01_tc", "01_01_ts", 0, 0, gogrinder.Timestamp(time.Now()),
		gogrinder.Elapsed(600 * time.Millisecond), "something is wrong!"},
		gogrinder.Elapsed(500 * time.Millisecond), 10240, http.StatusOK}
	hmr.Update(hm)

	// check that datapoint was reported
	if exp, got := 600.0, readSummaryVec(hmr.elapsed,
		prometheus.Labels{"teststep": "01_01_ts"})[0].GetValue(); exp != got {
		t.Errorf("Expected elapsed %d, got %d.", exp, got)
	}
	if exp, got := 500.0, readSummaryVec(hmr.firstByte,
		prometheus.Labels{"teststep": "01_01_ts"})[0].GetValue(); exp != got {
		t.Errorf("Expected firstByte %d, got %d.", exp, got)
	}
	if exp, got := 10.0, readSummaryVec(hmr.bytes,
		prometheus.Labels{"teststep": "01_01_ts"})[0].GetValue(); exp != got {
		t.Errorf("Expected kb %d, got %d.", exp, got)
	}
	if exp, got := 1.0, readCounterVec(hmr.error,
		prometheus.Labels{"teststep": "01_01_ts"}); exp != got {
		t.Errorf("Expected error counter %f, got %f.", exp, got)
	}
	if exp, got := 1.0, readCounterVec(hmr.code,
		prometheus.Labels{"teststep": "01_01_ts", "code": "200"}); exp != got {
		t.Errorf("Expected code counter %f, got %f.", exp, got)
	}
}

func TestCheckHttpMetricReporterImplementsReporterInterface(t *testing.T) {
	mr := NewHttpMetricReporter()
	if _, ok := interface{}(mr).(gogrinder.Reporter); !ok {
		t.Errorf("HttpMetricReporter does not implement the Reporter interface!")
	}
}

func TestCheckHttpMetricImplementsMetricInterface(t *testing.T) {
	mr := HttpMetric{}
	if _, ok := interface{}(mr).(gogrinder.Metric); !ok {
		t.Errorf("HttpMetric does not implement the Metric interface!")
	}
}

/* FIXME
func TestCheckHttpMetricEmbedsMeta(t *testing.T) {
	//mr := HttpMetric{gogrinder.Meta{}, time.Duration(0), 0, 0}
	mr := HttpMetric{}
	if _, ok := (gogrinder.Metric(mr)).(gogrinder.Meta); !ok {
		t.Errorf("HttpMetric does not embed Meta!")
	}
}
*/
