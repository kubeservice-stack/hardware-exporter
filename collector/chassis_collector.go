package collector

import (
	"fmt"
	"hardware_exporter/lib/gofish"
	redfish2 "hardware_exporter/lib/gofish/redfish"
	"strings"
	"sync"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

// ChassisSubsystem is the chassis subsystem
var (
	ChassisSubsystem                = "chassis"
	ChassisLabelNames               = []string{"server_assetTag", "server_sn", "mfr", "chassis_id", "resource"}
	ChassisTemperatureLabelNames    = []string{"server_assetTag", "server_sn", "mfr", "chassis_id", "sensor", "sensor_id", "resource"}
	ChassisFanLabelNames            = []string{"server_assetTag", "server_sn", "fan_id", "fan_slot", "resource"}
	ChassisPowerSupplyLabelNames    = []string{"server_assetTag", "server_sn", "power_id", "power_name", "mfr", "model", "power_sn", "resource"}
	ChassisNetworkAdapterLabelNames = []string{"server_assetTag", "server_sn", "netAdapter_id", "netAdapter_model", "netAdapter_sn", "netAdapter_slot", "resource"}
	ChassisNetworkPortLabelNames    = []string{"server_assetTag", "server_sn", "netAdapter_id", "Port_id", "port_mac", "port_speed", "port_type", "resource"}
	chassisMetrics                  = map[string]chassisMetric{
		"chassis_health": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, ChassisSubsystem, "health"),
				"health of chassis, 1(OK),2(Warning),3(Critical)",
				ChassisLabelNames,
				nil,
			),
		},
		"chassis_state": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, ChassisSubsystem, "state"),
				"state of chassis,1(Enabled),2(Disabled),3(StandbyOffinline),4(StandbySpare),5(InTest),6(Starting),7(Absent),8(UnavailableOffline),9(Deferring),10(Quiesced),11(Updating)",
				ChassisLabelNames,
				nil,
			),
		},
		"chassis_temperature_sensor_health": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, ChassisSubsystem, "temperature_sensor_health"),
				"temperature sensor on this chassis component,1(OK),2(Warning),3(Critical)",
				ChassisTemperatureLabelNames,
				nil,
			),
		},
		"chassis_temperature_sensor_state": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, ChassisSubsystem, "temperature_sensor_state"),
				"status state of temperature on this chassis component,1(Enabled),2(Disabled),3(StandbyOffinline),4(StandbySpare),5(InTest),6(Starting),7(Absent),8(UnavailableOffline),9(Deferring),10(Quiesced),11(Updating)",
				ChassisTemperatureLabelNames,
				nil,
			),
		},
		"chassis_temperature_celsius": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, ChassisSubsystem, "temperature_celsius"),
				"celsius of temperature on this chassis component",
				ChassisTemperatureLabelNames,
				nil,
			),
		},
		"chassis_fan_health": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, ChassisSubsystem, "fan_health"),
				"fan health on this chassis component,1(OK),2(Warning),3(Critical)",
				ChassisFanLabelNames,
				nil,
			),
		},
		"chassis_fan_state": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, ChassisSubsystem, "fan_state"),
				"fan state on this chassis component,1(Enabled),2(Disabled),3(StandbyOffinline),4(StandbySpare),5(InTest),6(Starting),7(Absent),8(UnavailableOffline),9(Deferring),10(Quiesced),11(Updating)",
				ChassisFanLabelNames,
				nil,
			),
		},
		"chassis_fan_rpm": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, ChassisSubsystem, "fan_rpm_percentage"),
				"fan rpm percentage on this chassis component",
				ChassisFanLabelNames,
				nil,
			),
		},
		"chassis_power_powersupply_state": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, ChassisSubsystem, "power_powersupply_state"),
				"powersupply state of chassis component,1(Enabled),2(Disabled),3(StandbyOffinline),4(StandbySpare),5(InTest),6(Starting),7(Absent),8(UnavailableOffline),9(Deferring),10(Quiesced),11(Updating)",
				ChassisPowerSupplyLabelNames,
				nil,
			),
		},
		"chassis_power_powersupply_health_status": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, ChassisSubsystem, "power_powersupply_health_status"),
				"powersupply health of chassis component,1(OK),2(Warning),3(Critical)",
				ChassisPowerSupplyLabelNames,
				nil,
			),
		},
		"chassis_power_line_input_voltage": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, ChassisSubsystem, "power_line_input_voltage"),
				"line_input_voltage of powersupply on this chassis",
				ChassisPowerSupplyLabelNames,
				nil,
			),
		},
		"chassis_power_powersupply_power_capacity_watts": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, ChassisSubsystem, "power_powersupply_power_capacity_watts"),
				"power_capacity_watts of powersupply on this chassis",
				ChassisPowerSupplyLabelNames,
				nil,
			),
		},
		"chassis_networkAdapter_state": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, ChassisSubsystem, "chassis_networkAdapter_state"),
				"chass networkAdapter state,1(Enabled),2(Disabled),3(StandbyOffinline),4(StandbySpare),5(InTest),6(Starting),7(Absent),8(UnavailableOffline),9(Deferring),10(Quiesced),11(Updating)",
				ChassisNetworkAdapterLabelNames,
				nil,
			),
		},
		"system_networkAdapterHealth_state": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, ChassisSubsystem, "system_networkAdapterHealth_state"),
				"networkAdapter health of chassis component,1(OK),2(Warning),3(Critical)",
				ChassisNetworkAdapterLabelNames,
				nil,
			),
		},
		"network_port_linkStatus": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, ChassisSubsystem, "network_port_linkStatus"),
				"system network port link status，1(LinkUp),2(NoLink),3(LinkDown)",
				ChassisNetworkPortLabelNames,
				nil,
			),
		},
	}
)

