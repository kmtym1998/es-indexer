package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kmtym1998/es-indexer/elasticsearch"
	"github.com/kmtym1998/es-indexer/logger"
	"github.com/kmtym1998/es-indexer/node"
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

	defer func() {
		if err := recover(); err != nil {
			l.Error("panic occurred", errors.New(fmt.Sprint(err)))
		}
	}()

	rootDir := flag.String("rootDir", "", "オーディオファイルのルートディレクトリ")
	flag.Parse()

	if lo.FromPtr(rootDir) == "" {
		err := errors.New("rootDir is required")
		l.Error(err.Error(), err)

		return
	}

	l.Debug(*rootDir)

	var documentList node.AudioFileList
	if err := filepath.WalkDir(*rootDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			l.Error(fmt.Sprintf("error walking on: %s", path), err)
			return err
		}

		if d.IsDir() || lo.Contains([]string{".DS_Store", "Thumbs.db"}, d.Name()) {
			return nil
		}

		l.Debug("loading...", slog.String("path", path))

		documentNode, err := node.NewAudioFileNode(path)
		if err != nil {
			if err == node.ErrNotAudioFile {
				l.Warning("not audio file", slog.String("path", path))

				return nil
			}

			l.Error("failed to create audio file node", err)
		}

		documentList = append(documentList, lo.FromPtr(documentNode))

		return nil
	}); err != nil {
		l.Error("failed to walk directory", err)

		return
	}

	// debug
	b, _ := json.Marshal(documentList)
	f, _ := os.Create("./tmp/audio_files.json")
	defer f.Close()
	fmt.Fprint(f, string(b))

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
