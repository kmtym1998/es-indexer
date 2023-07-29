package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os"

	"github.com/kmtym1998/es-indexer/elasticsearch"
	"github.com/kmtym1998/es-indexer/logger"
	"github.com/kmtym1998/es-indexer/node"
	"github.com/pkg/errors"
	"golang.org/x/exp/slog"

	"github.com/ktnyt/go-moji"
)

func main() {
	ctx := context.Background()
	l := logger.New(logger.Opts{
		Level: slog.LevelDebug,
	})
	l = l.WithCtx(ctx)

	defer func() {
		if err := recover(); err != nil {
			l.Error("panic occurred", errors.New(fmt.Sprint(err)))
		}
	}()

	f, err := os.Open("./data/address_ken_all.csv")
	if err != nil {
		l.Error("failed to open csv", err)
	}
	defer f.Close()

	r := csv.NewReader(f)

	var documentList node.AddressList
	var errList []error
	for i := 1; ; i++ {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			errList = append(errList, errors.Wrap(err, "failed to read csv"))
			continue
		}

		func() {
			defer func() {
				if err := recover(); err != nil {
					l.Warning("failed to parse csv row", errors.New(fmt.Sprint(err)))
				}
			}()

			documentList = append(documentList, node.Address{
				ID:               i,
				ZipCode:          record[2],
				PrefectureKana:   moji.Convert(record[3], moji.HK, moji.ZK),
				MunicipalityKana: moji.Convert(record[4], moji.HK, moji.ZK),
				TownKana:         moji.Convert(record[5], moji.HK, moji.ZK),
				Prefecture:       record[6],
				Municipality:     record[7],
				Town:             record[8],
				Concat:           fmt.Sprintf("%s%s%s", record[6], record[7], record[8]),
				ConcatKana:       fmt.Sprintf("%s%s%s", moji.Convert(record[3], moji.HK, moji.ZK), moji.Convert(record[4], moji.HK, moji.ZK), moji.Convert(record[5], moji.HK, moji.ZK)),
			})
		}()
	}

	if len(errList) > 0 {
		for _, err := range errList {
			l.Error("failed to read csv", err)
		}
	}

	es, err := elasticsearch.NewClient(l)
	if err != nil {
		l.Error("failed to create elasticsearch client", err)
	}

	if err := es.BulkInsert(ctx, documentList); err != nil {
		if errList, ok := err.(interface{ Unwrap() []error }); ok {
			for _, err := range errList.Unwrap() {
				fmt.Println(err)
			}
		}

		l.Error("failed to bulk insert", err)
	}
}
