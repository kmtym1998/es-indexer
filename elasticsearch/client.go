package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	e "errors"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/kmtym1998/es-indexer/logger"
	"golang.org/x/exp/slog"

	"github.com/cenkalti/backoff/v4"
	"github.com/dustin/go-humanize"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esutil"
)

type client struct {
	client *elasticsearch.Client
	logger logger.Logger
}

type DocumentNode interface {
	NodeIdentifier() string
}

type DocumentNodeList interface {
	IndexName() string
	ToList() []DocumentNode
}

func NewClient(l logger.Logger) (*client, error) {
	retryBackoff := backoff.NewExponentialBackOff()
	ELASTIC_USER_PASSWORD := os.Getenv("ELASTIC_USER_PASSWORD")

	esClient, err := elasticsearch.NewClient(elasticsearch.Config{
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
		return nil, errors.Wrap(err, "failed to create elasticsearch client")
	}

	return &client{
		client: esClient,
		logger: l,
	}, nil
}

func (es *client) BulkInsert(ctx context.Context, nodeList DocumentNodeList) error {
	bi, err := esutil.NewBulkIndexer(esutil.BulkIndexerConfig{
		Index:         nodeList.IndexName(),
		Client:        es.client,
		NumWorkers:    runtime.NumCPU(),
		FlushBytes:    int(5e+6),
		FlushInterval: 30 * time.Second,
	})
	if err != nil {
		return errors.Wrap(err, "failed to create bulk indexer")
	}
	es.logger.Debug("new bulk indexer created")

	es.logger.Debug("start bulk insert")

	var errList []error
	start := time.Now().UTC()
	for _, node := range nodeList.ToList() {
		body, err := json.Marshal(node)
		if err != nil {
			return errors.Wrap(err, "json marshal error")
		}

		if err := bi.Add(
			ctx,
			esutil.BulkIndexerItem{
				Index:      nodeList.IndexName(),
				Action:     "index",
				DocumentID: node.NodeIdentifier(),
				Body:       bytes.NewReader(body),
				OnFailure: func(
					failureCtx context.Context,
					item esutil.BulkIndexerItem,
					res esutil.BulkIndexerResponseItem,
					err error,
				) {
					errMsg := fmt.Sprintf("failed to index document: %s", res.DocumentID)
					l := es.logger.WithCtx(failureCtx)
					if b, err := json.Marshal(res); err == nil {
						l = l.With(slog.Any("errResponse", string(b)))
					}

					if err != nil {
						errList = append(errList, errors.Wrap(err, errMsg))
					} else {
						errList = append(errList, errors.Wrap(err, errMsg))
					}

					l.Error(errMsg, err)
				},
			},
		); err != nil {
			errList = append(errList, errors.Wrap(err, "failed to add bulk indexer item"))
		}
	}

	if len(errList) > 0 {
		return e.Join(errList...)
	}

	if err := bi.Close(ctx); err != nil {
		es.logger.Warning("failed to close bulk indexer", err)
	}

	biStats := bi.Stats()
	durFields := slog.Duration("duration", time.Since(start).Truncate(time.Millisecond))
	statField := slog.Any("stats", biStats)

	if biStats.NumFailed > 0 {
		es.logger.Warning(
			fmt.Sprintf(
				"some documents failed to index [%s]",
				nodeList.IndexName(),
			),
			durFields,
			statField,
		)
	} else {
		es.logger.Info(
			fmt.Sprintf(
				"successfully indexed [%s] documents in %s",
				nodeList.IndexName(),
				humanize.IBytes(uint64(biStats.NumFlushed)),
			),
			durFields,
			statField,
		)
	}

	return nil
}
