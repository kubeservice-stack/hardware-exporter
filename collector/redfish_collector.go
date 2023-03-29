package collector

import (
	"bytes"
	"errors"
	"fmt"
	"hardware_exporter/lib/gofish"
	gofishcommon "hardware_exporter/lib/gofish/common"
	redfish "hardware_exporter/lib/gofish/redfish"
	hosts "hardware_exporter/lib/hosts"
	"net"
	_ "net/http/pprof"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

// Metric name parts.
const (
	// Exporter namespace.
	namespace = "hardware"
	// Subsystem(s).
	exporter = "exporter"
	// Math constant for picoseconds to seconds.
	picoSeconds = 1e12
)

// Metric descriptors.
var (
	totalScrapeDurationDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, exporter, "collector_duration_seconds"),
		"Collector time duration.",
		nil, nil,
	)
)

// RedfishCollector collects redfish metrics. It implements prometheus.Collector.
type RedfishCollector struct {
	redfishClient *gofish.APIClient
	collectors    map[string]prometheus.Collector
	redfishUp     prometheus.Gauge
}

// NewRedfishCollector return RedfishCollector
func NewRedfishCollector(hostsConfigPath string, logger log.Logger) *RedfishCollector {
	var collectors map[string]prometheus.Collector
	serversInfo := hosts.ReadConfigIni(hostsConfigPath, logger)
	redfishClient, err := generateConfigIniRedfishClient(serversInfo, logger)
	if err != nil {
		level.Error(logger).Log("err", err)
		os.Exit(1)
	}
	level.Info(logger).Log("msg", "Generate redfish client successfully!")
	if err != nil {
		level.Error(logger).Log("err", "error creating redfish client")
	} else {
		chassisCollector := NewChassisCollector(namespace, redfishClient, logger)
		systemCollector := NewSystemCollector(namespace, redfishClient, logger)
		collectors = map[string]prometheus.Collector{"chassis": chassisCollector, "system": systemCollector}
	}

	return &RedfishCollector{
		redfishClient: redfishClient,
		collectors:    collectors,
		redfishUp: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "",
				Name:      "up",
				Help:      "redfish up",
			},
		),
	}
}

// Describe implements prometheus.Collector.
func (r *RedfishCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, collector := range r.collectors {
		collector.Describe(ch)
	}

}

// Collect implements prometheus.Collector.
func (r *RedfishCollector) Collect(ch chan<- prometheus.Metric) {

	scrapeTime := time.Now()
	if r.redfishClient != nil {
		defer r.redfishClient.Logout()
		r.redfishUp.Set(1)
		wg := &sync.WaitGroup{}
		wg.Add(len(r.collectors))

		defer wg.Wait()
		for _, collector := range r.collectors {
			go func(collector prometheus.Collector) {
				defer wg.Done()
				collector.Collect(ch)
			}(collector)
		}
	} else {
		r.redfishUp.Set(0)
	}

	ch <- r.redfishUp
	ch <- prometheus.MustNewConstMetric(totalScrapeDurationDesc, prometheus.GaugeValue, time.Since(scrapeTime).Seconds())
}

func generateConfigIniRedfishClient(serverInfo *hosts.ServerMap, logger log.Logger) (*gofish.APIClient, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return nil, errors.New("hostname error I can't generate redfish client, Please check your hosts file!")
	}
	for serverHostName, bmcInfo := range *serverInfo {
		if strings.EqualFold(hostname, serverHostName) {
			bmcInfoSplit := strings.Split(bmcInfo, "@hardware@")
			url, err := generateRedfishURL(bmcInfoSplit[0])
			if err != nil {
				return nil, errors.New("I can't generate redfish client, Please check your hosts file!")
			}
			redfishconfig := gofish.ClientConfig{
				Endpoint:  url,
				Username:  bmcInfoSplit[1],
				Password:  bmcInfoSplit[2],
				Insecure:  true,
				BasicAuth: true,
			}
			redfishClient, err := gofish.Connect(redfishconfig)
			if err != nil {
				return nil, err
			}
			return redfishClient, nil
		}
	}
	return nil, errors.New("I can't generate redfish client, Please check your hosts file!")
}

