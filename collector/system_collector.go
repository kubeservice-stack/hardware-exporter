package collector

import (
	"fmt"
	"hardware_exporter/lib/gofish"
	redfish "hardware_exporter/lib/gofish/redfish"
	"strings"
	"sync"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

// SystemSubsystem is the system subsystem
var (
	SystemSubsystem                   = "system"
	SystemLabelNames                  = []string{"server_assetTag", "server_sn", "server_name", "mfr", "system_id", "hw_model", "bios_version", "server_uid", "resource"}
	SystemMemoryLabelNames            = []string{"server_assetTag", "server_sn", "memory_id", "memory_name", "memory_type", "mfr", "memory_sn", "resource"}
	SystemProcessorLabelNames         = []string{"server_assetTag", "server_sn", "processor_id", "mfr", "model", "processor_sn", "MaxSpeedMHz", "ProcessorArchitecture", "resource"}
	SystemStorageControllerLabelNames = []string{"server_assetTag", "server_sn", "storage_controller_uid", "firmwareVersion", "storage_model", "resource"}
	SystemDriveLabelNames             = []string{"server_assetTag", "server_sn", "drive_id", "drive_name", "mfr", "Model", "drive_media_type", "resource"}
	SystemVolumeLabelNames            = []string{"server_assetTag", "server_sn", "volume_id", "volume", "resource"}

	systemMetrics = map[string]systemMetric{
		"system_state": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, SystemSubsystem, "state"),
				"system state,1(Enabled),2(Disabled),3(StandbyOffinline),4(StandbySpare),5(InTest),6(Starting),7(Absent),8(UnavailableOffline),9(Deferring),10(Quiesced),11(Updating)",
				SystemLabelNames,
				nil,
			),
		},
		"system_health_status": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, SystemSubsystem, "health_status"),
				"system health,1(OK),2(Warning),3(Critical)",
				SystemLabelNames,
				nil,
			),
		},
		"system_processor_summary_state": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, SystemSubsystem, "processor_summary_state"),
				"system overall processor state,1(Enabled),2(Disabled),3(StandbyOffinline),4(StandbySpare),5(InTest),6(Starting),7(Absent),8(UnavailableOffline),9(Deferring),10(Quiesced),11(Updating)",
				SystemLabelNames,
				nil,
			),
		},
		"system_processor_summary_health_status": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, SystemSubsystem, "processor_summary_health_status"),
				"system overall processor health,1(OK),2(Warning),3(Critical)",
				SystemLabelNames,
				nil,
			),
		},
		"system_processor_summary_count": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, SystemSubsystem, "processor_summary_count"),
				"system total processor count",
				SystemLabelNames,
				nil,
			),
		},
		"system_memory_summary_size": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, SystemSubsystem, "memory_summary_size"),
				"system total memory size, GiB",
				SystemLabelNames,
				nil,
			),
		},

		"system_processor_state": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, SystemSubsystem, "processor_state"),
				"system processor state,1(Enabled),2(Disabled),3(StandbyOffinline),4(StandbySpare),5(InTest),6(Starting),7(Absent),8(UnavailableOffline),9(Deferring),10(Quiesced),11(Updating)",
				SystemProcessorLabelNames,
				nil,
			),
		},
		"system_processor_health_status": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, SystemSubsystem, "processor_health_status"),
				"system processor health state,1(OK),2(Warning),3(Critical)",
				SystemProcessorLabelNames,
				nil,
			),
		},
		"system_processor_total_threads": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, SystemSubsystem, "processor_total_threads"),
				"system processor total threads",
				SystemProcessorLabelNames,
				nil,
			),
		},
		"system_processor_total_cores": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, SystemSubsystem, "processor_total_cores"),
				"system processor total cores",
				SystemProcessorLabelNames,
				nil,
			),
		},
		"system_total_memory_state": {
			prometheus.NewDesc(
				prometheus.BuildFQName(namespace, SystemSubsystem, "total_memory_state"),
				"system overall memory state,1(Enabled),2(Disabled),3(StandbyOffinline),4(StandbySpare),5(InTest),6(Starting),7(Absent),8(UnavailableOffline),9(Deferring),10(Quiesced),11(Updating)",
				SystemLabelNames,
				nil,
			),
		},
		"system_total_memory_health_state": {
			prometheus.NewDesc(
				prometheus.BuildFQName(namespace, SystemSubsystem, "total_memory_health_state"),
				"system overall memory health,1(OK),2(Warning),3(Critical)",
				SystemLabelNames,
				nil,
			),
		},
		"system_total_memory_size": {
			prometheus.NewDesc(
				prometheus.BuildFQName(namespace, SystemSubsystem, "total_memory_size"),
				"system total memory size, GiB",
				SystemLabelNames,
				nil,
			),
		},
		"system_memory_state": {
			prometheus.NewDesc(
				prometheus.BuildFQName(namespace, SystemSubsystem, "memory_state"),
				"system memory state,1(Enabled),2(Disabled),3(StandbyOffinline),4(StandbySpare),5(InTest),6(Starting),7(Absent),8(UnavailableOffline),9(Deferring),10(Quiesced),11(Updating)",
				SystemMemoryLabelNames,
				nil,
			),
		},
		"system_memory_health_state": {
			prometheus.NewDesc(
				prometheus.BuildFQName(namespace, SystemSubsystem, "memory_health_state"),
				"system memory  health state,1(OK),2(Warning),3(Critical)",
				SystemMemoryLabelNames,
				nil,
			),
		},
		"system_memory_capacity": {
			prometheus.NewDesc(
				prometheus.BuildFQName(namespace, SystemSubsystem, "memory_capacity"),
				"system memory capacity, MiB",
				SystemMemoryLabelNames,
				nil,
			),
		},
		"system_storageController_state": {
			prometheus.NewDesc(
				prometheus.BuildFQName(namespace, SystemSubsystem, "system_storageController_state"),
				"system storageController state,1(Enabled),2(Disabled),3(StandbyOffinline),4(StandbySpare),5(InTest),6(Starting),7(Absent),8(UnavailableOffline),9(Deferring),10(Quiesced),11(Updating)",
				SystemStorageControllerLabelNames,
				nil,
			),
		},
		"system_storageController_health_state": {
			prometheus.NewDesc(
				prometheus.BuildFQName(namespace, SystemSubsystem, "system_storageController_health_state"),
				"system storageController health state,1(OK),2(Warning),3(Critical)",
				SystemStorageControllerLabelNames,
				nil,
			),
		},

		"system_storage_volumeState": {
			prometheus.NewDesc(
				prometheus.BuildFQName(namespace, SystemSubsystem, "storage_volume_state"),
				"system storage volume state,1(Enabled),2(Disabled),3(StandbyOffinline),4(StandbySpare),5(InTest),6(Starting),7(Absent),8(UnavailableOffline),9(Deferring),10(Quiesced),11(Updating)",
				SystemVolumeLabelNames,
				nil,
			),
		},
		"system_storage_volumeHealthState": {
			prometheus.NewDesc(
				prometheus.BuildFQName(namespace, SystemSubsystem, "storage_volume_health_state"),
				"system storage volume health state,1(OK),2(Warning),3(Critical)",
				SystemVolumeLabelNames,
				nil,
			),
		},
		"system_storage_volumeCapacity": {
			prometheus.NewDesc(
				prometheus.BuildFQName(namespace, SystemSubsystem, "storage_volume_capacity"),
				"system storage volume capacity,Bytes",
				SystemVolumeLabelNames,
				nil,
			),
		},
		"system_storage_drive_state": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, SystemSubsystem, "storage_drive_state"),
				"system storage drive state,1(Enabled),2(Disabled),3(StandbyOffinline),4(StandbySpare),5(InTest),6(Starting),7(Absent),8(UnavailableOffline),9(Deferring),10(Quiesced),11(Updating)",
				SystemDriveLabelNames,
				nil,
			),
		},
		"system_storage_drive_health_state": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, SystemSubsystem, "storage_drive_health_state"),
				"system storage volume health state,1(OK),2(Warning),3(Critical)",
				SystemDriveLabelNames,
				nil,
			),
		},
		"system_storage_drive_capacity": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, SystemSubsystem, "storage_drive_capacity"),
				"system storage drive capacity,Bytes",
				SystemDriveLabelNames,
				nil,
			),
		},
	}
)

