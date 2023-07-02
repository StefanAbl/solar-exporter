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
	power = prometheus.NewDesc(
		"solar_power_watt", "The current power in watts", nil, nil,
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
		power, prometheus.GaugeValue, extractPower(string(response)),
	)
}

func extractPower(response string) float64 {
	var re = regexp.MustCompile(`var webdata_now_p = "(\d{1,3})"`)
	matches := re.FindStringSubmatch(response)
	fmt.Printf("%s", matches[1])
	power, err := strconv.Atoi(matches[1])
	if err != nil {
		log.Fatal(err)
	}
	return float64(power)
}

func (s Collector) call(path string, method string) ([]byte, error) {
	client := &http.Client{
		Timeout: time.Second,
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
