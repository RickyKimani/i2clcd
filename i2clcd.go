/*Written by Ricky Kimani*/
package i2clcd

import (
	"machine"
	"time"
)

type I2CLCD struct {
	bus       *machine.I2C
	addr      uint8
	cols      uint8
	rows      uint8
	backlight bool
	display   bool
	cursor    bool
	blink     bool
}

const (
	// commands
	LCD_CLEARDISPLAY   = 0x01
	LCD_RETURNHOME     = 0x02
	LCD_ENTRYMODESET   = 0x04
	LCD_DISPLAYCONTROL = 0x08
	LCD_CURSORSHIFT    = 0x10
	LCD_FUNCTIONSET    = 0x20
	LCD_SETCGRAMADDR   = 0x40
	LCD_SETDDRAMADDR   = 0x80

	// flags for display entry mode
	LCD_ENTRYRIGHT          = 0x00
	LCD_ENTRYLEFT           = 0x02
	LCD_ENTRYSHIFTINCREMENT = 0x01
	LCD_ENTRYSHIFTDECREMENT = 0x00

	// flags for display on/off control
	LCD_DISPLAYON  = 0x04
	LCD_DISPLAYOFF = 0x00
	LCD_CURSORON   = 0x02
	LCD_CURSOROFF  = 0x00
	LCD_BLINKON    = 0x01
	LCD_BLINKOFF   = 0x00

	// backlight control
	LCD_BACKLIGHT   = 0x08
	LCD_NOBACKLIGHT = 0x00

	// flags for display/cursor shift
	LCD_MOVELEFT    = 0x00
	LCD_MOVERIGHT   = 0x04
	LCD_SCROLLLEFT  = 0x18
	LCD_SCROLLRIGHT = 0x1C
)

// Create a new I2CLCD instance
func NewI2CLCD(bus *machine.I2C, addr, cols, rows uint8) *I2CLCD {
	return &I2CLCD{
		bus:       bus,
		addr:      addr,
		cols:      cols,
		rows:      rows,
		backlight: true,
		display:   true,
		cursor:    false,
		blink:     false,
	}
}

// Send a command to the LCD
func (lcd *I2CLCD) sendCommand(cmd byte) {
	lcd.send(cmd, 0)
}

// Send data to the LCD
func (lcd *I2CLCD) sendData(data byte) {
	lcd.send(data, 1)
}

// Send a byte to the LCD
func (lcd *I2CLCD) send(value byte, mode byte) {
	highNibble := value & 0xF0
	lowNibble := (value << 4) & 0xF0
	lcd.write4Bits(highNibble | mode)
	lcd.write4Bits(lowNibble | mode)
}

// Write 4 bits to the LCD
func (lcd *I2CLCD) write4Bits(value byte) {
	lcd.expanderWrite(value)
	lcd.pulseEnable(value)
}

// Write a byte to the I2C expander
func (lcd *I2CLCD) expanderWrite(data byte) {
	backlight := byte(0x00)
	if lcd.backlight {
		backlight = LCD_BACKLIGHT
	}
	lcd.bus.Tx(uint16(lcd.addr), []byte{data | backlight}, nil)
}

// Pulse the enable line
func (lcd *I2CLCD) pulseEnable(data byte) {
	lcd.expanderWrite(data | 0x04) // Enable bit high
	time.Sleep(1 * time.Millisecond)
	lcd.expanderWrite(data & ^byte(0x04)) // Enable bit low
	time.Sleep(1 * time.Millisecond)
}

// Initialize the LCD
func (lcd *I2CLCD) Init() {
	time.Sleep(50 * time.Millisecond) // Allow time for power-on

	// Initialize display
	lcd.sendCommand(0x03)
	time.Sleep(5 * time.Millisecond)
	lcd.sendCommand(0x03)
	time.Sleep(5 * time.Millisecond)
	lcd.sendCommand(0x03)
	time.Sleep(1 * time.Millisecond)
	lcd.sendCommand(0x02)

	var functionSet byte = LCD_FUNCTIONSET | 0x20 // Basic command set
	if lcd.rows > 1 {
		functionSet |= 0x08 // 2-line mode
	}
	lcd.sendCommand(functionSet)

	lcd.sendCommand(LCD_DISPLAYCONTROL | LCD_DISPLAYON)
	lcd.sendCommand(LCD_ENTRYMODESET | LCD_ENTRYLEFT) // Ensure text displays correctly
	lcd.sendCommand(LCD_CLEARDISPLAY)
	time.Sleep(2 * time.Millisecond)

	lcd.Backlight()
}

