package es

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/elastic/go-elasticsearch/v8"
)

type Writer[T any] struct {
	client *elasticsearch.Client
	index  string
}

func (w *Writer[T]) WriteBatch(models []T) error {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)

	for _, model := range models {
		meta := map[string]any{
			"index": map[string]any{
				"_index": w.index,
			},
		}
		if err := enc.Encode(meta); err != nil {
			return err
		}
		if err := enc.Encode(model); err != nil {
			return err
		}
	}

	res, err := w.client.Bulk(
		bytes.NewReader(buf.Bytes()),
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.IsError() {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("bulk error: %s", body)
	}
	return nil
}

func NewWriter[T any](client *elasticsearch.Client, index string) *Writer[T] {
	return &Writer[T]{
		client: client,
		index:  index,
	}
}
