package weather

// Client will have methods to return weather related data
type Client interface {
	Current(loc Location) (*Conditions, error)
	Hourly(loc Location) (*HourlyForecast, error)
}

type Location struct {
	State string
	City  string
}

type HourlyForecast struct {
	Conditions []Conditions
}

type Conditions struct {
	TempC         int
	TempF         int
	Humidity      int
	Precipitation int
}
