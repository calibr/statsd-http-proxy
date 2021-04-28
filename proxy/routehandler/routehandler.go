package routehandler

import (
	"net/http"

	"github.com/calibr/statsd-http-proxy/proxy/statsdclient"
	"log"
)

// RouteHandler as a collection of route handlers
type RouteHandler struct {
	statsdClient statsdclient.StatsdClientInterface
	metricPrefix string
	keyPartHeader string
}

// NewRouteHandler creates collection of route handlers
func NewRouteHandler(
	statsdClient statsdclient.StatsdClientInterface,
	metricPrefix string,
	keyPartHeader string,
) *RouteHandler {
	// prepare metric prefix
	if metricPrefix != "" && (metricPrefix)[len(metricPrefix)-1:] != "." {
		metricPrefix = metricPrefix + "."
	}

	// build route handler
	routeHandler := RouteHandler{
		statsdClient,
		metricPrefix,
		keyPartHeader,
	}

	return &routeHandler
}

// GetFullyQualifiedMetricKey return metric key with passed suffix and pre-configured prefix
func (routeHandler *RouteHandler) getFullyQualifiedMetricKey(metricKeySuffix string) string {
	return routeHandler.metricPrefix + metricKeySuffix
}

//HandleMetric reads count, gauge, timing and set metrics from HTTP and sent them to StatsD
func (routeHandler *RouteHandler) HandleMetric(
	w http.ResponseWriter,
	r *http.Request,
	metricType string,
	metricKeySuffix string,
	keyPartHeader string,
) {
	// get fully qualified metric key
	metricKey := routeHandler.getFullyQualifiedMetricKey(metricKeySuffix)
	var metricKeys [2]string
	metricKeys[0] = metricKey
	metricKeys[1] = ""

	if keyPartHeader != "" {
		keyPartValue := r.Header.Get(keyPartHeader)
		log.Printf("Value: %s", keyPartValue)
		if keyPartValue != "" {
			metricKeys[1] = "by_country." + keyPartValue + "." + metricKey
		}
	}

	// run handler
	switch metricType {
	case "count":
		routeHandler.handleCountRequest(w, r, metricKeys)
	case "gauge":
		routeHandler.handleGaugeRequest(w, r, metricKeys)
	case "timing":
		routeHandler.handleTimingRequest(w, r, metricKeys)
	case "set":
		routeHandler.handleSetRequest(w, r, metricKeys)
	}
}
