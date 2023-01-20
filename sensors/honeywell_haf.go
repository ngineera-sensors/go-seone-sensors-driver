package sensors

import (
	"encoding/binary"
	"log"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"periph.io/x/conn/v3/i2c"
)

const (
	// Honeywell Zephyr HAF
	// https://www.mouser.fr/datasheet/2/187/HWSC_S_A0001296200_1-3073185.pdf
	//
	//
	HAF_I2C_ADDR        uint16        = 0x49
	HAF_MEASURE_TIME_MS time.Duration = 150 * time.Millisecond // Minimum measure time is ~10ms
)

func MeasureHAFLoop(bus i2c.Bus, client mqtt.Client) {
	var err error
	for {
		err = bus.Tx(HAF_I2C_ADDR, []byte{0}, nil)
		if err != nil {
			sendNaN(HAF_VALUE_BROADCAST_MQTT_TOPIC_PATH, client)
			continue
		}

		time.Sleep(HAF_MEASURE_TIME_MS)

		data := make([]byte, 2)
		err = bus.Tx(HAF_I2C_ADDR, nil, data)
		if err != nil {
			log.Fatal(err)
		}

		rawValue := binary.BigEndian.Uint16(data)

		// 100 is the hard-coded Full Scale Flow for HAFBLF0100C4AX5
		// 16384 is to account for first 2 bits being always 00
		flowValue := 100 * ((float64(rawValue) / 16384) - 0.5) / 0.4
		log.Println("Flowrate value is: ", flowValue)

		msg := SensorValue{
			Value:     flowValue,
			Timestamp: time.Now().UnixMilli(),
		}
		err = PublishJsonMsg(HAF_VALUE_BROADCAST_MQTT_TOPIC_PATH, msg, client)
		if err != nil {
			log.Println(err)
		}
	}
}
