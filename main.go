package main

import (
	"fmt"
	"log"
	"time"

	"github.com/mhoppenstedt/gospi/mcp23s17"
	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/physic"
	"periph.io/x/periph/conn/spi"
	"periph.io/x/periph/host"
	"periph.io/x/periph/host/bcm283x"
	"periph.io/x/periph/host/sysfs"
)

// chipAddrU* constants hold the hardware address associated with each MCP chip.
const (
	chipAddrU1 = 0x40 // U1 Matrix 1 - Bank 1
	chipAddrU2 = 0x42 // U2 Matrix 1 - Bank 2
	chipAddrU3 = 0x44 // U3 Matrix 1 - Bank 3
	chipAddrU4 = 0x48 // U4 Matrix 2 - Bank 1
	chipAddrU5 = 0x4A // U5 Matrix 2 - Bank 2
	chipAddrU6 = 0x4C // U6 Matrix 2 - Bank 3
	ioConByte  = 0x0C // 0000 1100 (Bit 3 == IoCon.HaEn to enable address pins, Bit 2 == ODR which configures the INT pin as open drain)
)

// spiAddressMap is a slice of all chipAddr variables used for looping.
var spiAddressMap = []uint8{
	0x40, // b01000000 // U1 Matrix 1 - Bank 1
	0x42, // b01000010 // U2 Matrix 1 - Bank 2
	0x44, // b01000100 // U3 Matrix 1 - Bank 3
	0x48, // b01001000 // U4 Matrix 2 - Bank 1
	0x4A, // b01001010 // U5 Matrix 2 - Bank 2
	0x4C, // b01001100 // U6 Matrix 2 - Bank 3
}

type spiIoDriver struct {
	conn spi.Conn
}

func main() {
	// Setup the RPI3
	_, err := host.Init()
	if err != nil {
		log.Fatalf("failed to initialize periph: %v", err)
	}

	// Initializes the MCP chips by setting the RESET line to low then back high.
	resetMCPChips()

	// Create a new SPI port.
	spiPort, err := sysfs.NewSPI(0, 0)
	if err != nil {
		fmt.Printf("failed to init spi port %v", err)
		return
	}

	// Create a new SPI connection.
	spiDriver := spiIoDriver{}
	spiDriver.conn, err = spiPort.Connect(100*physic.KiloHertz, spi.Mode1, 8)
	if err != nil {
		fmt.Printf("failed to connect to spi port %v", err)
		return
	}

	// Configure MCP chips by changing register values
	spiDriver.configureMCPChips()

	// Loop through chipAddr* to set relays to latched.
	for pin := 0; pin < 16; pin++ {
		spiDriver.setRelayState(chipAddrU6, uint8(pin), 0x01)
		time.Sleep(time.Millisecond * time.Duration(100))
	}

	// Loop through chipAddr* to set relays to not latched.
	for pin := 0; pin < 16; pin++ {
		spiDriver.setRelayState(chipAddrU6, uint8(pin), 0x0)
		time.Sleep(time.Millisecond * time.Duration(100))
	}

	// Close SPI port
	err = spiPort.Close()
	if err != nil {
		fmt.Printf("failed to close spi port %v", err)
		return
	}
}

// resetMCPChips send an low signal to both the Health LED and IO_RESET lines then a high signal.
// This toggles the LED once to show activity and resets all MCP chips to default values.
func resetMCPChips() {
	// Reset all of the MCP IO drivers
	bcm283x.GPIO2.Out(gpio.Low) // IO Reset to cycle power to MCP chips
	bcm283x.GPIO5.Out(gpio.Low) // Health LED on fixture
	time.Sleep(time.Millisecond * time.Duration(500))
	bcm283x.GPIO2.Out(gpio.High) // IO Reset to cycle power to MCP chips
	bcm283x.GPIO5.Out(gpio.High) //Health LED on fixture
}

