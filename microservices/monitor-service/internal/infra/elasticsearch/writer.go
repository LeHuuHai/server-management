package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/LeHuuHai/server-management/microservices/monitor-service/internal/domain/repo"
	"github.com/elastic/go-elasticsearch/v8"
)

type Writer[T any] struct {
	client *elasticsearch.Client
	index  string
}

func NewESWriter[T any](client *elasticsearch.Client, index string) repo.LogsRepositoryInterface[T] {
	return &Writer[T]{
		client: client,
		index:  index,
	}
}

func (w *Writer[T]) WriteBatch(ctx context.Context, models []T) error {
	if len(models) == 0 {
		return nil
	}

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
		w.client.Bulk.WithContext(ctx),
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
