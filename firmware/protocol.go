package pcb

type Protocol int

type (
	SPI     Protocol
	JTAG    Protocol
	ICSP    Protocol
	I2C     Protocol
	Serial  Protocol
	OneWire Protocol // NOTE: Is this not serial?

	USB  Protocol
	GPIO Protocol
	UART Protocol
)

const (
	Tx UART = iota
	Rx
)

const (
	MISO SPI = iota
)

// TODO: We need to describe that both a pin is a UART pin, but it is also a
// sub-pin type of that UART protocol, or SPI protocol.
