package tz

default_group: string

group: [string]: {
	actor: [ID=_]: {
		name:          string | *ID
		city?:         string
		country_code?: string
		time_zone?:    string
	}
}
