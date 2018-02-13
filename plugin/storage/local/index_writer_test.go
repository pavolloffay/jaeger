package local

import (
	"testing"
	"fmt"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/prometheus/tsdb/index"

	"github.com/prometheus/tsdb/labels"
	"github.com/prometheus/tsdb/chunks"
	"os"
	"github.com/prometheus/tsdb"
	"github.com/prometheus/client_golang/prometheus"
)

func TestIndex(t *testing.T) {
	indexDir := "/tmp/index3"
	os.RemoveAll(indexDir)
	mp := index.NewMemPostings()
	mp.Add(15, labels.New(labels.Label{Name:"re", Value:"ka"}))

	writer, err := index.NewWriter(indexDir)
	require.NoError(t, err)
	// this is required
	writer.AddSymbols(map[string]struct{}{"re":{}, "ka":{}})
	err = writer.AddSeries(15, labels.FromMap(map[string]string{"re":"ka"}), chunks.Meta{})
	require.NoError(t, err)
	err = writer.WritePostings("re", "ka", mp.Get("re", "ka"))
	require.NoError(t, err)
	err = writer.Close()
	require.NoError(t, err)

	reader, err := index.NewFileReader(indexDir, 2)
	require.NoError(t, err)
	pos, err := reader.Postings("re", "ka")
	fmt.Println(pos.Next())
	require.NoError(t, err)

	writer, err = index.NewWriter(indexDir)
	require.NoError(t, err)
	writer.AddSymbols(map[string]struct{}{"re":{}, "ka":{}})
	err = writer.AddSeries(15, labels.FromMap(map[string]string{"re":"ka"}), chunks.Meta{Ref:15})

	require.NoError(t, err)
	mp = index.NewMemPostings()
	mp.Add(15, labels.New(labels.Label{Name:"re", Value:"ka"}))
	err = writer.WritePostings("re1", "ka1", mp.Get("re", "ka"))
	require.NoError(t, err)
	err = writer.Close()
	require.NoError(t, err)

	mp.Iter(func(label labels.Label, postings index.Postings) error {
		// TODO write here
		return nil
	})
}



func TestAdd(t *testing.T) {
	indexDir := "/tmp/index2"
	os.RemoveAll(indexDir)

	writer, err := index.NewWriter(indexDir)
	require.NoError(t, err)
	writer.AddSymbols(map[string]struct{}{"re":{}, "ka":{}})
	err = writer.AddSeries(15, labels.FromMap(map[string]string{"re":"ka"}), chunks.Meta{Ref:15})
	require.NoError(t, err)
	mp := index.NewMemPostings()
	mp.Add(15, labels.New(labels.Label{Name:"re", Value:"ka"}))
	err = writer.WritePostings("re", "ka", mp.Get("re", "ka"))
	require.NoError(t, err)
	err = writer.Close()
	require.NoError(t, err)

	reader, err := index.NewFileReader(indexDir, 2)

	require.NoError(t, err)
	pos, err := reader.Postings("re", "ka")
	fmt.Println(pos.Next())
	require.NoError(t, err)

	writer, err = index.NewWriter(indexDir)
	require.NoError(t, err)
	writer.AddSymbols(map[string]struct{}{"re":{}, "ka":{}})
	err = writer.AddSeries(15, labels.FromMap(map[string]string{"re":"ka"}), chunks.Meta{Ref:15})
	require.NoError(t, err)
	mp = index.NewMemPostings()
	mp.Add(15, labels.New(labels.Label{Name:"re", Value:"ka"}))
	err = writer.WritePostings("re1", "ka1", mp.Get("re", "ka"))
	require.NoError(t, err)
	err = writer.Close()
	require.NoError(t, err)

	labels := labels.New(labels.Label{Name:"n1", Value:"v2"})
	var chnks = make([]chunks.Meta, 0, 1)
	err = reader.Series(2, &labels, &chnks)
	require.NoError(t, err)

	pos, err = reader.Postings("n1", "n2")
	fmt.Println(pos.Next())
	require.NoError(t, err)

	assert.NotNil(t, pos)
	err = reader.Close()
	require.NoError(t, err)

	reader, err = index.NewFileReader(indexDir, 2)
	require.NoError(t, err)
	pos, err = reader.Postings("n1", "n2")
	fmt.Println(pos.Next())
	require.NoError(t, err)
}

func TestWriteRead(t *testing.T) {
	indexDir := "/tmp/index3"
	defer os.RemoveAll(indexDir)

	postings := index.NewMemPostings()
	postings.Add(1, labels.New(labels.Label{Name:"l1", Value:"v1"}))
	postings.Add(1, labels.New(labels.Label{Name:"l2", Value:"v2"}))
	postings.Add(2, labels.New(labels.Label{Name:"l3", Value:"v3"}))

	writer, err := index.NewWriter(indexDir)
	require.NoError(t, err)

	symbols := map[string]struct{}{}

	err = postings.Iter(func(label labels.Label, postings index.Postings) error {
		symbols[label.Name] = struct{}{}
		symbols[label.Value] = struct{}{}
		fmt.Println(label)
		return writer.WritePostings(label.Name, label.Value, postings)
	})
	require.NoError(t, err)


	block := tsdb.Block{}
	head := tsdb.NewHead(prometheus.Registerer())
	head.Appender().Commit()
	i := head.Index()
	p := i.Postings("re", "re")
	h := tsdb.Head{}
	tsdb.Open()


	err = writer.Close()
	require.NoError(t, err)
}
