package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"kmtym1998/es-zip-codes/model"
	"log"
	"os"
	"runtime"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/dustin/go-humanize"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esutil"
)

type esservice struct{
	client *elasticsearch.Client
}

func NewClient() (*esservice, error) {
	retryBackoff := backoff.NewExponentialBackOff()
	ELASTIC_USER_PASSWORD := os.Getenv("ELASTIC_USER_PASSWORD")
	es, err := elasticsearch.NewClient(elasticsearch.Config{
		Username:      "elastic",
		Password:      ELASTIC_USER_PASSWORD,
		RetryOnStatus: []int{502, 503, 504, 429},
		RetryBackoff: func(i int) time.Duration {
			if i == 1 {
				retryBackoff.Reset()
			}
			return retryBackoff.NextBackOff()
		},
		MaxRetries: 5,
	})
	if err != nil {
		log.Printf("elasticsearch.NewClient:%v", err)
		return nil, err
	}

	return &esservice{client: es}, nil
}

func (es *esservice) BulkInsertProducts(indexName string, addresses []*model.Address) error {
	bi, err := esutil.NewBulkIndexer(esutil.BulkIndexerConfig{
		Index:         indexName,
		Client:        es.client,
		NumWorkers:    runtime.NumCPU(),
		FlushBytes:    int(5e+6),
		FlushInterval: 30 * time.Second,
	})
	if err != nil {
		return err
	}

	start := time.Now().UTC()
	for _, a := range addresses {
		data, err := json.Marshal(a)
		if err != nil {
			return err
		}
		err = bi.Add(
			context.Background(),
			esutil.BulkIndexerItem{
				Index:      indexName,
				Action:     "index",
				DocumentID: fmt.Sprint(a.ID),
				Body:       bytes.NewReader(data),
				OnFailure: func(ctx context.Context, item esutil.BulkIndexerItem, res esutil.BulkIndexerResponseItem, err error) {
					if err != nil {
						log.Printf("ERROR index: %s", err)
					} else {
						log.Printf("ERROR index: %s: %s", res.Error.Type, res.Error.Reason)
					}
				},
			},
		)
		if err != nil {
			log.Printf("ES Bulk insert error: %s", err)
			return err
		}
	}
	if err := bi.Close(context.Background()); err != nil {
		return err
	}
	biStats := bi.Stats()
	dur := time.Since(start)
	if biStats.NumFailed > 0 {
		log.Fatalf(
			"Indexed [%s] documents with [%s] errors in %s (%s docs/sec)",
			humanize.Comma(int64(biStats.NumFlushed)),
			humanize.Comma(int64(biStats.NumFailed)),
			dur.Truncate(time.Millisecond),
			humanize.Comma(int64(1000.0/float64(dur/time.Millisecond)*float64(biStats.NumFlushed))),
		)
	} else {
		log.Printf(
			"Successfully indexed [%s] documents in %s (%s docs/sec)",
			humanize.Comma(int64(biStats.NumFlushed)),
			dur.Truncate(time.Millisecond),
			humanize.Comma(int64(1000.0/float64(dur/time.Millisecond)*float64(biStats.NumFlushed))),
		)
	}

	return nil
}