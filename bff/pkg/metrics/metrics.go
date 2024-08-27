package metrics

import (
	"fmt"
)

// type TimeSeries map[int64]float64
// type MetricsRecord map[string]float64
// type AggregatedMetrics map[int64]MetricsRecord
// AggregatedMetrics == map[int64]map[string]float64

type Repository interface {
	Execute(query string, options map[string]string) (*map[int64]map[string]float64, error)
	ExecuteRaw(query string, options map[string]string) ([]map[string]interface{}, error)
}

type Service interface {
	Execute(metrics map[string]string, options map[string]string) (map[int64]map[string]float64, error)
	ExecuteRaw(query string, options map[string]string) ([]map[string]interface{}, error)
}

type service_imp struct {
	repo Repository
}

func CreateService(r Repository) Service {
	return service_imp{r}
}

func (s service_imp) Execute(metrics map[string]string, options map[string]string) (map[int64]map[string]float64, error) {
	fmt.Println("Executing metrics queries")

	aggregate := make(map[int64]map[string]float64)
	// errors := make(map[string]error)

	// TODO do this in paralell, make sure accessing the map is threadsafe
	// TODO Or find out if the prometheus lib can aggregate them itself
	for metric, query_string := range metrics {
		timeSeries, err := s.repo.Execute(query_string, options)
		if err != nil {
			fmt.Printf("Error executing query for metric: %s, %s\n", metric, err)
			return aggregate, err
		}

		var isSingleSeries bool

		for _, value := range *timeSeries {
			// if the record object returned at each time only contains one string float pair, ie 1 value per time
			if len(value) == 1 {
				isSingleSeries = true
			}
			break
		}

		for time, record := range *timeSeries {
			if _, ok := aggregate[time]; !ok {
				aggregate[time] = make(map[string]float64)
			}

			for label, value := range record {
				if isSingleSeries {
					aggregate[time][metric] = value
				} else {
					aggregate[time][label] = value
				}
			}
		}

	}

	return aggregate, nil
}

func (s service_imp) ExecuteRaw(query string, options map[string]string) ([]map[string]interface{}, error) {

	result, err := s.repo.ExecuteRaw(query, options)
	if err != nil {
		fmt.Printf("Error executing query %s\n", err)
		return nil, err
	}

	return result, nil
}
