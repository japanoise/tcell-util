package main

import (
	"fmt"
	"log"
	"math/rand"

	"github.com/gdamore/tcell"
	termutil "github.com/japanoise/tcell-util"
)

func main() {
	s, err := tcell.NewScreen()
	if err != nil {
		log.Fatalf("%+v", err)
	}
	if err := s.Init(); err != nil {
		log.Fatalf("%+v", err)
	}

	// Set default text style
	s.SetStyle(tcell.StyleDefault)

	// Clear screen
	s.Clear()

	defer s.Fini()

	color := tcell.ColorDefault
	testPrompt := "Prompting"
	testScroll := "Scrolling through text"
	testKey := "Prompting for characters"
	testColor := "Selecting colors"
	quit := "Quit"
	choices := []string{
		testPrompt,
		testScroll,
		testKey,
		testColor,
		quit,
	}
	text := []string{
		"Magni enim ipsa maiores.",
		" Et exercitationem quaerat iure.",
		" Asperiores consequatur laboriosam est nihil id necessitatibus ad.",
		"",
		"Sed error repudiandae magni et suscipit cupiditate enim provident.",
		" Aut in vero rerum quia voluptate.",
		" Consectetur totam omnis et aut.",
		"",
		"Est nihil quia itaque adipisci.",
		" Dolores consequuntur minus vitae ipsum aut libero et natus.",
		" Magnam quo quis aperiam voluptatibus ut.",
		" Libero et atque aspernatur illum corporis sit est.",
		" Odio ipsam quisquam id autem.",
		"",
		"Est enim et est molestias ratione.",
		" In qui quaerat sed impedit non recusandae.",
		" Sapiente quibusdam est necessitatibus quod voluptatum.",
		" Voluptas dolor consequatur non blanditiis nostrum necessitatibus.",
		"",
		"Sunt doloribus est tempora in eligendi corporis animi voluptatibus.",
		" Qui amet temporibus iure.",
		" Porro et enim dicta earum odio quia rem.",
		"",
	}
	text = append(text, text...)
	text = append(text, text...)

	for {
		idx := termutil.ChoiceIndexCallback(
			s, "What do you want to test?", choices, 0,
			func(screen tcell.Screen, sel, x, y int) {
				termutil.PrintStringStyle(
					screen,
					x-10, y-rand.Intn(10)-2,
					fmt.Sprintf("Choice %v", sel),
					tcell.StyleDefault.Foreground(color))
			})
		if idx < 0 {
			continue
		}
		switch choices[idx] {
		case testPrompt:
			s.Clear()
			char := "Nothing!"
			buf := ""
			termutil.PromptWithCallback(
				s, "Type something",
				func(screen tcell.Screen, x, y int) {
					screen.Clear()
					termutil.PrintStringStyle(
						screen,
						x-termutil.RunewidthStr(buf), y-11,
						buf,
						tcell.StyleDefault.Foreground(
							tcell.ColorBlack+tcell.Color(rand.Intn(16)),
						))
					termutil.PrintStringStyle(
						screen, x-10, y-10, char,
						tcell.StyleDefault.Foreground(
							tcell.ColorBlack+tcell.Color(rand.Intn(16)),
						))
				},
				func(buffer, key string) {
					char = key
					buf = buffer
				})
		case testScroll:
			termutil.DisplayScreenMessage(
				s,
				text...,
			)
		case testKey:
			s.Clear()
			looping := true
			for looping {
				looping = termutil.YesNo(
					s, "Do you want to be asked y/n again?",
					func(screen tcell.Screen, x, y int) {
						termutil.PrintStringStyle(
							screen, rand.Intn(x), rand.Intn(y), "Wow!",
							tcell.StyleDefault.Foreground(
								tcell.ColorBlack+tcell.Color(rand.Intn(16)),
							))
					})
			}
		case testColor:
			color = termutil.PickColor(s, "Pick a color!")
		case quit:
			return
		}
	}
}
