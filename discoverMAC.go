package main

import (
	"fmt"
	"log"
	"os"

	"github.com/paypal/gatt"
	"github.com/paypal/gatt/examples/option"
)

var deviceName string


func convertToReverseHex(bleMacAddress string) string {
	// This functions converts MAC Address to reverse order in hex format.
	// Useful to change Raspberry PI mac address.
	// E.g. "AB:CD:EF" converts to "0xEF 0xCD 0xAB" 
	var hexFormat string
	var tmpHex string
	count := 0
	lengthAddr := len(bleMacAddress)
	for i, ch := range bleMacAddress {
		LastCase:
		if count >= 2 {
			hexFormat = "0x" + tmpHex + " " + hexFormat
			tmpHex = ""
			count = 0
		} else {
			tmpHex = tmpHex + string(ch)
			if (i >= lengthAddr - 1) {
				count = 2
				goto LastCase
			} else {
				count = count + 1
			}
		}
	}
	return hexFormat
}

func createFile(bleMacAddress string) {
	fileName := "macSpoof.sh"
	f, err := os.Create(fileName)

	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()

	hexFormat := convertToReverseHex(bleMacAddress)
	
	_, err2 := f.WriteString("#!/bin/bash\necho \"This will change RPI bluetooth MAC address. (Restart RPI to reset)\"\nsudo hcitool cmd 0x04 0x009\nsudo hcitool cmd 0x3f 0x001 " + hexFormat)

	if err2 != nil {
		log.Fatal(err2)
	}

	fmt.Println("| A shell script called '" + fileName + "' has been created. Run its commands to change the RPI bluetooth MAC address")
}

func onStateChanged(d gatt.Device, s gatt.State) {
	fmt.Println("State:", s)
	switch s {
	case gatt.StatePoweredOn:
		fmt.Println("scanning...")
		d.Scan([]gatt.UUID{}, false)
		return
	default:
		d.StopScanning()
	}
}

func onPeriphDiscovered(p gatt.Peripheral, a *gatt.Advertisement, rssi int) {
	if a.LocalName == deviceName {
		createFile(p.ID())
		fmt.Println("Dongle that was discovered: ")
		fmt.Printf("Peripheral ID:%s, NAME:(%s)\n", p.ID(), p.Name())
		fmt.Println("RSSI              =", rssi)
		fmt.Println("Local Name        =", a.LocalName)
		fmt.Println("TX Power Level    =", a.TxPowerLevel)
		fmt.Println("Manufacturer Data =", a.ManufacturerData)
		fmt.Println("Service Data      =", a.ServiceData)
		fmt.Println("> ... CTRL + C to end program.")
	}
}

func main() {
	d, err := gatt.NewDevice(option.DefaultClientOptions...)
	if err != nil {
		log.Fatalf("Failed to open device, err: %s\n", err)
		return
	}
	// Set name of bluetooth device you want to find
	deviceName = "VHM-ble"

	// Register handlers.
	d.Handle(gatt.PeripheralDiscovered(onPeriphDiscovered))
	d.Init(onStateChanged)
	select {}

}
