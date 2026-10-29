# Design

## Context

The project is empty apart from OpenSpec. Go 1.27 is installed. See proposal.md for the motivation and specs/talk-timer/spec.md for the required behaviour. The brief is "truly basic", so the design favours one small file over structure.

## Goals / Non-Goals

**Goals:**
- One `main.go` of a few hundred lines at most, with one dependency.
- Digits that read clearly from a couple of metres on a laptop screen.

**Non-Goals:**
- Configuration files, themes, sounds, overtime count-up, or seconds-only arguments.
- Automated tests. The user asked for something truly basic and agreed to a build-and-run check only.
- Windows-specific handling beyond what `tcell` gives for free.

## Decisions

**`tcell` for the terminal.** It gives full-screen mode, key events, resize events, colours and terminal restore on exit in one well-used library.
- Bubble Tea was rejected because its model/update/view framework is more than a single redrawn screen needs.
- Raw ANSI codes were rejected because raw mode, resize signals and terminal restore would need hand-written code.

**Arguments with the stdlib `flag` package.** Define `-total-time` and `-warning-time` as int flags. After `flag.Parse()`, if neither flag was set and exactly two positional arguments remain, parse them with `strconv.Atoi`. Then validate the ranges.
- Mixing flags and positional arguments is not supported and is rejected with the usage message.
- Using no third-party CLI library keeps the program to one dependency.

**Timing from a deadline, not a tick counter.** While running, the program keeps an end time (`time.Now() + remaining`). On pause it stores `remaining = time.Until(end)`.
- Displayed seconds are `ceil(remaining)`, so `10:00` shows for the first second after start, and `00:00` appears only when time is actually up.
- Colour is chosen from the displayed seconds, so the colour changes in step with the digits.
- A `time.Ticker` at about 100 ms wakes the loop to redraw. The ticker only triggers redraws and never adds or subtracts time, so it can't drift.

**One event loop.** A goroutine forwards `screen.PollEvent()` into a channel. The main loop `select`s on that channel and the ticker, updates the state (paused/running, remaining, end time), then redraws the whole screen. With no shared mutable state there are no locks.

**Flashing.** At zero the digits are drawn only when `time.Now().UnixMilli()/250` is even. This gives an on/off cycle of about twice a second, driven by the same ticker.

**7-segment rendering by sampling shapes at quarter-cell resolution.** The first version filled whole cells with `█`, which looked too blocky to pass for a 7-segment display.
- Each digit 0 to 9 maps to a 7-bit mask (segments a to g).
- Each segment is described geometrically as a bar of thickness `th` with pointed (45°) ends, shortened by a gap `g` at each end so neighbouring segments don't touch. Coordinates use units where a cell is 1 wide and 2 tall, so shapes are square on screen.
- Every cell is split into 2×2 quarters, and each quarter is drawn with one of the 16 quadrant block characters (`▘▝▀▖▌▞▛▗▚▐▜▄▙▟█`). A quarter is lit when at least half of it lies inside a lit segment, using 4×4 samples. This removes most single-quarter specks along the slanted edges.
- Ghostty and iTerm2 draw these characters natively, so edges join cleanly between cells.
- The colon is two square dots of side `th`, at about 1/3 and 2/3 of the digit height.
- Rejected:
  - Braille (2×4 dots per cell), because the dots have gaps and look stippled rather than solid.
  - Sextant characters (2×3 per cell), because font support is patchier.
  - Slanted digits, because they came out jagged at terminal resolution and the user chose upright.
  - Faint "ghost" unlit segments, which the user declined.

**Scaling.** The layout is `D D : D D`, with the gap between elements a quarter of the digit width. The digit width `w` (cells) is the largest that fits about `5.2w` across the screen with a small margin. Height is `5w/6` rows, which gives about a 1:1.7 width:height digit on screen. If that height doesn't fit, it's taken from the screen height instead. Stroke thickness is `0.22w` and the gap is `max(0.04w, 0.5)`. There's a minimum size of 6×5 cells, and the layout is centred. On a very small terminal the digits are drawn at the minimum size and may be clipped. That's acceptable for a full-screen tool.

## Risks / Trade-offs

- [Small terminal windows clip the display] → The speaker runs it full-screen. At the minimum size the layout needs only about 32×7 cells.
- [A terminal font without quadrant characters would show gaps or boxes] → Ghostty and iTerm2, the terminals in use, draw block elements themselves rather than using the font.
- [Pointed segment tips can leave a single quarter-cell at small sizes] → This is the true shape at that resolution. It shrinks relative to the digit at full-screen sizes.
- [Terminals without true yellow/red] → Use the basic ANSI colours (`tcell.ColorWhite`, `ColorYellow`, `ColorRed`), which every terminal supports.
- [Laptop sleeps during a talk] → Out of scope. The deadline approach means the display is correct on wake.
