package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"ophrys/pkg/engine"

	"github.com/gorilla/mux"
)

type OphrysEngineHandler struct {
	e *engine.Engine
	f func(e *engine.Engine, w http.ResponseWriter, r *http.Request) error
}

func (oeh *OphrysEngineHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := oeh.f(oeh.e, w, r)
	if err != nil {
		log.Printf(err.Error())
	}
}

type HttpAPI struct {
	port int
	id   string
}

func (api *HttpAPI) Id() string {
	return api.id
}

func NewHttpAPI(port int) *HttpAPI {
	return &HttpAPI{port: port}
}

func (api *HttpAPI) Engage(e *engine.Engine) error {
	r := mux.NewRouter()
	r.Handle("/stream/subscribe", &OphrysEngineHandler{e: e, f: subscribeStream}).Methods(http.MethodPost)
	r.Handle("/stream/unsubscribe", &OphrysEngineHandler{e: e, f: unsubscribeStream}).Methods(http.MethodPost)
	r.Handle("/subscriptions", &OphrysEngineHandler{e: e, f: subscriptionsList}).Methods(http.MethodGet)
	r.Handle("/workers", &OphrysEngineHandler{e: e, f: workersList}).Methods(http.MethodGet)

	return http.ListenAndServe(fmt.Sprintf(":%d", api.port), r)
}

func subscribeStream(e *engine.Engine, w http.ResponseWriter, r *http.Request) error {
	var p map[string]interface{}
	json.NewDecoder(r.Body).Decode(&p)

	var stream string = p["stream"].(string)
	var providerId string = p["providerId"].(string)

	subscription := <-(*e.GetProvider(providerId)).Subscribe(stream)

	subscriptionJSON, err := json.Marshal(subscription)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(subscriptionJSON)

	if err != nil {
		return err
	}

	return nil
}

func unsubscribeStream(e *engine.Engine, w http.ResponseWriter, r *http.Request) error {
	var p map[string]interface{}
	json.NewDecoder(r.Body).Decode(&p)

	var stream string = p["stream"].(string)
	var providerId string = p["providerId"].(string)

	unsubscription := <-(*e.GetProvider(providerId)).Unsubscribe(stream)

	unsubscriptionJSON, err := json.Marshal(unsubscription)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(unsubscriptionJSON)

	if err != nil {
		return err
	}

	return nil
}

func subscriptionsList(e *engine.Engine, w http.ResponseWriter, r *http.Request) error {
	var p map[string]interface{}
	json.NewDecoder(r.Body).Decode(&p)

	var providerId string = p["providerId"].(string)

	subscriptionList := <-(*e.GetProvider(providerId)).SubscriptionsList()

	subscriptionListJSON, err := json.Marshal(subscriptionList)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(subscriptionListJSON)

	if err != nil {
		return err
	}

	return nil
}
func workersList(e *engine.Engine, w http.ResponseWriter, r *http.Request) error {

	workersJSON, err := json.Marshal(e.Workers())
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(workersJSON)

	if err != nil {
		return err
	}

	return nil
}
