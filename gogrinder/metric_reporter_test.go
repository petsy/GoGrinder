package gogrinder

import (
	"testing"

	time "github.com/finklabs/ttime"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

func TestCheckMetricReporterImplementsReporterInterface(t *testing.T) {
	mr := NewMetricReporter()
	if _, ok := interface{}(mr).(Reporter); !ok {
		t.Errorf("MetricReporter does not implement the Reporter interface!")
	}
}

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
	mr := NewMetricReporter()

	// add datapoint
	m := Meta{"01_tc", "01_01_ts", 0, 0, Timestamp(time.Now()),
		Elapsed(600 * time.Millisecond), "something went wrong!"}
	mr.Update(m)

	// check that datapoint was reported
	if exp, got := 600.0, readSummaryVec(mr.elapsed,
		prometheus.Labels{"teststep": "01_01_ts"})[0].GetValue(); exp != got {
		t.Errorf("Expected elapsed %d, got %d.", exp, got)
	}
	if exp, got := 1.0, readCounterVec(mr.error,
		prometheus.Labels{"teststep": "01_01_ts"}); exp != got {
		t.Errorf("Expected error counter %f, got %f.", exp, got)
	}
}
