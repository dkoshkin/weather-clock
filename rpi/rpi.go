package rpi

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// AnalogWrite write PWM value adjusting to be between 0 and 255
func AnalogWrite(pin, value int) error {
	switch {
	case value > 255:
		value = 255
	case value < 0:
		value = 0
	}

	fmt.Printf("writing value: %d, on pin %d\n", value, pin)

	cmd := exec.Command("pigs", "p", strconv.Itoa(pin), strconv.Itoa(value))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

type Notifier struct {
	handle string
	Pipe   string
}

func NewNotifier() (*Notifier, error) {
	var outb, errb bytes.Buffer
	cmd := exec.Command("pigs", "NO")
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	if _ = cmd.Run(); errb.String() != "" {
		return nil, fmt.Errorf("error creating notifier: %s\n", strings.TrimSpace(errb.String()))
	}
	handle := strings.TrimSpace(outb.String())
	if handle == "" {
		return nil, fmt.Errorf("did not get a valid handler response: %s\n", outb.String())
	}
	pipe := filepath.Join("/dev", fmt.Sprintf("pigpio%s", handle))
	notifier := Notifier{
		handle: handle,
		Pipe:   pipe,
	}

	return &notifier, nil
}

func (n Notifier) Begin(trigger chan int, pin uint) error {
	cmd := exec.Command("pigs", "NB", n.handle, fmt.Sprintf("0x%x", 1<<pin))
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("did not get a valid handler response: %v\n", err)
	}

	f, err := os.OpenFile(n.Pipe, os.O_RDONLY, 0600)
	defer f.Close()
	if err != nil {
		return fmt.Errorf("could not open pipe: %v\n", err)
	}
	_, err = io.ReadFull(f, nil)
	if err != nil {
		return fmt.Errorf("initial read error: %v\n", err)
	}
	lastPress := time.Now()
	for {
		// 12 bytes for down, 12 bytes for up
		buffer := make([]byte, 24)
		n, err := io.ReadAtLeast(f, buffer, 24)
		//n, err := f.Read(buffer)
		if err != nil {
			fmt.Printf("could not read file: %v\n", err)
		}
		// 100ms debounce
		// bytes 8-11 hold state, 0 base slice
		// byte 8: 0-7
		// byte 9: 8-15
		// byte 10: 16:23
		// byte 11: 24:31
		// byte>>(pin%8) will be even when HIGH
		if time.Since(lastPress).Seconds() > 0.1 && n == 24 && (buffer[8+(pin/8)]>>(pin%8)%2) != 0 {
			fmt.Printf("read event: %b\n", buffer)
			trigger <- 1
			lastPress = time.Now()
		}
	}
}
