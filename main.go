package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/babbarshaer/usage-api/usage"
)

type Router struct {
	processor usage.UsageProcessor
}

func (router Router) authenticateUser(r *http.Request) (usage.User, error) {

	username, password, ok := r.BasicAuth()
	if !ok {
		return usage.User{}, fmt.Errorf("Unable to extract authentication information")
	}

	user, err := router.processor.Storage.GetUser(username, password)
	if err != nil {
		return usage.User{}, fmt.Errorf("Unable to locate the user")
	}

	return user, nil
}

func (router Router) pingHandler(rw http.ResponseWriter, r *http.Request) {
	rw.Write([]byte(`{"response": "pong!!"}`))
}

func (router Router) getUsageLimitsHandler(rw http.ResponseWriter, r *http.Request) {

	fmt.Println("Received a request to fetch the usage for the customer")

	user, err := router.authenticateUser(r)
	if err != nil {
		fmt.Println(err)
		rw.Header().Set("WWW-Authenticate", `Basic realm="Usage"`)
		rw.WriteHeader(401)
		rw.Write([]byte("401 Unauthorized\n"))
		return
	}

	limits, err := router.processor.GetLimitsForUser(user.UserId)

	if err != nil {

		fmt.Printf("Error while fetching the limits for the user: %s", err.Error())
		http.Error(rw, `{"error": "Internal Server Error"}`, 500)
		return
	}

	byt, _ := json.Marshal(limits)
	rw.Write(byt)
}

func (router Router) getDataHandler(rw http.ResponseWriter, r *http.Request) {

	fmt.Println("Received a request to fetch data for the user")

	user, err := router.authenticateUser(r)
	if err != nil {
		rw.Header().Set("WWW-Authenticate", `Basic realm="Usage"`)
		rw.WriteHeader(401)
		rw.Write([]byte("401 Unauthorized\n"))
		return
	}

	values := r.URL.Query()

	if len(values["resolution"]) == 0 || len(values["count"]) == 0 || len(values["start"]) == 0 {
		rw.WriteHeader(400)
		rw.Write([]byte(`{"error": {"code": 400, "reason": "Missing mandatory query params"}}`))
		return
	}

	badRequest := false

	if strings.TrimSpace(values["resolution"][0]) != "M" && strings.TrimSpace(values["resolution"][0]) != "D" {
		badRequest = true
	}

	if _, err := time.Parse("2006-01-02", strings.TrimSpace(values["start"][0])); err != nil {
		fmt.Println("Failed start")
		badRequest = true
	}

	if val, err := strconv.Atoi(strings.TrimSpace(values["count"][0])); err != nil || val <= 0 {
		fmt.Println("Failed count")
		badRequest = true
	}

	if badRequest {
		rw.WriteHeader(400)
		rw.Write([]byte(`{"error": {"code": 400, "reason": "Bad Request"}}`))
		return
	}

	count, _ := strconv.Atoi(strings.TrimSpace(values["count"][0]))
	resolution := strings.TrimSpace(values["resolution"][0])
	start := strings.TrimSpace(values["start"][0])

	payload, err := router.processor.GetDataForUser(user.UserId, count, resolution, start)
	if err != nil {

		fmt.Println(err)
		rw.WriteHeader(500)
		rw.Write([]byte(`{"error": {"code": 500, "reason": {"Internal Server Error"}}`))
		return
	}

	response := struct {
		Data [][]interface{} `json:"data"`
	}{
		Data: payload,
	}

	byt, _ := json.Marshal(response)
	rw.Write(byt)
}

func main() {

	fmt.Println("Starting with the TLS server")

	// Stage1: Setup the configuration
	// parameters to be used by the processor.
	config := usage.Config{
		DBLocation: "./usage/resource/usage_prod.db",
	}

	// Stage2: Set up the processor which will be used
	// by the router which invokes the handler functions for the endpoints.
	processor, err := usage.NewProcessor(config)
	if err != nil {
		panic(err)
	}

	router := Router{processor: processor}

	http.HandleFunc("/ping", router.pingHandler)
	http.HandleFunc("/limits", router.getUsageLimitsHandler)
	http.HandleFunc("/data", router.getDataHandler)

	// Stage3: Bootup the TLS Server.
	err = http.ListenAndServeTLS(":8081", "./cert/cert.pem", "cert/key.pem", nil)
	if err != nil {
		panic(err)
	}
}
