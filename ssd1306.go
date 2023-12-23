package main

import (
	"time"

	"github.com/serfreeman1337/go-ch347"
)

const RST = ch347.GPIO5 // SCS1
const DC = ch347.GPIO1  // MISO

type SSD1306 struct {
	c    *ch347.IO
	buf  []byte
	w, h int
	y, x int
}

func NewSSD1306(c *ch347.IO) *SSD1306 {
	return &SSD1306{
		c:   c,
		buf: make([]byte, 128*8),
		w:   128,
		h:   64,
	}
}

// Init inits SSD1306 128x64 SPI display.
func (s *SSD1306) Init() error {
	// For 128x64.
	mux := byte(64 - 1)
	com_pins := byte(0x12)
	contrast := byte(0x00) // byte(0xff)

	// Trigger RST sequence.
	s.c.WritePin(RST, true, true)
	time.Sleep(1 * time.Millisecond)

	s.c.WritePin(RST, true, false)
	time.Sleep(10 * time.Millisecond)

	s.c.WritePin(RST, true, true)

	// Init sequence.
	w := []byte{
		0xae,       // SSD1306_CMD_DISPLAY_OFF
		0xd5, 0x80, // SSD1306_CMD_SET_DISPLAY_CLK_DIV // follow with 0x80
		0xa8, mux, // SSD1306_CMD_SET_MUX_RATIO //  follow with 0x3F = 64 MUX
		0xd3, 0x00, // SSD1306_CMD_SET_DISPLAY_OFFSET // // follow with 0x00
		0x40,       // SSD1306_CMD_SET_DISPLAY_START_LINE
		0x8D, 0x14, // SSD1306_CMD_SET_CHARGE_PUMP // follow with 0x14
		0x20, 0x00, // SSD1306_CMD_SET_MEMORY_ADDR_MODE // SSD1306_CMD_SET_HORI_ADDR_MODE
		0xa1,           // SSD1306_CMD_SET_SEGMENT_REMAP_1
		0xc8,           // SSD1306_CMD_SET_COM_SCAN_MODE
		0xda, com_pins, // SSD1306_CMD_SET_COM_PIN_MAP
		0x81, contrast, // SSD1306_CMD_SET_CONTRAST
		0xd9, 0xf1, // SSD1306_CMD_SET_PRECHARGE // follow with 0xF1
		0xd8, 0x40, // SSD1306_CMD_SET_VCOMH_DESELCT
		0xa4, // SSD1306_CMD_DISPLAY_RAM
		0xa6, // SSD1306_CMD_DISPLAY_NORMAL
		0xaf, // SSD1306_CMD_DISPLAY_ON

		//
		0x21, 0x00, 0x7f, // SSD1306_CMD_SET_COLUMN_RANGE // follow with 0x00 and 0x7F = COL127
		0x22, 0x00, 0x07, // SSD1306_CMD_SET_PAGE_RANGE // follow with 0x00 and 0x07 = PAGE7
	}

	err := s.writeCMD(w)
	if err != nil {
		return err
	}

	return nil
}

// Write performs conversion to SSD1306 format and displays buffer every 8192 bytes written.
func (s *SSD1306) Write(p []byte) (int, error) {
	var page, pageRow, pageCol int

	for _, a := range p {
		page = s.y / 8
		pageRow = s.y % 8
		pageCol = s.x

		// Set pixel bit.
		if a > 64 { // True. Threshold value. Set pixel bit if intensity of that pixel is greater than 127.
			s.buf[page*128+pageCol] |= (1 << pageRow)
		} else { // False.
			s.buf[page*128+pageCol] &= ^(1 << pageRow)
		}

		s.x++
		if s.x > 127 {
			s.x = 0
			s.y++
			if s.y > 63 {
				s.y = 0
				err := s.Display()

				if err != nil {
					return 0, err
				}
			}
		}
	}

	return len(p), nil
}

func (s *SSD1306) Display() error {
	s.c.SetCS(true)
	err := s.c.SPI(s.buf, nil)
	s.c.SetCS(false)

	return err
}

func (s *SSD1306) On() error {
	return s.writeCMD([]byte{0xaf})
}

func (s *SSD1306) Off() error {
	return s.writeCMD([]byte{0xae})
}

func (s *SSD1306) writeCMD(p []byte) error {
	s.c.WritePin(DC, true, false) // Switch to cmd mode.

	s.c.SetCS(true)
	err := s.c.SPI(p, nil)
	s.c.SetCS(false)

	s.c.WritePin(DC, true, true) // Switch back to data mode.

	return err
}
