//These are some nice functions torn out of Gomacs which I think are better
//suited to be out of the project for reuse. It's imported as termutil.
package termutil

import (
	"errors"
	"unicode/utf8"

	"github.com/gdamore/tcell/v2"
)

//Get a string from the user. They can use typical emacs-ish editing commands,
//or press C-c or C-g to cancel.
func Prompt(screen tcell.Screen, prompt string, refresh func(tcell.Screen, int, int)) string {
	return PromptWithCallback(screen, prompt, refresh, nil)
}

//As prompt, but calls a function after every keystroke.
func PromptWithCallback(screen tcell.Screen, prompt string, refresh func(tcell.Screen, int, int), callback func(string, string)) string {
	if callback == nil {
		return DynamicPromptWithCallback(screen, prompt, refresh, nil)
	} else {
		return DynamicPromptWithCallback(screen, prompt, refresh, func(a, b string) string {
			callback(a, b)
			return a
		})
	}
}

//As prompt, but calls a function after every keystroke that can modify the query.
func DynamicPromptWithCallback(screen tcell.Screen, prompt string, refresh func(tcell.Screen, int, int), callback func(string, string) string) string {
	return EditDynamicWithCallback(screen, "", prompt, refresh, callback)
}

// Edit takes a default value and a refresh function. It allows the
// user to edit the default value. It returns what the user entered.
func Edit(screen tcell.Screen, defval, prompt string, refresh func(tcell.Screen, int, int)) string {
	return EditDynamicWithCallback(screen, defval, prompt, refresh, nil)
}

// EditDynamicWithCallback takes a default value, prompt, refresh
// function, and callback. It allows the user to edit the default
// value. It returns what the user entered.
func EditDynamicWithCallback(screen tcell.Screen, defval, prompt string, refresh func(tcell.Screen, int, int), callback func(string, string) string) string {
	var buffer string
	var bufpos, cursor, offset int
	if defval == "" {
		buffer = ""
		bufpos = 0
		cursor = 0
		offset = 0
	} else {
		x, _ := screen.Size()
		buffer = defval
		bufpos = len(buffer)
		if RunewidthStr(buffer) > x {
			cursor = x - 1
			offset = len(buffer) + 1 - x
		} else {
			offset = 0
			cursor = RunewidthStr(buffer)
		}
	}
	iw := RunewidthStr(prompt + ": ")
	for {
		buflen := len(buffer)
		x, y := screen.Size()
		if refresh != nil {
			refresh(screen, x, y)
		}
		ClearLine(screen, x, y-1)
		for iw+cursor >= x {
			offset++
			cursor--
		}
		for iw+cursor < iw {
			offset--
			cursor++
		}
		t, _ := trimString(buffer, offset)
		PrintString(screen, 0, y-1, prompt+": "+t)
		screen.ShowCursor(iw+cursor, y-1)
		screen.Show()
		ev := screen.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventKey:
			key := ParseTcellEvent(ev)
			switch key {
			case "LEFT", "C-b":
				if bufpos > 0 {
					r, rs := utf8.DecodeLastRuneInString(buffer[:bufpos])
					bufpos -= rs
					cursor -= Runewidth(r)
				}
			case "RIGHT", "C-f":
				if bufpos < buflen {
					r, rs := utf8.DecodeRuneInString(buffer[bufpos:])
					bufpos += rs
					cursor += Runewidth(r)
				}
			case "C-a":
				fallthrough
			case "Home":
				bufpos = 0
				cursor = 0
				offset = 0
			case "C-e":
				fallthrough
			case "End":
				bufpos = buflen
				if RunewidthStr(buffer) > x {
					cursor = x - 1
					offset = buflen + 1 - x
				} else {
					offset = 0
					cursor = RunewidthStr(buffer)
				}
			case "C-c":
				fallthrough
			case "C-g":
				if callback != nil {
					result := callback(buffer, key)
					if result != buffer {
						offset = 0
						buffer, buflen, bufpos, cursor = recalcBuffer(result)
					}
				}
				return defval
			case "RET":
				if callback != nil {
					result := callback(buffer, key)
					if result != buffer {
						offset = 0
						buffer, buflen, bufpos, cursor = recalcBuffer(result)
					}
				}
				return buffer
			case "C-d":
				fallthrough
			case "deletechar":
				if bufpos < buflen {
					r, rs := utf8.DecodeRuneInString(buffer[bufpos:])
					bufpos += rs
					cursor += Runewidth(r)
				} else {
					if callback != nil {
						result := callback(buffer, key)
						if result != buffer {
							offset = 0
							buffer, buflen, bufpos, cursor = recalcBuffer(result)
						}
					}
					continue
				}
				fallthrough
			case "DEL", "C-h":
				if buflen > 0 {
					if bufpos == buflen {
						r, rs := utf8.DecodeLastRuneInString(buffer)
						buffer = buffer[0 : buflen-rs]
						bufpos -= rs
						cursor -= Runewidth(r)
					} else if bufpos > 0 {
						r, rs := utf8.DecodeLastRuneInString(buffer[:bufpos])
						buffer = buffer[:bufpos-rs] + buffer[bufpos:]
						bufpos -= rs
						cursor -= Runewidth(r)
					}
				}
			case "C-u":
				buffer = ""
				buflen = 0
				bufpos = 0
				cursor = 0
				offset = 0
			case "M-DEL":
				if buflen > 0 && bufpos > 0 {
					delto := backwordWordIndex(buffer, bufpos)
					buffer = buffer[:delto] + buffer[bufpos:]
					buflen = len(buffer)
					bufpos = delto
					cursor = RunewidthStr(buffer[:bufpos])
				}
			case "M-d":
				if buflen > 0 && bufpos < buflen {
					delto := forwardWordIndex(buffer, bufpos)
					buffer = buffer[:bufpos] + buffer[delto:]
					buflen = len(buffer)
				}
			case "M-b":
				if buflen > 0 && bufpos > 0 {
					bufpos = backwordWordIndex(buffer, bufpos)
					cursor = RunewidthStr(buffer[:bufpos])
				}
			case "M-f":
				if buflen > 0 && bufpos < buflen {
					bufpos = forwardWordIndex(buffer, bufpos)
					cursor = RunewidthStr(buffer[:bufpos])
				}
			default:
				if utf8.RuneCountInString(key) == 1 {
					r, _ := utf8.DecodeLastRuneInString(buffer)
					buffer = buffer[:bufpos] + key + buffer[bufpos:]
					bufpos += len(key)
					cursor += Runewidth(r)
				}
			}
			if callback != nil {
				result := callback(buffer, key)
				if result != buffer {
					offset = 0
					buffer, buflen, _, _ = recalcBuffer(result)
					bufpos = buflen
					cursor = RunewidthStr(buffer)
				}
			}
		}
	}
}

