package collector

import (
	"crypto/tls"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"time"
)

type Collector struct {
	address string
	user    string
	pw      string
}

var (
	regexPower      = `var webdata_now_p = "(\d{1,4})"`
	regexYieldToday = `var webdata_today_e = "([\d.]{1,5})"`
	regexYieldTotal = `var webdata_total_e = "([\d.]{1,5})"`
	power           = prometheus.NewDesc(
		"solar_power_watt", "The current power in watts", nil, nil,
	)
	yieldToday = prometheus.NewDesc(
		"solar_power_yield_today_kwh", "The total yield today in KWh", nil, nil,
	)
	yieldTotal = prometheus.NewDesc(
		"solar_power_yield_total_kwh", "The total yield in KWh", nil, nil,
	)
)

func NewCollector(address string, user string, pw string) *Collector {
	return &Collector{
		address: address,
		user:    user,
		pw:      pw,
	}
}
func (s *Collector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(s, ch)
}
func (s *Collector) Collect(metrics chan<- prometheus.Metric) {
	response, err := s.call("status.html", "GET")
	if err != nil {
		log.Fatal(err)
	}
	metrics <- prometheus.MustNewConstMetric(
		power, prometheus.GaugeValue, extractValue(string(response), regexPower),
	)
	metrics <- prometheus.MustNewConstMetric(
		yieldToday, prometheus.CounterValue, extractValue(string(response), regexYieldToday),
	)
	metrics <- prometheus.MustNewConstMetric(
		yieldTotal, prometheus.CounterValue, extractValue(string(response), regexYieldTotal),
	)
}

func extractValue(response string, regex string) float64 {
	var re = regexp.MustCompile(regex)
	matches := re.FindStringSubmatch(response)
	if len(matches) == 0 {
		log.Fatalf("Could not finde value. Response was: \n %s", response)
	}
	//fmt.Printf("%s", matches[1])
	value, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		log.Fatal(err)
	}
	return value
}

func (s Collector) call(path string, method string) ([]byte, error) {
	client := &http.Client{
		Timeout: time.Second * 10,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	req, err := http.NewRequest(method, s.address+path, nil)
	if err != nil {
		return nil, fmt.Errorf("got error %s", err.Error())
	}
	req.SetBasicAuth(s.user, s.pw)
	response, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("got error %s", err.Error())
	}
	body, err := ioutil.ReadAll(response.Body)
	defer response.Body.Close()
	return body, err
}