// ChassisCollector implements the prometheus.Collector.
type ChassisCollector struct {
	redfishClient         *gofish.APIClient
	metrics               map[string]chassisMetric
	collectorScrapeStatus *prometheus.GaugeVec
	Log                   log.Logger
}

type chassisMetric struct {
	desc *prometheus.Desc
}

// NewChassisCollector returns a collector that collecting chassis statistics
func NewChassisCollector(namespace string, redfishClient *gofish.APIClient, logger log.Logger) *ChassisCollector {
	// get service from gofish client

	return &ChassisCollector{
		redfishClient: redfishClient,
		metrics:       chassisMetrics,
		Log:           logger,
		collectorScrapeStatus: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "collector_scrape_status",
				Help:      "collector_scrape_status",
			},
			[]string{"collector"},
		),
	}
}

// Describe implemented prometheus.Collector
func (c *ChassisCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, metric := range c.metrics {
		ch <- metric.desc
	}
	c.collectorScrapeStatus.Describe(ch)
}

// Collect implemented prometheus.Collector
func (c *ChassisCollector) Collect(ch chan<- prometheus.Metric) {
	service := c.redfishClient.Service
	var serialNumber string
	// get a list of chassis from service
	if systems, err := service.Systems(); err != nil {
		level.Error(c.Log).Log("operation", "service.Systems()", "err", err)
	} else {
		for _, system := range systems {
			serialNumber = system.SerialNumber
		}
	}
	if chassises, err := service.Chassis(); err != nil {
		level.Error(c.Log).Log("operation", "service.Chassis()", "err", err)
	} else {
		// process the chassises
		for _, chassis := range chassises {
			level.Info(c.Log).Log("Chassis", chassis.ID, "msg", "Chassis collector scrape started")
			assetTag := chassis.AssetTag
			systemManufacturer := "Unknown"
			if chassis.Manufacturer != "" {
				tmpStr := strings.Split(chassis.Manufacturer, " ")
				systemManufacturer = tmpStr[0]
			}

			chassisID := chassis.ID
			chassisStatus := chassis.Status
			chassisStatusState := chassisStatus.State
			chassisStatusHealth := chassisStatus.Health
			ChassisLabelValues := []string{assetTag, serialNumber, systemManufacturer, chassisID, "chassis"}

			if chassisStatusHealthValue, ok := parseCommonStatusHealth(chassisStatusHealth); ok {
				ch <- prometheus.MustNewConstMetric(c.metrics["chassis_health"].desc, prometheus.GaugeValue, chassisStatusHealthValue, ChassisLabelValues...)
			}
			if chassisStatusStateValue, ok := parseCommonStatusState(chassisStatusState); ok {
				ch <- prometheus.MustNewConstMetric(c.metrics["chassis_state"].desc, prometheus.GaugeValue, chassisStatusStateValue, ChassisLabelValues...)
			}

			chassisThermal, err := chassis.Thermal()
			if err != nil {
				level.Error(c.Log).Log("operation", "chassis.Thermal()", "err", "error getting thermal data from chassis", "err_info", err)
			} else if chassisThermal == nil {
				level.Error(c.Log).Log("operation", "chassis.Thermal()", "err", "no thermal data found")
			} else {
				// process temperature
				chassisTemperatures := chassisThermal.Temperatures
				wg1 := &sync.WaitGroup{}
				wg1.Add(len(chassisTemperatures))

				for _, chassisTemperature := range chassisTemperatures {
					go parseServerTemperatures(ch, assetTag, serialNumber, systemManufacturer, chassisID, chassisTemperature, wg1)
				}
				wg1.Wait()
				// process fans
				chassisFans := chassisThermal.Fans
				wg2 := &sync.WaitGroup{}
				wg2.Add(len(chassisFans))
				for _, chassisFan := range chassisFans {
					go parseChassisFan(ch, assetTag, serialNumber, systemManufacturer, chassisID, chassisFan, wg2)
				}
				wg2.Wait()
			}
			chassisPowerInfo, err := chassis.Power()
			if err != nil {
				level.Error(c.Log).Log("msg", "chassis.Power()", "error", err)
			} else if chassisPowerInfo == nil {
				level.Info(c.Log).Log("msg", "chassis.Power()", "err", "no power data found")
			} else {
				// powerSupply
				chassisPowerInfoPowerSupplies := chassisPowerInfo.PowerSupplies
				wg3 := &sync.WaitGroup{}
				wg3.Add(len(chassisPowerInfoPowerSupplies))
				for _, chassisPowerInfoPowerSupply := range chassisPowerInfoPowerSupplies {
					go parseChassisPowerInfoPowerSupply(ch, assetTag, serialNumber, chassisID, chassisPowerInfoPowerSupply, wg3)
				}
				wg3.Wait()
			}

			netWorkAdapters, err := chassis.NetworkAdapters()
			if err != nil {
				level.Error(c.Log).Log("msg", "chassis.NetworkAdapters()", "error", err)
			} else if chassisPowerInfo == nil {
				level.Info(c.Log).Log("msg", "chassis.NetworkAdapters()", "err", "no NetworkAdapters data found")
			} else {
				// Process netWorAdapter
				wg4 := &sync.WaitGroup{}
				wg4.Add(len(netWorkAdapters))
				for _, netWorkAdapter := range netWorkAdapters {
					go parseNetWorkAdapter(ch, assetTag, serialNumber, netWorkAdapter, wg4)
					ports, err := netWorkAdapter.NetworkPorts()
					if err != nil {
						level.Error(c.Log).Log("msg", "netWorkAdapter.NetworkPorts()", "error", err)
					} else if chassisPowerInfo == nil {
						level.Info(c.Log).Log("msg", "netWorkAdapter.NetworkPorts()", "err", "no NetworkPorts data found")
					} else {
						wg5 := &sync.WaitGroup{}
						wg5.Add(len(ports))
						for _, port := range ports {
							go parseNetworkPorts(ch, assetTag, serialNumber, netWorkAdapter.ID, port, wg5)
						}
						wg5.Wait()
					}
				}
				wg4.Wait()
			}
			level.Info(c.Log).Log("Chassis", chassisID, "msg", "Chassis collector scrape completed")
		}
	}

	c.collectorScrapeStatus.WithLabelValues("chassis").Set(float64(1))
}

