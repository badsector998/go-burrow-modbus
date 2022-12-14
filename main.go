package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
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

	handlerSlave1 := modbus.NewRTUClientHandler("COM3")
	handlerSlave1.BaudRate = 9600
	handlerSlave1.DataBits = 8
	handlerSlave1.Parity = "N"
	handlerSlave1.StopBits = 1
	handlerSlave1.SlaveId = 1
	handlerSlave1.Timeout = 10 * time.Second

	handlerSlave2 := modbus.NewRTUClientHandler("COM3")
	handlerSlave2.SlaveId = 2
	handlerSlave2.BaudRate = 9600
	handlerSlave2.DataBits = 8
	handlerSlave2.Parity = "N"
	handlerSlave2.StopBits = 1
	handlerSlave2.Timeout = 10 * time.Second

	handlerSlave3 := modbus.NewRTUClientHandler("COM3")
	handlerSlave3.SlaveId = 3
	handlerSlave2.BaudRate = 9600
	handlerSlave2.DataBits = 8
	handlerSlave2.Parity = "N"
	handlerSlave2.StopBits = 1
	handlerSlave3.Timeout = 10 * time.Second

	// handler.Logger = log.New(os.Stdout, "test: ", log.LstdFlags)
	// Connect manually so that multiple requests are handled in one connection session

	fmt.Println("Measurement Ready.... ")
	loc, _ := time.LoadLocation("Asia/Jakarta")

	for {
		err := handlerSlave2.Connect()
		if err != nil {
			log.Panic("Error on connect 2: ", err.Error())
		}
		client2 := modbus.NewClient(handlerSlave2)
		codRes := DecodeMessageHoldingRegister(client2, 1, 2)
		tssRes := DecodeMessageHoldingRegister(client2, 3, 2)
		bodRes := DecodeMessageHoldingRegister(client2, 5, 2)
		phRes := DecodeMessageHoldingRegister(client2, 9, 2)
		tempRes := DecodeMessageHoldingRegister(client2, 11, 2)
		handlerSlave2.Close()
		time.Sleep(1 * time.Second)

		err = handlerSlave1.Connect()
		if err != nil {
			log.Panic("Error on connect 1: ", err.Error())
		}
		client1 := modbus.NewClient(handlerSlave1)
		tdsRes := DecodeMessageInputRegister(client1, 1, 2)
		turbidityRes := DecodeMessageInputRegister(client1, 3, 2)
		handlerSlave1.Close()
		time.Sleep(1 * time.Second)

		// handlerSlave1.SlaveId = 3
		// err = handlerSlave1.Connect()
		// if err != nil {
		// 	log.Panic("Error on connect 3: ", err.Error())
		// }
		// doRes := DecodeMessageInputRegister(client1, 1, 2)
		// handlerSlave1.Close()
		// time.Sleep(1 * time.Second)

		err = handlerSlave3.Connect()
		if err != nil {
			log.Panic("Error on connect 3: ", err.Error())
		}
		client3 := modbus.NewClient(handlerSlave3)
		doRes := DecodeMessageInputRegister(client3, 1, 2)
		handlerSlave3.Close()
		time.Sleep(1 * time.Second)

		fmt.Println("COD Result : ", codRes)
		fmt.Println("TSS Result : ", tssRes)
		fmt.Println("BOD Result : ", bodRes)
		fmt.Println("pH Result : ", phRes)
		fmt.Println("Temperature Result : ", tempRes)
		fmt.Println("TDS Result : ", tdsRes)
		fmt.Println("Turbidity Result : ", turbidityRes)
		fmt.Println("DO Result : ", doRes)

		current_time := time.Now().UTC().In(loc)
		// year, month, day := current_time.Date()
		// current_date := fmt.Sprintf("%d-%d-%d", year, int(month), day)
		current_date := current_time.Format("2006-01-02")
		// hour, minute, second := current_time.Clock()
		// current_clock := fmt.Sprintf("%d:%d:%d", hour, minute, second)
		current_clock := current_time.Format("15:04:05")

		url_api := "https://ppkl.menlhk.go.id/onlimo/uji/connect/uji_data_onlimo"
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

		req, err := http.NewRequest("POST", url_api, bytes.NewBuffer(jsonPayload))
		if err != nil {
			fmt.Println("Error on creating the request, ", err.Error())
		}
		req.Header.Set("Content-Type", "application/json; charset=UTF-8")

		httpClient := &http.Client{}
		httpRespose, err := httpClient.Do(req)
		if err != nil {
			fmt.Println("Error on sending payload, ", err.Error())
		}

		fmt.Println("Response status : ", httpRespose.Status)
		fmt.Println("Response header : ", httpRespose.Header)
		body, _ := ioutil.ReadAll(httpRespose.Body)
		fmt.Println("Response body : ", string(body))

		httpRespose.Body.Close()

		time.Sleep(1 * time.Hour)
	}
}

