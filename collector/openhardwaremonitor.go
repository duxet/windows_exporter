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

	Clock       *prometheus.Desc
	Control     *prometheus.Desc
	FanSpeed    *prometheus.Desc
	Flow        *prometheus.Desc
	Level       *prometheus.Desc
	Load        *prometheus.Desc
	Temperature *prometheus.Desc
	Voltage     *prometheus.Desc
}

// NewOpenHardwareMonitorCollector ...
func NewOpenHardwareMonitorCollector(logger log.Logger) (Collector, error) {
	const subsystem = "open_hardware_monitor"
	return &SensorCollector{
		logger: log.With(logger, "collector", subsystem),
		Clock: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "sensor_clock_mhz"),
			"Clock speed from an OpenHardwareMonitor sensor in megahertz",
			[]string{"parent", "index", "name"},
			nil,
		),
		Control: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "sensor_control_percent"),
			"Duty cycle from an OpenHardwareMonitor sensor in percent",
			[]string{"parent", "index", "name"},
			nil,
		),
		FanSpeed: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "sensor_fan_speed_rpm"),
			"Fan speed from an OpenHardwareMonitor sensor in revolutions per minute",
			[]string{"parent", "index", "name"},
			nil,
		),
		Flow: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "sensor_flow_liters_per_hour"),
			"Flow from an OpenHardwareMonitor sensor in liters per hour",
			[]string{"parent", "index", "name"},
			nil,
		),
		Level: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "sensor_level_percent"),
			"Generic level measure from an OpenHardwareMonitor sensor in percent",
			[]string{"parent", "index", "name"},
			nil,
		),
		Load: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "sensor_load_percent"),
			"Load from an OpenHardwareMonitor sensor in percentages",
			[]string{"parent", "index", "name"},
			nil,
		),
		Temperature: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "sensor_temperature_degrees"),
			"Temperature from an OpenHardwareMonitor sensor in degrees Celsius",
			[]string{"parent", "index", "name"},
			nil,
		),
		Voltage: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "sensor_voltage_volts"),
			"Voltage from an OpenHardwareMonitor sensor in volts",
			[]string{"parent", "index", "name"},
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

	descMap := map[string]*prometheus.Desc{
		"Clock":       c.Clock,
		"Control":     c.Control,
		"Fan":         c.FanSpeed,
		"Flow":        c.Flow,
		"Level":       c.Level,
		"Load":        c.Load,
		"Temperature": c.Temperature,
		"Voltage":     c.Voltage,
	}

	for _, info := range dst {
		if desc, exists := descMap[info.SensorType]; exists {
			ch <- prometheus.MustNewConstMetric(
				desc,
				prometheus.GaugeValue,
				float64(info.Value),
				info.Parent,
				strconv.Itoa(int(info.Index)),
				info.Name,
			)
		}
	}

	return nil, nil
}
