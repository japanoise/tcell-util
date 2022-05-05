//go:build !windows
// +build !windows

package termutil

import (
	"fmt"

	"github.com/gdamore/tcell"
)

//Parses a tcell.EventKey event and returns it as an emacs-ish keybinding string
//(e.g. "C-c", "LEFT", "TAB", etc.)
func ParseTcellEvent(ev *tcell.EventKey) string {
	if ev.Key() != tcell.KeyRune {
		prefix := ""
		if ev.Modifiers()&tcell.ModAlt != 0 {
			prefix = "M-"
		}
		switch ev.Key() {
		case tcell.KeyBackspace2:
			return prefix + "DEL"
		case tcell.KeyTab:
			return prefix + "TAB"
		case tcell.KeyEnter:
			return prefix + "RET"
		case tcell.KeyDown:
			return prefix + "DOWN"
		case tcell.KeyUp:
			return prefix + "UP"
		case tcell.KeyLeft:
			return prefix + "LEFT"
		case tcell.KeyRight:
			return prefix + "RIGHT"
		case tcell.KeyPgDn:
			return prefix + "next"
		case tcell.KeyPgUp:
			return prefix + "prior"
		case tcell.KeyHome:
			return prefix + "Home"
		case tcell.KeyEnd:
			return prefix + "End"
		case tcell.KeyDelete:
			return prefix + "deletechar"
		case tcell.KeyInsert:
			return prefix + "insert"
		case tcell.KeyEsc:
			return prefix + "ESC"
		case tcell.KeyCtrlUnderscore:
			if ev.Modifiers()&tcell.ModAlt != 0 {
				return "C-M-_"
			} else {
				return "C-_"
			}
		case tcell.KeyCtrlSpace:
			if ev.Modifiers()&tcell.ModAlt != 0 {
				return "C-M-@" // ikr, weird. but try: C-h c, C-SPC. it's C-@.
			} else {
				return "C-@"
			}
		}
		if ev.Key() <= 0x1A {
			if ev.Modifiers()&tcell.ModAlt != 0 {
				return fmt.Sprintf("C-M-%c", 96+ev.Key())
			} else {
				return fmt.Sprintf("C-%c", 96+ev.Key())
			}
		} else if ev.Key() <= tcell.KeyF1 && ev.Key() >= tcell.KeyF12 {
			if ev.Modifiers()&tcell.ModAlt != 0 {
				return fmt.Sprintf("M-f%d", 1+(tcell.KeyF1-ev.Key()))
			} else {
				return fmt.Sprintf("f%d", 1+(tcell.KeyF1-ev.Key()))
			}
		}
	} else if ev.Rune() == ' ' {
		if ev.Modifiers()&tcell.ModAlt != 0 {
			return "M-SPC"
		} else {
			return " "
		}
	} else if ev.Modifiers()&tcell.ModAlt != 0 {
		return fmt.Sprintf("M-%c", ev.Rune())
	}
	return string(ev.Rune())
}
