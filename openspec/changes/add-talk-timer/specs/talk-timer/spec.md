# Spec Delta

## Purpose

A full-screen terminal countdown timer for a speaker, showing remaining talk time in large 7-segment-style digits whose colour signals how close the talk is to running out.

## ADDED Requirements

### Requirement: Time arguments
The program SHALL accept a total time and a warning time, both whole minutes, either as the flags `-total-time` and `-warning-time` or as two positional arguments in that order.

#### Scenario: Flags
- **WHEN** the user runs `talk-timer -total-time 10 -warning-time 3`
- **THEN** the timer starts with a total of 10 minutes and a warning time of 3 minutes

#### Scenario: Positional arguments
- **WHEN** the user runs `talk-timer 10 3`
- **THEN** the timer starts with a total of 10 minutes and a warning time of 3 minutes

### Requirement: Argument validation
The program SHALL exit with a non-zero status and a usage message, without entering full-screen mode, when the arguments are missing, not integers, the total is outside 1 to 99, or the warning is negative or not less than the total.

#### Scenario: Missing argument
- **WHEN** the user runs `talk-timer 10`
- **THEN** the program prints usage and exits non-zero

#### Scenario: Not a number
- **WHEN** the user runs `talk-timer ten 3`
- **THEN** the program prints usage and exits non-zero

#### Scenario: Total out of range
- **WHEN** the user runs `talk-timer 100 3` or `talk-timer 0 0`
- **THEN** the program prints usage and exits non-zero

#### Scenario: Warning not less than total
- **WHEN** the user runs `talk-timer 5 5`
- **THEN** the program prints usage and exits non-zero

### Requirement: Full-screen 7-segment display
The program SHALL fill the terminal with the remaining time as `MM:SS` in a 7-segment-style font with a solid colon, centred, at the largest size that fits, with no other text on screen.

#### Scenario: Initial display
- **WHEN** the program starts with a total of 10 minutes
- **THEN** the screen shows `10:00` in 7-segment-style digits filling the terminal

#### Scenario: Terminal resized
- **WHEN** the terminal window is resized
- **THEN** the digits are redrawn centred at the largest size that fits the new size

### Requirement: Segment shapes
Each digit SHALL be drawn upright as seven separate segments, each a bar with pointed ends, separated from neighbouring segments by a visible gap, as on a real 7-segment display. Unlit segments SHALL be blank.

#### Scenario: Digit 8
- **WHEN** the display shows the digit `8`
- **THEN** all seven segments are drawn as separate pointed bars with gaps at the corners where they meet

#### Scenario: Unlit segments
- **WHEN** the display shows the digit `1`
- **THEN** only the two right-hand segments are drawn and the other five positions are blank

### Requirement: Starts paused
The timer SHALL start paused, showing the total time, until the user presses space.

#### Scenario: No key pressed
- **WHEN** the program has started and no key has been pressed
- **THEN** the display stays at the total time

### Requirement: Space starts and pauses
Pressing space SHALL toggle the timer between counting down and paused, except at zero, where space SHALL do nothing.

#### Scenario: Start
- **WHEN** the timer is paused above zero and the user presses space
- **THEN** the timer counts down from the displayed time

#### Scenario: Pause
- **WHEN** the timer is counting down and the user presses space
- **THEN** the display freezes at the current remaining time

#### Scenario: Resume
- **WHEN** the timer is paused mid-countdown and the user presses space
- **THEN** counting continues from the frozen time

#### Scenario: Space at zero
- **WHEN** the timer shows `00:00` and the user presses space
- **THEN** nothing changes

### Requirement: Reset
Pressing `r` SHALL set the timer back to the total time, paused, from any state.

#### Scenario: Reset while running
- **WHEN** the timer is counting down and the user presses `r`
- **THEN** the display shows the total time, white and paused, until space is pressed

#### Scenario: Reset at zero
- **WHEN** the timer shows `00:00` flashing red and the user presses `r`
- **THEN** the display shows the total time, white and paused

### Requirement: Quit
Pressing `q`, Esc or Ctrl-C SHALL exit the program and restore the terminal.

#### Scenario: Quit key
- **WHEN** the user presses `q`, Esc or Ctrl-C in any state
- **THEN** the program exits and the terminal returns to its previous state

### Requirement: Accurate countdown
The countdown SHALL follow real elapsed time while running, so the displayed time does not drift over the length of a talk, and SHALL stop at `00:00`.

#### Scenario: Elapsed time
- **WHEN** the timer has been running for 60 seconds of wall-clock time from `10:00`
- **THEN** the display shows `09:00`

#### Scenario: Reaching zero
- **WHEN** the remaining time reaches zero
- **THEN** the display shows `00:00` and does not count further

### Requirement: Colour by remaining time
The digits SHALL be white while remaining time is above the warning time, yellow when remaining time is at or below the warning time and above zero, and red at zero.

#### Scenario: Normal
- **WHEN** the total is 10 minutes, the warning is 3 minutes, and the display shows `05:00`
- **THEN** the digits are white

#### Scenario: Warning
- **WHEN** the warning is 3 minutes and the display shows `03:00` or less, above zero
- **THEN** the digits are yellow

#### Scenario: Zero warning time
- **WHEN** the warning time is 0
- **THEN** the digits stay white until zero

### Requirement: Flashing at zero
At `00:00` the red digits SHALL flash, alternating visible and blank roughly twice a second, until the user resets or quits.

#### Scenario: Time up
- **WHEN** the display reaches `00:00`
- **THEN** the red digits flash on and off continuously