// Clear the display
func (lcd *I2CLCD) Clear() {
	lcd.sendCommand(LCD_CLEARDISPLAY)
	time.Sleep(2 * time.Millisecond)
}

// Return the cursor to the home position
func (lcd *I2CLCD) Home() {
	lcd.sendCommand(LCD_RETURNHOME)
	time.Sleep(2 * time.Millisecond)
}

// Print text to the LCD
func (lcd *I2CLCD) Print(text string) {
	for _, char := range text {
		lcd.sendData(byte(char))
	}
}

// Set the cursor position
func (lcd *I2CLCD) SetCursor(col, row uint8) {
	if row >= lcd.rows {
		row = lcd.rows - 1 // Clamp to max row
	}
	addr := col + (row * 0x40)
	lcd.sendCommand(LCD_SETDDRAMADDR | addr)
}

// Turn the display on
func (lcd *I2CLCD) DisplayOn() {
	lcd.display = true
	lcd.sendCommand(LCD_DISPLAYCONTROL | LCD_DISPLAYON)
}

// Turn the display off
func (lcd *I2CLCD) DisplayOff() {
	lcd.display = false
	lcd.sendCommand(LCD_DISPLAYCONTROL | LCD_DISPLAYOFF)
}

// Turn the cursor on
func (lcd *I2CLCD) CursorOn() {
	lcd.cursor = true
	lcd.sendCommand(LCD_DISPLAYCONTROL | LCD_DISPLAYON | LCD_CURSORON)
}

// Turn the cursor off
func (lcd *I2CLCD) CursorOff() {
	lcd.cursor = false
	lcd.sendCommand(LCD_DISPLAYCONTROL | LCD_DISPLAYON | LCD_CURSOROFF)
}

// Turn the cursor blink on
func (lcd *I2CLCD) BlinkOn() {
	lcd.blink = true
	lcd.sendCommand(LCD_DISPLAYCONTROL | LCD_DISPLAYON | LCD_BLINKON)
}

// Turn the cursor blink off
func (lcd *I2CLCD) BlinkOff() {
	lcd.blink = false
	lcd.sendCommand(LCD_DISPLAYCONTROL | LCD_DISPLAYON | LCD_BLINKOFF)
}

// Turn the backlight on
func (lcd *I2CLCD) Backlight() {
	lcd.backlight = true
	lcd.expanderWrite(0x00) // Refresh backlight setting
}

// Turn the backlight off
func (lcd *I2CLCD) NoBacklight() {
	lcd.backlight = false
	lcd.expanderWrite(0x00) // Refresh backlight setting
}

// Create a custom character
func (lcd *I2CLCD) CreateChar(location byte, charmap []byte) {
	location &= 0x07 // We only have 8 locations 0-7
	lcd.sendCommand(LCD_SETCGRAMADDR | (location << 3))
	for i := 0; i < 8; i++ {
		lcd.sendData(charmap[i])
	}
}

func (lcd *I2CLCD) ScrollDisplayLeft() {
	lcd.sendCommand(LCD_SCROLLLEFT)
}

func (lcd *I2CLCD) ScrollDisplayRight() {
	lcd.sendCommand(LCD_SCROLLRIGHT)
}

func (lcd *I2CLCD) LeftToRight() {
	lcd.sendCommand(LCD_ENTRYMODESET | LCD_ENTRYLEFT)
}

func (lcd *I2CLCD) RightToLeft() {
	lcd.sendCommand(LCD_ENTRYMODESET | LCD_ENTRYRIGHT)
}

func (lcd *I2CLCD) ShiftIncrement() {
	lcd.sendCommand(LCD_ENTRYMODESET | LCD_ENTRYSHIFTINCREMENT)
}

func (lcd *I2CLCD) ShiftDecrement() {
	lcd.sendCommand(LCD_ENTRYMODESET | LCD_ENTRYSHIFTDECREMENT)
}

func (lcd *I2CLCD) Autoscroll() {
	lcd.sendCommand(LCD_ENTRYMODESET | LCD_ENTRYSHIFTINCREMENT)
}

func (lcd *I2CLCD) NoAutoscroll() {
	lcd.sendCommand(LCD_ENTRYMODESET | LCD_ENTRYSHIFTDECREMENT)
}
