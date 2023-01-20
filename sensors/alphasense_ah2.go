package sensors

import (
	"encoding/binary"
	"log"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"periph.io/x/conn/v3/i2c"
)

const (

	// AlphaSense AH2 (ppb) is an analog device
	// Using ADC with i2C interface: LTC2485
	//
	// https://www.analog.com/media/en/technical-documentation/data-sheets/2485fd.pdf
	// Page 12
	PID_I2C_ADDR uint16 = 0x24
	// Page 5, line t_conv_1 @ 60Hz Rejection Mode
	PID_MEASURE_TIME_MS time.Duration = 150 * time.Millisecond
)

func MeasurePIDLoop(bus i2c.Bus, client mqtt.Client) {
	var err error

	for {
		time.Sleep(PID_MEASURE_TIME_MS)

		// Data actually contains 24bits (3 bytes) with MSB being the sign
		// We still need to read all 4 bytes in order to trigger the STOP condition
		// See page 12 of the specsheet
		beData := make([]byte, 4)
		err = bus.Tx(PID_I2C_ADDR, nil, beData)
		if err != nil {
			sendNaN(HAF_VALUE_BROADCAST_MQTT_TOPIC_PATH, client)
			continue
		}

		adcValue := binary.BigEndian.Uint32(beData) ^ 0x80000000
		voltValue := float64(adcValue) * 5.0 / 2147483648.0

		log.Printf("PID value: %d [%.4f V]", adcValue, voltValue)

		msg := SensorValue{
			Value:     float64(adcValue),
			Timestamp: time.Now().UnixMilli(),
		}
		err = PublishJsonMsg(AH2_VALUE_BROADCAST_MQTT_TOPIC_PATH, msg, client)
		if err != nil {
			log.Println(err)
		}
	}
}