func recalcBuffer(result string) (string, int, int, int) {
	rlen := len(result)
	return result, rlen, 0, 0
}

func backwordWordIndex(buffer string, bufpos int) int {
	r, rs := utf8.DecodeLastRuneInString(buffer[:bufpos])
	ret := bufpos - rs
	r, rs = utf8.DecodeLastRuneInString(buffer[:ret])
	for ret > 0 && WordCharacter(r) {
		ret -= rs
		r, rs = utf8.DecodeLastRuneInString(buffer[:ret])
	}
	return ret
}

func forwardWordIndex(buffer string, bufpos int) int {
	r, rs := utf8.DecodeRuneInString(buffer[bufpos:])
	ret := bufpos + rs
	r, rs = utf8.DecodeRuneInString(buffer[ret:])
	for ret < len(buffer) && WordCharacter(r) {
		ret += rs
		r, rs = utf8.DecodeRuneInString(buffer[ret:])
	}
	return ret
}

//Allows the user to select one of many choices displayed on-screen.
//Takes a title, choices, and default selection. Returns an index into the choices
//array; or def (default)
func ChoiceIndex(screen tcell.Screen, title string, choices []string, def int) int {
	return ChoiceIndexCallback(screen, title, choices, def, nil)
}

//As ChoiceIndex, but calls a function after drawing the interface,
//passing it the current selected choice, screen width, and screen height.
func ChoiceIndexCallback(screen tcell.Screen, title string, choices []string, def int, f func(tcell.Screen, int, int, int)) int {
	selection := def
	nc := len(choices) - 1
	if selection < 0 || selection > nc {
		selection = 0
	}
	offset := 0
	cx := 0
	for {
		sx, sy := screen.Size()
		screen.HideCursor()
		screen.Clear()
		PrintString(screen, 0, 0, title)
		for selection < offset {
			offset -= 5
			if offset < 0 {
				offset = 0
			}
		}
		for selection-offset >= sy-1 {
			offset += 5
			if offset >= nc {
				offset = nc
			}
		}
		for i, s := range choices[offset:] {
			ts, _ := trimString(s, cx)
			PrintString(screen, 3, i+1, ts)
			if cx > 0 {
				PrintString(screen, 2, i+1, "???")
			}
		}
		PrintString(screen, 1, (selection+1)-offset, ">")
		if f != nil {
			f(screen, selection, sx, sy)
		}
		screen.Show()
		ev := screen.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventKey:
			key := ParseTcellEvent(ev)
			switch key {
			case "C-v":
				fallthrough
			case "next":
				selection += sy - 5
				if selection >= len(choices) {
					selection = len(choices) - 1
				}
			case "M-v":
				fallthrough
			case "prior":
				selection -= sy - 5
				if selection < 0 {
					selection = 0
				}
			case "C-c":
				fallthrough
			case "C-g":
				return def
			case "UP", "C-p":
				if selection > 0 {
					selection--
				}
			case "DOWN", "C-n":
				if selection < len(choices)-1 {
					selection++
				}
			case "LEFT", "C-b":
				if cx > 0 {
					cx--
				}
			case "RIGHT", "C-f":
				cx++
			case "C-a", "Home":
				cx = 0
			case "M-<":
				selection = 0
			case "M->":
				selection = len(choices) - 1
			case "RET":
				return selection
			}
		}
	}
}

