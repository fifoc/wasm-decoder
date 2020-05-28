package main

import (
	"fmt"
	"math"
	"strings"
	"syscall/js"
	"time"
)

func main() {
	c := make(chan bool)
	js.Global().Set("renderFIF", js.FuncOf(callRenderFIF))
	<-c
}

type ByteReadr struct {
	offset int
	bytes  []byte
}

func (r *ByteReadr) readByte() byte {
	a := r.bytes[r.offset]
	r.offset = r.offset + 1
	return a
}
func (r *ByteReadr) readString(len int) string {
	a := r.bytes[r.offset : r.offset+len]
	r.offset = r.offset + len
	return string(a)
}
func (r *ByteReadr) readBytes(len int) []byte {
	a := r.bytes[r.offset : r.offset+len]
	r.offset = r.offset + len
	return a
}

func itob(b byte, fg int, bg int) int {
	if b > 0 {
		return fg
	}
	return bg
}

func drawBraille(c js.Value, braille byte, x int, y int, fg int, bg int) {
	var r1 int = itob(braille&0x08, fg, bg)
	var r2 int = itob(braille&0x10, fg, bg)
	var r3 int = itob(braille&0x20, fg, bg)
	var r4 int = itob(braille&0x80, fg, bg)

	var l1 int = itob(braille&0x01, fg, bg)
	var l2 int = itob(braille&0x02, fg, bg)
	var l3 int = itob(braille&0x04, fg, bg)
	var l4 int = itob(braille&0x40, fg, bg)

	draw(c, x, y-3, l1)
	draw(c, x+1, y-3, r1)

	draw(c, x, y-2, l2)
	draw(c, x+1, y-2, r2)

	draw(c, x, y-1, l3)
	draw(c, x+1, y-1, r3)

	draw(c, x, y, l4)
	draw(c, x+1, y, r4)
}

func draw(c js.Value, x int, y int, color int) {
	/*r := color & 0x0000FF
	g := color & 0x00FF
	b := color & 0xFF

	c.Set("fillStyle", "rgb(" + strconv.Itoa(r) + "," + strconv.Itoa(g) + "," + strconv.Itoa(b) +")")*/

	c.Set("fillStyle", "#"+fmt.Sprintf("%06x", color))
	c.Call("fillRect", x, y, 1, 1)
}
func callRenderFIF(this js.Value, inputs []js.Value) interface{} {
	go renderFIF(this, inputs)
	return nil
}
func renderFIF(this js.Value, inputs []js.Value) interface{} {
	if len(inputs) >= 4 {
		js.Global().Get("document").Call("getElementById", inputs[3].String()).Set("innerText", "RENDERING...")
	}
	delayEnabled := false
	delay := time.Millisecond
	if len(inputs) >= 5 {
		delayEnabled = true
		delay = time.Duration(inputs[4].Int()) * time.Millisecond
	}


	canvasV := inputs[0].String()
	var cv js.Value = js.
		Global().
		Get("document").
		Call("getElementById", canvasV)

	length := inputs[2].Int()
	var aa []byte = make([]byte, length)

	nowe := time.Now()
	js.CopyBytesToGo(aa, inputs[1])
	nowg := time.Now().UnixNano() - nowe.UnixNano()

	now := time.Now()
	br := ByteReadr{
		offset: 0,
		bytes:  aa,
	}
	var fastIF = br.readString(6)
	println(fastIF)
	println(strings.Compare(fastIF, "FastIF"))
	if strings.Compare(fastIF, "FastIF") != 0 {
		panic("Not FIF")
	}

	var w int = int(br.readByte()) * 2
	var h int = int(br.readByte()) * 4

	if w > 320 {
		println("Width exceeds RPH-set standars")
	}
	if h > 200 {
		println("Width exceeds RPH-set standars")
	}

	cv.Set("width", w)
	cv.Set("height", h)
	var canvas js.Value = cv.Call("getContext", "2d")
	run := true
	__len := len(br.bytes)

	background := 0
	foreground := 0

	for run {
		cmd := br.readByte()

		switch cmd {
		case 0x01:
			{
				r := br.readByte()
				g := br.readByte()
				b := br.readByte()
				background = int(b) + (int(g) << 8) + (int(r) << 16)

			}
		case 0x02:
			{
				r := br.readByte()
				g := br.readByte()
				b := br.readByte()
				foreground = int(b) + (int(g) << 8) + (int(r) << 16)
			}
		case 0x10:
			{
				x := br.readByte()
				y := br.readByte()
				size := br.readByte()

				data := br.readBytes(int(size))
				for i := 0; i < int(size); i++ {
					b := data[i]
					drawBraille(canvas, b, int(x)*2+(i*2), int(y)*4, foreground, background)
				}

			}
		case 0x11:
			{
				x := br.readByte()
				y := br.readByte()

				w := br.readByte()
				h := br.readByte()

				lower := br.readByte()

				for xr := 0; xr < int(w); xr++ {
					for yr := 0; yr < int(h); yr++ {
						drawBraille(canvas, lower, int(x)*2+xr*2, int(y)*4+yr*4, foreground, background)
					}
				}
			}
		case 0x12:
			{
				br.offset++
			}
		case 0x13:
			{
				x := br.readByte()
				y := br.readByte()
				size := br.readByte()

				data := br.readBytes(int(size))

				for yr := 0; yr < int(size); yr++ {
					a := data[yr]
					drawBraille(canvas, a, int(x)*2, int(y)*4 + yr*4, foreground, background)
				}

			}
		case 0x20:
			{
				run = false
			}
		}
		if br.offset > __len+1 {
			println("Non-compliant encoder was used. No 0x20 at the end.")
			break
		}
		if delayEnabled {
			time.Sleep(delay)
		}
	}

	taken := now.UnixNano() - time.Now().UnixNano()
	if len(inputs) >= 4 {
		js.Global().Get("document").Call("getElementById", inputs[3].String()).Set("innerText", fmt.Sprintf("Render time: %.2fms, Copy time: %.2fms", math.Abs(float64(taken))/1000000, math.Abs(float64(nowg))/1000000)) // DONT QUESTION THE A B S
	}

	if !js.Global().Get("finished").IsNull() {
		js.Global().Call("finished")
	}

	return nil
}
