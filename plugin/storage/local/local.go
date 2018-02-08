package local

import (
	"encoding/json"
	"strconv"

	"github.com/prometheus/tsdb/index"
	"github.com/prometheus/tsdb/labels"
	"github.com/dgraph-io/badger"

	"github.com/jaegertracing/jaeger/model"
	"github.com/jaegertracing/jaeger/storage/spanstore"
	"time"
	"math/rand"
	"fmt"
)

const (
	traceIdLabel = "traceid"
	serviceNameLabel = "serviceName"
	operationNamelabel = "operationName"

	servicesDBKey = "services"
	operationsDBKey = "operations"
)

type Storage struct {
	db *badger.DB
	mp *index.MemPostings
}

func (s *Storage) Start() error {
	fmt.Printf("aaaarrr")
	opts := badger.DefaultOptions
	dir := fmt.Sprintf("/tmp/%d", rand.NewSource(time.Now().UnixNano()).Int63())
	fmt.Printf("Creating storage %s\n", dir)
	opts.Dir = dir
	opts.ValueDir = dir
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
		err = addValToList(txn, serviceOperationDBKey(span.Process.ServiceName), span.OperationName)
		if err != nil {
			return err
		}

		lset := tagsToLabels(span.Tags)
		traceIdLabel := labels.Label{Name: traceIdLabel, Value: span.TraceID.String()}
		// k:v -> traceid1, traceid2
		s.mp.Add(uint64(span.TraceID.High), append(lset, traceIdLabel))
		// "trace":traceId -> spanid1, spanid2
		// "serviceName":s -> spanid
		// "operationName":op -> spanid
		s.mp.Add(uint64(span.SpanID), []labels.Label{
			traceIdLabel,
			{Name:serviceNameLabel, Value:span.Process.ServiceName},
			{Name:operationNamelabel, Value:span.OperationName},
		})
		// services
		// TODO probably move to a different place?
		s.mp.EnsureOrder()
		return nil
	})
}

func addValToList(txn *badger.Txn, key string, val string) error {
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
		err = json.Unmarshal(val, &services)
		if err != nil {
			return err
		}
	}
	json, err := json.Marshal(append(services, val))
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
	return s.getList(serviceOperationDBKey(service))
}

func serviceOperationDBKey(service string) string {
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
	// TODO really strconv?
	return strconv.FormatUint(id, 10)
}

func (s *Storage) GetTrace(traceID model.TraceID) (*model.Trace, error) {
	p := s.mp.Get(traceIdLabel, traceID.String())
	return s.getTrace(p)
}

func (s *Storage) getTrace(p index.Postings) (*model.Trace, error) {
	t := &model.Trace{}
	s.db.View(func(txn *badger.Txn) error {
		for p.Next() {
			item, err := txn.Get([]byte(idToString(p.At())))
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
	var tIds []uint64
	for k, v := range q.Tags {
		p := s.mp.Get(k, v)
		id := p.At()
		tIds = append(tIds, id)
		for p.Next() {
			// TODO now doing union
			tIds = append(tIds, p.At())
		}
	}

	var traces []*model.Trace
	for _, id := range tIds {
		// get spanids of the trace
		p := s.mp.Get("traceid", idToString(id))
		t, err := s.getTrace(p)
		if err != nil {
			return nil, err
		}
		traces = append(traces, t)
	}
	return traces, nil
}
