package usage

import "time"

type MinMaxTimestamp struct {
	Minimum time.Time `json:"minimum"`
	Maximum time.Time `json:"maximum"`
}

type MinMaxConsumption struct {
	Minimum int `json:"minimum"`
	Maximum int `json:"maximum"`
}

type MinMaxTemperature struct {
	Minimum int `json:"minimum"`
	Maximum int `json:"maximum"`
}

type Limits struct {
	MinMaxTimestamp   `json:"timestamp"`
	MinMaxConsumption `json:"consumption"`
	MinMaxTemperature `json:"temperature"`
}

type DailyMonthlyLimits struct {
	Daily   Limits `json:"daily"`
	Monthly Limits `json:"monthly"`
}