// configureMCPChips loops through spiAddressMap to set the IOCON byte and
// to set the direction of IoDirA and IoDirB both to outputs.
func (s *spiIoDriver) configureMCPChips() {
	fmt.Println("Configuring MCP chips using IOCON byte.")
	for _, addr := range spiAddressMap {
		s.spiWrite(addr, mcp23s17.IoCon, ioConByte)
	}

	fmt.Println("Configuring MCP chips for output using IoDirA/B byte.")
	for _, addr := range spiAddressMap {
		s.spiWrite(addr, mcp23s17.IoDirA, 0x00)
		s.spiWrite(addr, mcp23s17.IoDirB, 0x00)
	}
}

// setRelayState sets the relays state to on or latched.
func (s *spiIoDriver) setRelayState(address, pin, value uint8) {
	var ioDirReg uint8
	var olatReg uint8

	// Checks to see if pin passed in is less than or equal to 7. If so need to subtract 8
	// in order to use 0-7 on PORTB instead of PORTA. Sets register addresses based on pin value.
	if pin >= 0 && pin <= 7 {
		ioDirReg = mcp23s17.IoDirA
		olatReg = mcp23s17.OLatA
	} else if pin >= 8 && pin < 16 {
		pin = pin % 8
		ioDirReg = mcp23s17.IoDirB
		olatReg = mcp23s17.OLatB
	} else {
		fmt.Printf("Pin number is incorrect. Wanting 0-15. Recieved %v\n", pin)
		return
	}

	// Checks to see if IoDirA/B register is set for output.
	ioDirRegValue, _ := s.spiRead(address, ioDirReg)
	if ioDirRegValue != 0 {
		fmt.Printf("Address %#x, direction register incorrect (not set for output) value: %#x\n", address, ioDirRegValue)
		return
	}

	// setMask is used to set a 1 value in a specific place the OLatA/B register
	// clearMask is the inverse of setMask and is used to set a 0 value in a specific place the OLatA/B register
	setMask := 0x01 << pin
	clearMask := setMask ^ 0xFF

	// Must read the value in the OLatA/B register to see what current bits are set. This way we only change the bit we want
	prevRegValue, _ := s.spiRead(address, olatReg)

	// If the value being set is 0, then we "clear" the bit by ANDing clearMask and prevRegValue
	// If the value being set is not 0 (will on only ever be 1 or 0), then we "set" the bit by ANDing setMask with prevRegValue
	var newRegValue uint8
	if value == 0 {
		newRegValue = prevRegValue & uint8(clearMask)
	} else {
		newRegValue = prevRegValue | uint8(setMask)
	}

	// Write the new value to the OLAT register
	fmt.Printf("Writing value %#x to address %#x register %#x\n", newRegValue, address, olatReg)
	fmt.Printf("Register %#x OLD value was %#x\n", olatReg, prevRegValue)
	s.spiWrite(address, olatReg, newRegValue)
	newRegValue, _ = s.spiRead(address, olatReg)
	fmt.Printf("Register %#x NEW value is  %#x\n", olatReg, newRegValue)
}

// spiWrite writes value to the register of a chip using address.
func (s *spiIoDriver) spiWrite(address, register, value uint8) error {
	write := []byte{address, register, value}
	read := make([]byte, len(write))
	if err := s.conn.Tx(write, read); err != nil {
		return err
	}
	time.Sleep(1 * time.Millisecond)
	return nil
}

// spiRead reads the value of register of a chip using address.
func (s *spiIoDriver) spiRead(address, register uint8) (uint8, error) {
	write := []byte{address | mcp23s17.ReadCmd, register, 0xDB} // 0xDB is a dummy byte needed for a READ of a register
	read := make([]byte, len(write))
	if err := s.conn.Tx(write, read); err != nil {
		return 0, err
	}
	// fmt.Printf("Add[%#x]\tReg:%#x\t\tVal:%#x\n", write[0]&0xFE, write[1], read[2])
	time.Sleep(1 * time.Millisecond)
	return read[2], nil
}
