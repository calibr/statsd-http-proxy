package proxy

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/calibr/statsd-http-proxy/proxy/routehandler"
	"github.com/calibr/statsd-http-proxy/proxy/router"
	"github.com/calibr/statsd-http-proxy/proxy/statsdclient"
)

// Server is a proxy server between HTTP REST API and UDP Connection to StatsD
type Server struct {
	httpAddress  string
	httpServer   *http.Server
	statsdClient statsdclient.StatsdClientInterface
	tlsCert      string
	tlsKey       string
}

// NewServer creates new instance of StatsD HTTP Proxy
func NewServer(
	httpHost string,
	httpPort int,
	httpReadTimeout int,
	httpWriteTimeout int,
	httpIdleTimeout int,
	statsdHost string,
	statsdPort int,
	tlsCert string,
	tlsKey string,
	metricPrefix string,
	tokenSecret string,
	verbose bool,
	httpRouterName string,
	statsdClientName string,
	keyPartHeader string,
) *Server {
	log.Printf("Creating a HTTP server")

	// configure logging
	var logOutput io.Writer
	if verbose == true {
		logOutput = os.Stderr
	} else {
		logOutput = ioutil.Discard
	}

	log.SetOutput(logOutput)

	logger := log.New(logOutput, "", log.LstdFlags)

	// create StatsD Client
	var statsdClient statsdclient.StatsdClientInterface
	switch statsdClientName {
	case "Cactus":
		statsdClient = statsdclient.NewCactusClient(statsdHost, statsdPort)
	case "GoMetric":
		statsdClient = statsdclient.NewGoMetricClient(statsdHost, statsdPort)
	default:
		panic("Passed statsd client not supported")
	}

	// build route handler
	routeHandler := routehandler.NewRouteHandler(
		statsdClient,
		metricPrefix,
		keyPartHeader,
	)

	// build router
	var httpServerHandler http.Handler
	switch httpRouterName {
	case "HttpRouter":
		httpServerHandler = router.NewHTTPRouter(routeHandler, tokenSecret, keyPartHeader)
	case "GorillaMux":
		httpServerHandler = router.NewGorillaMuxRouter(routeHandler, tokenSecret)

	default:
		panic("Passed HTTP router not supported")
	}

	// get HTTP server address to bind
	httpAddress := fmt.Sprintf("%s:%d", httpHost, httpPort)

	// create http server
	httpServer := &http.Server{
		Addr:           httpAddress,
		Handler:        httpServerHandler,
		ErrorLog:       logger,
		ReadTimeout:    time.Duration(httpReadTimeout) * time.Second,
		WriteTimeout:   time.Duration(httpWriteTimeout) * time.Second,
		IdleTimeout:    time.Duration(httpIdleTimeout) * time.Second,
		MaxHeaderBytes: 1 << 11,
	}

	statsdHTTPProxyServer := Server{
		httpAddress,
		httpServer,
		statsdClient,
		tlsCert,
		tlsKey,
	}

	return &statsdHTTPProxyServer
}

// Listen starts listening HTTP connections
func (proxyServer *Server) Listen() {
	log.Printf("Listening")
	// prepare for gracefull shutdown
	gracefullStopSignalHandler := make(chan os.Signal, 1)
	signal.Notify(gracefullStopSignalHandler, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// start HTTP/HTTPS proxy to StatsD
	go func() {
		log.Printf("Starting HTTP server at %s", proxyServer.httpAddress)

		// open StatsD connection
		proxyServer.statsdClient.Open()
		defer proxyServer.statsdClient.Close()

		// open HTTP connection
		var err error
		if len(proxyServer.tlsCert) > 0 && len(proxyServer.tlsKey) > 0 {
			err = proxyServer.httpServer.ListenAndServeTLS(proxyServer.tlsCert, proxyServer.tlsKey)
		} else {
			err = proxyServer.httpServer.ListenAndServe()
		}

		if err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
			log.Fatal("Can not start HTTP server")
		}
	}()

	<-gracefullStopSignalHandler

	// Gracefull shutdown
	log.Printf("Stopping HTTP server")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		cancel()
	}()

	if err := proxyServer.httpServer.Shutdown(ctx); err != nil {
		log.Fatalf("HTTP Server Shutdown Failed:%+v", err)
	}

	log.Printf("HTTP server stopped successfully")
}
