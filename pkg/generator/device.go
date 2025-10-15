package generator

import (
	"math"
	"math/rand"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"procodus.dev/demo-app/pkg/iot"
)

type IoTDevice struct {
	Timestamp  time.Time
	DeviceID   string  `fake:"{uuid}"`
	Location   string  `fake:"{city}, {state}"`
	MacAddress string  `fake:"{macaddress}"`
	IPAddress  string  `fake:"{ipv4address}"`
	Firmware   string  `fake:"{appversion}"`
	Latitude   float64 `fake:"{latitude}"`
	Longitude  float64 `fake:"{longitude}"`
}

type IoTDataGenerator struct {
	deviceID         string
	baselineTemp     float64
	baselineHumidity float64
	baselinePressure float64
	noise            float64
	pressureTrend    float64 // Simulates weather system movement
	lastPressure     float64
}

func NewIoTDevice() *IoTDevice {
	var device IoTDevice
	err := gofakeit.Struct(&device)
	if err != nil {
		return nil
	}
	device.Timestamp = time.Now()
	return &device
}

func NewIoTGenerator(deviceID string) *IoTDataGenerator {
	return &IoTDataGenerator{
		deviceID:         deviceID,
		baselineTemp:     20.0 + rand.Float64()*10,         // 20-30°C
		baselineHumidity: 50.0 + rand.Float64()*20,         // 50-70%
		baselinePressure: 1013.0 + (rand.Float64()-0.5)*20, // 1003-1023 hPa
		noise:            rand.Float64() * 2,
		pressureTrend:    (rand.Float64() - 0.5) * 0.5, // Slow trend
		lastPressure:     1013.0,
	}
}

// GenerateTemperature with daily pattern.
func (g *IoTDataGenerator) GenerateTemperature(t time.Time) float64 {
	hour := float64(t.Hour())

	// Daily cycle (peak around 2-3 PM)
	dailyCycle := 5 * math.Sin((hour-6)*math.Pi/12)

	// Random noise
	noise := (rand.Float64() - 0.5) * g.noise

	// Occasional anomalies (5% chance)
	anomaly := 0.0
	if rand.Float64() < 0.05 {
		anomaly = (rand.Float64() - 0.5) * 15 // ±7.5°C spike
	}

	return g.baselineTemp + dailyCycle + noise + anomaly
}

// GenerateHumidity with inverse temperature correlation.
func (g *IoTDataGenerator) GenerateHumidity(t time.Time, temperature float64) float64 {
	hour := float64(t.Hour())

	// Daily cycle (inverse of temperature - higher at night)
	dailyCycle := -3 * math.Sin((hour-6)*math.Pi/12)

	// Inverse correlation with temperature
	// When temp is higher than baseline, humidity tends to be lower
	tempEffect := -(temperature - g.baselineTemp) * 1.5

	// Random noise (humidity is less noisy than temperature)
	noise := (rand.Float64() - 0.5) * g.noise * 0.5

	// Seasonal/weather pattern (slower changes)
	weatherPattern := 10 * math.Sin(float64(t.Unix())/(86400*7)) // Weekly cycle

	// Occasional anomalies (rain, etc.) - 3% chance
	anomaly := 0.0
	if rand.Float64() < 0.03 {
		anomaly = rand.Float64() * 20 // Humidity spike (rain)
	}

	humidity := g.baselineHumidity + dailyCycle + tempEffect + noise + weatherPattern + anomaly

	// Clamp between realistic bounds (20-95%)
	return math.Max(20, math.Min(95, humidity))
}

// GeneratePressure with slow trending behavior.
func (g *IoTDataGenerator) GeneratePressure(t time.Time) float64 {
	// Pressure changes slowly - simulate weather systems
	// Use random walk with trend

	// Small random change (±0.5 hPa per reading)
	randomChange := (rand.Float64() - 0.5) * 0.5

	// Apply trend (simulates high/low pressure system movement)
	trendChange := g.pressureTrend

	// Occasionally reverse trend (10% chance)
	if rand.Float64() < 0.1 {
		g.pressureTrend = -g.pressureTrend + (rand.Float64()-0.5)*0.2
	}

	// Very slow sinusoidal pattern (multi-day cycle)
	dayOfYear := float64(t.YearDay())
	seasonalPattern := 5 * math.Sin(dayOfYear*2*math.Pi/365)

	// Time-of-day effect (very subtle - pressure slightly higher in morning/evening)
	hour := float64(t.Hour())
	diurnalCycle := 0.5 * math.Sin((hour-3)*math.Pi/12)

	// Calculate new pressure based on last pressure (random walk)
	newPressure := g.lastPressure + randomChange + trendChange + diurnalCycle*0.1

	// Add baseline and seasonal pattern
	newPressure = g.baselinePressure + (newPressure-g.baselinePressure)*0.7 + seasonalPattern

	// Clamp to realistic bounds (980-1040 hPa)
	newPressure = math.Max(980, math.Min(1040, newPressure))

	// Occasional weather front (rapid pressure change) - 2% chance
	if rand.Float64() < 0.02 {
		frontChange := (rand.Float64() - 0.5) * 10 // ±5 hPa
		newPressure += frontChange
		g.pressureTrend = frontChange * 0.3 // Trend follows the front
	}

	g.lastPressure = newPressure
	return newPressure
}

// GenerateCorrelatedReading - generates readings with realistic correlations.
func (g *IoTDataGenerator) GenerateCorrelatedReading(t time.Time) *iot.SensorReading {
	// Generate temperature first
	temperature := g.GenerateTemperature(t)

	// Humidity is correlated with temperature
	humidity := g.GenerateHumidity(t, temperature)

	// Pressure is independent but slow-changing
	pressure := g.GeneratePressure(t)

	// Battery slowly drains over time
	hoursRunning := time.Since(t.Add(-720 * time.Hour)).Hours() // Assume started 30 days ago
	batteryDrain := hoursRunning / (720 * 1.2) * 100            // Drains over ~36 days
	battery := 100 - batteryDrain - rand.Float64()*2            // Add small random variation
	battery = math.Max(5, math.Min(100, battery))

	return &iot.SensorReading{
		DeviceId:     g.deviceID,
		Timestamp:    t.Unix(),
		Temperature:  math.Round(temperature*100) / 100, // 2 decimal places
		Humidity:     math.Round(humidity*100) / 100,
		Pressure:     math.Round(pressure*100) / 100,
		BatteryLevel: math.Round(battery*10) / 10, // 1 decimal place
	}
}
