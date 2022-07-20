package main

import (
	"fmt"
	"github.com/jurgen-kluft/go-conbee/sensors"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
	"time"
)

var (
	gaugeValueBool = map[bool]float64{
		false: 0,
		true:  1,
	}
)

type deconzCollector struct {
	logger  *zap.Logger
	sensors *sensors.Sensors

	batteryMetric            *prometheus.GaugeVec
	temperatureMetric        *prometheus.GaugeVec
	humidityMetric           *prometheus.GaugeVec
	pressureMetric           *prometheus.GaugeVec
	lightLevelMetric         *prometheus.GaugeVec
	presenceMetric           *prometheus.GaugeVec
	energyPowerMetric        *prometheus.GaugeVec // Watt
	energyConsumptionMetric  *prometheus.GaugeVec // kWh
	sensorLastSeenSecondsAgo *prometheus.GaugeVec
}

var variableGroupLabelNames = []string{
	"name",
	"model_id",
	"sw_version",
	"manufacturer_name",
}

func NewDeconzCollector(namespace string, log *zap.Logger, s *sensors.Sensors) prometheus.Collector {
	return &deconzCollector{
		logger:  log,
		sensors: s,

		batteryMetric: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "battery",
				Name:      "level",
				Help:      "Battery level percentage (0-100)",
			},
			variableGroupLabelNames,
		),

		temperatureMetric: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "climate",
				Name:      "temperature",
				Help:      "Temperature (Celsius)",
			},
			variableGroupLabelNames,
		),
		pressureMetric: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "climate",
				Name:      "pressure",
				Help:      "Pressure (Pascal)",
			},
			variableGroupLabelNames,
		),
		humidityMetric: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "climate",
				Name:      "humidity",
				Help:      "Relative Humidity %",
			},
			variableGroupLabelNames,
		),

		lightLevelMetric: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "light",
				Name:      "level",
			},
			variableGroupLabelNames,
		),
		presenceMetric: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "light",
				Name:      "presence",
			},
			variableGroupLabelNames,
		),

		energyPowerMetric: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "energy",
				Name:      "power",
				Help:      "Power (Watt)",
			},
			variableGroupLabelNames,
		),
		energyConsumptionMetric: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "energy",
				Name:      "consumption",
				Help:      "Energy consumption (kWh)",
			},
			variableGroupLabelNames,
		),
		sensorLastSeenSecondsAgo: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "sensor",
				Name:      "last_seen_seconds_ago",
				Help:      "Seconds since last update",
			},
			variableGroupLabelNames,
		),
	}
}

func (c *deconzCollector) Describe(ch chan<- *prometheus.Desc) {
	c.batteryMetric.Describe(ch)
	c.temperatureMetric.Describe(ch)
	c.humidityMetric.Describe(ch)
	c.pressureMetric.Describe(ch)
	c.lightLevelMetric.Describe(ch)
	c.presenceMetric.Describe(ch)
	c.energyPowerMetric.Describe(ch)
	c.energyConsumptionMetric.Describe(ch)
	c.sensorLastSeenSecondsAgo.Describe(ch)
}

func (c *deconzCollector) Collect(ch chan<- prometheus.Metric) {
	c.batteryMetric.Reset()
	c.temperatureMetric.Reset()
	c.humidityMetric.Reset()
	c.pressureMetric.Reset()
	c.lightLevelMetric.Reset()
	c.presenceMetric.Reset()
	c.energyPowerMetric.Reset()
	c.energyConsumptionMetric.Reset()
	c.sensorLastSeenSecondsAgo.Reset()

	sens, err := c.sensors.GetAllSensors()
	if err != nil {
		c.logger.Error("Failed to get sensor values", zap.Error(err))
		return
	}

	for _, l := range sens {
		c.logger.Info("sensor", zap.Any("sensor", l))

		if l.Name == "" {
			c.logger.Error("Sensor has no name", zap.Int("id", l.ID))
			return
		}

		labels := prometheus.Labels{
			"name":              l.Name,
			"model_id":          l.ModelID,
			"manufacturer_name": l.ManufacturerName,
			"sw_version":        l.SWVersion,
		}

		switch l.Type {
		case "ZHATemperature":
			c.temperatureMetric.With(labels).Set(float64(l.State.Temperature) / 100)
		case "ZHAHumidity":
			c.humidityMetric.With(labels).Set(float64(l.State.Humidity) / 100)
		case "ZHAPressure":
			c.pressureMetric.With(labels).Set(float64(l.State.Pressure))
		case "ZHALightLevel":
			c.lightLevelMetric.With(labels).Set(float64(l.State.LightLevel))
		case "ZHAPresence":
			c.presenceMetric.With(labels).Set(gaugeValueBool[l.State.Presence])
		case "ZHAPower":
			c.energyPowerMetric.With(labels).Set(float64(l.State.Power))
		case "ZHAConsumption":
			c.energyConsumptionMetric.With(labels).Set(float64(l.State.Consumption))
		}

		if l.Config.Battery > 0 {
			c.batteryMetric.With(labels).Set(float64(l.Config.Battery))
		}

		if l.State.LastUpdated != "" {
			if ts, err := time.Parse("2006-01-02T15:04:05.999", l.State.LastUpdated); err == nil {
				c.sensorLastSeenSecondsAgo.With(labels).Set(time.Now().Sub(ts).Seconds())
			} else {
				fmt.Println(l.State.LastUpdated, err)
			}
		}
	}

	c.batteryMetric.Collect(ch)
	c.temperatureMetric.Collect(ch)
	c.humidityMetric.Collect(ch)
	c.pressureMetric.Collect(ch)
	c.lightLevelMetric.Collect(ch)
	c.presenceMetric.Collect(ch)
	c.energyPowerMetric.Collect(ch)
	c.energyConsumptionMetric.Collect(ch)
	c.sensorLastSeenSecondsAgo.Collect(ch)
}