//Displays the prompt p and asks the user to say y or n. Returns true if y; false
//if no.
func YesNo(screen tcell.Screen, p string, refresh func(tcell.Screen, int, int)) bool {
	ret, _ := yesNoChoice(screen, p, false, refresh)
	return ret
}

//Same as YesNo, but will return a non-nil error if the user presses C-g.
func YesNoCancel(screen tcell.Screen, p string, refresh func(tcell.Screen, int, int)) (bool, error) {
	return yesNoChoice(screen, p, true, refresh)
}

// Asks the user to press one of a set of keys. Returns the one which they pressed.
func PressKey(screen tcell.Screen, p string, refresh func(tcell.Screen, int, int), keys ...string) string {
	var plen int
	pm := p + " ("
	for i, key := range keys {
		if i != 0 {
			pm += "/"
		}
		pm += key
	}
	pm += ")"
	plen = utf8.RuneCountInString(pm) + 1
	x, y := screen.Size()
	if refresh != nil {
		refresh(screen, x, y)
	}
	ClearLine(screen, x, y-1)
	PrintString(screen, 0, y-1, pm)
	screen.ShowCursor(plen, y-1)
	screen.Show()
	for {
		ev := screen.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventResize:
			x, y = screen.Size()
			if refresh != nil {
				refresh(screen, x, y)
			}
			ClearLine(screen, x, y-1)
			PrintString(screen, 0, y-1, pm)
			screen.ShowCursor(plen, y-1)
			screen.Show()
		case *tcell.EventKey:
			pressedkey := ParseTcellEvent(ev)
			for _, key := range keys {
				if key == pressedkey {
					return key
				}
			}
		}
	}
}

func yesNoChoice(screen tcell.Screen, p string, allowcancel bool, refresh func(tcell.Screen, int, int)) (bool, error) {
	if allowcancel {
		key := PressKey(screen, p, refresh, "y", "n", "C-g")
		switch key {
		case "y":
			return true, nil
		case "n":
			return false, nil
		case "C-g", "C-c":
			return false, errors.New("User cancelled")
		}
	}
	key := PressKey(screen, p, refresh, "y", "n")
	return key == "y", nil
}

func PickColor(screen tcell.Screen, prompt string) tcell.Color {
	idx := 0
	for {
		sx, sy := screen.Size()
		pillWidth := sx / 16
		screen.Clear()
		PrintString(screen, 0, 0, prompt)
		if sy < 16 {
			PrintString(screen, sx-26, 0, "Warning: Screen too short")
		}
		if sx < 16 {
			PrintString(screen, sx-27, 0, "Warning: Screen too narrow")
		}
		for i := 0; i < 256; i++ {
			for j := 0; j < pillWidth; j++ {
				if i == idx {
					screen.SetContent(
						((i%16)*pillWidth)+j,
						1+(i/16),
						'=', nil,
						tcell.StyleDefault.Foreground(
							tcell.ColorBlack+tcell.Color(i)))
				} else {
					screen.SetContent(
						((i%16)*pillWidth)+j,
						1+(i/16),
						' ', nil,
						tcell.StyleDefault.Background(
							tcell.ColorBlack+tcell.Color(i)))
				}
			}
		}
		screen.ShowCursor((idx%16)*pillWidth, 1+(idx/16))
		screen.Show()

		ev := screen.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventKey:
			key := ParseTcellEvent(ev)
			switch key {
			case "M-<":
				idx = 0
			case "M->":
				idx = 255
			case "UP", "C-p":
				idx -= 16
			case "DOWN", "C-n":
				idx += 16
			case "M-UP", "M-p":
				idx -= 64
			case "M-DOWN", "M-n":
				idx += 64
			case "M-LEFT", "M-b":
				idx -= 4
			case "M-RIGHT", "M-f":
				idx += 4
			case "LEFT", "C-b":
				idx--
			case "RIGHT", "C-f":
				idx++
			case "HOME", "C-a":
				for idx%16 != 0 {
					idx--
				}
			case "END", "C-e":
				for idx%16 != 15 {
					idx++
				}
			case "RET":
				return tcell.ColorBlack + tcell.Color(idx)
			}
			if idx < 0 {
				idx = 0
			} else if idx > 255 {
				idx = 255
			}
		}
	}
}
