package sensors

import (
	"fmt"
	"log"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"periph.io/x/conn/v3/i2c"
)

const (
	// https://prod-edam.honeywell.com/content/dam/honeywell-edam/sps/siot/en-us/products/sensors/humidity-with-temperature-sensors/common/documents/sps-siot-i2c-comms-humidicon-tn-009061-2-en-ciid-142171.pdf
	HIH_I2C_ADDR uint16 = 0x27
	// Page 3 (section 3.0) (36.65ms)
	HIH_MEASURE_TIME_MS     time.Duration = 150 * time.Millisecond // minimum measure time is 37ms
	HIH_STATUS_NORMAL       byte          = 0
	HIH_STATUS_STALE        byte          = 1
	HIH_STATUS_COMMAND_MODE byte          = 2
	HIH_STATUS_ERROR        byte          = 3
)

func MeasureHIHLoop(bus i2c.Bus, client mqtt.Client) {
	var err error
	for {
		err = bus.Tx(HIH_I2C_ADDR, []byte{0}, nil)
		if err != nil {
			sendNaN(HAF_VALUE_BROADCAST_MQTT_TOPIC_PATH, client)
			continue
		}

		time.Sleep(HIH_MEASURE_TIME_MS)

		data := make([]byte, 4)
		err = bus.Tx(HIH_I2C_ADDR, nil, data)
		if err != nil {
			log.Fatal(err)
		}
		// log.Println("Read data: ", data)
		status := data[0] >> 6 // Status is in first 2 bits

		switch status {
		case HIH_STATUS_ERROR:
			log.Println(fmt.Errorf("HIH in ERROR state; Continuing.."))
			continue
		case HIH_STATUS_STALE:
			log.Println("HIH is STALE; Continuing..")
			continue
		}

		// Convert bytes into uint16
		var intData []uint16 = make([]uint16, 4)
		for i, b := range data {
			intData[i] = uint16(b)
		}
		raw_humidity := ((intData[0] & 0b00111111) << 8) | intData[1] // first byte has only 6 signficant bits (2 first bits is status)
		raw_temperature := ((intData[2] << 8) | intData[3]) >> 2      // last 2 bits of last byte to be ignored

		// log.Printf("Raw temperature: %d; Raw humidity: %d", raw_temperature, raw_humidity)

		humidity := float64(raw_humidity) / 16382                // 14 bits precision has 16383 levels
		temperature := (float64(raw_temperature)/16382)*165 - 40 // Honeywell calibration: gain=165, offset=-40

		log.Printf("Temperature: %.2f; Humidity: %.4f", temperature, humidity)

		ts := time.Now().UnixMilli()

		msgH := SensorValue{
			Value:     humidity,
			Timestamp: ts,
		}
		err = PublishJsonMsg(HIH_HUMIDITY_VALUE_BROADCAST_MQTT_TOPIC_PATH, msgH, client)
		if err != nil {
			log.Println(err)
		}

		msgT := SensorValue{
			Value:     temperature,
			Timestamp: ts,
		}
		err = PublishJsonMsg(HIH_TEMPERATURE_VALUE_BROADCAST_MQTT_TOPIC_PATH, msgT, client)
		if err != nil {
			log.Println(err)
		}
	}
}
