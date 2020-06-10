package pcb

// NOTE: We are building an overlay layer over TinyGo to provide the
// functionality we feel better models working with hardware without requiring
// changes to TinyGo so they can act as both a proposal and a heavier weight
// alternative approach that makes working with hardware, at least for
// Multiverse OS developers, feel more natural.

// This will be used for all of our open soure harware firmware.

// NOTE: Rather than talk about boards, we are going to deal with chips. Chips
// have pinouts, we can use the datasheets to automatically generate pinout
// information. And we will use chips with development boards, which take those
// pinouts and extend or implement the pinout of the chips.
//
// This will result in a bit heaiver code, but it will more realistically
// reflect the hardware we are working with; also it will support greater
// variety of development breakout boards, custom boards, and most importantly,
// make that stuff natural and easy to work with, without needing to build it
// form scratch everytime. We want to call in our breakoutboard and chip, and
// then start describing what sensors are plugged in, what data its sending over
// what pins and how.
//
// This may also allow us to begin to explore ways to representing FPGA designs
// as well. Which may end up being like building up from a logical NAND gate
// array into actual computers.
type PCB struct {
	Chips []*Chip
}

type Chip struct {
	Pins []*Pin
}

type PinType int

const (
	Digital PinType = iota
	Analog
)

type PinFlow int

const (
	Input PinFlow = iota
	Output
)

// NOTE: Pins have a lot of possible layered types when you look at pinouts. We
// have physical pins that layer ontop of that digital/analog/pwm and further
// USB, SPI, UARt, etc. We want to have this be natural to work with and simple
// to implement quickly.

type Pin struct {
	Type PinType
}

type Protocol int

type UART Protocol

const (
	Tx UART = iota
	Rx
)

// TODO: We need to describe that both a pin is a UART pin, but it is also a
// sub-pin type of that UART protocol, or SPI protocol.
