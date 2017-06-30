include keys.env

build:
	env GOOS=linux GOARCH=arm go build

sed:
	./keys.env
	sed -e 's/__WU_API_KEY__/$(WU_API_KEY)/g' weather-clock.service.sed  > weather-clock.service

deploy: build
	ssh pi@raspberrypi.local sudo systemctl stop weather-clock
	ssh pi@raspberrypi.local sudo systemctl restart pigpio
	scp weather-clock pi@raspberrypi.local:~/
	ssh pi@raspberrypi.local sudo systemctl restart weather-clock


init: sed
	scp weather-clock.service pi@raspberrypi.local:/home/pi/weather-clock.service
	ssh pi@raspberrypi.local sudo cp /home/pi/weather-clock.service /lib/systemd/system/weather-clock.service
	scp pigpio.service pi@raspberrypi.local:/home/pi/pigpio.service
	ssh pi@raspberrypi.local sudo cp /home/pi/pigpio.service /lib/systemd/system/pigpio.service
	ssh pi@raspberrypi.local sudo systemctl daemon-reload
	ssh pi@raspberrypi.local sudo systemctl enable pigpio
	ssh pi@raspberrypi.local sudo systemctl enable weather-clock
	ssh pi@raspberrypi.local sudo systemctl restart pigpio
	ssh pi@raspberrypi.local sudo systemctl restart weather-clock	