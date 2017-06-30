package weather

import (
	"encoding/json"
	"fmt"
	"net/http"
)

const url = "http://api.wunderground.com/api/%s/hourly/q/%s/%s.json"

// wuClient implements Client using https://www.wunderground.com/
type wuClient struct {
	apiKey string
}

// NewWUClient returns a Client backed by wunderground
func NewWUClient(key string) Client {
	return &wuClient{apiKey: key}
}

// Hourly return current data using https://www.wunderground.com/weather/api/d/docs?d=data/hourly&MR=1
func (c *wuClient) Hourly(loc Location) (*HourlyForecast, error) {
	reqURL := fmt.Sprintf(url, c.apiKey, loc.State, loc.City)
	fmt.Println(reqURL)
	resp, err := http.Get(reqURL)
	if err != nil {
		return nil, err
	}
	// Note: Weather Underground responds with 200 even if API key is wrong
	// TODO detect using wrong api key
	wu := WUHourly{}
	err = json.NewDecoder(resp.Body).Decode(&wu)
	if err != nil {
		return nil, err
	}

	hr := &HourlyForecast{}
	for _, h := range wu.HourlyForecast {
		hr.Conditions = append(hr.Conditions, Conditions{
			TempC:         h.Temp.Metric,
			TempF:         h.Temp.English,
			Humidity:      h.Humidity,
			Precipitation: h.Pop,
		})
	}
	return hr, nil
}

func (c *wuClient) Current(loc Location) (*Conditions, error) {
	hourly, err := c.Hourly(loc)
	if err != nil {
		return nil, fmt.Errorf("could not get hourly conditions: %v", err)
	}

	if len(hourly.Conditions) == 0 {
		return nil, fmt.Errorf("could not get current conditions, empty forecast")
	}
	// the first element is tha latest conditions
	hr := hourly.Conditions[0]
	current := &Conditions{
		TempC:         hr.TempC,
		TempF:         hr.TempF,
		Humidity:      hr.Humidity,
		Precipitation: hr.Precipitation,
	}

	return current, nil
}

type WUHourly struct {
	Response struct {
		Version        string `json:"version"`
		TermsofService string `json:"termsofService"`
		Features       struct {
			Hourly int `json:"hourly"`
		} `json:"features"`
	} `json:"response"`
	HourlyForecast []struct {
		// FCTTIME struct {
		// 	Hour                   string `json:"hour"`
		// 	HourPadded             string `json:"hour_padded"`
		// 	Min                    string `json:"min"`
		// 	MinUnpadded            string `json:"min_unpadded"`
		// 	Sec                    string `json:"sec"`
		// 	Year                   string `json:"year"`
		// 	Mon                    string `json:"mon"`
		// 	MonPadded              string `json:"mon_padded"`
		// 	MonAbbrev              string `json:"mon_abbrev"`
		// 	Mday                   string `json:"mday"`
		// 	MdayPadded             string `json:"mday_padded"`
		// 	Yday                   string `json:"yday"`
		// 	Isdst                  string `json:"isdst"`
		// 	Epoch                  string `json:"epoch"`
		// 	Pretty                 string `json:"pretty"`
		// 	Civil                  string `json:"civil"`
		// 	MonthName              string `json:"month_name"`
		// 	MonthNameAbbrev        string `json:"month_name_abbrev"`
		// 	WeekdayName            string `json:"weekday_name"`
		// 	WeekdayNameNight       string `json:"weekday_name_night"`
		// 	WeekdayNameAbbrev      string `json:"weekday_name_abbrev"`
		// 	WeekdayNameUnlang      string `json:"weekday_name_unlang"`
		// 	WeekdayNameNightUnlang string `json:"weekday_name_night_unlang"`
		// 	Ampm                   string `json:"ampm"`
		// 	Tz                     string `json:"tz"`
		// 	Age                    string `json:"age"`
		// 	UTCDATE                string `json:"UTCDATE"`
		// } `json:"FCTTIME"`
		Temp struct {
			English int `json:"english,string"`
			Metric  int `json:"metric,string"`
		} `json:"temp"`
		// Dewpoint struct {
		// 	English int `json:"english"`
		// 	Metric  int `json:"metric"`
		// } `json:"dewpoint"`
		// Condition string `json:"condition"`
		// Icon      string `json:"icon"`
		// IconURL   string `json:"icon_url"`
		// Fctcode   int    `json:"fctcode"`
		// Sky       string `json:"sky"`
		// Wspd      struct {
		// 	English int `json:"english"`
		// 	Metric  int `json:"metric"`
		// } `json:"wspd"`
		// Wdir struct {
		// 	Dir     string `json:"dir"`
		// 	Degrees int    `json:"degrees"`
		// } `json:"wdir"`
		// Wx        string `json:"wx"`
		// Uvi       int    `json:"uvi"`
		Humidity int `json:"humidity,string"`
		// Windchill struct {
		// 	English int `json:"english"`
		// 	Metric  int `json:"metric"`
		// } `json:"windchill"`
		// Heatindex struct {
		// 	English int `json:"english"`
		// 	Metric  int `json:"metric"`
		// } `json:"heatindex"`
		// Feelslike struct {
		// 	English int `json:"english"`
		// 	Metric  int `json:"metric"`
		// } `json:"feelslike"`
		// Qpf struct {
		// 	English float32 `json:"english"`
		// 	Metric  int     `json:"metric"`
		// } `json:"qpf"`
		// Snow struct {
		// 	English float32 `json:"english"`
		// 	Metric  string  `json:"metric"`
		// } `json:"snow"`
		Pop int `json:"pop,string"`
		// Mslp struct {
		// 	English float32 `json:"english"`
		// 	Metric  int     `json:"metric"`
		// } `json:"mslp"`
	} `json:"hourly_forecast"`
}
