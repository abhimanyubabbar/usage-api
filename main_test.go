package main

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/babbarshaer/usage-api/usage"
)

var processor usage.UsageProcessor
var router Router

var testUsers map[int]usage.User

// Setup the storage layer which will be used to setup the
// dummy information in the system.
func setup() error {

	// Stage1: Setup the test users for which
	// we will be adding the data.
	testUsers = make(map[int]usage.User)
	for _, val := range []int{1, 2} {

		testUsers[val] = usage.User{
			UserId:   val,
			UserName: fmt.Sprintf("username%d", val),
			Password: fmt.Sprintf("password%d", val),
		}
	}

	// Stage2: Create the necessary structs for the
	// invoking the endpoints through handlers.
	var err error

	config := usage.Config{
		DBLocation: "./usage/resource/usage_test.db",
	}
	processor, err = usage.NewProcessor(config)

	if err != nil {
		return err
	}

	// Stage3: Setup the router to which we have attached
	// the handlers.
	router.processor = processor

	// Stage4: Finally create the users in the datbase.
	// But before that truncate the existing table.
	_, err = router.processor.Storage.DB.Exec(`DELETE FROM user`)
	_, err = router.processor.Storage.DB.Exec(`DELETE FROM days`)
	_, err = router.processor.Storage.DB.Exec(`DELETE FROM months`)

	if err != nil {
		return err
	}

	for _, user := range testUsers {

		err = router.processor.Storage.AddNewUser(user.UserId,
			user.UserName,
			user.Password)

		if err != nil {
			return err
		}
	}

	return nil
}

// TestMain is the function that gets called
// by the test environment. It can used to setup
// necessary data for the test
func TestMain(m *testing.M) {

	fmt.Println("Starting with the setting up data for the test")
	err := setup()

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	code := m.Run()
	os.Exit(code)
}

