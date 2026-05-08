package main

import (
	"context"
	"explorium/codify/cmd"
	"log/slog"
	"net/http"
	_ "net/http/pprof"
	"os"

	fang "charm.land/fang/v2"
)

func main() {
	if os.Getenv("CODIFY_PROFILE") != "" {
		if err := http.ListenAndServe("localhost:6060", nil); err != nil {
			slog.Error("Failed to listen pprof on localhost:6060", err)
		}
	}

	// 日志没好之前，先丢弃打印，防止打印输出到控制台
	slog.SetDefault(slog.New(slog.DiscardHandler))

	if err := fang.Execute(
		context.Background(),
		cmd.NewCodifyCommand(),
		fang.WithNotifySignal(os.Interrupt),
	); err != nil {
		os.Exit(1)
	}
}
