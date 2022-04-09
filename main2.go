// +build

package main

import (
	"fmt"
	"log"

	"github.com/paypal/gatt"
	"github.com/paypal/gatt/examples/option"
	"github.com/paypal/gatt/examples/service"
)

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
			log.Println("Wrote:", string(data))
			fmt.Fprintf(x, "13.V\r>")
			var data2 = ([]byte("13.5V\r>"))
			x.Write(data2)
			return gatt.StatusSuccess
		})
	return s
}

func main() {
	d, err := gatt.NewDevice(option.DefaultServerOptions...)
	if err != nil {
		log.Fatalf("Failed to open device, err: %s", err)
	}

	// Register optional handlers.
	d.Handle(
		gatt.CentralConnected(func(c gatt.Central) { fmt.Println("Connect: ", c.ID()) }),
		gatt.CentralDisconnected(func(c gatt.Central) { fmt.Println("Disconnect: ", c.ID()) }),
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

			fmt.Println("HEHRHERHEHREHRHERHHREHRHE")

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

	d.Init(onStateChanged)
	select {}
}