func parseServerTemperatures(ch chan<- prometheus.Metric, assetTag, serialNumber, systemManufacturer, chassisID string, chassisTemperature redfish2.Temperature, wg *sync.WaitGroup) {
	defer wg.Done()
	chassisTemperatureSensorName := chassisTemperature.Name
	chassisTemperatureSensorID := fmt.Sprintf("%v", chassisTemperature.SensorNumber)
	chassisTemperatureStatusHealth := chassisTemperature.Status.Health
	chassisTemperatureStatusState := chassisTemperature.Status.State

	chassisTemperatureLabelvalues := []string{assetTag, serialNumber, systemManufacturer, chassisID, chassisTemperatureSensorName, chassisTemperatureSensorID, "Temperature"}

	if chassisFanStausHealthValue, ok := parseCommonStatusHealth(chassisTemperatureStatusHealth); ok {
		ch <- prometheus.MustNewConstMetric(chassisMetrics["chassis_temperature_sensor_health"].desc, prometheus.GaugeValue, chassisFanStausHealthValue, chassisTemperatureLabelvalues...)
	}
	if chassisTemperatureStatusStateValue, ok := parseCommonStatusState(chassisTemperatureStatusState); ok {
		ch <- prometheus.MustNewConstMetric(chassisMetrics["chassis_temperature_sensor_state"].desc, prometheus.GaugeValue, chassisTemperatureStatusStateValue, chassisTemperatureLabelvalues...)
	}

	chassisTemperatureReadingCelsius := chassisTemperature.ReadingCelsius
	ch <- prometheus.MustNewConstMetric(chassisMetrics["chassis_temperature_celsius"].desc, prometheus.GaugeValue, float64(chassisTemperatureReadingCelsius), chassisTemperatureLabelvalues...)
}

