package usage

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var schemas = []string{

	`CREATE TABLE IF NOT EXISTS user (
		user_id INTEGER PRIMARY KEY,
		username TEXT NOT NULL UNIQUE,
		password TEXT NOT NULL
	)`,
	`CREATE TABLE IF NOT EXISTS days (
		day_id INTEGER PRIMARY KEY,
		user_id INTEGER NOT NULL,
		timestamp TEXT NOT NULL,
		consumption INTEGER NOT NULL,
		temperature INTEGER NOT NULL,
		FOREIGN KEY(user_id) REFERENCES user(user_id) ON DELETE CASCADE
	)`,
	`CREATE TABLE IF NOT EXISTS months (
		month_id INTEGER PRIMARY KEY,
		user_id INTEGER NOT NULL,
		timestamp TEXT NOT NULL,
		consumption INTEGER NOT NULL,
		temperature INTEGER NOT NULL,
		FOREIGN KEY(user_id) REFERENCES user(user_id) ON DELETE CASCADE
	)`,
}

type UsageStorage struct {
	DB *sql.DB
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

func (storage UsageStorage) AddNewUser(userId int, username string, password string) error {

	q := `INSERT INTO user(user_id, username, password) VALUES (?, ?, ?)`
	_, err := storage.DB.Exec(q, userId, username, password)
	return err
}

func (storage UsageStorage) GetUser(username string, password string) (User, error) {

	user := User{}

	q := `SELECT user_id, username, password FROM user WHERE username=? AND password=?`
	err := storage.DB.QueryRow(q, username, password).Scan(&user.UserId,
		&user.UserName,
		&user.Password)

	if err != nil {
		return User{}, err
	}

	return user, nil
}

func (storage UsageStorage) AddDailyLimit(userId, dayId, temperature, consumption int, timestamp string) error {

	q := `INSERT INTO days (user_id, day_id, timestamp, consumption, temperature) VALUES (?, ?, ?, ?, ?)`

	_, err := storage.DB.Exec(q, userId, dayId, timestamp, consumption, temperature)
	return err
}

func (storage UsageStorage) AddMonthlyLimit(userId, monthId, temperature, consumption int, timestamp string) error {

	q := `INSERT INTO months (user_id, month_id, timestamp, consumption, temperature) VALUES (?, ?, ?, ?, ?)`

	_, err := storage.DB.Exec(q, userId, monthId, timestamp, consumption, temperature)
	return err
}

func (storage UsageStorage) GetDailyLimits(userId int) (Limits, error) {

	fmt.Printf("Received request to fetch the daily limits for the user: %d\n", userId)

	q := `SELECT COALESCE(min(timestamp), "0001-01-01 00:00:00"), COALESCE(max(timestamp), "0001-01-01 00:00:00"),
	COALESCE(min(consumption), 0), COALESCE(max(consumption), 0),
	COALESCE(min(temperature), 0), COALESCE(max(temperature), 0) from days where user_id = ?`

	mmTimestamp := MinMaxTimestamp{}
	mmConsumption := MinMaxConsumption{}
	mmTemperature := MinMaxTemperature{}

	var timestampMin []byte
	var timestampMax []byte

	err := storage.DB.QueryRow(q, userId).Scan(&timestampMin, &timestampMax,
		&mmConsumption.Minimum, &mmConsumption.Maximum,
		&mmTemperature.Minimum, &mmTemperature.Maximum)

	if err != nil && err != sql.ErrNoRows {
		return Limits{}, err
	}

	t, _ := time.Parse("2006-01-02 15:04:05", string(timestampMin))
	mmTimestamp.Minimum = t.Format("2006-01-02")

	t, _ = time.Parse("2006-01-02 15:04:05", string(timestampMax))
	mmTimestamp.Maximum = t.Format("2006-01-02")

	return Limits{
		MinMaxTimestamp:   mmTimestamp,
		MinMaxConsumption: mmConsumption,
		MinMaxTemperature: mmTemperature,
	}, nil
}

func (storage UsageStorage) GetMonthlyLimits(userId int) (Limits, error) {

	fmt.Printf("Received request to fetch monthly limits for the user: %d\n", userId)

	q := `SELECT COALESCE(min(timestamp), "0001-01-01 00:00:00"), COALESCE(max(timestamp), "0001-01-01 00:00:00"),
	COALESCE(min(consumption), 0), COALESCE(max(consumption), 0),
	COALESCE(min(temperature), 0), COALESCE(max(temperature), 0) from months where user_id = ?`

	mmTimestamp := MinMaxTimestamp{}
	mmConsumption := MinMaxConsumption{}
	mmTemperature := MinMaxTemperature{}

	var timestampMin []byte
	var timestampMax []byte

	err := storage.DB.QueryRow(q, userId).Scan(&timestampMin, &timestampMax,
		&mmConsumption.Minimum, &mmConsumption.Maximum,
		&mmTemperature.Minimum, &mmTemperature.Maximum)

	if err != nil && err != sql.ErrNoRows {
		return Limits{}, err
	}

	t, _ := time.Parse("2006-01-02 15:04:05", string(timestampMin))
	mmTimestamp.Minimum = t.Format("2006-01-02")

	t, _ = time.Parse("2006-01-02 15:04:05", string(timestampMax))
	mmTimestamp.Maximum = t.Format("2006-01-02")

	return Limits{
		MinMaxTimestamp:   mmTimestamp,
		MinMaxConsumption: mmConsumption,
		MinMaxTemperature: mmTemperature,
	}, nil
}

func (storage UsageStorage) GetMonthlyUserData(userId int, count int, start string) ([][]interface{}, error) {

	var response [][]interface{}

	q := `SELECT timestamp, temperature, consumption from months WHERE user_id = ? and timestamp >= ? LIMIT ?`
	rows, err := storage.DB.Query(q, userId, start, count)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {

		var temperature, consumption int
		var timestamp []byte

		if err := rows.Scan(&timestamp, &temperature, &consumption); err != nil {
			return nil, err
		}

		t, _ := time.Parse("2006-01-02 15:04:05", string(timestamp))

		response = append(response, []interface{}{
			t.Format("2006-01-02"),
			temperature,
			consumption,
		})
	}

	return response, err
}

func (storage UsageStorage) GetDailyUserData(userId int, count int, start string) ([][]interface{}, error) {

	var response [][]interface{}

	q := `SELECT timestamp, temperature, consumption from days WHERE user_id = ? and timestamp >= ? LIMIT ?`
	rows, err := storage.DB.Query(q, userId, start, count)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {

		var temperature, consumption int
		var timestamp []byte

		if err := rows.Scan(&timestamp, &temperature, &consumption); err != nil {
			return nil, err
		}

		t, _ := time.Parse("2006-01-02 15:04:05", string(timestamp))

		response = append(response, []interface{}{
			t.Format("2006-01-02"),
			temperature,
			consumption,
		})
	}

	return response, err
}
