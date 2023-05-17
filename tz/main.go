package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/DavidGamba/dgtools/run"
	"github.com/DavidGamba/go-getoptions"
	"github.com/charmbracelet/lipgloss"
)

var Logger = log.New(io.Discard, "", log.LstdFlags)

func main() {
	os.Exit(program(os.Args))
}

// Time format used for printing
// Choose between "15:04" or "03:04 PM"
var HourMinuteFormat = "15:04"

// TODO
var HourFormat = "15"

func program(args []string) int {
	opt := getoptions.New()
	opt.SetUnknownMode(getoptions.Pass)
	opt.Bool("verbose", false, opt.GetEnv("TZ_VERBOSE"))
	opt.Bool("standard", false, opt.Alias("analog", "civilian", "12-hour", "12h", "am-pm"), opt.Description("Use standard 12 hour AM/PM time format"))
	opt.SetCommandFn(Run)
	opt.HelpCommand("help", opt.Alias("?"))

	list := opt.NewCommand("list", "list all timezones")
	list.SetCommandFn(ListRun)

	remaining, err := opt.Parse(args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		return 1
	}
	if opt.Called("verbose") {
		Logger.SetOutput(os.Stderr)
	}
	if opt.Called("standard") {
		HourMinuteFormat = "03:04 PM"
		HourFormat = "03"
	}
	Logger.Println(remaining)

	ctx, cancel, done := getoptions.InterruptContext()
	defer func() { cancel(); <-done }()

	err = opt.Dispatch(ctx, remaining)
	if err != nil {
		if errors.Is(err, getoptions.ErrorHelpCalled) {
			return 1
		}
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		return 1
	}
	return 0
}

func Run(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	Logger.Printf("Running")

	cmd := []string{"echo", "hello", "world"}
	err := run.CMD(cmd...).Log().Run()
	if err != nil {
		return fmt.Errorf("failed: %w", err)
	}
	return nil
}

type ActorTime struct {
	Actor    string
	Time     time.Time
	Location string
	Offset   int // in seconds
	Display  string
}

type ActorMap map[int][]ActorTime

// List of locations can be found in "/usr/share/zoneinfo" in both Linux and macOS
func ListRun(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	locations := []string{
		"Australia/Sydney",
		"Asia/Tokyo",
		"Asia/Hong_Kong",
		"Asia/Shanghai",
		"Europe/Berlin",
		"Europe/Paris",
		"Europe/Madrid",
		"Europe/London",
		"UTC",
		"America/Bogota",
		"America/New_York",
		"America/Toronto",
		"America/Chicago",
		"America/Costa_Rica",
		"America/Edmonton",
		"America/Los_Angeles",
	}

	am := make(ActorMap)
	count := 0
	for _, location := range locations {
		loc, err := time.LoadLocation(location)
		if err != nil {
			return fmt.Errorf("failed to load '%s': %w", location, err)
		}
		now := time.Now().In(loc)
		_, offset := now.Zone()
		at := ActorTime{
			Actor:    location,
			Location: location,
			Time:     now,
			Offset:   offset,
			Display:  fmt.Sprintf("@%s (%s)", location, now.Format("MST")),
		}
		Logger.Printf("@%s: %s %d", at.Actor, at.Time.Format("01/02 15:04 MST"), offset/3600)
		if _, ok := am[offset]; !ok {
			am[offset] = []ActorTime{at}
		} else {
			am[offset] = append(am[offset], at)
		}
		count++
	}
	Logger.Printf("Total: %d\n", count)

	PrintActors(am)
	return nil
}

func PrintActors(am ActorMap) {
	// p := PurpleYellow
	// p := BlueGreen
	p := BlueYellow

	offsets := []int{}
	for offset := range am {
		offsets = append(offsets, offset)
	}
	sort.Ints(offsets)
	Logger.Println(len(offsets))
	count := 0
	for _, offset := range offsets {
		display := []string{}
		for _, at := range am[offset] {
			count++
			display = append(display, at.Display)
		}
		fmt.Printf("%s %s\n", am[offset][0].Time.Format(HourMinuteFormat), strings.Join(display, ", "))
		PrintActorsLine(p, am[offset])
		PrintHours(p, am[offset][0].Time, am[offsets[0]][0].Time)
	}
	Logger.Printf("Total: %d", count)
}

func PrintActorsLine(p Palette, att []ActorTime) {
	display := []string{}
	for _, at := range att {
		display = append(display, at.Display)
	}

	var timeStyle = lipgloss.NewStyle().
		Bold(true).
		PaddingLeft(0).
		PaddingRight(0)

	t := att[0].Time

	fmt.Printf("%s %s   %s\n", ClockEmoji[t.Hour()], timeStyle.Render(t.Format(HourMinuteFormat)), strings.Join(display, ", "))
}

