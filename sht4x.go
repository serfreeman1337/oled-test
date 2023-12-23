package main

import (
	"time"

	"github.com/serfreeman1337/go-ch347"
)

type SHT4X struct {
	c *ch347.IO
}

func (s *SHT4X) Measure() (t float32, rh float32, err error) {
	r := make([]byte, 6)
	r[0] = 0xfd // High accurracy.

	// Send measure command.
	err = s.c.I2C(0x44, r[:1], nil)
	if err != nil {
		return
	}

	// Wait 10ms for measurements.
	time.Sleep(10 * time.Millisecond)

	// Read measurements data.
	err = s.c.I2C(0x44, nil, r)
	if err != nil {
		return
	}

	// Perform conversion.
	t = -45 + 175*float32(((int(r[0])<<8)|int(r[1])))/65535
	rh = -6 + 125*float32((int(r[3])<<8)|int(r[4]))/65535
	return
}
