// Copyright 2020, OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package elastic_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.elastic.co/apm/model"
	"go.elastic.co/apm/transport/transporttest"
	"go.elastic.co/fastjson"
	"go.opentelemetry.io/collector/model/pdata"

	"github.com/open-telemetry/opentelemetry-collector-contrib/exporter/elasticexporter/internal/translator/elastic"
)

func TestEncodeMetrics(t *testing.T) {
	var w fastjson.Writer
	var recorder transporttest.RecorderTransport
	elastic.EncodeResourceMetadata(pdata.NewResource(), &w)

	instrumentationLibraryMetrics := pdata.NewInstrumentationLibraryMetrics()
	metrics := instrumentationLibraryMetrics.Metrics()
	appendMetric := func(name string, dataType pdata.MetricDataType) pdata.Metric {
		metric := metrics.AppendEmpty()
		metric.SetName(name)
		metric.SetDataType(dataType)
		return metric
	}

	timestamp0 := time.Unix(123, 0).UTC()
	timestamp1 := time.Unix(456, 0).UTC()

	var expectDropped int

	metric := appendMetric("int_gauge_metric", pdata.MetricDataTypeGauge)
	intGauge := metric.Gauge()
	intGauge.DataPoints().EnsureCapacity(4)
	idp := intGauge.DataPoints().AppendEmpty()
	idp.SetTimestamp(pdata.TimestampFromTime(timestamp0))
	idp.SetIntVal(1)
	idp = intGauge.DataPoints().AppendEmpty()
	idp.SetTimestamp(pdata.TimestampFromTime(timestamp1))
	idp.SetIntVal(2)
	idp.LabelsMap().InitFromMap(map[string]string{"k": "v"})
	idp = intGauge.DataPoints().AppendEmpty()
	idp.SetTimestamp(pdata.TimestampFromTime(timestamp1))
	idp.SetIntVal(3)
	idp = intGauge.DataPoints().AppendEmpty()
	idp.SetTimestamp(pdata.TimestampFromTime(timestamp1))
	idp.SetIntVal(4)
	idp.LabelsMap().InitFromMap(map[string]string{"k": "v2"})

	metric = appendMetric("double_gauge_metric", pdata.MetricDataTypeGauge)
	doubleGauge := metric.Gauge()
	doubleGauge.DataPoints().EnsureCapacity(4)
	ddp := doubleGauge.DataPoints().AppendEmpty()
	ddp.SetTimestamp(pdata.TimestampFromTime(timestamp0))
	ddp.SetDoubleVal(5)
	ddp = doubleGauge.DataPoints().AppendEmpty()
	ddp.SetTimestamp(pdata.TimestampFromTime(timestamp1))
	ddp.SetDoubleVal(6)
	ddp.LabelsMap().InitFromMap(map[string]string{"k": "v"})
	ddp = doubleGauge.DataPoints().AppendEmpty()
	ddp.SetTimestamp(pdata.TimestampFromTime(timestamp1))
	ddp.SetDoubleVal(7)
	ddp = doubleGauge.DataPoints().AppendEmpty()
	ddp.SetTimestamp(pdata.TimestampFromTime(timestamp1))
	ddp.SetDoubleVal(8)
	ddp.LabelsMap().InitFromMap(map[string]string{"k": "v2"})

	metric = appendMetric("int_sum_metric", pdata.MetricDataTypeSum)
	intSum := metric.Sum()
	intSum.DataPoints().EnsureCapacity(3)
	is := intSum.DataPoints().AppendEmpty()
	is.SetTimestamp(pdata.TimestampFromTime(timestamp0))
	is.SetIntVal(9)
	is = intSum.DataPoints().AppendEmpty()
	is.SetTimestamp(pdata.TimestampFromTime(timestamp1))
	is.SetIntVal(10)
	is.LabelsMap().InitFromMap(map[string]string{"k": "v"})
	is = intSum.DataPoints().AppendEmpty()
	is.SetTimestamp(pdata.TimestampFromTime(timestamp1))
	is.SetIntVal(11)
	is.LabelsMap().InitFromMap(map[string]string{"k2": "v"})

	metric = appendMetric("double_sum_metric", pdata.MetricDataTypeSum)
	doubleSum := metric.Sum()
	doubleSum.DataPoints().EnsureCapacity(3)
	ds := doubleSum.DataPoints().AppendEmpty()
	ds.SetTimestamp(pdata.TimestampFromTime(timestamp0))
	ds.SetDoubleVal(12)
	ds = doubleSum.DataPoints().AppendEmpty()
	ds.SetTimestamp(pdata.TimestampFromTime(timestamp1))
	ds.SetDoubleVal(13)
	ds.LabelsMap().InitFromMap(map[string]string{"k": "v"})
	ds = doubleSum.DataPoints().AppendEmpty()
	ds.SetTimestamp(pdata.TimestampFromTime(timestamp1))
	ds.SetDoubleVal(14)
	ds.LabelsMap().InitFromMap(map[string]string{"k2": "v"})

	// Histograms are currently not supported, and will be ignored.
	metric = appendMetric("double_histogram_metric", pdata.MetricDataTypeHistogram)
	metric.Histogram().DataPoints().AppendEmpty()
	expectDropped++

	dropped, err := elastic.EncodeMetrics(metrics, instrumentationLibraryMetrics.InstrumentationLibrary(), &w)
	require.NoError(t, err)
	assert.Equal(t, expectDropped, dropped)
	sendStream(t, &w, &recorder)

	payloads := recorder.Payloads()
	assert.Equal(t, []model.Metrics{{
		Timestamp: model.Time(timestamp0),
		Samples: map[string]model.Metric{
			"double_gauge_metric": {Value: 5},
			"double_sum_metric":   {Value: 12},
			"int_gauge_metric":    {Value: 1},
			"int_sum_metric":      {Value: 9},
		},
	}, {
		Timestamp: model.Time(timestamp1),
		Samples: map[string]model.Metric{
			"double_gauge_metric": {Value: 7},
			"int_gauge_metric":    {Value: 3},
		},
	}, {
		Timestamp: model.Time(timestamp1),
		Labels:    model.StringMap{{Key: "k", Value: "v"}},
		Samples: map[string]model.Metric{
			"double_gauge_metric": {Value: 6},
			"double_sum_metric":   {Value: 13},
			"int_gauge_metric":    {Value: 2},
			"int_sum_metric":      {Value: 10},
		},
	}, {
		Timestamp: model.Time(timestamp1),
		Labels:    model.StringMap{{Key: "k", Value: "v2"}},
		Samples: map[string]model.Metric{
			"double_gauge_metric": {Value: 8},
			"int_gauge_metric":    {Value: 4},
		},
	}, {
		Timestamp: model.Time(timestamp1),
		Labels:    model.StringMap{{Key: "k2", Value: "v"}},
		Samples: map[string]model.Metric{
			"double_sum_metric": {Value: 14},
			"int_sum_metric":    {Value: 11},
		},
	}}, payloads.Metrics)

	assert.Empty(t, payloads.Errors)
}
