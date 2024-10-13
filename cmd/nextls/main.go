package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/gopherd/core/builder"
	"github.com/sourcegraph/jsonrpc2"

	_ "github.com/mkideal/next/cmd/nextls/internal/handler/next"
	_ "github.com/mkideal/next/cmd/nextls/internal/handler/npl"
)

var flags struct {
	log struct {
		file  string
		level string
	}
}

func main() {
	logFile, _ := os.UserHomeDir()
	if logFile != "" {
		logFile = filepath.Join(logFile, ".next", "log", "nextls.log")
	}
	flag.StringVar(&flags.log.file, "log.file", logFile, "Log file path for debug")
	flag.StringVar(&flags.log.level, "log.level", "INFO", "Log level")
	flag.Parse()

	if flag.NArg() == 1 && flag.Arg(0) == "version" {
		fmt.Fprintln(os.Stdout, builder.Info())
		os.Exit(0)
	}

	var enableLogging bool
	if flags.log.file != "" && flags.log.level != "" {
		var level slog.Level
		if err := level.UnmarshalText([]byte(flags.log.level)); err == nil {
			dir := filepath.Dir(flags.log.file)
			if os.MkdirAll(dir, 0755) == nil {
				if file, err := os.OpenFile(flags.log.file, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644); err == nil {
					defer file.Close()
					slog.SetDefault(slog.New(slog.NewTextHandler(file, &slog.HandlerOptions{
						Level: level,
					})))
					enableLogging = true
				}
			}
		}
	}
	if !enableLogging {
		slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))
	}

	server := &Server{}
	conn := jsonrpc2.NewConn(
		context.Background(),
		jsonrpc2.NewBufferedStream(stdrwc{}, jsonrpc2.VSCodeObjectCodec{}),
		jsonrpc2.HandlerWithError(server.Handle),
	)
	slog.Debug("Connection established, waiting for requests")
	<-conn.DisconnectNotify()
	slog.Debug("Connection closed, server shutting down")
}

type stdrwc struct{}

func (stdrwc) Read(p []byte) (n int, err error) {
	return os.Stdin.Read(p)
}

func (stdrwc) Write(p []byte) (n int, err error) {
	return os.Stdout.Write(p)
}

func (stdrwc) Close() error {
	return nil
}
