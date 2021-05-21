package testdata

import "go.opentelemetry.io/collector/consumer/pdata"

var (
	resourceAttributes1 = map[string]pdata.AttributeValue{"resource-attr": pdata.NewAttributeValueString("resource-attr-val-1")}
	spanEventAttributes = map[string]pdata.AttributeValue{"span-event-attr": pdata.NewAttributeValueString("span-event-attr-val")}
)

const (
	TestLabelKey1       = "label-1"
	TestLabelValue1     = "label-value-1"
	TestLabelKey2       = "label-2"
	TestLabelValue2     = "label-value-2"
)

func initMetricLabels1(dest pdata.StringMap) {
	dest.InitFromMap(map[string]string{TestLabelKey1: TestLabelValue1})
}

func initMetricLabels2(dest pdata.StringMap) {
	dest.InitFromMap(map[string]string{TestLabelKey2: TestLabelValue2})
}

func initResourceAttributes1(dest pdata.AttributeMap) {
	dest.InitFromMap(resourceAttributes1)
}

func initSpanEventAttributes(dest pdata.AttributeMap) {
	dest.InitFromMap(spanEventAttributes)
}
