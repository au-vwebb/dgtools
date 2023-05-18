package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/DavidGamba/dgtools/run"
	"github.com/DavidGamba/go-getoptions"
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

	cities := opt.NewCommand("cities", "filter cities list")
	cities.Bool("all", false, opt.Alias("a"), opt.Description("Show all cities"))
	cities.String("country-code", "", opt.Alias("c"), opt.Description("Filter by country code"))
	cities.SetCommandFn(CitiesRun)

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

	p := NewPalette("BlueYellow")
	PrintActors(am, short, p)
	return nil
}

func CitiesRun(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	Logger.Printf("Cities")
	all := opt.Value("all").(bool)
	countryCode := opt.Value("country-code").(string)

	if len(args) == 0 && !all {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", "Missing city name")
		fmt.Fprintf(os.Stderr, "%s", opt.Help(getoptions.HelpSynopsis))
		return getoptions.ErrorHelpCalled
	}

	nameQuery := strings.Join(args, " ")
	cc := NewCities()
	_, err := cc.Search(nameQuery, countryCode)
	if err != nil {
		return fmt.Errorf("failed search: %w", err)
	}

	return nil
}
