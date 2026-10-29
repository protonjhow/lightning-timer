package main

import (
	"flag"
	"fmt"
	"math"
	"os"
	"strconv"
	"time"

	"github.com/gdamore/tcell/v2"
)

// Segment bits: a=top, b=top-right, c=bottom-right, d=bottom, e=bottom-left, f=top-left, g=middle.
const (
	segA = 1 << iota
	segB
	segC
	segD
	segE
	segF
	segG
)

var digitSegments = [10]int{
	segA | segB | segC | segD | segE | segF,
	segB | segC,
	segA | segB | segD | segE | segG,
	segA | segB | segC | segD | segG,
	segB | segC | segF | segG,
	segA | segC | segD | segF | segG,
	segA | segC | segD | segE | segF | segG,
	segA | segB | segC,
	segA | segB | segC | segD | segE | segF | segG,
	segA | segB | segC | segD | segF | segG,
}

func usage(msg string) {
	if msg != "" {
		fmt.Fprintln(os.Stderr, "error:", msg)
	}
	fmt.Fprintln(os.Stderr, "usage: talk-timer -total-time MINUTES -warning-time MINUTES")
	fmt.Fprintln(os.Stderr, "       talk-timer MINUTES MINUTES")
	os.Exit(2)
}

func parseArgs() (total, warning int) {
	flag.Usage = func() { usage("") }
	flag.IntVar(&total, "total-time", 0, "total time in minutes (1-99)")
	flag.IntVar(&warning, "warning-time", 0, "warning time in minutes (less than total)")
	flag.Parse()

	set := 0
	flag.Visit(func(*flag.Flag) { set++ })
	switch {
	case set == 2 && flag.NArg() == 0:
	case set == 0 && flag.NArg() == 2:
		var err1, err2 error
		total, err1 = strconv.Atoi(flag.Arg(0))
		warning, err2 = strconv.Atoi(flag.Arg(1))
		if err1 != nil || err2 != nil {
			usage("times must be whole numbers of minutes")
		}
	default:
		usage("give both times, either as flags or as two positional arguments")
	}

	if total < 1 || total > 99 {
		usage("total time must be between 1 and 99 minutes")
	}
	if warning < 0 || warning >= total {
		usage("warning time must be 0 or more and less than the total time")
	}
	return total, warning
}

// Quadrant block characters indexed by lit quarters: 1=top-left, 2=top-right, 4=bottom-left, 8=bottom-right.
var quadrants = []rune(" ▘▝▀▖▌▞▛▗▚▐▜▄▙▟█")

// plot draws a w×h cell box at (x, y), lighting each quarter cell that is at least half inside the shape.
// Shape coordinates are in units where a cell is 1 wide and 2 tall, so the units are square on screen.
func plot(s tcell.Screen, x, y, w, h int, st tcell.Style, inside func(x, y float64) bool) {
	const n = 4 // samples per side of a quarter cell
	for row := range h {
		for col := range w {
			bits := 0
			for k := range 4 {
				qx, qy := float64(col)+0.5*float64(k%2), float64(2*row)+float64(k/2)
				hits := 0
				for i := range n * n {
					if inside(qx+(float64(i%n)+0.5)*0.5/n, qy+(float64(i/n)+0.5)/n) {
						hits++
					}
				}
				if 2*hits >= n*n {
					bits |= 1 << k
				}
			}
			if bits != 0 {
				s.SetContent(x+col, y+row, quadrants[bits], nil, st)
			}
		}
	}
}

