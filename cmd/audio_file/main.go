package main

import (
	"context"
	"flag"

	"github.com/kmtym1998/es-indexer/indexer"
	"github.com/kmtym1998/es-indexer/logger"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"golang.org/x/exp/slog"
)

func main() {
	ctx := context.Background()
	l := logger.New(logger.Opts{
		Level: slog.LevelDebug,
	})
	l = l.WithCtx(ctx)

	rootDir := flag.String("rootDir", "", "オーディオファイルのルートディレクトリ")
	flag.Parse()

	if lo.FromPtr(rootDir) == "" {
		err := errors.New("rootDir is required")
		l.Error(err.Error(), err)

		return
	}

	l.Debug(*rootDir)

	if err := indexer.NewAudioIndexer(ctx).Run(*rootDir); err != nil {
		l.Error("failed to index audio files", err)
	}
}
