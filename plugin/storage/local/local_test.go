package local

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jaegertracing/jaeger/model"
	"fmt"
)


func TestStoreGet(t *testing.T) {
	storage := &Storage{}
	storage.Start()

	p1 := &model.Process{ServiceName:"bar"}
	p2 := &model.Process{ServiceName:"baz"}
	span1 := &model.Span{
		SpanID: 1,
		TraceID: model.TraceID{Low: 1, High:2},
		OperationName:"foo",
		Tags:[]model.KeyValue{{Key:"k", VStr: "v"}},
		Process: p1}
	span2 := &model.Span{
		SpanID: 2,
		ParentSpanID: span1.SpanID,
		TraceID: model.TraceID{Low: 2, High:2},
		OperationName:"foo",
		Tags:[]model.KeyValue{{Key:"k", VStr: "v"}},
		Process:p2}

	err := storage.WriteSpan(span1)
	require.NoError(t, err)
	err = storage.WriteSpan(span2)
	require.NoError(t, err)

	trace, err := storage.GetTrace(span1.TraceID)
	require.NoError(t, err)
	printTrace(trace)

	assert.NotNil(t, trace)
	assert.Equal(t, 2, len(trace.Spans))

	services, err := storage.GetServices()
	require.NoError(t, err)
	assert.Equal(t, 2, len(services))

	ops, err := storage.GetOperations(span1.Process.ServiceName)
	require.NoError(t, err)
	assert.Equal(t, 1, len(ops))
	assert.Equal(t, []string{span1.OperationName}, ops)

	storage.db.Close()
}

func printTrace(t *model.Trace) {
	for _, p := range t.Spans {
		fmt.Printf("%+v\n", p)
	}
}
