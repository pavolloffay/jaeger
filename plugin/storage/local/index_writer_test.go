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