var ClockEmoji = map[int]string{
	0:  "🕛",
	12: "🕛",
	24: "🕛",

	1:  "🕐",
	13: "🕐",

	2:  "🕑",
	14: "🕑",

	3:  "🕒",
	15: "🕒",

	4:  "🕓",
	16: "🕓",

	5:  "🕔",
	17: "🕔",

	6:  "🕕",
	18: "🕕",

	7:  "🕖",
	19: "🕖",

	8:  "🕗",
	20: "🕗",

	9:  "🕘",
	21: "🕘",

	10: "🕙",
	22: "🕙",

	11: "🕚",
	23: "🕚",
}

type Palette struct {
	Night     string
	Dawn      string
	Morning   string
	Noon      string // work hours
	AfterNoon string
	Dusk      string
	Evening   string

	FgNight     string
	FgDawn      string
	FgMorning   string
	FgNoon      string // work hours
	FgAfterNoon string
	FgDusk      string
	FgEvening   string

	Highlight   string
	FgHighlight string
}

var PurpleYellow = Palette{
	Night:     "#3d0066",
	Dawn:      "#5c0099",
	Morning:   "#c86bfa",
	Noon:      "#fdc500",
	AfterNoon: "#ffd500",
	Dusk:      "#ffee32",
	Evening:   "#03071e",
}

var BlueGreen = Palette{
	Night:     "#003e7f",
	Dawn:      "#0068af",
	Morning:   "#5495e1",
	Noon:      "#dddddd",
	AfterNoon: "#44a75e",
	Dusk:      "#008239",
	Evening:   "#00540e",
}

var BlueYellow = Palette{
	Night:     "#003e7f",
	Dawn:      "#0068af",
	Morning:   "#5495e1",
	Noon:      "#dddddd",
	AfterNoon: "#fef4d7",
	Dusk:      "#f9b16e",
	Evening:   "#00540e",

	FgNight:     "#FFFFFF",
	FgDawn:      "#FFFFFF",
	FgMorning:   "#646970",
	FgNoon:      "#000000",
	FgAfterNoon: "#646970",
	FgDusk:      "#FFFFFF",
	FgEvening:   "#FFFFFF",

	Highlight:   "#fafa00",
	FgHighlight: "#000000",
}

func PrintHours(p Palette, t, base time.Time) {
	// Logger.Printf("t: %s", t.Format("-07"))
	// Logger.Printf("base: %s", base.Format("-07"))
	// Logger.Printf("last: %s", last.Format("-07"))
	x := t.Hour()
	h := t.Hour()
	h -= 4

	var hStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(p.FgHighlight)).
		Background(lipgloss.Color(p.Highlight)).
		PaddingLeft(0).
		PaddingRight(0)

	for i := 0; i < 24; i++ {
		if h >= 24 {
			h = h - 24
		}
		if h < 0 {
			h = 24 + h
		}
		switch {
		case h >= 22 && h < 24:
			PrintBlock(fmt.Sprintf("%02d", h), p.FgNight, p.Night, h == x, hStyle)
		case h >= 0 && h < 3:
			PrintBlock(fmt.Sprintf("%02d", h), p.FgNight, p.Night, h == x, hStyle)
		case h >= 3 && h < 5:
			PrintBlock(fmt.Sprintf("%02d", h), p.FgNight, p.Night, h == x, hStyle)
		case h >= 5 && h < 7:
			PrintBlock(fmt.Sprintf("%02d", h), p.FgDawn, p.Dawn, h == x, hStyle)
		case h >= 7 && h < 9:
			PrintBlock(fmt.Sprintf("%02d", h), p.FgMorning, p.Morning, h == x, hStyle)
		case h >= 9 && h < 15:
			PrintBlock(fmt.Sprintf("%02d", h), p.FgNoon, p.Noon, h == x, hStyle)
		case h >= 15 && h < 17:
			PrintBlock(fmt.Sprintf("%02d", h), p.FgAfterNoon, p.AfterNoon, h == x, hStyle)
		case h >= 17 && h < 22:
			PrintBlock(fmt.Sprintf("%02d", h), p.FgDusk, p.Dusk, h == x, hStyle)
		case h >= 17 && h < 20:
			PrintBlock(fmt.Sprintf("%02d", h), p.FgDusk, p.Dusk, h == x, hStyle)
		case h >= 20 && h < 22:
			PrintBlock(fmt.Sprintf("%02d", h), p.FgEvening, p.Evening, h == x, hStyle)
		}
		h++
	}
	fmt.Println()

}

func PrintBlock(hour string, fg, bg string, highlight bool, hStyle lipgloss.Style) {
	var normal = lipgloss.NewStyle().
		Foreground(lipgloss.Color(fg)).
		Background(lipgloss.Color(bg)).
		PaddingLeft(2).
		PaddingRight(2)

	if highlight {
		// hour = fmt.Sprintf("%s", aurora.Index(0, aurora.BgIndex(229, hour)))
		hour = hStyle.Render(hour)
	} else {
		// hour = fmt.Sprintf("%s", aurora.Index(fg, aurora.BgIndex(bg, hour)))
		hour = normal.Render(hour)
	}
	fmt.Printf("%s", hour)
	// fmt.Printf("%s%s%s", aurora.Index(fg, aurora.BgIndex(bg, "  ")), hour, aurora.Index(fg, aurora.BgIndex(bg, "  ")))
}
