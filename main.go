package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"time"

	"github.com/goburrow/modbus"
)

func main() {
	// Modbus TCP

	handlerSlave1 := modbus.NewTCPClientHandler("localhost:502")
	handlerSlave1.SlaveId = byte(1)
	handlerSlave1.Timeout = 10 * time.Second

	handlerSlave2 := modbus.NewTCPClientHandler("localhost:502")
	handlerSlave2.SlaveId = byte(2)
	handlerSlave2.Timeout = 10 * time.Second

	handlerSlave3 := modbus.NewTCPClientHandler("localhost:502")
	handlerSlave3.SlaveId = byte(3)
	handlerSlave3.Timeout = 10 * time.Second

	// handler.Logger = log.New(os.Stdout, "test: ", log.LstdFlags)
	// Connect manually so that multiple requests are handled in one connection session
	err := handlerSlave2.Connect()
	if err != nil {
		log.Panic("Error on connect : ", err.Error())
	}
	defer handlerSlave2.Close()

	client1 := modbus.NewClient(handlerSlave1)
	client2 := modbus.NewClient(handlerSlave2)
	client3 := modbus.NewClient(handlerSlave3)
	fmt.Println("Reading measurement on universal channel 1, using 2 as quantity")

	codRes := DecodeMessageHoldingRegister(client2, 1, 2)
	tssRes := DecodeMessageHoldingRegister(client2, 3, 2)
	bodRes := DecodeMessageHoldingRegister(client2, 5, 2)
	phRes := DecodeMessageHoldingRegister(client2, 9, 2)
	tempRes := DecodeMessageHoldingRegister(client2, 11, 2)
	tdsRes := DecodeMessageInputRegister(client1, 1, 2)
	turbidityRes := DecodeMessageInputRegister(client1, 3, 2)
	doRes := DecodeMessageInputRegister(client3, 1, 2)

	fmt.Println("COD Result : ", codRes)
	fmt.Println("TSS Result : ", tssRes)
	fmt.Println("BOD Result : ", bodRes)
	fmt.Println("pH Result : ", phRes)
	fmt.Println("Temperature Result : ", tempRes)
	fmt.Println("TDS Result : ", tdsRes)
	fmt.Println("Turbidity Result : ", turbidityRes)
	fmt.Println("DO Result : ", doRes)

}

func DecodeMessageHoldingRegister(client modbus.Client, addr, quantity uint16) float32 {
	var finalRes float32

	read, err := client.ReadHoldingRegisters(addr, quantity)
	if err != nil {
		fmt.Printf("Error reading reading address %d with quantity %d. Overriding Value to 0", addr, quantity)
		finalRes = 0
		return finalRes
	}

	buf := bytes.NewReader(read)
	if err = binary.Read(buf, binary.BigEndian, &finalRes); err != nil {
		fmt.Printf("Error reading reading address %d with quantity %d. Overriding Value to 0", addr, quantity)
		finalRes = 0
		return finalRes
	}

	return finalRes
}

func DecodeMessageInputRegister(client modbus.Client, addr, quantity uint16) float32 {
	var finalRes float32

	read, err := client.ReadInputRegisters(addr, quantity)
	if err != nil {
		fmt.Printf("Error reading reading address %d with quantity %d. Overriding Value to 0", addr, quantity)
		finalRes = 0
		return finalRes
	}

	fmt.Println("Bytes before word swapping : ", read)

	temp := read[0:2]
	temp2 := read[2:4]
	wordSwap := append(temp2, temp...)

	fmt.Println("Bytes after word swapping : ", wordSwap)

	buf := bytes.NewReader(wordSwap)
	if err = binary.Read(buf, binary.BigEndian, &finalRes); err != nil {
		fmt.Printf("Error reading reading address %d with quantity %d. Overriding Value to 0", addr, quantity)
		finalRes = 0
		return finalRes
	}

	return finalRes
}
