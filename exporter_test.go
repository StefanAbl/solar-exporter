package main

import (
	"io"
	"log"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestExporter(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc(
		"/status.html", func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, "test_data.html")
		},
	)
	testDataServer := http.Server{
		Handler: mux,
		Addr:    "localhost:12345",
	}
	go testDataServer.ListenAndServe()
	time.Sleep(time.Second)

	testListAddress := "localhost:23456"
	testServer := createServer(&testListAddress, "http://localhost:12345/", "", "")
	go testServer.ListenAndServe()

	resp, err := http.Get("http://" + testListAddress + "/metrics")
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()

	if !(resp.StatusCode == http.StatusOK) {
		t.Fatalf("Request to test server failed with status %s", resp.Status)
	}
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	body := string(bodyBytes)

	testContains(t, body, "solar_power_watt 9")
	testContains(t, body, "solar_power_yield_today_kwh 2.3")
	testContains(t, body, "solar_power_yield_total_kwh 19.9")
	testServer.Close()
	testDataServer.Close()

}

func testContains(t *testing.T, sut string, searchString string) {
	contains := strings.Contains(sut, searchString)
	if !contains {
		t.Fatalf("Expected string to contain %s but got: %s", searchString, sut)
	}
}