func DecodeMessageHoldingRegister(client modbus.Client, addr, quantity uint16) float32 {
	var finalRes float32

	read, err := client.ReadHoldingRegisters(addr, quantity)
	if err != nil {
		fmt.Printf("Error reading reading address %d with quantity %d. Overriding Value to 0 \n", addr, quantity)
		finalRes = 0
		return finalRes
	}

	buf := bytes.NewReader(read)
	if err = binary.Read(buf, binary.BigEndian, &finalRes); err != nil {
		fmt.Printf("Error reading reading address %d with quantity %d. Overriding Value to 0 \n", addr, quantity)
		finalRes = 0
		return finalRes
	}

	return finalRes
}

func DecodeMessageInputRegister(client modbus.Client, addr, quantity uint16) float32 {
	var finalRes float32

	read, err := client.ReadInputRegisters(addr, quantity)
	if err != nil {
		fmt.Printf("Error reading reading address %d with quantity %d. Overriding Value to 0. Error code: %v \n", addr, quantity, err.Error())
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

func DecodeMessageInputRegisterID3(client modbus.Client, addr, quantity uint16) float32 {
	var finalRes float32

	read, err := client.ReadInputRegisters(addr, quantity)
	if err != nil {
		fmt.Printf("Error reading reading address %d with quantity %d. Overriding Value to 0. Error code: %v \n", addr, quantity, err.Error())
		finalRes = 0
		return finalRes
	}

	fmt.Println("Bytes before word swapping CD AB : ", read)

	wordSwap := make([]byte, 0)
	// Default reading CD AB

	// DC BA
	// wordSwap = append(wordSwap, read[1])
	// wordSwap = append(wordSwap, read[0])
	// wordSwap = append(wordSwap, read[3])
	// wordSwap = append(wordSwap, read[2])

	// AB CD
	// wordSwap = append(wordSwap, read[2])
	// wordSwap = append(wordSwap, read[3])
	// wordSwap = append(wordSwap, read[0])
	// wordSwap = append(wordSwap, read[1])

	// BA DC
	wordSwap = append(wordSwap, read[3])
	wordSwap = append(wordSwap, read[2])
	wordSwap = append(wordSwap, read[1])
	wordSwap = append(wordSwap, read[0])

	// // AB CD
	// temp := read[0:2]
	// temp2 := read[2:4]
	// wordSwap = append(wordSwap, temp2...)
	// wordSwap = append(wordSwap, temp...)

	// fmt.Println("Wordswap AB CD : ", wordSwap)

	// // BA DC
	// temp1 := wordSwap[0]
	// wordSwap[0] = wordSwap[1]
	// wordSwap[1] = temp1

	// temp1 = wordSwap[2]
	// wordSwap[2] = wordSwap[3]
	// wordSwap[3] = temp1

	fmt.Println("Bytes after word swapping : ", wordSwap)

	buf := bytes.NewReader(wordSwap)
	if err = binary.Read(buf, binary.BigEndian, &finalRes); err != nil {
		fmt.Printf("Error reading reading address %d with quantity %d. Overriding Value to 0", addr, quantity)
		finalRes = 0
		return finalRes
	}

	return finalRes
}
