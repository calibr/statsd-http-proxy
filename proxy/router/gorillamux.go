package router

import (
	"net/http"

	"github.com/calibr/statsd-http-proxy/proxy/middleware"
	"github.com/calibr/statsd-http-proxy/proxy/routehandler"
	"github.com/gorilla/mux"
)

// NewGorillaMuxRouter creates Gorilla Mux router
func NewGorillaMuxRouter(
	routeHandler *routehandler.RouteHandler,
	tokenSecret string,
) http.Handler {
	// build router
	router := mux.NewRouter().StrictSlash(true)

	// register common middlewares
	router.Use(middleware.ValidateCORS)

	// register http request handlers
	router.Handle(
		"/heartbeat",
		http.HandlerFunc(routeHandler.HandleHeartbeatRequest),
	)

	router.Handle(
		"/{type:(?:count|gauge|timing|set)}/{key}",
		middleware.ValidateJWT(
			http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					// get variables from path
					vars := mux.Vars(r)
					metricType := vars["type"]
					metricKeySuffix := vars["key"]

					routeHandler.HandleMetric(w, r, metricType, metricKeySuffix, "")
				},
			),
			tokenSecret,
		),
	).Methods(http.MethodPost, http.MethodOptions)

	return router
}
