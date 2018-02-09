package local

import (
	"testing"
	"time"
	"math/rand"
	"fmt"
	"os"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jaegertracing/jaeger/model"
	"github.com/jaegertracing/jaeger/storage/spanstore"
)


func withStorage(t *testing.T,fce func(s *Storage))  {
	dir := fmt.Sprintf("/tmp/%d", rand.NewSource(time.Now().UnixNano()).Int63())
	storage := NewStorage(StorageOptions{Directory:dir})
	storage.Start()
	fce(storage)
	err := storage.Close()
	require.NoError(t, err)
	err = os.RemoveAll(dir)
	require.NoError(t, err)
}

func TestWriteGet(t *testing.T) {
	withStorage(t, func(storage *Storage) {
		p1 := &model.Process{ServiceName:"bar"}
		p2 := &model.Process{ServiceName:"baz"}
		span1 := &model.Span{
			SpanID: 11,
			TraceID: model.TraceID{Low: 2, High:2},
			OperationName:"foo",
			Tags:[]model.KeyValue{{Key:"k", VStr: "v"}},
			Process: p1}
		span2 := &model.Span{
			SpanID: 22,
			ParentSpanID: span1.SpanID,
			TraceID: span1.TraceID,
			OperationName:"foo",
			Tags:[]model.KeyValue{{Key:"k", VStr: "v"}},
			Process:p2}
		span3 := &model.Span{
			SpanID: 3232,
			ParentSpanID: span1.SpanID,
			TraceID: model.TraceID{Low:4343, High:22},
			OperationName:"foo",
			Tags:[]model.KeyValue{{Key:"k", VStr: "v"}},
			Process:p2}

		err := storage.WriteSpan(span1)
		require.NoError(t, err)
		err = storage.WriteSpan(span2)
		require.NoError(t, err)
		err = storage.WriteSpan(span3)
		require.NoError(t, err)

		trace, err := storage.GetTrace(span1.TraceID)
		require.NoError(t, err)
		assert.Equal(t, 2, len(trace.Spans))

		services, err := storage.GetServices()
		require.NoError(t, err)
		assert.Equal(t, 2, len(services))

		ops, err := storage.GetOperations(span1.Process.ServiceName)
		require.NoError(t, err)
		assert.Equal(t, 1, len(ops))
		assert.Equal(t, []string{span1.OperationName}, ops)
	})
}

func TestFindTraces(t *testing.T) {
	withStorage(t, func(storage *Storage) {
		span1 := &model.Span{
			SpanID: 111,
			TraceID: model.TraceID{Low: 1, High:2},
			OperationName:"span1",
			Tags:[]model.KeyValue{{Key:"k1", VStr: "v1"}},
			Process: &model.Process{ServiceName:"service1"}}
		span2 := &model.Span{
			SpanID: 222,
			ParentSpanID: span1.SpanID,
			TraceID: model.TraceID{Low: 2, High:2},
			OperationName:"span2",
			Tags:[]model.KeyValue{{Key:"k2", VStr: "v2"}},
			Process: &model.Process{ServiceName:"service2"}}
		span3 := &model.Span{
			SpanID: 3232,
			ParentSpanID: span1.SpanID,
			TraceID: model.TraceID{Low:4343, High:22},
			OperationName:"foo",
			Tags:[]model.KeyValue{{Key:"k1", VStr: "v1"}},
			Process:&model.Process{ServiceName:"service1"}}

		err := storage.WriteSpan(span1)
		require.NoError(t, err)
		err = storage.WriteSpan(span2)
		require.NoError(t, err)
		err = storage.WriteSpan(span3)
		require.NoError(t, err)

		traces, err := storage.FindTraces(&spanstore.TraceQueryParameters{ServiceName:"service1"})
		require.NoError(t, err)
		assert.Equal(t, 2, len(traces))

		traces, err = storage.FindTraces(&spanstore.TraceQueryParameters{ServiceName:"service1",
		OperationName:"span1"})
		require.NoError(t, err)
		assert.Equal(t, 1, len(traces))

		traces, err = storage.FindTraces(&spanstore.TraceQueryParameters{ServiceName:"service1",
			OperationName:"span1",
			Tags:map[string]string{"k1":"v1"}})
		require.NoError(t, err)
		assert.Equal(t, 1, len(traces))

		traces, err = storage.FindTraces(&spanstore.TraceQueryParameters{ServiceName:"service1",
			OperationName:"span1",
			Tags:map[string]string{"k1":"v2"}})
		require.NoError(t, err)
		assert.Equal(t, 0, len(traces))
	})
}
