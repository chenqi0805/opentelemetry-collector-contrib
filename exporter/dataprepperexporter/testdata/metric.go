package testdata

import (
	"go.opentelemetry.io/collector/consumer/pdata"
	"time"
)

var (
	TestMetricStartTime      = time.Date(2020, 2, 11, 20, 26, 12, 321, time.UTC)
	TestMetricStartTimestamp = pdata.TimestampFromTime(TestMetricStartTime)

	TestMetricExemplarTime      = time.Date(2020, 2, 11, 20, 26, 13, 123, time.UTC)
	TestMetricExemplarTimestamp = pdata.TimestampFromTime(TestMetricExemplarTime)

	TestMetricTime      = time.Date(2020, 2, 11, 20, 26, 13, 789, time.UTC)
	TestMetricTimestamp = pdata.TimestampFromTime(TestMetricTime)
)

const (
	TestGaugeDoubleMetricName     = "gauge-double"
	TestGaugeIntMetricName        = "gauge-int"
	TestCounterDoubleMetricName   = "counter-double"
	TestCounterIntMetricName      = "counter-int"
	TestDoubleHistogramMetricName = "double-histogram"
	TestIntHistogramMetricName    = "int-histogram"
	TestDoubleSummaryMetricName   = "double-summary"
)

func GenerateMetricsOneEmptyResourceMetrics() pdata.Metrics {
	md := pdata.NewMetrics()
	md.ResourceMetrics().AppendEmpty()
	return md
}

func GenerateMetricsNoLibraries() pdata.Metrics {
	md := GenerateMetricsOneEmptyResourceMetrics()
	ms0 := md.ResourceMetrics().At(0)
	initResource1(ms0.Resource())
	return md
}

func GenerateMetricsOneEmptyInstrumentationLibrary() pdata.Metrics {
	md := GenerateMetricsNoLibraries()
	md.ResourceMetrics().At(0).InstrumentationLibraryMetrics().AppendEmpty()
	return md
}

func GenerateMetricsOneMetric() pdata.Metrics {
	md := GenerateMetricsOneEmptyInstrumentationLibrary()
	rm0ils0 := md.ResourceMetrics().At(0).InstrumentationLibraryMetrics().At(0)
	initCounterIntMetric(rm0ils0.Metrics().AppendEmpty())
	return md
}

func initCounterIntMetric(im pdata.Metric) {
	initMetric(im, TestCounterIntMetricName, pdata.MetricDataTypeIntSum)

	idps := im.IntSum().DataPoints()
	idp0 := idps.AppendEmpty()
	initMetricLabels1(idp0.LabelsMap())
	idp0.SetStartTimestamp(TestMetricStartTimestamp)
	idp0.SetTimestamp(TestMetricTimestamp)
	idp0.SetValue(123)
	idp1 := idps.AppendEmpty()
	initMetricLabels2(idp1.LabelsMap())
	idp1.SetStartTimestamp(TestMetricStartTimestamp)
	idp1.SetTimestamp(TestMetricTimestamp)
	idp1.SetValue(456)
}

func initMetric(m pdata.Metric, name string, ty pdata.MetricDataType) {
	m.SetName(name)
	m.SetDescription("")
	m.SetUnit("1")
	m.SetDataType(ty)
	switch ty {
	case pdata.MetricDataTypeIntSum:
		sum := m.IntSum()
		sum.SetIsMonotonic(true)
		sum.SetAggregationTemporality(pdata.AggregationTemporalityCumulative)
	case pdata.MetricDataTypeDoubleSum:
		sum := m.DoubleSum()
		sum.SetIsMonotonic(true)
		sum.SetAggregationTemporality(pdata.AggregationTemporalityCumulative)
	case pdata.MetricDataTypeIntHistogram:
		histo := m.IntHistogram()
		histo.SetAggregationTemporality(pdata.AggregationTemporalityCumulative)
	case pdata.MetricDataTypeHistogram:
		histo := m.Histogram()
		histo.SetAggregationTemporality(pdata.AggregationTemporalityCumulative)
	}
}
