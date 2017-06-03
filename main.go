package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/babbarshaer/usage-api/usage"
)

type Router struct {
	processor usage.UsageProcessor
}

func (router Router) pingHandler(rw http.ResponseWriter, r *http.Request) {
	rw.Write([]byte(`{"response": "pong!!"}`))
}

func (router Router) getUsageLimitsHandler(rw http.ResponseWriter, r *http.Request) {

	fmt.Println("Received a request to fetch the usage for the customer")
	limits, err := router.processor.GetLimitsForUser(1)

	if err != nil {

		fmt.Printf("Error while fetching the limits for the user: %s", err.Error())
		http.Error(rw, `{"error": "Internal Server Error"}`, 500)
		return
	}

	byt, _ := json.Marshal(limits)
	rw.Write(byt)
}

func main() {

	fmt.Println("Starting with the TLS server")

	// Stage1: Setup the configuration
	// parameters to be used by the processor while setting up.
	config := usage.Config{
		DBLocation: "./usage/resource/test_data_2017_02_24.db?parseTime=True",
	}

	// Stage2: Set up the processor which will be used
	// by the router when invoking the correct functions.
	processor, err := usage.NewProcessor(config)
	if err != nil {
		panic(err)
	}

	router := Router{processor: processor}

	http.HandleFunc("/ping", router.pingHandler)
	http.HandleFunc("/limits", router.getUsageLimitsHandler)

	err = http.ListenAndServeTLS(":8081", "./cert/cert.pem", "cert/key.pem", nil)
	if err != nil {
		panic(err)
	}
}
