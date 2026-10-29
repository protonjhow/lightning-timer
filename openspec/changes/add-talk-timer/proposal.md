# Proposal

## Why

A speaker giving a talk needs to see, at a glance from a step or two away, how much time they have left, without reading small text. A full-screen terminal countdown on a laptop facing the speaker, which changes colour as time runs low, does that with nothing to install beyond a single binary.

## What Changes

- New Go command-line program `talk-timer`, a single file using `tcell` for terminal handling.
- Takes two whole-minute integers, total time and warning time, either as flags (`-total-time 10 -warning-time 3`) or positionally (`10 3`).
- Shows the remaining time full-screen as `MM:SS` in a 7-segment-style font built from block characters, scaled to the terminal and redrawn on resize.
- Colour signals urgency: white normally, yellow at or below the warning time, red and flashing at `00:00`.
- Starts paused at the total time. Space starts and pauses, `r` resets to the total time paused, and `q`, Esc and Ctrl-C quit.

## Capabilities

### New Capabilities
- `talk-timer`: the full-screen countdown timer: argument handling, countdown and pause/reset behaviour, key controls, and the 7-segment colour-coded display.

### Modified Capabilities

None.

## Impact

- New files in the project root: `main.go`, `go.mod`, `go.sum`.
- New dependency: `github.com/gdamore/tcell/v2`.
- No existing code or specs are affected.
