package main

import (
	"flag"
	"fmt"
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

func fill(s tcell.Screen, x, y, w, h int, st tcell.Style) {
	for row := y; row < y+h; row++ {
		for col := x; col < x+w; col++ {
			s.SetContent(col, row, '█', nil, st)
		}
	}
}

func drawDigit(s tcell.Screen, d, x, y, w, h, t int, st tcell.Style) {
	v := 2 * t
	mid := y + (h-t)/2
	m := digitSegments[d]
	if m&segA != 0 {
		fill(s, x, y, w, t, st)
	}
	if m&segG != 0 {
		fill(s, x, mid, w, t, st)
	}
	if m&segD != 0 {
		fill(s, x, y+h-t, w, t, st)
	}
	if m&segF != 0 {
		fill(s, x, y, v, mid+t-y, st)
	}
	if m&segB != 0 {
		fill(s, x+w-v, y, v, mid+t-y, st)
	}
	if m&segE != 0 {
		fill(s, x, mid, v, y+h-mid, st)
	}
	if m&segC != 0 {
		fill(s, x+w-v, mid, v, y+h-mid, st)
	}
}

func draw(s tcell.Screen, secs int, st tcell.Style) {
	sw, sh := s.Size()
	// Layout is D gap D gap colon gap D gap D, with gap = w/4 and colon = w/4 wide.
	w := (sw - 2) * 4 / 21
	h := w * 7 / 8
	if h > sh-2 {
		h = sh - 2
		w = h * 8 / 7
	}
	w, h = max(w, 5), max(h, 5)
	t := max(1, h/7)
	gap := max(1, w/4)
	colon := 2 * t

	total := 4*w + 4*gap + colon
	x := (sw - total) / 2
	y := (sh - h) / 2

	digits := [4]int{secs / 600, secs / 60 % 10, secs % 60 / 10, secs % 10}
	for i, d := range digits {
		drawDigit(s, d, x, y, w, h, t, st)
		x += w + gap
		if i == 1 {
			fill(s, x, y+h/3-t/2, colon, t, st)
			fill(s, x, y+2*h/3-t/2, colon, t, st)
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
