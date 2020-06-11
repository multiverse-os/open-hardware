// +build atsamd21

package hub75

import (
	"device/arm"
	"device/sam"
	"machine"
	"runtime/volatile"
	"unsafe"
)

const dmaDescriptors = 2

//go:align 16
var dmaDescriptorSection [dmaDescriptors]dmaDescriptor

//go:align 16
var dmaDescriptorWritebackSection [dmaDescriptors]dmaDescriptor

type chipSpecificSettings struct {
	bus          *machine.SPI
	dmaChannel   uint8
	timer        *sam.TCC_Type
	timerChannel *volatile.Register32
}

type dmaDescriptor struct {
	btctrl   uint16
	btcnt    uint16
	srcaddr  unsafe.Pointer
	dstaddr  unsafe.Pointer
	descaddr unsafe.Pointer
}

func (d *Device) configureChip(dataPin, clockPin machine.Pin) {
	d.dmaChannel = 0
	d.bus = &machine.SPI0      // must be SERCOM0
	const triggerSource = 0x02 // SERCOM0_DMAC_ID_TX
	d.bus.Configure(machine.SPIConfig{
		Frequency: 8000000,
		Mode:      0,
	})

	// Enable DMA IRQ.
	arm.EnableIRQ(sam.IRQ_DMAC)

	// Init DMAC.
	// First configure the clocks, then configure the DMA descriptors. Those
	// descriptors must live in SRAM and must be aligned on a 16-byte boundary.
	// http://www.lucadavidian.com/2018/03/08/wifi-controlled-neo-pixels-strips/
	// https://svn.larosterna.com/oss/trunk/arduino/zerotimer/zerodma.cpp
	sam.PM.AHBMASK.SetBits(sam.PM_AHBMASK_DMAC_)
	sam.PM.APBBMASK.SetBits(sam.PM_APBBMASK_DMAC_)
	sam.DMAC.BASEADDR.Set(uint32(uintptr(unsafe.Pointer(&dmaDescriptorSection))))
	sam.DMAC.WRBADDR.Set(uint32(uintptr(unsafe.Pointer(&dmaDescriptorWritebackSection))))

	// Enable peripheral with all priorities.
	sam.DMAC.CTRL.SetBits(sam.DMAC_CTRL_DMAENABLE | sam.DMAC_CTRL_LVLEN0 | sam.DMAC_CTRL_LVLEN1 | sam.DMAC_CTRL_LVLEN2 | sam.DMAC_CTRL_LVLEN3)

	// Configure channel descriptor.
	dmaDescriptorSection[d.dmaChannel] = dmaDescriptor{
		btctrl: (1 << 0) | // VALID: Descriptor Valid
			(0 << 3) | // BLOCKACT=NOACT: Block Action
			(1 << 10) | // SRCINC: Source Address Increment Enable
			(0 << 11) | // DSTINC: Destination Address Increment Enable
			(1 << 12) | // STEPSEL=SRC: Step Selection
			(0 << 13), // STEPSIZE=X1: Address Increment Step Size
		btcnt:   24, // beat count
		dstaddr: unsafe.Pointer(&d.bus.Bus.DATA.Reg),
	}

	// Add channel.
	sam.DMAC.CHID.Set(d.dmaChannel)
	sam.DMAC.CHCTRLA.SetBits(sam.DMAC_CHCTRLA_SWRST)
	sam.DMAC.CHCTRLB.Set((sam.DMAC_CHCTRLB_LVL_LVL0 << sam.DMAC_CHCTRLB_LVL_Pos) | (sam.DMAC_CHCTRLB_TRIGACT_BEAT << sam.DMAC_CHCTRLB_TRIGACT_Pos) | (triggerSource << sam.DMAC_CHCTRLB_TRIGSRC_Pos))

	// Enable DMA block transfer complete interrupt.
	sam.DMAC.CHINTENSET.SetBits(sam.DMAC_CHINTENSET_TCMPL)

	// Next up, configure the timer/counter used for precisely timing the Output
	// Enable pin.
	// d.oe == D4 == PA07
	// PA07 is on TCC1 CC[1]
	machine.InitPWM()
	pwm := machine.PWM{d.oe}
	pwm.Configure()
	d.timer = sam.TCC1
	d.timerChannel = &d.timer.CC1

	// Enable an interrupt on CC1 match.
	d.timer.INTENSET.Set(sam.TCC_INTENSET_MC1)
	arm.EnableIRQ(sam.IRQ_TCC1)

	// Set to one-shot and count down.
	d.timer.CTRLBSET.SetBits(sam.TCC_CTRLBSET_ONESHOT | sam.TCC_CTRLBSET_DIR)
	for d.timer.SYNCBUSY.HasBits(sam.TCC_SYNCBUSY_CTRLB) {
	}

	// Enable TCC output.
	d.timer.CTRLA.SetBits(sam.TCC_CTRLA_ENABLE)
	for d.timer.SYNCBUSY.HasBits(sam.TCC_SYNCBUSY_ENABLE) {
	}
}

// startTransfer starts the SPI transaction to send the next row to the screen.
func (d *Device) startTransfer() {
	bitstring := d.displayBitstrings[d.row][d.colorBit]

	// For some reason, you have to provide the address just past the end of the
	// array instead of the address of the array.
	dmaDescriptorSection[d.dmaChannel].srcaddr = unsafe.Pointer(uintptr(unsafe.Pointer(&bitstring[0])) + uintptr(len(bitstring)))

	// Start the transfer.
	sam.DMAC.CHCTRLA.SetBits(sam.DMAC_CHCTRLA_ENABLE)
}

// startOutputEnableTimer will enable and disable the screen for a very short
// time, depending on which bit is currently enabled.
func (d *Device) startOutputEnableTimer() {
	// Multiplying the brightness by 3 to be consistent with the nrf52 driver
	// (48MHz vs 16MHz).
	count := (d.brightness * 3) << d.colorBit
	d.timerChannel.Set(0xffff - count)
	for d.timer.SYNCBUSY.HasBits(sam.TCC_SYNCBUSY_CC0 | sam.TCC_SYNCBUSY_CC1 | sam.TCC_SYNCBUSY_CC2 | sam.TCC_SYNCBUSY_CC3) {
	}
	d.timer.CTRLBSET.Set(sam.TCC_CTRLBSET_CMD_RETRIGGER << sam.TCC_CTRLBSET_CMD_Pos)
}

//go:export DMAC_IRQHandler
func dmacHandler() {
	// Clear interrupt flags, otherwise this interrupt will trigger
	// continuously.
	sam.DMAC.CHINTFLAG.Set(sam.DMAC_CHINTENCLR_TERR | sam.DMAC_CHINTENCLR_TCMPL | sam.DMAC_CHINTENCLR_SUSP)

	display.handleSPIEvent()
}

//go:export TCC1_IRQHandler
func tcc1Handler() {
	// Clear the interrupt flag.
	sam.TCC1.INTFLAG.Set(sam.TCC_INTFLAG_MC1)

	display.handleTimerEvent()
}
