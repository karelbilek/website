package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/karelbilek/website/fileserver"
	"github.com/karelbilek/website/gemini"
	"github.com/karelbilek/website/middleware"
)

const purl = "karelbilek.com"

//go:embed public
var staticfs embed.FS

const (
	defaultMaxConns = 128
	defaultTimeout  = 5
)

func main() {
	go mainProxy()

	var err error
	var flogger, dlogger *log.Logger

	flogger, err = setupLogger()
	if err != nil {
		log.Fatal(err)
	}

	dlogger, err = setupLogger()
	if err != nil {
		log.Fatal(err)
	}

	logprefix := "gem: "

	mux := gemini.NewMux()
	mux.Use(middleware.Logger(flogger, logprefix))
	subfs, err := fs.Sub(staticfs, "public")
	if err != nil {
		panic(err)
	}
	mux.Handle(gemini.HandlerFunc(fileserver.Serve(subfs)))

	server := &gemini.Server{
		Addr:            "0.0.0.0:1965",
		Hostname:        purl,
		TLSConfigLoader: setupCertificate(),
		Handler:         mux,
		MaxOpenConns:    defaultMaxConns,
		ReadTimeout:     time.Duration(defaultTimeout) * time.Second,
		Logger:          dlogger,
	}

	confirm := make(chan struct{}, 1)

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, gemini.ErrServerClosed) {
			log.Fatalf("ListenAndServe terminated unexpectedly: %v", err)
		}

		close(confirm)
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(defaultTimeout)*time.Second)
	if err := server.Shutdown(ctx); err != nil {
		cancel()
		log.Fatalf("ListenAndServe shutdown with error: %v", err)
	}

	<-confirm
	cancel()
}

//go:embed gemini-key.rsa
var geminiKeyRsa []byte

//go:embed gemini-cert.pem
var geminiCertPem []byte

func setupCertificate() func() (*tls.Config, error) {
	geminiKeyRsa := bytes.TrimSpace(geminiKeyRsa)
	geminiCertPem := bytes.TrimSpace(geminiCertPem)

	geminiCertPemS := geminiCertPem[0:50]
	return func() (*tls.Config, error) {
		cert, err := tls.X509KeyPair(geminiCertPem, geminiKeyRsa)
		if err != nil {
			return nil, fmt.Errorf("load x509 keypair: %w (%d %d %+v)", err, len(geminiCertPem), len(geminiKeyRsa), geminiCertPemS)
		}
		return gemini.TLSConfig(purl, cert), nil
	}
}

func setupLogger() (*log.Logger, error) {
	logger := log.New(os.Stdout, "", log.LUTC|log.Ldate|log.Ltime)

	return logger, nil
}

func setupFileLogging(logger *log.Logger, logpath string) (*os.File, error) {
	logfile, err := os.OpenFile(logpath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return logfile, fmt.Errorf("log file open: %w", err)
	}

	logger.SetOutput(logfile)
	return logfile, nil
}