func parseChassisFan(ch chan<- prometheus.Metric, assetTag, serialNumber string, systemManufacturer, chassisID string, chassisFan redfish2.Fan, wg *sync.WaitGroup) {
	defer wg.Done()
	chassisFanID := chassisFan.MemberID
	chassisFanSlotNum := chassisFan.MemberID
	chassisFanStaus := chassisFan.Status
	chassisFanStausHealth := chassisFanStaus.Health
	chassisFanStausState := chassisFanStaus.State
	chassisFanRPM := chassisFan.Reading

	chassisFanLabelvalues := []string{assetTag, serialNumber, chassisFanID, chassisFanSlotNum, "Fan"}

	if chassisFanStausHealthValue, ok := parseCommonStatusHealth(chassisFanStausHealth); ok {
		ch <- prometheus.MustNewConstMetric(chassisMetrics["chassis_fan_health"].desc, prometheus.GaugeValue, chassisFanStausHealthValue, chassisFanLabelvalues...)
	}
	if chassisFanStausStateValue, ok := parseCommonStatusState(chassisFanStausState); ok {
		ch <- prometheus.MustNewConstMetric(chassisMetrics["chassis_fan_state"].desc, prometheus.GaugeValue, chassisFanStausStateValue, chassisFanLabelvalues...)
	}
	ch <- prometheus.MustNewConstMetric(chassisMetrics["chassis_fan_rpm"].desc, prometheus.GaugeValue, float64(chassisFanRPM), chassisFanLabelvalues...)
}

func parseChassisPowerInfoPowerSupply(ch chan<- prometheus.Metric, assetTag, serialNumber string, chassisID string, chassisPowerInfoPowerSupply redfish2.PowerSupply, wg *sync.WaitGroup) {

	defer wg.Done()
	chassisPowerInfoPowerSupplyName := chassisPowerInfoPowerSupply.Name
	chassisPowerModle := chassisPowerInfoPowerSupply.Model
	chassisPowerManufacturer := chassisPowerInfoPowerSupply.Manufacturer
	chassisPowerSerialNum := chassisPowerInfoPowerSupply.SerialNumber
	chassisPowerInfoPowerSupplyID := chassisPowerInfoPowerSupply.MemberID
	if chassisPowerInfoPowerSupplyID == "" {
		chassisPowerInfoPowerSupplyID = chassisPowerInfoPowerSupply.SerialNumber
	}

	chassisPowerInfoPowerSupplyPowerCapacityWatts := chassisPowerInfoPowerSupply.PowerCapacityWatts
	chassisPowerInfoPowerSupplyLineInputVoltage := chassisPowerInfoPowerSupply.LineInputVoltage
	chassisPowerInfoPowerSupplyState := chassisPowerInfoPowerSupply.Status.State
	chassisPowerInfoPowerSupplyHealthStatus := chassisPowerInfoPowerSupply.Status.Health

	chassisPowerSupplyLabelvalues := []string{assetTag, serialNumber, chassisPowerInfoPowerSupplyID, chassisPowerInfoPowerSupplyName, chassisPowerManufacturer, chassisPowerModle, chassisPowerSerialNum, "Power"}

	if chassisPowerInfoPowerSupplyStateValue, ok := parseCommonStatusState(chassisPowerInfoPowerSupplyState); ok {
		ch <- prometheus.MustNewConstMetric(chassisMetrics["chassis_power_powersupply_state"].desc, prometheus.GaugeValue, chassisPowerInfoPowerSupplyStateValue, chassisPowerSupplyLabelvalues...)
	}
	if chassisPowerInfoPowerSupplyHealthStatusValue, ok := parseCommonStatusHealth(chassisPowerInfoPowerSupplyHealthStatus); ok {
		ch <- prometheus.MustNewConstMetric(chassisMetrics["chassis_power_powersupply_health_status"].desc, prometheus.GaugeValue, chassisPowerInfoPowerSupplyHealthStatusValue, chassisPowerSupplyLabelvalues...)
	}
	ch <- prometheus.MustNewConstMetric(chassisMetrics["chassis_power_powersupply_power_capacity_watts"].desc, prometheus.GaugeValue, float64(chassisPowerInfoPowerSupplyPowerCapacityWatts), chassisPowerSupplyLabelvalues...)
	ch <- prometheus.MustNewConstMetric(chassisMetrics["chassis_power_line_input_voltage"].desc, prometheus.GaugeValue, float64(chassisPowerInfoPowerSupplyLineInputVoltage), chassisPowerSupplyLabelvalues...)

}

