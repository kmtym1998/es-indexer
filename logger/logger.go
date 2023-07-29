package logger

import (
	"context"
	"os"

	"golang.org/x/exp/slog"
)

type Logger struct {
	ctx     context.Context
	logger  *slog.Logger
	onError ErrorHookFunc
}

type Opts struct {
	Level       slog.Level
	ReplaceAttr func(groups []string, a slog.Attr) slog.Attr
	OnError     ErrorHookFunc
}

type ErrorHookFunc func(ctx context.Context, msg string, err error, arg ...any)

func New(opts Opts) *Logger {
	return &Logger{
		ctx: context.Background(),
		logger: slog.New(
			slog.NewTextHandler(
				os.Stdout,
				&slog.HandlerOptions{Level: opts.Level, ReplaceAttr: opts.ReplaceAttr},
			),
		),
		onError: opts.OnError,
	}
}

func (l *Logger) WithCtx(ctx context.Context) *Logger {
	return &Logger{
		ctx:     ctx,
		logger:  l.logger,
		onError: l.onError,
	}
}

func (l *Logger) With(args ...any) *Logger {
	return &Logger{
		logger:  l.logger.With(args...),
		onError: l.onError,
	}
}

func (l *Logger) Debug(msg string, arg ...any) {
	l.logger.Debug(msg, arg...)
}

func (l *Logger) Info(msg string, arg ...any) {
	l.logger.Info(msg, arg...)
}

func (l *Logger) Warning(msg string, arg ...any) {
	l.logger.Warn(msg, arg...)
}

func (l *Logger) Error(msg string, err error, arg ...any) {
	l.logger.Error(msg)

	go func() {
		// エラーログ出力後なにかやりたい時 (sentry に送るとか) は OnError() を呼び元から渡す
		if l.onError != nil {
			l.onError(l.ctx, msg, err)
		}
	}()
}
