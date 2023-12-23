package main

import (
	"flag"
	"fmt"
	"image"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/serfreeman1337/go-ch347"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

const (
	luxFrom = "20:30:00" // Enable lux check from this time.
	luxTo   = "07:30:00" // Disable lux check after this time.
	luxMin  = 0.15       // Min lux to display to be on.
)

func main() {
	var ifname string
	flag.StringVar(&ifname, "i", "", "interface name")
	flag.Parse()

	if ifname == "" {
		flag.Usage()
		return
	}

	c, err := FindCH347IO()
	if err != nil {
		panic(err)
	}

	// Configure I2C.
	err = c.SetI2C(ch347.I2CMode3)
	if err != nil {
		panic(err)
	}

	// Configure SPI.
	err = c.SetSPI(ch347.SPIMode0, ch347.SPIClock0, ch347.SPIByteOrderMSB)
	if err != nil {
		panic(err)
	}

	// Turn off ACT led.
	c.WritePin(ch347.GPIO4, true, true)

	// Configure SSD1306 display.
	s := NewSSD1306(c)
	err = s.Init()
	if err != nil {
		panic(err)
	}

	ns, err := NewNetifStats(ifname)
	if err != nil {
		panic(err)
	}
	defer ns.Close()

	// Load fonts.
	// ttf, err := os.Open("/usr/share/fonts/noto/NotoSansMono-Regular.ttf")
	ttf, err := os.Open("/usr/share/fonts/TTF/JetBrainsMono-Regular.ttf")
	if err != nil {
		panic(err)
	}
	defer ttf.Close()

	f, _ := opentype.ParseReaderAt(ttf)
	face, _ := opentype.NewFace(f, &opentype.FaceOptions{
		Size:    13,
		DPI:     96,
		Hinting: font.HintingFull,
	})
	faceSmall, _ := opentype.NewFace(f, &opentype.FaceOptions{
		Size:    9,
		DPI:     96,
		Hinting: font.HintingFull,
	})

	// Setup image drawer for SSD1306.
	img := image.NewGray(image.Rect(0, 0, 128, 64))
	d := font.Drawer{
		Dst: img,
		Src: image.White,
	}

	alignLeft := func(text string) {
		_, advance := d.BoundString(text)
		d.Dot.X = fixed.I(img.Rect.Max.X) - advance
	}

	var text string
	textSpacing := fixed.I(6)

	ifstats := IfStats{}

	// For human readable representation of an IEC size.
	var (
		div uint64
		exp int
		v   float64
	)
	mm := "KMGTPE"
	byteCountIEC := func(b uint64) {
		const unit uint64 = 1024

		div, exp = unit, 0
		for n := b / unit; n >= unit; n /= unit {
			div *= unit
			exp++
		}
	}

	// Read ifstats at first to setup inital previous bytes count needed for rate calculation.
	ns.Read(&ifstats)
	prevRx, prevTx := ifstats.RxBytes, ifstats.TxBytes
	var rate uint64

	// SHT45 readings.
	sht45 := &SHT4X{c: c}
	var t, rh float32

	// TSL2591 readings
	tsl := &TSL2519{c: c}
	tsl.SetGainIntegration(TSL2591GainHigh, TSL2591IntegrationTime100ms)
	var (
		luxTime time.Time
		withLux bool
		lux     float32
	)

	show := true

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	tick := time.NewTicker(1 * time.Second)

	for {
		select {
		case <-stop: // Turn off display on Ctrl+C.
			s.Off()
			fmt.Println("")
			return
		case <-tick.C:
		}

		err = ns.Read(&ifstats)
		if err != nil {
			panic(err)
		}

		if show {
			// Draw main text with a regular font.
			d.Face = face

			// RX Speed
			rate = ifstats.RxBytes - prevRx
			byteCountIEC(rate)

			// Yep, draw arrow on the left.
			d.Dot.X, d.Dot.Y = 0, d.Face.Metrics().CapHeight
			text = "↓"
			d.DrawString(text)

			// Draw text aligned to the right.
			d.Dot.X = 0
			v = float64(rate) / float64(div)
			text = fmt.Sprintf("%.1f %ciB/s", v, mm[exp])
			alignLeft(text)
			d.DrawString(text)

			// TX Speed
			rate = ifstats.TxBytes - prevTx
			byteCountIEC(rate)

			d.Dot.X, d.Dot.Y = 0, d.Dot.Y+d.Face.Metrics().CapHeight+textSpacing
			text = "↑"
			d.DrawString(text)

			d.Dot.X = 0
			v = float64(rate) / float64(div)
			text = fmt.Sprintf("%.1f %ciB/s", v, mm[exp])
			alignLeft(text)
			d.DrawString(text)

			// Draw support text with a small font.
			d.Face = faceSmall

			// RX Total
			byteCountIEC(ifstats.RxBytes)

			d.Dot.X, d.Dot.Y = 0, d.Dot.Y+d.Face.Metrics().CapHeight+textSpacing
			v = float64(ifstats.RxBytes) / float64(div)
			if v < 100.0 {
				text = "%.3f"
			} else {
				text = "%.2f"
			}
			text = fmt.Sprintf("↓"+text+" %ciB", v, mm[exp])
			alignLeft(text)
			d.DrawString(text)

			// TX Total
			byteCountIEC(ifstats.TxBytes)

			d.Dot.X, d.Dot.Y = 0, d.Dot.Y+d.Face.Metrics().CapHeight+textSpacing
			v = float64(ifstats.TxBytes) / float64(div)
			if v < 100.0 {
				text = "%.3f"
			} else {
				text = "%.2f"
			}
			text = fmt.Sprintf("↑"+text+" %ciB", v, mm[exp])
			alignLeft(text)
			d.DrawString(text)

			// SHT45 readings in the bottom left corner.
			t, rh, err = sht45.Measure()
			if err == nil {
				d.Dot.X = 0
				text = fmt.Sprintf("%.01f°C", t)
				d.DrawString(text)

				d.Dot.X, d.Dot.Y = 0, d.Dot.Y-d.Face.Metrics().CapHeight-textSpacing
				text = fmt.Sprintf("%.01f%%", rh)
				d.DrawString(text)
			} else {
				fmt.Println(time.Now(), "sht45 err", err)
			}

			// Now display.
			_, err = s.Write(img.Pix)
			if err != nil {
				fmt.Println(time.Now(), "display err", err)
			}

			// And clear the image.
			clear(img.Pix)
		}

		prevRx, prevTx = ifstats.RxBytes, ifstats.TxBytes

		if time.Now().After(luxTime) {
			if withLux && !show { // Turning off lux check, so bring display back on.
				show = true
				s.On()
			}

			luxTime, withLux = NextLuxTime(luxFrom, luxTo)
		}

		if withLux {
			lux, err = tsl.Lux()
			if err == nil {
				if lux < luxMin {
					if show {
						show = false
						s.Off()
					}
				} else {
					if !show {
						show = true
						s.On()
					}
				}
			} else {
				fmt.Println(time.Now(), "lux err", err)
			}
		}

		// Let's blink the hard way.
		// ok, err = c.ReadPin(ch347.GPIO4)
		// if err == nil {
		// 	err = c.WritePin(ch347.GPIO4, true, !ok)
		// 	if err != nil {
		// 		fmt.Println(time.Now(), "pin write err", err)
		// 	}
		// } else {
		// 	fmt.Println(time.Now(), "pin read err", err)
		// }
	}
}

func NextLuxTime(from, to string) (time.Time, bool) {
	now := time.Now()

	// Time from.
	p := strings.Split(from, ":")
	hh, _ := strconv.Atoi(p[0])
	mm, _ := strconv.Atoi(p[1])
	ss, _ := strconv.Atoi(p[2])
	timeFrom := time.Date(now.Year(), now.Month(), now.Day(), hh, mm, ss, now.Nanosecond(), now.Location())

	// Time to.
	p = strings.Split(to, ":")
	hh, _ = strconv.Atoi(p[0])
	mm, _ = strconv.Atoi(p[1])
	ss, _ = strconv.Atoi(p[2])
	timeTo := time.Date(now.Year(), now.Month(), now.Day(), hh, mm, ss, now.Nanosecond(), now.Location())

	if now.Before(timeTo) {
		// Lux before timeTo.
		return timeTo, true
	} else if now.Before(timeFrom) {
		// Don't lux until timeFrom.
		return timeFrom, false
	} else {
		// Lux until next day timeTo.
		return timeTo.Add(24 * time.Hour), true
	}
}
