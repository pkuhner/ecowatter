// Binary ecowatt
package main

import (
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/alexedwards/flow"

	"github.com/pkuhner/ecowatter/internal/api"
	"github.com/pkuhner/ecowatter/internal/ecowatt"
)

func main() {
	//ecowattAPIBaseURL, err := url.Parse("https://digital.iservices.rte-france.com/open_api/ecowatt/v4/")
	ecowattAPIBaseURL, err := url.Parse("https://digital.iservices.rte-france.com/open_api/ecowatt/v4/sandbox/")
	if err != nil {
		log.Printf("Couldn't parse Ecowatt API base URL: %v", err)
	}

	ecowattBearerTokenEndpoint, err := url.Parse("https://digital.iservices.rte-france.com/token/oauth/")
	if err != nil {
		log.Printf("Couldn't parse RTE Data OAuth endpoint: %v", err)
	}

	ecowattBearerTokenLifetime := 7200
	ecowattAPIRatelimit := 3
	ecowattAuthorizationToken := os.Getenv("ECOWATTER_AUTH_TOKEN")

	ecw := ecowatt.New(ecowattAPIBaseURL, ecowattAPIRatelimit, ecowattBearerTokenEndpoint, ecowattBearerTokenLifetime, ecowattAuthorizationToken)
	go ecw.Start()

	srv := api.New(ecw)
	mux := flow.New()
	mux.Use(api.SetContentTypeJSON)
	mux.HandleFunc("/signals", srv.ListSignals, "GET")
	mux.HandleFunc("/signals/:day", srv.ListDaySignal, "GET")
	srv.Router = mux

	err = http.ListenAndServe(":8080", srv.Router)
	log.Fatal(err)
}
