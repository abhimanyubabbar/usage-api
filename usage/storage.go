package usage

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var schemas = []string{

	`CREATE TABLE IF NOT EXISTS days (
		day_id INTEGER PRIMARY KEY,
		user_id INTEGER NOT NULL,
		timestamp TEXT NOT NULL,
		consumption INTEGER NOT NULL,
		temperature INTEGER NOT NULL
	)`,
	`CREATE TABLE IF NOT EXISTS months (
		month_id INTEGER PRIMARY KEY,
		user_id INTEGER NOT NULL,
		timestamp TEXT NOT NULL,
		consumption INTEGER NOT NULL,
		temperature INTEGER NOT NULL
	)`,
	`CREATE TABLE IF NOT EXISTS user (
		user_id INTEGER PRIMARY KEY,
		username TEXT NOT NULL,
		password TEXT NOT NULL
	)`,
}

type UsageStorage struct {
	db *sql.DB
}

func connectToDB(location string) (*sql.DB, error) {

	db, err := sql.Open("sqlite3", location)

	if err != nil {
		return nil, err
	}

	return db, db.Ping()
}

func NewStorage(location string) (UsageStorage, error) {

	fmt.Println("Received request to create the storage layer")

	// Stage 1: Get hold of connection to the database.
	db, err := connectToDB(location)

	if err != nil {
		return UsageStorage{},
			fmt.Errorf("Unable to create the storage layer: %s", err.Error())
	}

	// Stage 2: Load all the schemas in the database.
	for _, schema := range schemas {

		if _, err := db.Exec(schema); err != nil {
			return UsageStorage{},
				fmt.Errorf("Unable to create the schema in the storage layer: %s, error: %s", schema, err.Error())
		}
	}

	return UsageStorage{db}, nil
}

func (storage UsageStorage) GetDailyLimits(userId int) (Limits, error) {

	fmt.Printf("Received request to fetch the daily limits for the user: %d\n", userId)

	q := `SELECT min(timestamp), max(timestamp), 
	min(consumption), max(consumption), 
	min(temperature), max(temperature) from days where user_id = ?`

	mmTimestamp := MinMaxTimestamp{}
	mmConsumption := MinMaxConsumption{}
	mmTemperature := MinMaxTemperature{}

	var timestampMin []byte
	var timestampMax []byte

	err := storage.db.QueryRow(q, userId).Scan(&timestampMin, &timestampMax,
		&mmConsumption.Minimum, &mmConsumption.Maximum,
		&mmTemperature.Minimum, &mmTemperature.Maximum)

	if err != nil && err != sql.ErrNoRows {
		return Limits{}, err
	}

	// mmTimestamp.Minimum, _:= time.Parse("2006-01-02 15:04:05", string(timestampMin))
	// mmTimestamp.Maximum, _:= time.Parse("2006-01-02 15:04:05", string(timestampMax))

	mmTimestamp.Minimum, _ = time.Parse("2006-01-02 15:04:05", string(timestampMin))
	mmTimestamp.Maximum, _ = time.Parse("2006-01-02 15:04:05", string(timestampMax))

	return Limits{
		MinMaxTimestamp:   mmTimestamp,
		MinMaxConsumption: mmConsumption,
		MinMaxTemperature: mmTemperature,
	}, nil
}

func (storage UsageStorage) GetMonthlyLimits(userId int) (Limits, error) {

	fmt.Printf("Received request to fetch monthly limits for the user: %d\n", userId)

	q := `SELECT min(timestamp), max(timestamp), 
	min(consumption), max(consumption), 
	min(temperature), max(temperature) from months where user_id = ?`

	mmTimestamp := MinMaxTimestamp{}
	mmConsumption := MinMaxConsumption{}
	mmTemperature := MinMaxTemperature{}

	var timestampMin []byte
	var timestampMax []byte

	err := storage.db.QueryRow(q, userId).Scan(&timestampMin, &timestampMax,
		&mmConsumption.Minimum, &mmConsumption.Maximum,
		&mmTemperature.Minimum, &mmTemperature.Maximum)

	if err != nil && err != sql.ErrNoRows {
		return Limits{}, err
	}

	mmTimestamp.Minimum, _ = time.Parse("2006-01-02 15:04:05", string(timestampMin))
	mmTimestamp.Maximum, _ = time.Parse("2006-01-02 15:04:05", string(timestampMax))

	return Limits{
		MinMaxTimestamp:   mmTimestamp,
		MinMaxConsumption: mmConsumption,
		MinMaxTemperature: mmTemperature,
	}, nil
}
