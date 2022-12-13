package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/goburrow/modbus"
)

type StationData struct {
	ID_Stasiun string  `json:"IDStasiun"`
	Tanggal    string  `json:"Tanggal"`
	Jam        string  `json:"Jam"`
	Suhu       float32 `json:"Suhu"`
	TDS        float32 `json:"TDS"`
	DO         float32 `json:"DO"`
	PH         float32 `json:"PH"`
	Turbidity  float32 `json:"Turbidity"`
	Kedalaman  float32 `json:"Kedalaman"`
	Nitrat     float32 `json:"Nitrat"`
	Amonia     float32 `json:"Amonia"`
	COD        float32 `json:"COD"`
	BOD        float32 `json:"BOD"`
	TSS        float32 `json:"TSS"`
}

type Payload struct {
	Data      StationData `json:"Data"`
	ApiKey    string      `json:"apikey"`
	ApiSecret string      `json:"apisecret"`
}

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
	fmt.Println("Reading measurement.... ")

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

	current_time := time.Now()
	year, month, day := current_time.Date()
	current_date := fmt.Sprintf("%d-%d-%d", year, int(month), day)
	//current_date := current_time.Format("2022-01-01")
	hour, minute, second := current_time.Clock()
	current_clock := fmt.Sprintf("%d:%d:%d", hour, minute, second)
	//current_clock := current_time.Format("15:01:01")

	//url_api := "https://ppkl.menlhk.go.id/onlimo/uji/connect/uji_data_onlimo"
	id_station := "indosense"
	apikey := "uji@forbesmarshallindonesia"
	apisecret := "4ede9dea-b352-4815-823f-8525c1563663"

	data := &StationData{
		ID_Stasiun: id_station,
		Tanggal:    current_date,
		Jam:        current_clock,
		Suhu:       tempRes,
		TDS:        tdsRes,
		DO:         doRes,
		PH:         phRes,
		Turbidity:  turbidityRes,
		Kedalaman:  0,
		Nitrat:     0,
		Amonia:     0,
		COD:        codRes,
		BOD:        bodRes,
		TSS:        tssRes,
	}

	payload := &Payload{
		Data:      *data,
		ApiKey:    apikey,
		ApiSecret: apisecret,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		fmt.Println("Error occured on marshalling payload, ", err.Error())
	}
	fmt.Println(string(jsonPayload))

	// req, err := http.NewRequest("POST", url_api, bytes.NewBuffer(jsonPayload))
	// if err != nil {
	// 	fmt.Println("Error on creating the request, ", err.Error())
	// }
	// req.Header.Set("Content-Type", "application/json; charset=UTF-8")

	// httpClient := &http.Client{}
	// httpRespose, err := httpClient.Do(req)
	// if err != nil {
	// 	fmt.Println("Error on sending payload, ", err.Error())
	// }
	// defer httpRespose.Body.Close()

	// fmt.Println("Response status : ", httpRespose.Status)
	// fmt.Println("Response header : ", httpRespose.Header)
	// body, _ := ioutil.ReadAll(httpRespose.Body)
	// fmt.Println("Response body : ", string(body))

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

	// fmt.Println("Bytes before word swapping : ", read)

	temp := read[0:2]
	temp2 := read[2:4]
	wordSwap := append(temp2, temp...)

	// fmt.Println("Bytes after word swapping : ", wordSwap)

	buf := bytes.NewReader(wordSwap)
	if err = binary.Read(buf, binary.BigEndian, &finalRes); err != nil {
		fmt.Printf("Error reading reading address %d with quantity %d. Overriding Value to 0", addr, quantity)
		finalRes = 0
		return finalRes
	}

	return finalRes
}
