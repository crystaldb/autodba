package metrics

import (
	"fmt"
)

// type TimeSeries map[int64]float64
// type MetricsRecord map[string]float64
// type AggregatedMetrics map[int64]MetricsRecord
// AggregatedMetrics == map[int64]map[string]float64

type Repository interface {
	Execute(query string, options map[string]string) (*map[int64]float64, error)
}

type Service interface {
	Execute(metrics map[string]string, options map[string]string) (map[int64]map[string]float64, error)
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

		for time, value := range *timeSeries {
			if _, ok := aggregate[time]; !ok {
				aggregate[time] = make(map[string]float64)
			}
			aggregate[time][metric] = value
		}
	}

	return aggregate, nil
}
