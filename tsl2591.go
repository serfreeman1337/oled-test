package main

import (
	"time"

	"github.com/serfreeman1337/go-ch347"
)

const (
	TSL2591CmdBit     byte = 0xa0
	TSL2591RegControl byte = 0x01
	TSL2591RegCH0Low  byte = 0x14
)

type TSL2519Gain uint8

const (
	TSL2591GainLow  TSL2519Gain = 0x00 // Low gain (1x).
	TSL2591GainMed  TSL2519Gain = 0x10 // Medium gain (25x).
	TSL2591GainHigh TSL2519Gain = 0x20 // High gain (428x).
	TSL2591GainMax  TSL2519Gain = 0x30 // Max gain (9876x).
)

type TSL2591IntegrationTime uint8

const (
	TSL2591IntegrationTime100ms TSL2591IntegrationTime = iota
	TSL2591IntegrationTime200ms
	TSL2591IntegrationTime300ms
	TSL2591IntegrationTime400ms
	TSL2591IntegrationTime500ms
	TSL2591IntegrationTime600ms
)

type TSL2519 struct {
	c     *ch347.IO
	gain  TSL2519Gain
	atime TSL2591IntegrationTime
}

func (t *TSL2519) Lux() (lux float32, err error) {
	var atime, again, cpl float32

	switch t.gain {
	case TSL2591GainLow:
		again = 1.0
	case TSL2591GainMed:
		again = 25.0
	case TSL2591GainHigh:
		again = 428.0
	case TSL2591GainMax:
		again = 9876.0
	}

	atime = float32(t.atime+1) * 100.0

	cpl = (atime * again) / 408.0 // TSL2591_LUX_DF

	err = t.enable()
	if err != nil {
		return
	}

	time.Sleep(time.Duration(atime)*time.Millisecond + 20*time.Millisecond)

	r := make([]byte, 4)
	err = t.c.I2C(0x29, []byte{TSL2591CmdBit | TSL2591RegCH0Low}, r)
	if err != nil {
		return
	}

	err = t.disable()

	ch0 := (int(r[1]) << 8) | int(r[0])
	ch1 := (int(r[3]) << 8) | int(r[2])

	if ch0 == 0 {
		return
	}

	lux = (float32(ch0) - float32(ch1)) * (1.0 - (float32(ch1) / float32(ch0))) / cpl
	return
}

func (t *TSL2519) SetGainIntegration(gain TSL2519Gain, atime TSL2591IntegrationTime) error {
	err := t.c.I2C(0x29, []byte{TSL2591CmdBit | TSL2591RegControl, byte(gain) | byte(atime)}, nil)
	if err != nil {
		return err
	}
	t.gain, t.atime = gain, atime
	return nil
}

func (t *TSL2519) enable() error {
	return t.c.I2C(0x29, []byte{TSL2591CmdBit, 0x03}, nil)
}

func (t *TSL2519) disable() error {
	return t.c.I2C(0x29, []byte{TSL2591CmdBit, 0x00}, nil)
}
