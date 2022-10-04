// Package api provides the actual ecowatter HTTP API
package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/alexedwards/flow"

	"github.com/pkuhner/ecowatter/internal/ecowatt"
)

type API struct {
	Router  *flow.Mux
	Ecowatt *ecowatt.Ecowatt
}

func New(ecowatt *ecowatt.Ecowatt) *API {
	return &API{
		Ecowatt: ecowatt,
	}
}

func SetContentTypeJSON(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func (a *API) ListSignals(w http.ResponseWriter, r *http.Request) {
	signals, err := a.Ecowatt.GetSignals()
	if err != nil {
		http.Error(w, "Couldn't retrieve signals", http.StatusInternalServerError)
		return
	}

	res, err := json.Marshal(signals)
	if err != nil {
		http.Error(w, "Couldn't marshal signals", http.StatusInternalServerError)
		return
	}

	w.Write(res)
}

func (a *API) ListDaySignal(w http.ResponseWriter, r *http.Request) {
	dayStr := flow.Param(r.Context(), "day")
	if dayStr == "" {
		http.Error(w, "Day cannot be empty", http.StatusBadRequest)
		return
	}

	day, err := strconv.ParseInt(dayStr, 10, 0)
	if err != nil {
		http.Error(w, "Day is not an integer", http.StatusBadRequest)
		return
	}

	if !(day >= 0 && day <= 3) {
		http.Error(w, "Valid values for day are 0 (today), 1 (tomorrow), 2, and 3", http.StatusBadRequest)
		return
	}

	signal, err := a.Ecowatt.GetSignal(int(day))
	if err != nil {
		http.Error(w, "Couldn't retrieve signal", http.StatusInternalServerError)
		return
	}

	res, err := json.Marshal(signal)
	if err != nil {
		http.Error(w, "Couldn't marshal signal", http.StatusInternalServerError)
		return
	}

	w.Write(res)
}
