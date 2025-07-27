package color

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// ANSI color codes
const (
	// Reset
	Reset = "\033[0m"

	// Text styles
	Bold      = "\033[1m"
	Dim       = "\033[2m"
	Italic    = "\033[3m"
	Underline = "\033[4m"
	Blink     = "\033[5m"
	Reverse   = "\033[7m"
	Strike    = "\033[9m"

	// Foreground colors
	Black   = "\033[30m"
	Red     = "\033[31m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Blue    = "\033[34m"
	Magenta = "\033[35m"
	Cyan    = "\033[36m"
	White   = "\033[37m"

	// Bright foreground colors
	BrightBlack   = "\033[90m"
	BrightRed     = "\033[91m"
	BrightGreen   = "\033[92m"
	BrightYellow  = "\033[93m"
	BrightBlue    = "\033[94m"
	BrightMagenta = "\033[95m"
	BrightCyan    = "\033[96m"
	BrightWhite   = "\033[97m"

	// Background colors
	BgBlack   = "\033[40m"
	BgRed     = "\033[41m"
	BgGreen   = "\033[42m"
	BgYellow  = "\033[43m"
	BgBlue    = "\033[44m"
	BgMagenta = "\033[45m"
	BgCyan    = "\033[46m"
	BgWhite   = "\033[47m"

	// Bright background colors
	BgBrightBlack   = "\033[100m"
	BgBrightRed     = "\033[101m"
	BgBrightGreen   = "\033[102m"
	BgBrightYellow  = "\033[103m"
	BgBrightBlue    = "\033[104m"
	BgBrightMagenta = "\033[105m"
	BgBrightCyan    = "\033[106m"
	BgBrightWhite   = "\033[107m"
)

// Color represents a color with optional styling
type Color struct {
	foreground string
	background string
	bold       bool
	dim        bool
	italic     bool
	underline  bool
	blink      bool
	reverse    bool
	strike     bool
}

// New creates a new Color instance
func New() *Color {
	return &Color{}
}

// SetForeground sets the foreground color
func (c *Color) SetForeground(color string) *Color {
	c.foreground = color
	return c
}

// SetBackground sets the background color
func (c *Color) SetBackground(color string) *Color {
	c.background = color
	return c
}

// SetBold sets or unsets bold styling
func (c *Color) SetBold(bold bool) *Color {
	c.bold = bold
	return c
}

// SetDim sets or unsets dim styling
func (c *Color) SetDim(dim bool) *Color {
	c.dim = dim
	return c
}

// SetItalic sets or unsets italic styling
func (c *Color) SetItalic(italic bool) *Color {
	c.italic = italic
	return c
}

// SetUnderline sets or unsets underline styling
func (c *Color) SetUnderline(underline bool) *Color {
	c.underline = underline
	return c
}

// SetBlink sets or unsets blink styling
func (c *Color) SetBlink(blink bool) *Color {
	c.blink = blink
	return c
}

// SetReverse sets or unsets reverse styling
func (c *Color) SetReverse(reverse bool) *Color {
	c.reverse = reverse
	return c
}

// SetStrike sets or unsets strikethrough styling
func (c *Color) SetStrike(strike bool) *Color {
	c.strike = strike
	return c
}

// buildCode builds the ANSI escape code based on current settings
func (c *Color) buildCode() string {
	var codes []string

	if c.foreground != "" {
		codes = append(codes, strings.TrimPrefix(c.foreground, "\033["))
		codes[len(codes)-1] = strings.TrimSuffix(codes[len(codes)-1], "m")
	}

	if c.background != "" {
		codes = append(codes, strings.TrimPrefix(c.background, "\033["))
		codes[len(codes)-1] = strings.TrimSuffix(codes[len(codes)-1], "m")
	}

	if c.bold {
		codes = append(codes, "1")
	}
	if c.dim {
		codes = append(codes, "2")
	}
	if c.italic {
		codes = append(codes, "3")
	}
	if c.underline {
		codes = append(codes, "4")
	}
	if c.blink {
		codes = append(codes, "5")
	}
	if c.reverse {
		codes = append(codes, "7")
	}
	if c.strike {
		codes = append(codes, "9")
	}

	if len(codes) == 0 {
		return ""
	}

	return "\033[" + strings.Join(codes, ";") + "m"
}

// Sprint returns the colored string
func (c *Color) Sprint(text string) string {
	if !isColorEnabled() {
		return text
	}

	code := c.buildCode()
	if code == "" {
		return text
	}

	return code + text + Reset
}

// Print prints the colored text
func (c *Color) Print(text string) {
	fmt.Print(c.Sprint(text))
}

// Println prints the colored text with a newline
func (c *Color) Println(text string) {
	fmt.Println(c.Sprint(text))
}

// Printf prints formatted colored text
func (c *Color) Printf(format string, args ...interface{}) {
	fmt.Print(c.Sprint(fmt.Sprintf(format, args...)))
}

// Convenience functions for common colors with optional bold
func RedText(text string, bold bool) string {
	c := New().SetForeground(Red).SetBold(bold)
	return c.Sprint(text)
}

func GreenText(text string, bold bool) string {
	c := New().SetForeground(Green).SetBold(bold)
	return c.Sprint(text)
}

func BlueText(text string, bold bool) string {
	c := New().SetForeground(Blue).SetBold(bold)
	return c.Sprint(text)
}

func YellowText(text string, bold bool) string {
	c := New().SetForeground(Yellow).SetBold(bold)
	return c.Sprint(text)
}

func MagentaText(text string, bold bool) string {
	c := New().SetForeground(Magenta).SetBold(bold)
	return c.Sprint(text)
}

func CyanText(text string, bold bool) string {
	c := New().SetForeground(Cyan).SetBold(bold)
	return c.Sprint(text)
}

func WhiteText(text string, bold bool) string {
	c := New().SetForeground(White).SetBold(bold)
	return c.Sprint(text)
}

// RGB creates a color from RGB values (0-255)
func RGB(r, g, b int) string {
	return fmt.Sprintf("\033[38;2;%d;%d;%dm", r, g, b)
}

// BgRGB creates a background color from RGB values (0-255)
func BgRGB(r, g, b int) string {
	return fmt.Sprintf("\033[48;2;%d;%d;%dm", r, g, b)
}

// Hex creates a color from hex string (e.g., "#FF0000" or "FF0000")
func Hex(hex string) string {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) != 6 {
		return ""
	}

	r, err := strconv.ParseInt(hex[0:2], 16, 0)
	if err != nil {
		return ""
	}
	g, err := strconv.ParseInt(hex[2:4], 16, 0)
	if err != nil {
		return ""
	}
	b, err := strconv.ParseInt(hex[4:6], 16, 0)
	if err != nil {
		return ""
	}

	return RGB(int(r), int(g), int(b))
}

