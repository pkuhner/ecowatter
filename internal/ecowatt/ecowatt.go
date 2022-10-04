package ecowatt

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"sort"
	"sync"
	"time"
)

type Ecowatt struct {
	Client              *http.Client
	BaseURL             *url.URL
	BearerToken         BearerToken
	BearerTokenEndpoint *url.URL
	BearerTokenLifetime int
	Ratelimit           int
	AuthorizationToken  string
	LastcallTime        *time.Time
	Signals             *Signals
	mx                  sync.RWMutex
}

type BearerToken struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

type SignalsResponse struct {
	Signals Signals `json:"signals"`
}

type Signals []*Signal

type Signal struct {
	GenerationFichier time.Time `json:"GenerationFichier"`
	Jour              time.Time `json:"jour"`
	DValue            int       `json:"dvalue"`
	Message           string    `json:"message"`
	Values            []*Value  `json:"values"`
}

type Value struct {
	Pas   int `json:"pas"`
	Value int `json:"hvalue"`
}

func New(baseURL *url.URL, ratelimit int, bearerTokenEndpoint *url.URL, bearerTokenLifetime int, authorizationToken string) *Ecowatt {
	return &Ecowatt{
		Client:              &http.Client{},
		BaseURL:             baseURL,
		Ratelimit:           ratelimit,
		AuthorizationToken:  authorizationToken,
		BearerTokenEndpoint: bearerTokenEndpoint,
		BearerTokenLifetime: bearerTokenLifetime,
	}
}

func (ecw *Ecowatt) Start() {
	bearerTokenAge := time.Time{}

	for {
		log.Print("Updating signals...")
		if int(time.Since(bearerTokenAge)/time.Second) > ecw.BearerTokenLifetime {
			log.Print("Retrieving a new bearer token...")
			var tokenErr error
			tokenErr = ecw.getBearerToken()
			bearerTokenAge = time.Now()
			if tokenErr != nil {
				log.Printf("Couldn't retrieve a bearer token: %v", tokenErr)
				time.Sleep(time.Duration(ecw.Ratelimit+1) * time.Second)
				continue
			}
			log.Print("Done.")
		} else {
			log.Print("\tBearer token is still valid...")
		}

		signalsErr := ecw.updateSignals()
		if signalsErr != nil {
			log.Printf("\tCouldn't update signals: %v", signalsErr)
		}

		log.Printf("Done. Sleeping %d seconds...", ecw.Ratelimit+1)
		time.Sleep(time.Duration(ecw.Ratelimit+1) * time.Second)
	}
}

func (ecw *Ecowatt) GetSignals() (*Signals, error) {
	return ecw.getSignals()
}

func (ecw *Ecowatt) GetSignal(day int) (*Signal, error) {
	return ecw.getSignalForDay(day)
}

func (ecw *Ecowatt) getBearerToken() error {
	var result BearerToken

	req, err := http.NewRequest("GET", ecw.BearerTokenEndpoint.String(), nil)
	req.Header.Add("Authorization", "Basic "+ecw.AuthorizationToken)
	resp, err := ecw.Client.Do(req)
	if err != nil {
		return fmt.Errorf("Couldn't GET bearer token endpoint: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("Couldn't read response: %w", err)
	}

	err = json.Unmarshal(body, &result)
	if err != nil {
		return fmt.Errorf("Couldn't unmarshal body: %w", err)
	}

	ecw.BearerToken = result

	return nil
}

func (ecw *Ecowatt) updateSignals() error {
	var result SignalsResponse

	req, err := http.NewRequest("GET", ecw.BaseURL.JoinPath("/signals").String(), nil)
	req.Header.Add("Authorization", ecw.BearerToken.TokenType+" "+ecw.BearerToken.AccessToken)

	log.Printf("\tGET %s...", req.URL)
	resp, err := ecw.Client.Do(req)
	if err != nil {
		return fmt.Errorf("Couldn't GET signals: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Non-OK HTTP status: %d", resp.StatusCode)
	}

	now := time.Now()
	ecw.LastcallTime = &now

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("Couldn't read response: %w", err)
	}

	err = json.Unmarshal(body, &result)
	if err != nil {
		return fmt.Errorf("Couldn't unmarshal body: %w", err)
	}

	signals := result.Signals

	sort.Slice(signals, func(i, j int) bool {
		return signals[i].Jour.Before(signals[j].Jour)
	})

	ecw.mx.Lock()
	ecw.Signals = &signals
	ecw.mx.Unlock()

	return nil
}

func (ecw *Ecowatt) getSignals() (*Signals, error) {
	ecw.mx.RLock()
	defer ecw.mx.RUnlock()

	return ecw.Signals, nil
}

func (ecw *Ecowatt) getSignalForDay(day int) (*Signal, error) {
	ecw.mx.RLock()
	if !(day >= 0 && day <= len(*ecw.Signals)) {
		return nil, fmt.Errorf("Day %d does not exist in signals", day)
	}
	signal := (*ecw.Signals)[day]
	ecw.mx.RUnlock()

	return signal, nil
}
