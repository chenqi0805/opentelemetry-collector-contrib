package testdata

import (
	"go.opentelemetry.io/collector/consumer/pdata"
	"time"
)

var (
	TestSpanStartTime      = time.Date(2020, 2, 11, 20, 26, 12, 321, time.UTC)
	TestSpanStartTimestamp = pdata.TimestampFromTime(TestSpanStartTime)

	TestSpanEventTime      = time.Date(2020, 2, 11, 20, 26, 13, 123, time.UTC)
	TestSpanEventTimestamp = pdata.TimestampFromTime(TestSpanEventTime)

	TestSpanEndTime      = time.Date(2020, 2, 11, 20, 26, 13, 789, time.UTC)
	TestSpanEndTimestamp = pdata.TimestampFromTime(TestSpanEndTime)
)

func GenerateTracesOneEmptyResourceSpans() pdata.Traces {
	td := pdata.NewTraces()
	td.ResourceSpans().AppendEmpty()
	return td
}

func GenerateTracesNoLibraries() pdata.Traces {
	td := GenerateTracesOneEmptyResourceSpans()
	rs0 := td.ResourceSpans().At(0)
	initResource1(rs0.Resource())
	return td
}

func GenerateTracesOneEmptyInstrumentationLibrary() pdata.Traces {
	td := GenerateTracesNoLibraries()
	td.ResourceSpans().At(0).InstrumentationLibrarySpans().AppendEmpty()
	return td
}

func GenerateTracesOneSpan() pdata.Traces {
	td := GenerateTracesOneEmptyInstrumentationLibrary()
	rs0ils0 := td.ResourceSpans().At(0).InstrumentationLibrarySpans().At(0)
	fillSpanOne(rs0ils0.Spans().AppendEmpty())
	return td
}

func fillSpanOne(span pdata.Span) {
	span.SetName("operationA")
	span.SetStartTimestamp(TestSpanStartTimestamp)
	span.SetEndTimestamp(TestSpanEndTimestamp)
	span.SetDroppedAttributesCount(1)
	evs := span.Events()
	ev0 := evs.AppendEmpty()
	ev0.SetTimestamp(TestSpanEventTimestamp)
	ev0.SetName("event-with-attr")
	initSpanEventAttributes(ev0.Attributes())
	ev0.SetDroppedAttributesCount(2)
	ev1 := evs.AppendEmpty()
	ev1.SetTimestamp(TestSpanEventTimestamp)
	ev1.SetName("event")
	ev1.SetDroppedAttributesCount(2)
	span.SetDroppedEventsCount(1)
	status := span.Status()
	status.SetCode(pdata.StatusCodeError)
	status.SetMessage("status-cancelled")
}