// segmentTest returns whether a point in a w×h digit lies on one of the lit segments.
// Each segment is a bar of thickness th with pointed ends, separated from its neighbours by gap g.
func segmentTest(mask int, w, h, th, g float64) func(x, y float64) bool {
	l, r, t, m, b := th/2, w-th/2, th/2, h/2, h-th/2
	hseg := func(x, y, yc float64) bool {
		d := math.Abs(y - yc)
		return d <= th/2 && x >= l+g+d && x <= r-g-d
	}
	vseg := func(x, y, xc, y0, y1 float64) bool {
		d := math.Abs(x - xc)
		return d <= th/2 && y >= y0+g+d && y <= y1-g-d
	}
	return func(x, y float64) bool {
		return mask&segA != 0 && hseg(x, y, t) ||
			mask&segG != 0 && hseg(x, y, m) ||
			mask&segD != 0 && hseg(x, y, b) ||
			mask&segF != 0 && vseg(x, y, l, t, m) ||
			mask&segB != 0 && vseg(x, y, r, t, m) ||
			mask&segE != 0 && vseg(x, y, l, m, b) ||
			mask&segC != 0 && vseg(x, y, r, m, b)
	}
}

func draw(s tcell.Screen, secs int, st tcell.Style) {
	sw, sh := s.Size()
	// Layout is D gap D gap colon gap D gap D, with gap = w/4 and colon about w/5, so about 5.2w in total.
	w := (sw - 2) * 5 / 26
	h := w * 5 / 6
	if h > sh-2 {
		h = sh - 2
		w = h * 6 / 5
	}
	w, h = max(w, 6), max(h, 5)
	fw, fh := float64(w), float64(2*h)
	th := 0.22 * fw
	g := max(0.04*fw, 0.5)
	gap := max(1, w/4)
	colon := max(1, int(th+0.5))

	total := 4*w + 4*gap + colon
	x := (sw - total) / 2
	y := (sh - h) / 2

	dots := func(px, py float64) bool {
		near := func(c float64) bool { return math.Abs(py-c) <= th/2 }
		return near(fh*0.32) || near(fh*0.68)
	}
	digits := [4]int{secs / 600, secs / 60 % 10, secs % 60 / 10, secs % 10}
	for i, d := range digits {
		plot(s, x, y, w, h, st, segmentTest(digitSegments[d], fw, fh, th, g))
		x += w + gap
		if i == 1 {
			plot(s, x, y, colon, h, st, dots)
			x += colon + gap
		}
	}
}

func main() {
	totalMin, warnMin := parseArgs()
	total := time.Duration(totalMin) * time.Minute
	warnSecs := warnMin * 60

	s, err := tcell.NewScreen()
	if err == nil {
		err = s.Init()
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	defer s.Fini()
	s.HideCursor()

	events := make(chan tcell.Event)
	go func() {
		for {
			ev := s.PollEvent()
			if ev == nil {
				return
			}
			events <- ev
		}
	}()
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	remaining := total
	running := false
	var end time.Time

	for {
		if running {
			remaining = max(0, time.Until(end))
			if remaining == 0 {
				running = false
			}
		}
		secs := int((remaining + time.Second - 1) / time.Second)

		s.Clear()
		switch {
		case secs == 0:
			if time.Now().UnixMilli()/250%2 == 0 {
				draw(s, 0, tcell.StyleDefault.Foreground(tcell.ColorRed))
			}
		case secs <= warnSecs:
			draw(s, secs, tcell.StyleDefault.Foreground(tcell.ColorYellow))
		default:
			draw(s, secs, tcell.StyleDefault.Foreground(tcell.ColorWhite))
		}
		s.Show()

		select {
		case <-ticker.C:
		case ev := <-events:
			switch ev := ev.(type) {
			case *tcell.EventResize:
				s.Sync()
			case *tcell.EventKey:
				switch {
				case ev.Key() == tcell.KeyEscape, ev.Key() == tcell.KeyCtrlC,
					ev.Key() == tcell.KeyRune && ev.Rune() == 'q':
					return
				case ev.Key() == tcell.KeyRune && ev.Rune() == ' ':
					if running {
						remaining = max(0, time.Until(end))
						running = false
					} else if remaining > 0 {
						end = time.Now().Add(remaining)
						running = true
					}
				case ev.Key() == tcell.KeyRune && ev.Rune() == 'r':
					remaining = total
					running = false
				}
			}
		}
	}
}
