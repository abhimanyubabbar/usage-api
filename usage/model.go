package usage

type User struct {
	UserId   int    `db:"user_id"`
	UserName string `db:"username"`
	Password string `db:"password"`
}

type UserData struct {
	Timestamp   string `json:"timestamp"`
	Temperature int    `json:"temperature"`
	Consumption int    `json:"consumption"`
}

type MinMaxTimestamp struct {
	Minimum string `json:"minimum"`
	Maximum string `json:"maximum"`
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
