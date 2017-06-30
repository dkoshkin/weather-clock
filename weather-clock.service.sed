[Unit]
Description=weather-clock uses PIGPIO to control GPIO setting clock or temp modes
Requires=pigpio

[Service]
Environment=WU_API_KEY=__WU_API_KEY__
Type=simple
Restart=always
RestartSec=3
ExecStart=/home/pi/weather-clock

[Install]
WantedBy=multi-user.target