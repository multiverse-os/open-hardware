// +build atsamd51

package ili9341

import (
	"machine"
	"runtime/volatile"
)

type parallelDriver struct {
	d0 machine.Pin
	wr machine.Pin

	setPort *uint32
	setMask uint32

	clrPort *uint32
	clrMask uint32

	wrPortSet *uint32
	wrMaskSet uint32

	wrPortClr *uint32
	wrMaskClr uint32
}

func NewParallel(d0, wr, dc, cs, rst, rd machine.Pin) *Device {
	return &Device{
		dc:  dc,
		cs:  cs,
		rd:  rd,
		rst: rst,
		driver: &parallelDriver{
			d0: d0,
			wr: wr,
		},
	}
}

func (pd *parallelDriver) configure(config *Config) {
	output := machine.PinConfig{machine.PinOutput}
	for pin := pd.d0; pin < pd.d0+8; pin++ {
		pin.Configure(output)
		pin.Low()
	}
	pd.wr.Configure(output)
	pd.wr.High()

	pd.setPort, _ = pd.d0.PortMaskSet()
	pd.setMask = uint32(pd.d0) & 0x1f

	pd.clrPort, _ = (pd.d0).PortMaskClear()
	pd.clrMask = 0xFF << uint32(pd.d0)

	pd.wrPortSet, pd.wrMaskSet = pd.wr.PortMaskSet()
	pd.wrPortClr, pd.wrMaskClr = pd.wr.PortMaskClear()
}

//go:inline
func (pd *parallelDriver) write8(b byte) {
	volatile.StoreUint32(pd.clrPort, pd.clrMask)
	volatile.StoreUint32(pd.setPort, uint32(b)<<pd.setMask)
	volatile.StoreUint32(pd.wrPortClr, pd.wrMaskClr)
	volatile.StoreUint32(pd.wrPortSet, pd.wrMaskSet)
}

//go:inline
func (pd *parallelDriver) write16(data uint16) {
	pd.write8(byte(data >> 8))
	pd.write8(byte(data))
}

//go:inline
func (pd *parallelDriver) write16n(data uint16, n int) {
	for i := 0; i < n; i++ {
		pd.write8(byte(data >> 8))
		pd.write8(byte(data))
	}
}

//go:inline
func (pd *parallelDriver) write16sl(data []uint16) {
	for i, c := 0, len(data); i < c; i++ {
		pd.write8(byte(data[i] >> 8))
		pd.write8(byte(data[i]))
	}
}
