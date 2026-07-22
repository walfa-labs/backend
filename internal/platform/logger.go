package platform

import (
	"github.com/gookit/slog"
	"github.com/gookit/slog/handler"
	"github.com/gookit/rotatefile"
)

// NewLogger creates a production-ready slog.Logger.
//
// Production: writes to a daily-rotating file (logs/app.log, 10 MB cap,
// 7 backups, gzip-compressed) plus a console mirror, thresholded at Info.
// Development: colored console output at Debug level.
//
// The returned *slog.Logger is safe for concurrent use.
func NewLogger(env string) *slog.Logger {
	if env == "production" {
		fileH, err := handler.NewRotateFileHandler(
			"logs/app.log",
			rotatefile.EveryDay,
			handler.WithMaxSize(10*1024*1024), // 10 MB
			handler.WithBackupNum(7),
			handler.WithCompress(true),
			handler.WithLogLevel(slog.InfoLevel),
		)
		if err != nil {
			panic(err)
		}

		consoleH := handler.ConsoleWithMaxLevel(slog.InfoLevel)
		return slog.NewWithHandlers(fileH, consoleH)
	}

	// Dev: colored console, Debug and above (i.e. all levels).
	consoleH := handler.ConsoleWithMaxLevel(slog.DebugLevel)
	return slog.NewWithHandlers(consoleH)
}