func generateRedfishURL(s string) (string, error) {
	ip := net.ParseIP(s)
	if ip == nil {
		return "", errors.New(fmt.Sprintf("the ip address %v is invalid!", s))
	}
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '.':
			// ipv4
			redfishUrl := fmt.Sprintf("https://%s", s)
			return redfishUrl, nil
		case ':':
			// ipv6
			redfishUrl := fmt.Sprintf("https://[%s]", s)
			return redfishUrl, nil
		}
	}
	return "", errors.New(fmt.Sprintf("the ip address %v is invalid!", s))
}
func parseCommonStatusHealth(status gofishcommon.Health) (float64, bool) {
	if bytes.Equal([]byte(status), []byte("OK")) {
		return float64(1), true
	} else if bytes.Equal([]byte(status), []byte("Warning")) {
		return float64(2), true
	} else if bytes.Equal([]byte(status), []byte("Critical")) {
		return float64(3), true
	}
	return float64(0), false
}

func parseCommonStatusState(status gofishcommon.State) (float64, bool) {

	if bytes.Equal([]byte(status), []byte("")) {
		return float64(0), false
	} else if bytes.Equal([]byte(status), []byte("Enabled")) {
		return float64(1), true
	} else if bytes.Equal([]byte(status), []byte("Disabled")) {
		return float64(2), true
	} else if bytes.Equal([]byte(status), []byte("StandbyOffinline")) {
		return float64(3), true
	} else if bytes.Equal([]byte(status), []byte("StandbySpare")) {
		return float64(4), true
	} else if bytes.Equal([]byte(status), []byte("InTest")) {
		return float64(5), true
	} else if bytes.Equal([]byte(status), []byte("Starting")) {
		return float64(6), true
	} else if bytes.Equal([]byte(status), []byte("Absent")) {
		return float64(7), true
	} else if bytes.Equal([]byte(status), []byte("UnavailableOffline")) {
		return float64(8), true
	} else if bytes.Equal([]byte(status), []byte("Deferring")) {
		return float64(9), true
	} else if bytes.Equal([]byte(status), []byte("Quiesced")) {
		return float64(10), true
	} else if bytes.Equal([]byte(status), []byte("Updating")) {
		return float64(11), true
	}
	return float64(0), false
}

func parseCommonPowerState(status redfish.PowerState) (float64, bool) {
	if bytes.Equal([]byte(status), []byte("On")) {
		return float64(1), true
	} else if bytes.Equal([]byte(status), []byte("Off")) {
		return float64(2), true
	} else if bytes.Equal([]byte(status), []byte("PoweringOn")) {
		return float64(3), true
	} else if bytes.Equal([]byte(status), []byte("PoweringOff")) {
		return float64(4), true
	}
	return float64(0), false
}

func parseLinkStatus(status redfish.LinkStatus) (float64, bool) {
	if bytes.Equal([]byte(status), []byte("LinkUp")) {
		return float64(1), true
	} else if bytes.Equal([]byte(status), []byte("NoLink")) {
		return float64(2), true
	} else if bytes.Equal([]byte(status), []byte("LinkDown")) {
		return float64(3), true
	}
	return float64(0), false
}

func parseZteLinkStatus(status redfish.LinkStatus) (float64, bool) {
	if bytes.Equal([]byte(status), []byte("Up")) {
		return float64(1), true
	} else if bytes.Equal([]byte(status), []byte("No")) {
		return float64(2), true
	} else if bytes.Equal([]byte(status), []byte("Down")) {
		return float64(3), true
	}
	return float64(0), false
}
func boolToFloat64(data bool) float64 {

	if data {
		return float64(1)
	}
	return float64(0)

}
