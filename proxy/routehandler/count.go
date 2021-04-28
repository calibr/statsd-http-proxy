package routehandler

import (
	"net/http"
	"strconv"
	"log"
)

// Handle StatsD Count request
func (routeHandler *RouteHandler) handleCountRequest(w http.ResponseWriter, r *http.Request, keys [2]string) {
	// get count value
	var value = 1
	valuePostFormValue := r.PostFormValue("value")
	if valuePostFormValue != "" {
		var err error
		value, err = strconv.Atoi(valuePostFormValue)
		if err != nil {
			http.Error(w, "Invalid value specified", 400)
		}
	}

	// get sample rate
	var sampleRate float64 = 1
	sampleRatePostFormValue := r.PostFormValue("sampleRate")
	if sampleRatePostFormValue != "" {
		var err error
		sampleRate, err = strconv.ParseFloat(sampleRatePostFormValue, 32)
		if err != nil {
			http.Error(w, "Invalid sample rate specified", 400)
		}
	}

	// send request
	for _, key := range keys {
    if key != "" {
    	log.Printf("Sending %s %d", key, value)
    	routeHandler.statsdClient.Count(key, value, float32(sampleRate))
    }
	}
}
