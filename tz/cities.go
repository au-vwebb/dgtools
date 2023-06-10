package main

import (
	"embed"
	"encoding/csv"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

//go:embed cities-tz.tsv
var f embed.FS

type City struct {
	Name        string
	Admin1Name  string // Admin division 1
	CountryCode string
	TimeZone    string
	Population  string
}

// CityMap is a map of city names to City.
type CityMap struct {
	m      map[string][]*City
	loaded bool
}

func NewCities() *CityMap {
	return &CityMap{
		m:      make(map[string][]*City),
		loaded: false,
	}
}

// Returns a list of cities with the given name and optionally a country code.
func (c *CityMap) Get(name, countryCode string) ([]*City, error) {
	if !c.loaded {
		err := c.load()
		if err != nil {
			return []*City{}, fmt.Errorf("failed to load cities table: %w", err)
		}
	}
	cities, ok := c.m[name]
	if !ok {
		return []*City{}, fmt.Errorf("no cities found for '%s'", name)
	}
	if len(cities) <= 1 || countryCode == "" {
		return cities, nil
	}
	cc := []*City{}
	for _, city := range cities {
		if strings.EqualFold(countryCode, city.CountryCode) {
			cc = append(cc, city)
		}
	}
	return cc, nil
}

func (c *CityMap) Search(name, countryCode string) ([]City, error) {
	if !c.loaded {
		err := c.load()
		if err != nil {
			return []City{}, fmt.Errorf("failed to load cities table: %w", err)
		}
	}
	count := 0
	cc := []*City{}
	for n, cities := range c.m {
		for _, city := range cities {
			count++
			if strings.Contains(strings.ToLower(n), strings.ToLower(name)) {
				if countryCode == "" || strings.EqualFold(countryCode, city.CountryCode) {
					cc = append(cc, city)
				}
			}
		}
	}
	sort.Slice(cc, func(i, j int) bool {
		return cc[i].Name < cc[j].Name
	})
	PrintCities(cc)
	return []City{}, nil
}

func (c *CityMap) load() error {
	if c.loaded {
		return nil
	}

	tableFilename := "cities-tz.tsv"
	tableFH, err := f.Open(tableFilename)
	if err != nil {
		return fmt.Errorf("failed to open '%s': %w", tableFilename, err)
	}
	defer tableFH.Close()

	r := csv.NewReader(tableFH)
	r.Comma = '\t'
	r.FieldsPerRecord = -1
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read table: %w", err)
		}

		// Column widths are known:
		// select max(length(asciiname)) from admin1;
		// 37, 40, 2, 30, 8
		// name, admin1Name, countryCode, timeZone
		// Logger.Printf("%#v\n", record)
		p := message.NewPrinter(language.English)

		pop, err := strconv.Atoi(record[4])
		if err != nil {
			return fmt.Errorf("failed to parse population: %w", err)
		}
		population := p.Sprintf("%d\n", pop)
		if c.m[record[0]] == nil {
			c.m[record[0]] = []*City{{
				Name:        record[0],
				Admin1Name:  record[1],
				CountryCode: record[2],
				TimeZone:    record[3],
				Population:  population,
			}}
		} else {
			c.m[record[0]] = append(c.m[record[0]], &City{
				Name:        record[0],
				Admin1Name:  record[1],
				CountryCode: record[2],
				TimeZone:    record[3],
				Population:  population,
			})
		}
	}
	c.loaded = true

	return nil
}

func PrintCities(cities []*City) {
	light := lipgloss.NewStyle().Foreground(lipgloss.Color("#000000")).Background(lipgloss.Color("#eef2f3"))
	dark := lipgloss.NewStyle().Foreground(lipgloss.Color("#000000")).Background(lipgloss.Color("#dce3e6"))
	for i, city := range cities {
		if i%2 == 0 {
			fmt.Printf("%s%s%s%s%s\n",
				light.
					Width(41).
					PaddingLeft(1).
					PaddingRight(1).
					Render(city.Name),
				light.
					Width(42).
					PaddingLeft(1).
					PaddingRight(1).
					Render(city.Admin1Name),
				light.
					Width(2).
					PaddingLeft(1).
					PaddingRight(1).
					Render(city.CountryCode),
				light.
					Width(32).
					PaddingLeft(1).
					PaddingRight(1).
					Render(city.TimeZone),
				light.
					Width(12).
					PaddingLeft(1).
					PaddingRight(1).
					Render(city.Population),
			)
		} else {
			fmt.Printf("%s%s%s%s%s\n",
				dark.
					Width(41).
					PaddingLeft(1).
					PaddingRight(1).
					Render(city.Name),
				dark.
					Width(42).
					PaddingLeft(1).
					PaddingRight(1).
					Render(city.Admin1Name),
				dark.
					Width(2).
					PaddingLeft(1).
					PaddingRight(1).
					Render(city.CountryCode),
				dark.
					Width(32).
					PaddingLeft(1).
					PaddingRight(1).
					Render(city.TimeZone),
				dark.
					Width(12).
					PaddingLeft(1).
					PaddingRight(1).
					Render(city.Population),
			)
		}
	}
}
