package collector

import (
	"crypto/tls"
	"errors"
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
	regexPower      = `var webdata_now_p = "(\d{1,4}|)"`
	regexYieldToday = `var webdata_today_e = "([\d.]{1,5}|)"`
	regexYieldTotal = `var webdata_total_e = "([\d.]{1,5}|)"`
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
		log.Println(err)
	}
	if response != nil {
		if p, err := extractValue(string(response), regexPower); err == nil {
			metrics <- prometheus.MustNewConstMetric(
				power, prometheus.GaugeValue, p,
			)
		} else {
			log.Println(err)
		}
		if yDay, err := extractValue(string(response), regexYieldToday); err == nil {
			metrics <- prometheus.MustNewConstMetric(
				yieldToday, prometheus.CounterValue, yDay,
			)
		} else {
			log.Println(err)
		}
		if yTotal, err := extractValue(string(response), regexYieldTotal); err == nil {
			metrics <- prometheus.MustNewConstMetric(
				yieldTotal, prometheus.CounterValue, yTotal,
			)
		} else {
			log.Println(err)
		}
	}
}

func extractValue(response string, regex string) (float64, error) {
	var re = regexp.MustCompile(regex)
	matches := re.FindStringSubmatch(response)
	if len(matches) == 0 {
		return 0.0, errors.New(
			fmt.Sprintf(
				"Could not finde value for regex %s. Response was: \n %s \n", regex, response,
			),
		)
	}
	var value float64
	var err error
	if matches[1] == "" {
		value = 0.0
	} else {
		value, err = strconv.ParseFloat(matches[1], 64)
	}
	if err != nil {
		return 0, err
	}
	return value, nil
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