func parseNetWorkAdapter(ch chan<- prometheus.Metric, assetTag string, serialNumber string, netWorkAdapter *redfish2.NetworkAdapter, wg *sync.WaitGroup) {
	defer wg.Done()

	networkAdapterID := netWorkAdapter.ID
	networkAdapterModel := netWorkAdapter.Model
	netWorkAdapterState := netWorkAdapter.Status.State
	netWorkAdapterHealthState := netWorkAdapter.Status.Health

	var networkAdapterSlotNum string
	if netWorkAdapter.Oem.Public.SlotNumber != nil {
		networkAdapterSlotNum = fmt.Sprintf("%v", netWorkAdapter.Oem.Public.SlotNumber)
	} else if netWorkAdapter.Oem.Huawei.AssociatedResource != "" {
		networkAdapterSlotNum = fmt.Sprintf("%v", netWorkAdapter.Oem.Huawei.AssociatedResource)
	} else if netWorkAdapter.Oem.H3c.AssociatedResource != "" {
		networkAdapterSlotNum = fmt.Sprintf("%v", netWorkAdapter.Oem.H3c.AssociatedResource)
	}
	var networkAdapterSN string
	if netWorkAdapter.SerialNumber != "" {
		networkAdapterSN = fmt.Sprintf("%v", netWorkAdapter.SerialNumber)
	} else if netWorkAdapter.Oem.Huawei.AssociatedResource != "" {
		networkAdapterSN = fmt.Sprintf("%v", netWorkAdapter.Oem.Huawei.BoardIdHex)
	} else if netWorkAdapter.Oem.H3c.AssociatedResource != "" {
		networkAdapterSN = fmt.Sprintf("%v", netWorkAdapter.Oem.H3c.BoardIdHex)
	}

	chassisNetworkAdapterLabelValues := []string{assetTag, serialNumber, networkAdapterID, networkAdapterModel, networkAdapterSN, networkAdapterSlotNum, "NetCard"}

	if netWorkAdapterStateVaule, ok := parseCommonStatusState(netWorkAdapterState); ok {
		ch <- prometheus.MustNewConstMetric(chassisMetrics["chassis_networkAdapter_state"].desc, prometheus.GaugeValue, netWorkAdapterStateVaule, chassisNetworkAdapterLabelValues...)
	}
	if netWorkAdapterHealthStateVaule, ok := parseCommonStatusHealth(netWorkAdapterHealthState); ok {
		ch <- prometheus.MustNewConstMetric(chassisMetrics["system_networkAdapterHealth_state"].desc, prometheus.GaugeValue, netWorkAdapterHealthStateVaule, chassisNetworkAdapterLabelValues...)
	}

}

func parseNetworkPorts(ch chan<- prometheus.Metric, assetTag string, serialNumber string, netAdapter_id string, networkPort *redfish2.NetworkPort, wg *sync.WaitGroup) {
	defer wg.Done()
	networkPortID := networkPort.ID
	networkPortMac := networkPort.AssociatedNetworkAddresses[0]
	networkPortSpeed := networkPort.Oem.Public.CurrentSpeed
	networkPortLinkStatus := networkPort.LinkStatus

	var networkPortType string
	if networkPort.Oem.Public.PortType != "" {
		networkPortType = fmt.Sprintf("%v", networkPort.Oem.Public.PortType)
	} else if networkPort.Oem.Huawei.PortType != "" {
		networkPortType = fmt.Sprintf("%v", networkPort.Oem.Huawei.PortType)
	} else if networkPort.Oem.H3c.PortType != "" {
		networkPortType = fmt.Sprintf("%v", networkPort.Oem.H3c.PortType)
	}
	
	networkPortLabelValues := []string{assetTag, serialNumber, netAdapter_id, networkPortID, networkPortMac, networkPortSpeed, networkPortType, "EthInterface"}
	if networkPortLinkStatusValue, ok := parseZteLinkStatus(redfish2.LinkStatus(networkPortLinkStatus)); ok {
		ch <- prometheus.MustNewConstMetric(chassisMetrics["network_port_linkStatus"].desc, prometheus.GaugeValue, networkPortLinkStatusValue, networkPortLabelValues...)
	}
}
