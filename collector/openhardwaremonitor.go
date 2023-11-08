package collector

import (
	"strconv"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/yusufpapurcu/wmi"
)

// A openHardwareMonitorCollector is a Prometheus collector for WMI Sensor metrics
type SensorCollector struct {
	logger log.Logger

	SensorType *prometheus.Desc
	Identifier *prometheus.Desc
	Parent     *prometheus.Desc
	Name       *prometheus.Desc
	Value      *prometheus.Desc
	Max        *prometheus.Desc
	Min        *prometheus.Desc
	Index      *prometheus.Desc
}

// NewOpenHardwareMonitorCollector ...
func NewOpenHardwareMonitorCollector(logger log.Logger) (Collector, error) {
	const subsystem = "open_hardware_monitor"
	return &SensorCollector{
		logger: log.With(logger, "collector", subsystem),
		Value: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "sensor_value"),
			"Provides the value from an OpenHardwareMonitor sensor.",
			[]string{"parent", "index", "name", "sensor_type"},
			nil,
		),
	}, nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *SensorCollector) Collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) error {
	if desc, err := c.collect(ch); err != nil {
		_ = level.Error(c.logger).Log("failed collecting openHardwareMonitor metrics:", desc, err)
		return err
	}
	return nil
}

// OpenHardwareMonitor Sensor:
type Sensor struct {
	SensorType string
	Identifier string
	Parent     string
	Name       string
	Value      float32
	Max        float32
	Min        float32
	Index      uint32
}

func (c *SensorCollector) collect(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []Sensor
	q := queryAll(&dst, c.logger)
	if err := wmi.QueryNamespace(q, &dst, "root/OpenHardwareMonitor"); err != nil {
		return nil, err
	}

	for _, info := range dst {
		ch <- prometheus.MustNewConstMetric(
			c.Value,
			prometheus.GaugeValue,
			float64(info.Value),
			info.Parent,
			strconv.Itoa(int(info.Index)),
			info.Name,
			info.SensorType,
		)
	}

	return nil, nil
}
