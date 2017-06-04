package usage

import "fmt"

type Config struct {
	DBLocation string
}

type UsageProcessor struct {
	Storage UsageStorage
}

func NewProcessor(config Config) (UsageProcessor, error) {

	fmt.Println("Received a request to create a new processor")

	storage, err := NewStorage(config.DBLocation)
	if err != nil {
		return UsageProcessor{}, err
	}

	return UsageProcessor{
		Storage: storage,
	}, nil
}

// GetLimitsForUser fetches the daily and monthly limits for the
// temperature, consumption and timestamp for the provided user.
func (processor UsageProcessor) GetLimitsForUser(userId int) (DailyMonthlyLimits, error) {

	fmt.Printf("Received request to fetch usage limits for the user: %d\n", userId)

	dailyLimits, err := processor.Storage.GetDailyLimits(userId)

	if err != nil {
		return DailyMonthlyLimits{}, fmt.Errorf("Unable to fetch daily limits: %s", err.Error())
	}

	monthlyLimits, err := processor.Storage.GetMonthlyLimits(userId)

	if err != nil {
		return DailyMonthlyLimits{}, fmt.Errorf("Unable to fetch monthly limits: %s", err.Error())
	}

	return DailyMonthlyLimits{
		Daily:   dailyLimits,
		Monthly: monthlyLimits,
	}, nil
}

func (processor UsageProcessor) GetDataForUser(
	userId int,
	count int,
	resolution string,
	start string) ([][]interface{}, error) {

	if resolution == "M" {
		return processor.Storage.GetMonthlyUserData(userId, count, start)
	}

	return processor.Storage.GetDailyUserData(userId, count, start)
}
