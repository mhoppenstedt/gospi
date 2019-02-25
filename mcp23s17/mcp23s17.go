package mcp23s17

// Address Pins
const defaultAddress = 0x40

// Register addresses for MCP chips
const (
	IoDirA   = 0x00 // I/O direction A
	IoDirB   = 0x01 // I/O direction B
	IPolA    = 0x02 // I/O polarity A
	IPolB    = 0x03 // I/O polarity B
	GpIntEnA = 0x04 // interupt enable A
	GpIntEnB = 0x05 // interupt enable B
	DefValA  = 0x06 // register default value A (interupts)
	DefValB  = 0x07 // register default value B (interupts)
	IntConA  = 0x08 // interupt control A
	IntConB  = 0x09 // interupt control B
	IoCon    = 0x0A // I/O config (also 0xB)
	GpPuA    = 0x0C // port A pullups
	GpPuB    = 0x0D // port B pullups
	IntFA    = 0x0E // interupt flag A (where the interupt came from)
	IntFB    = 0x0F // interupt flag B
	IntCapA  = 0x10 // interupt capture A (value at interupt is saved here)
	IntCapB  = 0x11 // interupt capture B
	GpIoA    = 0x12 // port A
	GpIoB    = 0x13 // port B
	OLatA    = 0x14 // output latch A
	OLatB    = 0x15 // output latch B
)

// I/O config for MCP chips
const (
	BankOff      = 0x00 // addressing mode
	BankOn       = 0x80
	IntMirrorOn  = 0x40 // interupt mirror (INTa|INTb)
	IntMirrorOff = 0x00
	SeqOpOff     = 0x20 // incrementing address pointer
	SeqOpOn      = 0x00
	DisSlwOn     = 0x10 // slew rate
	DisSlwOff    = 0x00
	HaEnOn       = 0x08 // hardware addressing
	HaEnOff      = 0x00
	ODrOn        = 0x04 // open drain for interupts
	ODrOff       = 0x00
	IntPolHigh   = 0x02 // interupt polarity
	IntPolLow    = 0x00
)

// Commands
const (
	WriteCmd = 0
	ReadCmd  = 1
)

// Nibbles
const (
	LowerNibble = 0
	UpperNibble = 1
)
