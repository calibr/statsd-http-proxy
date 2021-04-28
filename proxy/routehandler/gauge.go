package routehandler

import (
	"net/http"
	"strconv"
)

// Handle StatsD Gauge request
func (routeHandler *RouteHandler) handleGaugeRequest(w http.ResponseWriter, r *http.Request, keys [2]string) {
	var key = keys[0]

	// get gauge shift
	shiftPostFormValue := r.PostFormValue("shift")
	if shiftPostFormValue != "" {
		// get value
		value, err := strconv.Atoi(shiftPostFormValue)
		if err != nil {
			http.Error(w, "Invalid gauge shift specified", 400)
		}
		// send request
		routeHandler.statsdClient.GaugeShift(key, value)
		return
	}

	// get gauge value
	var value = 1
	valuePostFormValue := r.PostFormValue("value")
	if valuePostFormValue != "" {
		// get value
		var err error
		value, err = strconv.Atoi(valuePostFormValue)
		if err != nil {
			http.Error(w, "Invalid gauge value specified", 400)
		}
	}

	// send gauge value request
	routeHandler.statsdClient.Gauge(key, value)
}
