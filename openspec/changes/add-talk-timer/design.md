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

**7-segment rendering from a segment table.** Each digit 0 to 9 maps to a 7-bit mask (segments a to g). A digit is drawn into a box of width `W` and height `H` by filling rectangles of full-block cells (`█`) for each lit segment:
- Horizontal segments are `T` rows thick.
- Vertical segments are `2T` columns thick, because terminal cells are about twice as tall as they are wide, so this keeps the strokes looking even.
- The colon is two filled squares (`2T` × `T` cells).

**Scaling.** The layout is `D D : D D` with a fixed gap between elements. `T` is computed as the largest value for which the whole layout (digit width about `6T` columns, height about `5T` rows plus gaps and colon) fits inside the screen with a small margin. `T` has a floor of 1, and the layout is centred. On a very small terminal the digits are simply drawn at the minimum size and may be clipped. That's acceptable for a full-screen tool.

## Risks / Trade-offs

- [Small terminal windows clip the display] → The speaker runs it full-screen. With `T` = 1 the layout needs only about 30×7 cells.
- [Terminals without true yellow/red] → Use the basic ANSI colours (`tcell.ColorWhite`, `ColorYellow`, `ColorRed`), which every terminal supports.
- [Laptop sleeps during a talk] → Out of scope. The deadline approach means the display is correct on wake.
