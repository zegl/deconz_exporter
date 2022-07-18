package main

import (
	"github.com/jurgen-kluft/go-conbee/sensors"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
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

	batteryMetric           *prometheus.GaugeVec
	temperatureMetric       *prometheus.GaugeVec
	humidityMetric          *prometheus.GaugeVec
	pressureMetric          *prometheus.GaugeVec
	lightLevelMetric        *prometheus.GaugeVec
	presenceMetric          *prometheus.GaugeVec
	energyPowerMetric       *prometheus.GaugeVec // Watt
	energyConsumptionMetric *prometheus.GaugeVec // kWh
}

var variableGroupLabelNames = []string{
	"name",
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
			"name": l.Name,
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
	}

	c.batteryMetric.Collect(ch)
	c.temperatureMetric.Collect(ch)
	c.humidityMetric.Collect(ch)
	c.pressureMetric.Collect(ch)
	c.lightLevelMetric.Collect(ch)
	c.presenceMetric.Collect(ch)
	c.energyPowerMetric.Collect(ch)
	c.energyConsumptionMetric.Collect(ch)
}
