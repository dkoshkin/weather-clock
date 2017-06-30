package main

import (
	"fmt"
	"os"
	"time"

	"io"

	"github.com/dkoshkin/weather-clock/rpi"
	"github.com/dkoshkin/weather-clock/weather"
	"github.com/sirupsen/logrus"
)

// pins
const (
	btnPin   = 26
	dial1Pin = 17
	dial2Pin = 27
	dial3Pin = 22
)

// modes
const (
	CLOCK = iota
	WEATHER
	TEST
)

func main() {
	// setup log
	var log = logrus.New()
	logFile, err := os.OpenFile("weather.log", os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		log.Out = os.Stdout
		log.Warnf("Could not create log file: %v", err)
	} else {
		log.Out = io.MultiWriter(os.Stdout, logFile)
	}
	defer logFile.Close()

	// default to clock mode
	var clockChan, weatherChan chan bool
	clockChan = make(chan bool)
	go clockMode(clockChan, log)

	btnChan := make(chan int)
	// listen for button presses to change modes
	go listenButtonPress(btnChan, log)

	prevMode := CLOCK
	mode := CLOCK
	for {
		log.Infoln("LOOPING")
		pressed := <-btnChan
		// there are 3 modes, only chancge on a press
		// pressed will block until an event occurs
		mode = ((mode + pressed) % 3)
		if prevMode != mode {
			log.Infof("mode changed from %d to %d", prevMode, mode)
			// end previous mode
			switch mode {
			case CLOCK:
				clockChan = make(chan bool)
				go clockMode(clockChan, log)
			case WEATHER:
				// previous mode either CLOCK
				clockChan <- true
				weatherChan = make(chan bool)
				go weatherMode(weatherChan, log)
			case TEST:
				weatherChan <- true
				go testMode(log)
			}
			prevMode = mode
		}
	}
}

func listenButtonPress(trigger chan int, log *logrus.Logger) {
	n, err := rpi.NewNotifier()
	if err != nil {
		panic(fmt.Sprintf("could not initialize button listener: %v", err))
	}
	log.Infof("started button listener on pipe: %s", n.Pipe)
	if err := n.Begin(trigger, btnPin); err != nil {
		panic(fmt.Sprintf("could not start button listener: %v", err))
	}
}

func clockMode(done chan bool, log *logrus.Logger) {
	log.Info("starting clock mode")
	ticker := time.NewTicker(1 * time.Second)
	now := time.Now()
	if err := setTime(now.Hour(), now.Minute(), now.Second()); err != nil {
		panic(fmt.Sprintf("error setting time: %v", err))
	}
	for {
		select {
		case now := <-ticker.C:
			log.Infof("clock ticking")
			if err := setTime(now.Hour(), now.Minute(), now.Second()); err != nil {
				panic(fmt.Sprintf("error setting time: %v", err))
			}
		case <-done:
			// teardown
			// stop ticker and exit
			log.Infof("stopped clock")
			ticker.Stop()
			return
		}
	}
}

func weatherMode(done chan bool, log *logrus.Logger) {
	log.Info("starting weather mode")
	// create weather client
	key := os.Getenv("WU_API_KEY")
	if key == "" {
		panic("WU_API_KEY must be set")
	}
	c := weather.NewWUClient(key)
	// set initial weather
	tempF, _, hum, precipitation, err := getCurrentWeather(c)
	if err != nil {
		panic(err.Error())
	}
	if err := setWeather(tempF, hum, precipitation); err != nil {
		panic(fmt.Sprintf("error settint weather conditions: %v", err))
	}
	ticker := time.NewTicker(10 * time.Second)
	for {
		select {
		case <-ticker.C:
			log.Infof("weather ticking")
			tempF, _, hum, precipitation, err := getCurrentWeather(c)
			if err != nil {
				panic(err.Error())
			}
			if err := setWeather(tempF, hum, precipitation); err != nil {
				panic(fmt.Sprintf("error settint weather conditions: %v", err))
			}
		case <-done:
			// teardown
			// stop ticker and exit
			log.Infof("stopped weather")
			ticker.Stop()
			return
		}
	}
}

func testMode(log *logrus.Logger) {
	pins := []int{dial1Pin, dial2Pin, dial3Pin}
	for _, p := range pins {
		if err := rpi.AnalogWrite(p, 240); err != nil {
			log.Errorf("Error setting second: %v", err)
		}
	}
}

// 0-239 PWM
// 0~24: hour
// 0~60: min, sec
func setTime(hour, min, sec int) error {
	if err := rpi.AnalogWrite(dial1Pin, hour*10); err != nil {
		return fmt.Errorf("Error setting hour: %v", err)
	}
	if err := rpi.AnalogWrite(dial2Pin, min*4); err != nil {
		return fmt.Errorf("Error setting minute: %v", err)
	}
	if err := rpi.AnalogWrite(dial3Pin, sec*4); err != nil {
		return fmt.Errorf("Error setting second: %v", err)
	}

	return nil
}

func getCurrentWeather(client weather.Client) (tempF int, tempC int, hum int, precipitation int, err error) {
	loc := weather.Location{
		State: "NJ",
		City:  "Hoboken",
	}
	current, err := client.Current(loc)
	if err != nil {
		return 0, 0, 0, 0, err
	}

	return current.TempF, current.TempC, current.Humidity, current.Precipitation, nil
}

// 0-239 PWM
// 0~100 tempC
// -20~40 tempF
// 0~100 hum
// 0~100 precipitation
func setWeather(tempF, hum, precipitation int) error {
	if err := rpi.AnalogWrite(dial1Pin, int(float64(tempF)*2.4)); err != nil {
		return fmt.Errorf("Error setting temp: %v", err)
	}
	if err := rpi.AnalogWrite(dial2Pin, int(float64(hum)*2.4)); err != nil {
		return fmt.Errorf("Error setting humidity: %v", err)
	}
	if err := rpi.AnalogWrite(dial3Pin, int(float64(precipitation)*2.4)); err != nil {
		return fmt.Errorf("Error setting precipitation: %v", err)
	}
	return nil
}
