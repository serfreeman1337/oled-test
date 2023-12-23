package main

import (
	"fmt"
	"time"

	"github.com/serfreeman1337/go-ch347"
	"github.com/sstallion/go-hid"
)

type HIDWithTimeout struct {
	*hid.Device
}

// Read overrided with ReadWithTimeout and with "Interrupted system call" error handling.
func (d *HIDWithTimeout) Read(p []byte) (n int, err error) {
	for {
		n, err = d.Device.ReadWithTimeout(p, 1*time.Second)
		if err == nil || err.Error() != "Interrupted system call" {
			return
		}
	}
}

// DevPath returns CH347 hidraw path.
//
// Allowed ifaces:
//   - 0 - UART
//   - 1 - SPI+I2C+GPIO
func DevPath(iface int) string {
	var devPath string

	// Don't forget to allow access to hidraw:
	// sudo chmod 777 /dev/hidraw{5,6}
	// hidraw numbers can be checked with the `dmesg` command.

	// Locate HID device.
	// ID 1a86:55dc QinHeng Electronics
	var devInfos []*hid.DeviceInfo
	hid.Enumerate(0x1a86, 0x55dc, func(info *hid.DeviceInfo) error {
		devInfos = append(devInfos, info)
		return nil
	})

	for _, di := range devInfos {
		// InterfaceNbr == 0 - UART
		// InterfaceNbr == 1 - SPI+I2C+GPIO
		if di.ProductStr == "HID To UART+SPI+I2C" && di.InterfaceNbr == iface {
			devPath = di.Path
			break
		}
	}

	return devPath
}

func FindCH347IO() (*ch347.IO, error) {
	devPath := DevPath(1)
	if len(devPath) == 0 {
		return nil, fmt.Errorf("ch347 not found")
	}

	dev, err := hid.OpenPath(devPath)
	if err != nil {
		return nil, err
	}

	return &ch347.IO{Dev: &HIDWithTimeout{dev}}, nil
}
