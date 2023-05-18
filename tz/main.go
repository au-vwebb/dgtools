package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
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
	opt.Bool("verbose", false, opt.GetEnv("TZ_VERBOSE"), opt.Description("Enable logging"))
	opt.Bool("standard", false, opt.Alias("analog", "civilian", "12-hour", "12h", "am-pm"), opt.Description("Use standard 12 hour AM/PM time format"))
	opt.Bool("short", false, opt.Alias("s"), opt.Description("Don't show timezone bars"))
	opt.SetCommandFn(ListRun)

	list := opt.NewCommand("list", "list all timezones")
	list.SetCommandFn(ListRun)

	opt.HelpCommand("help", opt.Alias("?"))
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
	short := opt.Value("short").(bool)

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

	PrintActors(am, short)
	return nil
}

func PrintActors(am ActorMap, short bool) {
	// p := PurpleYellow
	// p := BlueGreen
	p := NewPalette("BlueYellow")

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
	display := []string{}
	for _, at := range att {
		display = append(display, at.Display)
	}

	t := att[0].Time

	// Line length is 141

	fmt.Printf("%s %s   ", ClockEmoji[t.Hour()], p.Style(t.Hour()).Render(t.Format(HourMinuteFormat)))
	for _, d := range display {
		fmt.Printf("%s  ", p.LipglossPalette.Actor.Render(d))
	}
	fmt.Println()
}

func PrintHours(p *Palette, t, base time.Time) {
	// Logger.Printf("t: %s", t.Format("-07"))
	// Logger.Printf("base: %s", base.Format("-07"))
	// Logger.Printf("last: %s", last.Format("-07"))
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
