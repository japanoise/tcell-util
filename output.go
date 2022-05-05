package termutil

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/gdamore/tcell"
	"github.com/mattn/go-runewidth"
)

//Indicate whether the given rune is a word character
func WordCharacter(c rune) bool {
	return (c >= '0' && c <= '9') || (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c == '_') || c > 127
}

//Pass the screenwidth and a line number; this function will clear the given line.
func ClearLine(screen tcell.Screen, sx, y int) {
	for i := 0; i < sx; i++ {
		screen.SetContent(i, y, ' ', nil, tcell.StyleDefault)
	}
}

//Returns how many cells wide the given rune is.
func Runewidth(ru rune) int {
	if IsControl(ru) {
		return 2
	} else if ' ' <= ru && ru <= '~' {
		return 1
	}
	rw := runewidth.RuneWidth(ru)
	if rw <= 0 {
		return 1
	} else {
		return rw
	}
}

//Returns true if the rune is a control character or invalid rune
func IsControl(ru rune) bool {
	return unicode.IsControl(ru) || !utf8.ValidRune(ru)
}

//Returns how many cells wide the given string is
func RunewidthStr(s string) int {
	ret := 0
	for _, ru := range s {
		ret += Runewidth(ru)
	}
	return ret
}

//Prints the rune given on the screen. Uses reverse colors for control
//characters.
func PrintRune(screen tcell.Screen, x, y int, ru rune) {
	PrintRuneStyle(screen, x, y, ru, tcell.StyleDefault)
}

//Print the rune with reverse colors for control characters
func PrintRuneStyle(screen tcell.Screen, x, y int, ru rune, style tcell.Style) {
	if IsControl(ru) {
		if ru <= rune(26) {
			screen.SetContent(x, y, '^', nil, style.Reverse(true))
			screen.SetContent(x+1, y, '@'+ru, nil, style.Reverse(true))
		} else {
			screen.SetContent(x, y, '�', nil, style)
		}
	} else {
		screen.SetContent(x, y, ru, nil, style)
	}
}

//Prints the string given on the screen. Uses the above functions to choose how it
//appears.
func Printstring(screen tcell.Screen, s string, x, y int) {
	PrintStringStyle(screen, x, y, s, tcell.StyleDefault)
}

//Same as Printstring, but passes a color to PrintRune.
func PrintstringColored(screen tcell.Screen, style tcell.Style, s string, x, y int) {
	PrintStringStyle(screen, x, y, s, style)
}

//Print string with a style
func PrintStringStyle(screen tcell.Screen, x, y int, s string, style tcell.Style) {
	i := 0
	for _, ru := range s {
		PrintRuneStyle(screen, x+i, y, ru, style)
		i += Runewidth(ru)
	}
}

func pauseForAnyKey(screen tcell.Screen, currentRow int) {
	Printstring(screen, "<More>", 0, currentRow)
	screen.Show()
	ev := screen.PollEvent()
Loop:
	for {
		switch ev.(type) {
		case *tcell.EventKey:
			break Loop
		default:
			ev = screen.PollEvent()
		}
	}
	screen.Clear()
	screen.Show()
}

type lessRow struct {
	data string
	len  int
}

func lessDrawRows(screen tcell.Screen, sx, sy, cx, cy int, rows []lessRow, numrows int) {
	for i := 0; i < sy-1; i++ {
		ri := cy + i
		if ri >= 0 && ri < numrows {
			if cx < len(rows[ri].data) {
				ts, _ := trimString(rows[ri].data, cx)
				Printstring(screen, ts, 0, i)
			}
		}
	}
	for i := 0; i < sx; i++ {
		PrintRuneStyle(screen, i, sy-1, ' ', tcell.StyleDefault.Reverse(true))
	}
	PrintstringColored(screen, tcell.StyleDefault.Reverse(true), "^C, ^G, q to quit. Arrow keys/Vi keys/Emacs keys to move.", 0, sy-1)
	screen.Show()
}

//Prints all strings given to the screen, and allows the user to scroll through,
//rather like less(1).
func DisplayScreenMessage(screen tcell.Screen, messages ...string) {
	screen.HideCursor()
	rows := make([]lessRow, 0)
	for _, msg := range messages {
		for _, s := range strings.Split(msg, "\n") {
			renderstring := strings.Replace(s, "\t", "        ", -1)
			rows = append(rows, lessRow{renderstring, len(renderstring)})
		}
	}
	numrows := len(rows)
	cy := 0
	cx := 0
	done := false
	for !done {
		screen.Clear()
		sx, sy := screen.Size()
		if sy > numrows {
			cy = 0
		}
		lessDrawRows(screen, sx, sy, cx, cy, rows, numrows)

		ev := screen.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventKey:
			switch ParseTcellEvent(ev) {
			case "q", "C-c", "C-g":
				done = true
			case "DOWN", "j", "C-n":
				if cy < numrows+1-sy {
					cy++
				}
			case "UP", "k", "C-p":
				if cy > 0 {
					cy--
				}
			case "Home", "C-a":
				cx = 0
			case "LEFT", "h", "C-b":
				if cx > 0 {
					cx--
				}
			case "RIGHT", "l", "C-f":
				cx++
			case "next", "C-v":
				cy += sy - 2
				if cy > numrows+1-sy {
					cy = numrows + 1 - sy
				}
			case "prior", "M-v":
				cy -= sy - 2
				if cy < 0 {
					cy = 0
				}
			case "g", "M-<":
				cy = 0
			case "G", "M->":
				cy = numrows + 1 - sy
			case "/", "C-s":
				search := Prompt(screen, "Search", func(screen tcell.Screen, ssx, ssy int) {
					lessDrawRows(screen, ssx, ssy, cx, cy, rows, numrows)
				})
				screen.HideCursor()
				for offset, row := range rows[cy:] {
					if strings.Contains(row.data, search) {
						cy += offset
						break
					}
				}
			}
		}
	}
}

func trimString(s string, coloff int) (string, int) {
	if coloff == 0 {
		return s, 0
	}
	sr := []rune(s)
	if coloff < len(sr) {
		ret := string(sr[coloff:])
		return ret, strings.Index(s, ret)
	} else {
		return "", 0
	}
}
