// Package tmp102 implements a driver for the TMP102 digital temperature sensor.
//
// Datasheet: https://download.mikroe.com/documents/datasheets/tmp102-data-sheet.pdf

package tmp102 // import "tinygo.org/x/drivers/tmp102"

import (
	"machine"
)

// Device holds the already configured I2C bus and the address of the sensor.
type Device struct {
	bus     machine.I2C
	address uint8
}

// Config is the configuration for the TMP102.
type Config struct {
	Address uint8
}

// New creates a new TMP102 connection. The I2C bus must already be configured.
func New(bus machine.I2C) Device {
	return Device{
		bus: bus,
	}
}

// Configure initializes the sensor with the given parameters.
func (d *Device) Configure(cfg Config) {
	if cfg.Address == 0 {
		cfg.Address = Address
	}

	d.address = cfg.Address
}

// Reads the temperature from the sensor and returns it in celsius milli degrees (°C/1000).
func (d *Device) ReadTemperature() (temperature int32, err error) {

	tmpData := make([]byte, 2)

	err = d.bus.ReadRegister(d.address, RegTemperature, tmpData)

	if err != nil {
		return
	}

	temperatureSum := int32((int16(tmpData[0])<<8 | int16(tmpData[1])) >> 4)

	if (temperatureSum & int32(1<<11)) == int32(1<<11) {
		temperatureSum |= int32(0xf800)
	}

	temperature = temperatureSum * 625

	return temperature / 10, nil
}
