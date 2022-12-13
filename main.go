package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/goburrow/modbus"
)

func main() {
	// Modbus TCP
	handler := modbus.NewTCPClientHandler("localhost:5000")
	handler.Timeout = 10 * time.Second
	handler.Logger = log.New(os.Stdout, "test: ", log.LstdFlags)
	// Connect manually so that multiple requests are handled in one connection session
	err := handler.Connect()
	if err != nil {
		log.Panic("Error on connect : ", err.Error())
	}
	defer handler.Close()

	client := modbus.NewClient(handler)
	fmt.Println("Reading measurement on universal channel 1, using 3 as quantity")
	results, err := client.ReadHoldingRegisters(10, 5)
	if err != nil {
		log.Panic("error querying to address 202 : ", err.Error())
	}

	// fmt.Println("Reading measurement on universal channel 1 on each register 200, 201 and 202")
	// reg200, _ := client.ReadHoldingRegisters(200, 1)
	// reg201, _ := client.ReadHoldingRegisters(201, 1)
	// reg202, _ := client.ReadHoldingRegisters(202, 1)

	// results2, err := client.ReadHoldingRegisters(203, 3)
	// if err != nil {
	// 	log.Panic("error querying to address 203 : ", err.Error())
	// }

	fmt.Println("results value address 200 as universal channel 1 : ", results)
	// fmt.Println("results 2 value address 203 : ", results2)
	// fmt.Println("results for address 200 : ", reg200)
	// fmt.Println("results for address 201 : ", reg201)
	// fmt.Println("results for address 200 : ", reg202)

	val := results[2:]

	fmt.Println("result for getting value bytes : ", val)
	var result float64
	buf := bytes.NewReader(val)
	err = binary.Read(buf, binary.BigEndian, &result)
	if err != nil {
		log.Panic("error converting byte to float32, ", err.Error())
	}

	fmt.Println("Result reading address 200 : ", result)

}
