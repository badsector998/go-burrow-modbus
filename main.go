package main

import (
	"log"

	"github.com/goburrow/serial"
	modbus "github.com/things-go/go-modbus"
)

func main() {
	// Modbus TCP
	handler := modbus.NewRTUClientProvider(modbus.WithEnableLogger(), modbus.WithSerialConfig(serial.Config{
		Address:  "COM3",
		BaudRate: 9600,
		DataBits: 8,
		StopBits: 1,
		Parity:   "N",
		Timeout:  modbus.DefaultConnectTimeout,
	}))

	client := modbus.NewClient(handler)
	if err := client.Connect(); err != nil {
		log.Panic("Error on modbus connect")
	}
	defer client.Close()

	bytes, err := client.ReadHoldingRegistersBytes(2, 1, 2)
	if err != nil {
		log.Panic("Error reading holding register")
	}
	println(bytes)
}
