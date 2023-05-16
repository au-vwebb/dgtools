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

var Logger = log.New(os.Stderr, "", log.LstdFlags)

func main() {
	os.Exit(program(os.Args))
}

func program(args []string) int {
	opt := getoptions.New()
	opt.SetUnknownMode(getoptions.Pass)
	opt.Bool("quiet", false, opt.GetEnv("QUIET"))
	opt.SetCommandFn(Run)
	opt.HelpCommand("help", opt.Alias("?"))

	list := opt.NewCommand("list", "list all timezones")
	list.SetCommandFn(ListRun)

	remaining, err := opt.Parse(args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		return 1
	}
	if opt.Called("quiet") {
		Logger.SetOutput(io.Discard)
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
	Logger.Printf("Total: %d", count)
	fmt.Println()

	PrintActors(am)
	return nil
}

func PrintActors(am ActorMap) {
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
		fmt.Printf("%s\n", strings.Join(display, ", "))
		PrintHours(am[offset][0].Time, am[offsets[0]][0].Time)
	}
	Logger.Printf("Total: %d", count)
}

func PrintHours(t, base time.Time) {
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
		switch {
		case h >= 0 && h < 3:
			PrintBlock(fmt.Sprintf("%02d", h), "#FFFFFF", "#0A4B78", h == x)
		case h >= 3 && h < 7:
			PrintBlock(fmt.Sprintf("%02d", h), "#FFFFFF", "#135E96", h == x)
		case h >= 7 && h < 9:
			PrintBlock(fmt.Sprintf("%02d", h), "#646970", "#b8e6bf", h == x)
		case h >= 9 && h < 15:
			PrintBlock(fmt.Sprintf("%02d", h), "#000000", "#68de7c", h == x)
		case h >= 15 && h < 17:
			PrintBlock(fmt.Sprintf("%02d", h), "#646970", "#00ba37", h == x)
		case h >= 17 && h < 22:
			PrintBlock(fmt.Sprintf("%02d", h), "#FFFFFF", "#9ec2e6", h == x)
		case h >= 22 && h < 24:
			PrintBlock(fmt.Sprintf("%02d", h), "#FFFFFF", "#0A4B78", h == x)
		}
		h++
	}
	fmt.Println()

}

func PrintBlock(hour string, fg, bg string, highlight bool) {
	var style = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		PaddingLeft(0).
		PaddingRight(0)

	var normal = lipgloss.NewStyle().
		Foreground(lipgloss.Color(fg)).
		Background(lipgloss.Color(bg)).
		PaddingLeft(2).
		PaddingRight(2)

	if highlight {
		// hour = fmt.Sprintf("%s", aurora.Index(0, aurora.BgIndex(229, hour)))
		hour = style.Render(hour)
	} else {
		// hour = fmt.Sprintf("%s", aurora.Index(fg, aurora.BgIndex(bg, hour)))
		hour = normal.Render(hour)
	}
	fmt.Printf("%s", hour)
	// fmt.Printf("%s%s%s", aurora.Index(fg, aurora.BgIndex(bg, "  ")), hour, aurora.Index(fg, aurora.BgIndex(bg, "  ")))
}
