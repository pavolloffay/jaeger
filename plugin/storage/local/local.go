package local

import (
	"encoding/json"
	"fmt"
	"bytes"

	"github.com/prometheus/tsdb/index"
	"github.com/prometheus/tsdb/labels"
	"github.com/dgraph-io/badger"

	"github.com/jaegertracing/jaeger/model"
	"github.com/jaegertracing/jaeger/storage/spanstore"
)

const (
	defaultNumTraces = 100
	traceIdLabel       = "traceid"
	serviceNameLabel   = "serviceName"
	operationNameLabel = "operationName"

	servicesDBKey = "services"
	operationsDBKey = "operations"
)

type Storage struct {
	db *badger.DB
	mp *index.MemPostings
	options StorageOptions
}

type StorageOptions struct{
	Directory string
}

func NewStorage(options StorageOptions) *Storage {
	return &Storage{options:options}
}

func (s *Storage) Start() error {
	opts := badger.DefaultOptions
	opts.Dir = s.options.Directory
	opts.ValueDir = s.options.Directory
	db, err := badger.Open(opts)
	if err != nil {
		return err
	}
	s.mp = index.NewMemPostings()
	s.db = db
	return nil
}

func (s *Storage) WriteSpan(span *model.Span) error {
	return s.db.Update(func(txn *badger.Txn) error {
		json, err := json.Marshal(span)
		if err != nil {
			return err
		}
		 //maybe use traceid#spanid ?
		err = txn.Set([]byte(span.SpanID.String()), json)
		if err != nil {
			return err
		}

		err = addValToList(txn, servicesDBKey, span.Process.ServiceName)
		if err != nil {
			return err
		}
		err = addValToList(txn, serviceOperationKey(span.Process.ServiceName), span.OperationName)
		if err != nil {
			return err
		}

		lset := labels.New(labels.Label{
			Name:serviceNameLabel, Value:span.Process.ServiceName},
			labels.Label{Name: operationNameLabel, Value:span.OperationName,
		})
		lset = append(lset, tagsToLabels(span.Tags)...)
		// "serviceName":s -> traceid1
		// "operationName":op -> traceid1
		// tag1:val1 -> traceid1
		// TODO now writing only Low id
		s.mp.Add(uint64(span.TraceID.Low), lset)
		// "traceId":traceid1 -> spanid
		s.mp.Add(uint64(span.SpanID), []labels.Label{{Name: traceIdLabel, Value: span.TraceID.String()}})
		return nil
	})
}

func addValToList(txn *badger.Txn, key string, value string) error {
	item, err := txn.Get([]byte(key))
	if err != nil && err != badger.ErrKeyNotFound {
		return err
	}
	var services []string
	if err != badger.ErrKeyNotFound {
		val, err := item.Value()
		if err != nil {
			return err
		}
		// do not add if it is already there
		if bytes.Contains(val, []byte(value)) {
			return nil
		}
		err = json.Unmarshal(val, &services)
		if err != nil {
			return err
		}
	}
	json, err := json.Marshal(append(services, value))
	if err != nil {
		return err
	}
	err = txn.Set([]byte(key), json)
	if err != nil {
		return err
	}
	return nil
}

func tagsToLabels(tags model.KeyValues) labels.Labels {
	l := labels.Labels{}
	for _, t := range tags {
		// TODO handle all value types
		l = append(l, labels.Label{Name:t.Key, Value:t.VStr})
	}
	return l
}

func (s *Storage) GetServices() ([]string, error) {
	return s.getList(servicesDBKey)
}

func (s *Storage) GetOperations(service string) ([]string, error) {
	return s.getList(serviceOperationKey(service))
}

func serviceOperationKey(service string) string {
	return fmt.Sprintf("%s:%s", service, operationsDBKey)
}

func (s *Storage) getList(key string) ([]string, error) {
	var vals []string
	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			if err == badger.ErrKeyNotFound {
				return nil
			}
			return err
		}

		b, err := item.Value()
		if err != nil {
			return err
		}
		err = json.Unmarshal(b, &vals)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return vals, nil
}

func idToString(id uint64) string {
	return fmt.Sprintf("%x", id)
}

func (s *Storage) GetTrace(traceID model.TraceID) (*model.Trace, error) {
	return s.getTrace(traceID.String())
}

func (s *Storage) getTrace(id string) (*model.Trace, error) {
	ids := s.getFromIndex(traceIdLabel, id)
	t := &model.Trace{}
	s.db.View(func(txn *badger.Txn) error {
		for i := range ids {
			item, err := txn.Get([]byte(idToString(i)))
			if err  != nil {
				return err
			}
			val, err := item.Value()
			if err != nil {
				return err
			}
			s := &model.Span{}
			err = json.Unmarshal(val, s)
			if err != nil {
				return err
			}
			t.Spans = append(t.Spans, s)
		}
		return nil
	})
	return t, nil
}

func (s *Storage) FindTraces(q *spanstore.TraceQueryParameters) ([]*model.Trace, error) {
	if q.NumTraces == 0 {
		q.NumTraces = defaultNumTraces
	}

	// service
	ids := s.getFromIndex(serviceNameLabel, q.ServiceName)
	// operation
	if q.OperationName != "" {
		ids = intersection(ids, s.getFromIndex(operationNameLabel, q.OperationName))
	}
	// tags
	if q.Tags != nil || len(q.Tags) > 0 {
		for k, v := range q.Tags {
			var tids = s.getFromIndex(k ,v)
			ids = intersection(ids, tids)
		}
	}

	// TODO duration, time range
	// duration could be indexed in buckets
	//q.DurationMax
	//q.DurationMin
	//q.StartTimeMax
	//q.StartTimeMin

	traces := make([]*model.Trace, 0, q.NumTraces)
	for id := range ids {
		// get spanids of the trace
		t, err := s.getTrace(idToString(id))
		if err != nil {
			return nil, err
		}
		traces = append(traces, t)
		if len(traces) == q.NumTraces {
			break
		}
	}
	return traces, nil
}

func (s *Storage) getFromIndex(key string, value string) map[uint64]bool {
	s.mp.EnsureOrder()
	var ids = map[uint64]bool{}
	postings := s.mp.Get(key, value)
	for postings.Next() {
		tId := postings.At()
		ids[tId] = true
	}
	return ids
}

func intersection(a map[uint64]bool, b map[uint64]bool) map[uint64]bool {
	var res = map[uint64]bool{}
	for k := range a {
		if b[k] {
			res[k] = true
		}
	}
	return res
}

func (s *Storage) Close() error {
	return s.db.Close()
}
