// Package api provides the actual ecowatter HTTP API
package api

import (
	"encoding/json"
	"fmt"
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
	res, err := json.Marshal(a.Ecowatt.Signals)
	if err != nil {
		fmt.Fprintf(w, "error")
	}

	w.Write(res)
}

func (a *API) ListDaySignal(w http.ResponseWriter, r *http.Request) {
	dayStr := flow.Param(r.Context(), "day")
	if dayStr == "" {
		fmt.Fprintf(w, "error")
	}

	day, err := strconv.ParseInt(dayStr, 10, 0)

	signal, err := a.Ecowatt.SignalForDay(int(day))

	res, err := json.Marshal(signal)
	if err != nil {
		fmt.Fprintf(w, "error")
	}

	w.Write(res)
}
