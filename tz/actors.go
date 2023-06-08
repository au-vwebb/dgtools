package main

import (
	"fmt"
	"sort"
	"time"

	"github.com/charmbracelet/lipgloss"
)

type ActorTime struct {
	Actor        string
	Time         time.Time
	Location     string
	Offset       int // in seconds
	Display      string
	Abbreviation string
}

type ActorMap map[int][]ActorTime

func PrintActors(am ActorMap, short bool, p *Palette) {
	offsets := []int{}
	for offset := range am {
		offsets = append(offsets, offset)
	}
	sort.Ints(offsets)
	Logger.Println(len(offsets))
	for _, offset := range offsets {
		PrintActorsLine(p, am[offset])
		if !short {
			PrintHours(p, am[offset][0].Time, am[offsets[0]][0].Time)
			fmt.Println()
		}
	}
}

func PrintActorsLine(p *Palette, att []ActorTime) {
	// map[abbreviation] = display
	displayLine := map[string][]string{}
	for _, at := range att {
		if _, ok := displayLine[at.Abbreviation]; !ok {
			displayLine[at.Abbreviation] = []string{at.Display}
		} else {
			displayLine[at.Abbreviation] = append(displayLine[at.Abbreviation], at.Display)
			// TODO: who cares about the repeated sorting right?
			sort.Strings(displayLine[at.Abbreviation])
		}
	}

	t := att[0].Time

	// Line length is 141

	fmt.Printf("%s %s   ", ClockEmoji[t.Hour()], p.Style(t.Hour()).Render(t.Format(HourMinuteFormat)))
	for abb, list := range displayLine {
		fmt.Printf("(%s)   ", abb)
		for _, d := range list {
			fmt.Printf("%s  ", p.LipglossPalette.Actor.Render(d))
		}
	}
	fmt.Println()
}

func PrintHours(p *Palette, t, base time.Time) {
	x := t.Hour()
	h := t.Hour()
	h -= 4

	for i := 0; i < 24; i++ {
		if h >= 24 {
			h = h - 24
		}
		if h < 0 {
			h = 24 + h
		}
		PrintBlock(fmt.Sprintf("%02d", h), p.Style(h), h == x, p.LipglossPalette.Highlight)
		h++
	}
	fmt.Println()

}

func PrintBlock(hour string, style lipgloss.Style, highlight bool, hStyle lipgloss.Style) {
	normal := style.Copy()

	normal.
		PaddingLeft(2).
		PaddingRight(2)

	if highlight {
		hour = hStyle.Render(hour)
	} else {
		hour = normal.Render(hour)
	}
	fmt.Printf("%s", hour)
}
