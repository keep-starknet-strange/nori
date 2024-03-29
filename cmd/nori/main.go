package main

import (
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/BurntSushi/toml"
	"github.com/ethereum/go-ethereum/log"
	"github.com/keep-starknet-strange/nori"
	"golang.org/x/exp/slog"
)

var (
	GitVersion = ""
	GitCommit  = ""
	GitDate    = ""
)

func main() {
	// Set up logger with a default INFO level in case we fail to parse flags.
	// Otherwise the final critical log won't show what the parsing error was.
    log.SetDefault(log.NewLogger(log.LogfmtHandlerWithLevel(os.Stdout, log.LvlInfo)))

	log.Info("starting nori", "version", GitVersion, "commit", GitCommit, "date", GitDate)

	if len(os.Args) < 2 {
		log.Crit("must specify a config file on the command line")
	}

	config := new(nori.Config)
	if _, err := toml.DecodeFile(os.Args[1], config); err != nil {
		log.Crit("error reading config file", "err", err)
	}

	// update log level from config
    logLevelString := config.Server.LogLevel
    var logLevel slog.Level
    switch logLevelString {
    case "trace":
        logLevel = log.LevelTrace
    case "debug":
        logLevel = log.LevelDebug
    case "info":
        logLevel = log.LevelInfo
    case "warn":
        logLevel = log.LevelWarn
    case "error":
        logLevel = log.LevelError
    case "crit":
        logLevel = log.LevelCrit
    default:
        logLevel = log.LevelInfo
        log.Warn("invalid server.log_level set: " + logLevelString)
    }
    log.SetDefault(log.NewLogger(log.LogfmtHandlerWithLevel(os.Stdout, logLevel)))

	if config.Server.EnablePprof {
		log.Info("starting pprof", "addr", "0.0.0.0", "port", "6060")
		pprofSrv := StartPProf("0.0.0.0", 6060)
		log.Info("started pprof server", "addr", pprofSrv.Addr)
		defer func() {
			if err := pprofSrv.Close(); err != nil {
				log.Error("failed to stop pprof server", "err", err)
			}
		}()
	}

	_, shutdown, err := nori.Start(config)
	if err != nil {
		log.Crit("error starting nori", "err", err)
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	recvSig := <-sig
	log.Info("caught signal, shutting down", "signal", recvSig)
	shutdown()
}

func StartPProf(hostname string, port int) *http.Server {
	mux := http.NewServeMux()

	// have to do below to support multiple servers, since the
	// pprof import only uses DefaultServeMux
	mux.Handle("/debug/pprof/", http.HandlerFunc(pprof.Index))
	mux.Handle("/debug/pprof/cmdline", http.HandlerFunc(pprof.Cmdline))
	mux.Handle("/debug/pprof/profile", http.HandlerFunc(pprof.Profile))
	mux.Handle("/debug/pprof/symbol", http.HandlerFunc(pprof.Symbol))
	mux.Handle("/debug/pprof/trace", http.HandlerFunc(pprof.Trace))

	addr := net.JoinHostPort(hostname, strconv.Itoa(port))
	srv := &http.Server{
		Handler: mux,
		Addr:    addr,
	}

	go srv.ListenAndServe()

	return srv
}
