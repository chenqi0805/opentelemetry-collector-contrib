package testdata

import "go.opentelemetry.io/collector/consumer/pdata"

func initResource1(r pdata.Resource) {
	initResourceAttributes1(r.Attributes())
}
