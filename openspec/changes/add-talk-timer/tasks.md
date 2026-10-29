# Tasks

## 1. Setup

- [x] 1.1 Run `go mod init talk-timer` and `go get github.com/gdamore/tcell/v2`, and verify `go.mod` lists `tcell/v2`

## 2. Arguments

- [x] 2.1 In `main.go`, parse `-total-time`/`-warning-time` flags with a fallback to two positional ints, and validate total 1–99 and 0 ≤ warning < total. Print usage and exit non-zero on error. Verify that `go run . 10`, `go run . ten 3`, `go run . 100 3`, `go run . 0 0`, `go run . 5 5` and `go run . -total-time 10 3` each print usage and exit non-zero, before any screen is opened

## 3. Display

- [x] 3.1 Implement the 7-segment renderer: a segment mask table for digits 0–9, segment rectangles with `T`-row horizontal and `2T`-column vertical strokes, a two-square colon, and the largest `T` that fits, centred. Verify that `go build` succeeds and that a tmux pane (`tmux new -d -x 120 -y 40 'go run . 10 3'`, then `tmux capture-pane -p`) shows `10:00` in block digits filling the pane
- [x] 3.2 Redraw on resize events. Verify that after `tmux resize-window -x 60 -y 20` a fresh `capture-pane` shows the digits re-fitted to the smaller pane

## 4. Timer state and controls

- [x] 4.1 Implement the event loop: a goroutine forwarding `PollEvent`, a ~100 ms ticker, deadline-based remaining time with `ceil` display, and start paused. Verify in tmux that the display is unchanged at `10:00` after waiting 3 seconds with no key pressed
- [x] 4.2 Space toggles running/paused (and does nothing at zero), `r` resets to total and pauses, `q`/Esc/Ctrl-C quit and restore the terminal. Verify in tmux with `send-keys`:
  - space, then the display drops below `10:00` within 2 seconds
  - space again, then it stays frozen
  - `r`, then it shows `10:00` again
  - `q`, then the pane is back at a shell prompt (or the session has ended)
- [x] 4.3 Colour by displayed seconds: white above the warning time, yellow at or below it, red at zero, with red digits shown only on even 250 ms intervals at zero. Verify with `tmux capture-pane -e -p` on `talk-timer 2 1` after pressing space:
  - white escape codes before 60 s
  - yellow codes after about 65 s
  - red codes at `00:00` after about 125 s
  - successive captures at zero alternate between digits and blank

## 5. Final check

- [ ] 5.1 Run `go vet ./...` and `gofmt -l .` and verify both report nothing; then John runs `go build -o talk-timer . && ./talk-timer 10 3` full-screen on his laptop and confirms it reads clearly at speaking distance