// SystemCollector implemented prometheus.Collector
type SystemCollector struct {
	redfishClient         *gofish.APIClient
	metrics               map[string]systemMetric
	collectorScrapeStatus *prometheus.GaugeVec
	Log                   log.Logger
}

type systemMetric struct {
	desc *prometheus.Desc
}

// NewSystemCollector returns a collector that collecting memory statistics
func NewSystemCollector(namespace string, redfishClient *gofish.APIClient, logger log.Logger) *SystemCollector {
	return &SystemCollector{
		redfishClient: redfishClient,
		metrics:       systemMetrics,
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

// Describe implements prometheus.Collector.
func (s *SystemCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, metric := range s.metrics {
		ch <- metric.desc
	}
	s.collectorScrapeStatus.Describe(ch)

}

// Collect implements prometheus.Collector.
func (s *SystemCollector) Collect(ch chan<- prometheus.Metric) {
	//get service
	service := s.redfishClient.Service
	// get a list of systems from service
	if systems, err := service.Systems(); err != nil {
		level.Error(s.Log).Log("operation", "service.Systems()", "err", err)
	} else {
		for _, system := range systems {
			level.Info(s.Log).Log("System", system.ID, "msg", "System collector scrape started")
			// overall system metrics
			// server info
			SystemID := system.ID
			SystemAssetTag := system.AssetTag
			SystemUUID := system.UUID
			SystemBiosVersion := system.BIOSVersion
			SerialNumber := system.SerialNumber
			systemModel := system.Model
			serverName := system.HostName
			systemManufacturer := "Unknown"

			if system.Manufacturer != "" {
				tmpStr := strings.Split(system.Manufacturer, " ")
				systemManufacturer = tmpStr[0]
			}
			if systemManufacturer == "Dell" {
				SerialNumber = system.SKU
			}

			//common status
			systemState := system.Status.State
			systemHealthStatus := system.Status.Health

			systemLabelValues := []string{SystemAssetTag, SerialNumber, serverName, systemManufacturer, SystemID, systemModel, SystemBiosVersion, SystemUUID, "Server"}
			// system state health
			if systemStateValue, ok := parseCommonStatusState(systemState); ok {
				ch <- prometheus.MustNewConstMetric(s.metrics["system_state"].desc, prometheus.GaugeValue, systemStateValue, systemLabelValues...)
			}
			if systemHealthStatusValue, ok := parseCommonStatusHealth(systemHealthStatus); ok {
				ch <- prometheus.MustNewConstMetric(s.metrics["system_health_status"].desc, prometheus.GaugeValue, systemHealthStatusValue, systemLabelValues...)
			}

			// process processor
			processors, err := system.Processors()
			if err != nil {
				level.Error(s.Log).Log("msg", "system.Processors()", "error", err)
			} else if processors == nil {
				level.Info(s.Log).Log("msg", "system.Processors()", "err", "no processor data found")
			} else {
				wg1 := &sync.WaitGroup{}
				wg1.Add(len(processors))
				for _, processor := range processors {
					go parsePorcessor(ch, SystemAssetTag, SerialNumber, processor, wg1)
				}
				wg1.Wait()
			}

			//process Memory
			memories, err := system.Memory()
			if err != nil {
				level.Error(s.Log).Log("msg", "system.Memory()", "error", err)
			} else if memories == nil {
				level.Info(s.Log).Log("msg", "system.Memory()", "err", "no memory data found")
			} else {
				wg2 := &sync.WaitGroup{}
				wg2.Add(len(memories))
				for _, memory := range memories {
					go parseMemory(ch, SystemAssetTag, SerialNumber, memory, wg2)
				}
				wg2.Wait()
			}
			storages, err := system.Storage()
			if err != nil {
				level.Error(s.Log).Log("msg", "system.Storage()", "error", err)
			} else if storages == nil {
				level.Info(s.Log).Log("msg", "system.Storage()", "err", "no storage data found")
			} else {
				for _, storage := range storages {
					//process storageController
					wg3 := &sync.WaitGroup{}
					wg3.Add(len(storage.StorageControllers))
					for _, storageController := range storage.StorageControllers {
						go parseStorageController(ch, SystemAssetTag, SerialNumber, &storageController, wg3)
					}
					wg3.Wait()
					//process Drives
					drives, err := storage.Drives()
					if err != nil {
						level.Error(s.Log).Log("msg", "system.Drives()", "error", err)
					} else if drives == nil {
						level.Info(s.Log).Log("msg", "system.Drives()", "err", "no drive data found")
					} else {
						wg4 := &sync.WaitGroup{}
						wg4.Add(len(drives))
						for _, drive := range drives {
							go parseDrive(ch, SystemAssetTag, SerialNumber, drive, wg4)
						}
						wg4.Wait()
					}
					// process volumes
					volumes, err := storage.Volumes()
					if err != nil {
						level.Error(s.Log).Log("msg", "system.Volumes()", "error", err)
					} else if volumes == nil {
						level.Info(s.Log).Log("msg", "system.Drives()", "err", "no drive data found")
					} else {
						wg5 := &sync.WaitGroup{}
						wg5.Add(len(volumes))
						for _, volume := range volumes {
							go parseVolume(ch, SystemAssetTag, SerialNumber, volume, wg5)
						}
						wg5.Wait()
					}
				}
			}
			level.Info(s.Log).Log("System", SystemID, "msg", "System collector scrape completed")
		}
		s.collectorScrapeStatus.WithLabelValues("system").Set(float64(1))
	}

}

func parsePorcessor(ch chan<- prometheus.Metric, assetTag string, serialNumber string, processor *redfish.Processor, wg *sync.WaitGroup) {
	defer wg.Done()
	processorID := processor.ID
	processorModel := processor.Model
	processorTotalCores := processor.TotalCores
	processorTotalThreads := processor.TotalThreads
	processorState := processor.Status.State
	processorHealthStatus := processor.Status.Health
	processorManufacturer := processor.Manufacturer
	processorProcessorArchitecture := string(processor.ProcessorArchitecture)

	processorMaxSpeedMHz := fmt.Sprintf("%vMhz", processor.MaxSpeedMHz)

	var processorSerialNumber string
	if processor.Oem.H3C.SerialNumber != "" {
		processorSerialNumber = processor.Oem.H3C.SerialNumber
	} else if processor.Oem.Huawei.SerialNumber != "" {
		processorSerialNumber = processor.Oem.Huawei.SerialNumber
	} else {
		processorSerialNumber = processor.Oem.Public.SerialNumber
	}

	systemProcessorLabelValues := []string{assetTag, serialNumber, processorID, processorManufacturer, processorModel, processorSerialNumber, processorMaxSpeedMHz, processorProcessorArchitecture, "Processor"}

	if processorStateValue, ok := parseCommonStatusState(processorState); ok {
		ch <- prometheus.MustNewConstMetric(systemMetrics["system_processor_state"].desc, prometheus.GaugeValue, processorStateValue, systemProcessorLabelValues...)
	}
	if processorHealthStatusValue, ok := parseCommonStatusHealth(processorHealthStatus); ok {
		ch <- prometheus.MustNewConstMetric(systemMetrics["system_processor_health_status"].desc, prometheus.GaugeValue, processorHealthStatusValue, systemProcessorLabelValues...)
	}
	ch <- prometheus.MustNewConstMetric(systemMetrics["system_processor_total_threads"].desc, prometheus.GaugeValue, float64(processorTotalThreads), systemProcessorLabelValues...)
	ch <- prometheus.MustNewConstMetric(systemMetrics["system_processor_total_cores"].desc, prometheus.GaugeValue, float64(processorTotalCores), systemProcessorLabelValues...)
}

func parseMemory(ch chan<- prometheus.Metric, assetTag string, serialNumber string, memory *redfish.Memory, wg *sync.WaitGroup) {
	defer wg.Done()
	memoryName := memory.Name
	memoryID := memory.ID
	memoryDeviceType := string(memory.MemoryDeviceType)
	memoryManufacturer := memory.Manufacturer
	memoryCapacityMiB := memory.CapacityMiB
	memorySerialNumber := memory.SerialNumber
	memoryState := memory.Status.State
	memoryHealthState := memory.Status.Health
	systemMemoryLabelValues := []string{assetTag, serialNumber, memoryID, memoryName, memoryDeviceType, memoryManufacturer, memorySerialNumber, "Memory"}
	if memoryStateValue, ok := parseCommonStatusState(memoryState); ok {
		ch <- prometheus.MustNewConstMetric(systemMetrics["system_memory_state"].desc, prometheus.GaugeValue, memoryStateValue, systemMemoryLabelValues...)
	}
	if memoryHealthStateValue, ok := parseCommonStatusHealth(memoryHealthState); ok {
		ch <- prometheus.MustNewConstMetric(systemMetrics["system_memory_health_state"].desc, prometheus.GaugeValue, memoryHealthStateValue, systemMemoryLabelValues...)
	}
	ch <- prometheus.MustNewConstMetric(systemMetrics["system_memory_capacity"].desc, prometheus.GaugeValue, float64(memoryCapacityMiB), systemMemoryLabelValues...)

}

func parseStorageController(ch chan<- prometheus.Metric, assetTag string, serialNumber string, storageController *redfish.StorageController, wg *sync.WaitGroup) {
	defer wg.Done()
	storageControllerUID := storageController.SerialNumber
	storageControllerFirmwareVersion := storageController.FirmwareVersion
	storageControllerModel := storageController.Model
	storageControllerState := storageController.Status.State
	storageControllerHealthState := storageController.Status.Health
	storageStorageControllerLabelValues := []string{assetTag, serialNumber, storageControllerUID, storageControllerFirmwareVersion, storageControllerModel, "StorageController"}
	if storageControllerStateValue, ok := parseCommonStatusState(storageControllerState); ok {
		ch <- prometheus.MustNewConstMetric(systemMetrics["system_storageController_state"].desc, prometheus.GaugeValue, storageControllerStateValue, storageStorageControllerLabelValues...)
	}
	if storageControllerHealthStateValue, ok := parseCommonStatusHealth(storageControllerHealthState); ok {
		ch <- prometheus.MustNewConstMetric(systemMetrics["system_storageController_health_state"].desc, prometheus.GaugeValue, storageControllerHealthStateValue, storageStorageControllerLabelValues...)
	}

}

func parseDrive(ch chan<- prometheus.Metric, assetTag string, serialNumber string, drive *redfish.Drive, wg *sync.WaitGroup) {
	defer wg.Done()
	driveName := drive.Name
	driveUID := drive.SerialNumber
	driveModel := drive.Model
	driveManufacturer := drive.Manufacturer
	driveMediaType := drive.MediaType
	driveCapacityBytes := drive.CapacityBytes
	driveState := drive.Status.State
	driveHealthState := drive.Status.Health
	storageDriveLabelValues := []string{assetTag, serialNumber, driveUID, driveName, driveManufacturer, driveModel, string(driveMediaType), "Drive"}
	if driveStateValue, ok := parseCommonStatusState(driveState); ok {
		ch <- prometheus.MustNewConstMetric(systemMetrics["system_storage_drive_state"].desc, prometheus.GaugeValue, driveStateValue, storageDriveLabelValues...)
	}
	if driveHealthStateValue, ok := parseCommonStatusHealth(driveHealthState); ok {
		ch <- prometheus.MustNewConstMetric(systemMetrics["system_storage_drive_health_state"].desc, prometheus.GaugeValue, driveHealthStateValue, storageDriveLabelValues...)
	}
	ch <- prometheus.MustNewConstMetric(systemMetrics["system_storage_drive_capacity"].desc, prometheus.GaugeValue, float64(driveCapacityBytes), storageDriveLabelValues...)
}

func parseVolume(ch chan<- prometheus.Metric, assetTag string, serialNumber string, volume *redfish.Volume, wg *sync.WaitGroup) {
	defer wg.Done()
	volumeName := volume.Name
	volumeID := volume.ID
	volumeCapacityBytes := volume.CapacityBytes
	volumeState := volume.Status.State
	volumeHealthState := volume.Status.Health

	systemVolumeLabelValues := []string{assetTag, serialNumber, volumeID, volumeName, "LogicalDrive"}

	if volumeStateValue, ok := parseCommonStatusState(volumeState); ok {
		ch <- prometheus.MustNewConstMetric(systemMetrics["system_storage_volumeState"].desc, prometheus.GaugeValue, volumeStateValue, systemVolumeLabelValues...)
	}
	if volumeHealthStateValue, ok := parseCommonStatusHealth(volumeHealthState); ok {
		ch <- prometheus.MustNewConstMetric(systemMetrics["system_storage_volumeHealthState"].desc, prometheus.GaugeValue, volumeHealthStateValue, systemVolumeLabelValues...)
	}
	ch <- prometheus.MustNewConstMetric(systemMetrics["system_storage_volumeCapacity"].desc, prometheus.GaugeValue, float64(volumeCapacityBytes), systemVolumeLabelValues...)
}