// BgHex creates a background color from hex string
func BgHex(hex string) string {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) != 6 {
		return ""
	}

	r, err := strconv.ParseInt(hex[0:2], 16, 0)
	if err != nil {
		return ""
	}
	g, err := strconv.ParseInt(hex[2:4], 16, 0)
	if err != nil {
		return ""
	}
	b, err := strconv.ParseInt(hex[4:6], 16, 0)
	if err != nil {
		return ""
	}

	return BgRGB(int(r), int(g), int(b))
}

// isColorEnabled checks if color output is supported/enabled
func isColorEnabled() bool {
	// Check if NO_COLOR environment variable is set
	if os.Getenv("NO_COLOR") != "" {
		return false
	}

	// Check if FORCE_COLOR is set
	if os.Getenv("FORCE_COLOR") != "" {
		return true
	}

	// Check if we're in a terminal
	term := os.Getenv("TERM")
	if term == "" || term == "dumb" {
		return false
	}

	// Check if stdout is a terminal (basic check)
	if fileInfo, _ := os.Stdout.Stat(); fileInfo != nil {
		return (fileInfo.Mode() & os.ModeCharDevice) != 0
	}

	return true
}

// DisableColor globally disables color output
func DisableColor() {
	os.Setenv("NO_COLOR", "1")
}

// EnableColor globally enables color output
func EnableColor() {
	os.Unsetenv("NO_COLOR")
	os.Setenv("FORCE_COLOR", "1")
}
