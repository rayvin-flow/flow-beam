package main

import (
	"beam/beam"
	"context"
	"encoding/json"
	"fmt"
	"github.com/onflow/cadence"
	jsoncdc "github.com/onflow/cadence/encoding/json"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"
)

func serve(ctx context.Context, listenPort string) (err error) {
	mux := http.NewServeMux()

	mux.HandleFunc("/events", HandleRequest(HandleEventsRequest))
	mux.HandleFunc("/latest-block-height", HandleRequest(HandleLatestBlockHeightRequest))
	mux.HandleFunc("/execute-script", HandleRequest(ExecuteScript))

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", listenPort),
		Handler: mux,
	}
	log.Warn().Msg(fmt.Sprintf("Listening on port: %s", listenPort))

	go func() {
		if err = srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error().Msg(fmt.Sprintf("Listen error: %s", err))
		}
	}()

	log.Debug().Msg(fmt.Sprintf("ACCESS_NODES=%s", os.Getenv("ACCESS_NODES")))

	accessNodes := beam.GetAccessNodes()

	log.Debug().Msg("Access Nodes:")
	for _, node := range accessNodes {
		var legacy = 0
		if node.IsLegacy {
			legacy = 1
		}
		log.Debug().Msg(fmt.Sprintf("%d - %d: %s (Legacy=%d)", node.StartHeight, node.EndHeight, node.Address, legacy))
	}

	log.Debug().Msg("Server started")

	<-ctx.Done()

	log.Debug().Msg("Server stopped")

	ctxShutDown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		cancel()
	}()

	if err = srv.Shutdown(ctxShutDown); err != nil {
		log.Error().Msg(fmt.Sprintf("Server shutdown failed: %s", err))
	}

	log.Debug().Msg("Server exited properly")

	if err == http.ErrServerClosed {
		err = nil
	}

	return
}

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	logLevels := make(map[string]zerolog.Level)

	logLevels["DEBUG"] = zerolog.DebugLevel
	logLevels["INFO"] = zerolog.InfoLevel
	logLevels["WARN"] = zerolog.WarnLevel
	logLevels["ERROR"] = zerolog.ErrorLevel

	logLevel := os.Getenv("APP_LOG_LEVEL")

	if len(logLevel) > 0 {
		zerolog.SetGlobalLevel(logLevels[strings.ToUpper(logLevel)])
	}

	listenPort := os.Getenv("LISTEN_PORT")

	if len(listenPort) == 0 {
		listenPort = "8080"
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		oscall := <-c
		log.Debug().Msg(fmt.Sprintf("System call:%+v", oscall))
		cancel()
	}()

	if err := serve(ctx, listenPort); err != nil {
		log.Error().Msg(fmt.Sprintf("Failed to serve: %s", err))
	}
}

func HandleRequest(fn func(r *http.Request) (int, interface{})) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				type Error struct {
					Message interface{}
				}

				log.Error().Msg(fmt.Sprintf("%s", r))

				out, _ := json.Marshal(Error{Message: r})
				w.WriteHeader(500)
				w.Header().Add("Content-Type", "application/json")
				w.Write(out)
			}
		}()

		statusCode, response := fn(r)

		out, _ := json.Marshal(response)
		w.WriteHeader(statusCode)
		w.Header().Add("Content-Type", "application/json")
		w.Write(out)
	}
}

func HandleEventsRequest(r *http.Request) (int, interface{}) {
	log.Debug().Msg(fmt.Sprintf("HandleEventsRequest: %s", r.URL.Query()))
	start, _ := strconv.ParseUint(r.URL.Query()["start"][0], 10, 64)
	end, _ := strconv.ParseUint(r.URL.Query()["end"][0], 10, 64)
	eventType := r.URL.Query()["eventType"][0]

	response, err := beam.GetEvents(eventType, start, end)

	type ErrorResponse struct {
		ApiCalls uint32
		Error    interface{}
	}

	if err != nil {
		return 500, ErrorResponse{
			ApiCalls: response.ApiCalls,
			Error:    fmt.Sprintf("%s", err),
		}
	}

	return 200, response
}

func HandleLatestBlockHeightRequest(r *http.Request) (int, interface{}) {
	log.Debug().Msg("HandleLatestBlockHeightRequest")

	response, err := beam.GetLatestBlockHeight()

	type ErrorResponse struct {
		ApiCalls uint32
		Error    interface{}
	}

	if err != nil {
		return 500, ErrorResponse{
			ApiCalls: response.ApiCalls,
			Error:    fmt.Sprintf("%s", err),
		}
	}

	return 200, response
}

type ExecuteScriptBody struct {
	Script    string
	Arguments []interface{}
}

func ExecuteScript(r *http.Request) (int, interface{}) {
	b, err := ioutil.ReadAll(r.Body)

	if err != nil {
		panic(err)
	}

	log.Debug().Msg(fmt.Sprintf("ExecuteScript: %s", b))

	var body ExecuteScriptBody

	err = json.Unmarshal(b, &body)

	if err != nil {
		panic(err)
	}

	arguments := make([]cadence.Value, 0)

	for _, arg := range body.Arguments {
		b, err := json.Marshal(arg)

		if err != nil {
			panic(err)
		}

		cdc, err := jsoncdc.Decode(b)

		if err != nil {
			panic(err)
		}

		arguments = append(arguments, cdc)
	}

	if err != nil {
		panic(err)
	}

	response, err := beam.ExecuteScript(body.Script, arguments)

	type ErrorResponse struct {
		ApiCalls uint32
		Error    interface{}
	}

	if err != nil {
		return 500, ErrorResponse{
			ApiCalls: response.ApiCalls,
			Error:    fmt.Sprintf("%s", err),
		}
	}

	return 200, response
}
