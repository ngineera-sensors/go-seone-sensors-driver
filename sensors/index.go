package sensors

import (
	"log"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"periph.io/x/conn/v3/driver/driverreg"
	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/host/v3"
)

const (
	PRIMARY_I2C_BUS_NUMBER    = "1"
	I2C_READ_ERROR_SLEEP_TIME = 1000 * time.Millisecond

	HIH_HUMIDITY_VALUE_BROADCAST_MQTT_TOPIC_PATH    = "/sensors/hih-humidity/broadcast"
	HIH_TEMPERATURE_VALUE_BROADCAST_MQTT_TOPIC_PATH = "/sensors/hih-temperature/broadcast"
	HAF_VALUE_BROADCAST_MQTT_TOPIC_PATH             = "/sensors/haf/broadcast"
	AH2_VALUE_BROADCAST_MQTT_TOPIC_PATH             = "/sensors/ah2/broadcast"
)

var (
	SEONE_SN = ""
)

type SensorValue struct {
	Value     float64
	Timestamp int64 // milliseconds
}

func init() {
	sn, err := os.ReadFile(filepath.Join("config", "serialnumber.txt"))
	if err != nil {
		log.Fatal(err)
	}
	if len(sn) != 0 {
		log.Printf("Setting SEONE_SN value: %s", string(sn))
	}
	snStr := string(sn)
	snStr = strings.TrimSpace(snStr)
	SEONE_SN = snStr
}

func sendNaN(topic string, client mqtt.Client) {
	msg := SensorValue{
		Value:     math.NaN(),
		Timestamp: time.Now().UnixMilli(),
	}
	err := PublishJsonMsg(HAF_VALUE_BROADCAST_MQTT_TOPIC_PATH, msg, client)
	if err != nil {
		log.Println(err)
	}
	time.Sleep(I2C_READ_ERROR_SLEEP_TIME)
}

func MainLoop() {
	host.Init()

	if _, err := driverreg.Init(); err != nil {
		log.Fatal(err)
	}

	// Use i2creg I²C bus registry to find the first available I²C bus.
	bus, err := i2creg.Open(PRIMARY_I2C_BUS_NUMBER)
	if err != nil {
		log.Fatal(err)
	}
	defer bus.Close()

	client, err := NewMQTTClient()
	if err != nil {
		log.Printf("Error occurred while initializing MQTT client: %s. Exiting.", err)
		return
	}
	go MeasureHIHLoop(bus, client)
	go MeasureHAFLoop(bus, client)
	go MeasurePIDLoop(bus, client)
	for {
		time.Sleep(1 * time.Second)
	}
}