func TestGetLimitsForUnauthorizedRequest(t *testing.T) {

	req, err := http.NewRequest("GET", "/limits", nil)
	if err != nil {
		t.Errorf("Unable to create the request: %s", err.Error())
		return
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(router.getUsageLimitsHandler)

	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusUnauthorized {
		t.Errorf("handler returned code: %d, expected: %d", rr.Code, http.StatusUnauthorized)
	}

}

func TestGetDataForUnauthorizedRequest(t *testing.T) {

	req, err := http.NewRequest("GET", "/data", nil)
	if err != nil {
		t.Errorf("Unable to create the request: %s", err.Error())
		return
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(router.getDataHandler)

	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusUnauthorized {
		t.Errorf("handler returned code: %d, expected unauthorized response: %d", rr.Code, http.StatusUnauthorized)
	}
}

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

func TestGetDataForBadRequest(t *testing.T) {

	req, err := http.NewRequest("GET", "/data?start=2006-07-19", nil)
	if err != nil {
		t.Errorf("Unable to create the request: %s", err.Error())
		return
	}

	req.Header.Add("Authorization",
		fmt.Sprintf("Basic %s", basicAuth(testUsers[1].UserName, testUsers[1].Password)))

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(router.getDataHandler)

	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned code: %d, expected bad response: %d", rr.Code, http.StatusBadRequest)
	}
}

func TestGetLimitsForInvalidUser(t *testing.T) {

	invalidUsername := "invalidUsername"
	invalidPassword := "invalidPassword"

	req, err := http.NewRequest("GET", "/limits", nil)
	if err != nil {
		t.Errorf("Unable to create the request: %s", err.Error())
		return
	}

	req.Header.Add("Authorization",
		fmt.Sprintf("Basic %s", basicAuth(invalidUsername, invalidPassword)))

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(router.getUsageLimitsHandler)

	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusUnauthorized {
		t.Errorf("handler returned code: %d, expected bad response: %d", rr.Code, http.StatusUnauthorized)
	}
}

func TestGetDefaultData(t *testing.T) {

	validUser := testUsers[1]

	req, err := http.NewRequest("GET", "/limits", nil)
	if err != nil {
		t.Errorf("Unable to create the request: %s", err.Error())
		return
	}

	req.Header.Add("Authorization",
		fmt.Sprintf("Basic %s", basicAuth(validUser.UserName, validUser.Password)))

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(router.getUsageLimitsHandler)

	handler.ServeHTTP(rr, req)

	byt, _ := ioutil.ReadAll(rr.Body)
	expected := `{"daily":{"timestamp":{"minimum":"0001-01-01","maximum":"0001-01-01"},"consumption":{"minimum":0,"maximum":0},"temperature":{"minimum":0,"maximum":0}},"monthly":{"timestamp":{"minimum":"0001-01-01","maximum":"0001-01-01"},"consumption":{"minimum":0,"maximum":0},"temperature":{"minimum":0,"maximum":0}}}`

	if expected != string(byt) {
		t.Fatalf("Mismatch between the expected : %s and actual: %s default limits value", expected, string(byt))
	}

	req, err = http.NewRequest("GET", "/data?resolution=M&start=2014-01-03&count=1", nil)
	if err != nil {
		t.Errorf("Unable to create the request: %s", err.Error())
		return
	}

	req.Header.Add("Authorization",
		fmt.Sprintf("Basic %s", basicAuth(validUser.UserName, validUser.Password)))

	rr = httptest.NewRecorder()
	handler = http.HandlerFunc(router.getDataHandler)

	handler.ServeHTTP(rr, req)

	byt, _ = ioutil.ReadAll(rr.Body)
	expected = `{"data":null}`
	if expected != string(byt) {
		t.Fatalf("Mismatch between the expected : %s and actual: %s default limits value", expected, string(byt))
	}
}

func TestGetValidLimitsForAuthorizedUser(t *testing.T) {

	validUser := testUsers[1]

	// DATA FORMAT : (day_id, temperature, consumption, timestamp)
	dailyTestData := [][]interface{}{
		[]interface{}{1, -1, 10, "2014-02-05 12:02:13"},
		[]interface{}{2, -10, 100, "2015-02-05 12:02:13"},
		[]interface{}{3, 20, 89, "2017-02-05 12:02:13"},
	}

	defer func() {
		// NOTE: Each test should clean up after itself.
		processor.Storage.DB.Exec(`DELETE FROM days WHERE user_id = ?`, validUser.UserId)
	}()

	for _, data := range dailyTestData {

		err := processor.Storage.AddDailyLimit(validUser.UserId,
			data[0].(int),
			data[1].(int),
			data[2].(int),
			data[3].(string))

		if err != nil {
			t.Fatalf("Unable to add daily limit for the user: %d, error: %s", validUser.UserId, err.Error())
		}
	}

	req, err := http.NewRequest("GET", "/limits", nil)
	if err != nil {
		t.Fatalf("Unable to create the request: %s", err.Error())
	}

	req.Header.Add("Authorization",
		fmt.Sprintf("Basic %s", basicAuth(validUser.UserName, validUser.Password)))

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(router.getUsageLimitsHandler)

	handler.ServeHTTP(rr, req)
	byt, _ := ioutil.ReadAll(rr.Body)

	expected := `{"daily":{"timestamp":{"minimum":"2014-02-05","maximum":"2017-02-05"},"consumption":{"minimum":10,"maximum":100},"temperature":{"minimum":-10,"maximum":20}},"monthly":{"timestamp":{"minimum":"0001-01-01","maximum":"0001-01-01"},"consumption":{"minimum":0,"maximum":0},"temperature":{"minimum":0,"maximum":0}}}`

	actual := string(byt)

	if expected != actual {
		t.Fatalf("Unable to fetch the correct limits for the user, expected: %s, actual: %s", expected, actual)
	}
}

func TestGetDataForAuthorizedUser(t *testing.T) {

	validUser := testUsers[1]

	// DATA FORMAT : (day_id, temperature, consumption, timestamp)
	dailyTestData := [][]interface{}{
		[]interface{}{1, -1, 10, "2014-02-01 12:02:13"},
		[]interface{}{2, -10, 100, "2014-02-06 12:02:13"},
		[]interface{}{3, 20, 89, "2014-02-07 12:02:13"},
	}

	monthlyTestData := [][]interface{}{
		[]interface{}{1, -1, 10, "2014-02-05 12:02:13"},
		[]interface{}{2, -10, 100, "2014-03-05 12:02:13"},
		[]interface{}{3, 20, 89, "2014-04-05 12:02:13"},
	}

	defer func() {
		// NOTE: Each test should clean up after itself.
		processor.Storage.DB.Exec(`DELETE FROM days WHERE user_id = ?`, validUser.UserId)
		processor.Storage.DB.Exec(`DELETE FROM months WHERE user_id = ?`, validUser.UserId)
	}()

	for _, data := range dailyTestData {

		err := processor.Storage.AddDailyLimit(validUser.UserId,
			data[0].(int),
			data[1].(int),
			data[2].(int),
			data[3].(string))

		if err != nil {
			t.Fatalf("Unable to add daily limit for the user: %d, error: %s", validUser.UserId, err.Error())
		}
	}

	for _, data := range monthlyTestData {

		err := processor.Storage.AddMonthlyLimit(validUser.UserId,
			data[0].(int),
			data[1].(int),
			data[2].(int),
			data[3].(string))

		if err != nil {
			t.Fatalf("Unable to add daily limit for the user: %d, error: %s", validUser.UserId, err.Error())
		}
	}

	req, err := http.NewRequest("GET", "/data?start=2014-02-10&count=4&resolution=M", nil)
	if err != nil {
		t.Errorf("Unable to create the request: %s", err.Error())
		return
	}

	req.Header.Add("Authorization",
		fmt.Sprintf("Basic %s", basicAuth(validUser.UserName, validUser.Password)))

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(router.getDataHandler)

	handler.ServeHTTP(rr, req)
	byt, _ := ioutil.ReadAll(rr.Body)

	expected := `{"data":[["2014-03-05",-10,100],["2014-04-05",20,89]]}`
	actual := string(byt)

	if actual != expected {
		t.Fatalf("Unable to get monthly data for the user expected %s, actual: %s", expected, actual)
	}

	req, err = http.NewRequest("GET", "/data?start=2014-02-02&count=4&resolution=D", nil)
	if err != nil {
		t.Errorf("Unable to create the request: %s", err.Error())
		return
	}

	req.Header.Add("Authorization",
		fmt.Sprintf("Basic %s", basicAuth(validUser.UserName, validUser.Password)))

	rr = httptest.NewRecorder()
	handler = http.HandlerFunc(router.getDataHandler)

	handler.ServeHTTP(rr, req)
	byt, _ = ioutil.ReadAll(rr.Body)

	expected = `{"data":[["2014-02-06",-10,100],["2014-02-07",20,89]]}`
	actual = string(byt)

	if actual != expected {
		t.Fatalf("Unable to get monthly data for the user expected %s, actual: %s", expected, actual)
	}
}
