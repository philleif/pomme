package sparkline

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"strings"
)

var blocks = []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

// Braille patterns for vertical bar (bottom to top fill)
// Each braille char is 2x4 dots. We use left column only for single-width bars.
var braille = []rune{
	'⠀', // 0 - empty
	'⣀', // 1 - bottom 2 dots
	'⣤', // 2 - bottom 4 dots
	'⣶', // 3 - bottom 6 dots
	'⣿', // 4 - all 8 dots (full)
}

// Extended braille for 8-level resolution (matching block chars)
var brailleExt = []rune{
	'⠀', // 0
	'⢀', // 1
	'⣀', // 2
	'⣠', // 3
	'⣤', // 4
	'⣴', // 5
	'⣶', // 6
	'⣾', // 7
	'⣿', // 8 - full
}

// Subscript digits for compact display
var subscripts = []string{"₀", "₁", "₂", "₃", "₄", "₅", "₆", "₇", "₈", "₉", "₁₀", "₁₁", "₁₂"}

type Style int

const (
	StyleBlock Style = iota
	StyleBraille
	StyleKittyGraphics
)

type Options struct {
	MaxVal       int
	Style        Style
	ShowLabels   bool
	ShowValues   bool
	HighlightIdx int // -1 for no highlight (today is typically last index)
	ShowGoal     bool
	Width        int // For Kitty graphics: pixel width
	Height       int // For Kitty graphics: pixel height
}

func DefaultOptions() Options {
	return Options{
		MaxVal:       12,
		Style:        StyleBraille,
		ShowLabels:   false,
		ShowValues:   false,
		HighlightIdx: -1,
		ShowGoal:     false,
		Width:        140, // 7 days * 20px
		Height:       20,
	}
}

type Output struct {
	Sparkline string
	Labels    string
	Values    string
	Highlight string
}

func Generate(values []int, maxVal int) string {
	if len(values) == 0 {
		return ""
	}

	if maxVal <= 0 {
		maxVal = 1
		for _, v := range values {
			if v > maxVal {
				maxVal = v
			}
		}
	}

	result := make([]rune, len(values))
	for i, v := range values {
		if v <= 0 {
			result[i] = blocks[0]
		} else if v >= maxVal {
			result[i] = blocks[len(blocks)-1]
		} else {
			idx := (v * (len(blocks) - 1)) / maxVal
			result[i] = blocks[idx]
		}
	}

	return string(result)
}

func GenerateBraille(values []int, maxVal int) string {
	if len(values) == 0 {
		return ""
	}

	if maxVal <= 0 {
		maxVal = 12
	}

	result := make([]rune, len(values))
	levels := len(brailleExt) - 1

	for i, v := range values {
		if v <= 0 {
			result[i] = brailleExt[0]
		} else if v >= maxVal {
			result[i] = brailleExt[levels]
		} else {
			idx := (v * levels) / maxVal
			result[i] = brailleExt[idx]
		}
	}

	return string(result)
}

// GenerateBrailleSpaced creates a spaced braille sparkline for alignment with 3-char columns
func GenerateBrailleSpaced(values []int, maxVal int) string {
	if len(values) == 0 {
		return ""
	}

	if maxVal <= 0 {
		maxVal = 12
	}

	var result strings.Builder
	levels := len(brailleExt) - 1

	for i, v := range values {
		var char rune
		if v <= 0 {
			char = brailleExt[0]
		} else if v >= maxVal {
			char = brailleExt[levels]
		} else {
			idx := (v * levels) / maxVal
			char = brailleExt[idx]
		}
		result.WriteRune(char)
		// Add spacing to align with 3-char columns (braille char + 2 spaces)
		if i < len(values)-1 {
			result.WriteString("  ")
		}
	}

	return result.String()
}

func GenerateEnhanced(values []int, opts Options) Output {
	out := Output{}

	if len(values) == 0 {
		return out
	}

	maxVal := opts.MaxVal
	if maxVal <= 0 {
		maxVal = 12
	}

	switch opts.Style {
	case StyleBraille:
		out.Sparkline = GenerateBraille(values, maxVal)
	case StyleKittyGraphics:
		out.Sparkline = GenerateKittyGraphics(values, maxVal, opts.Width, opts.Height)
	default:
		out.Sparkline = Generate(values, maxVal)
	}

	if opts.ShowLabels {
		days := []string{"M", "T", "W", "T", "F", "S", "S"}
		// Align labels with sparkline chars (2 spaces per char for braille width)
		var labels []string
		for i := 0; i < len(values) && i < 7; i++ {
			labels = append(labels, days[(7-len(values)+i)%7])
		}
		out.Labels = strings.Join(labels, " ")
	}

	if opts.ShowValues {
		var vals []string
		for _, v := range values {
			if v < 10 {
				vals = append(vals, fmt.Sprintf("%d", v))
			} else {
				vals = append(vals, fmt.Sprintf("%d", v))
			}
		}
		out.Values = strings.Join(vals, " ")
	}

	if opts.HighlightIdx >= 0 && opts.HighlightIdx < len(values) {
		spaces := strings.Repeat(" ", opts.HighlightIdx*2)
		out.Highlight = spaces + "↑"
	}

	return out
}

func FromDayStats(intervals []int, goal int) string {
	return GenerateBraille(intervals, goal)
}

func Subscript(n int) string {
	if n >= 0 && n < len(subscripts) {
		return subscripts[n]
	}
	if n >= 10 {
		return fmt.Sprintf("₁%s", subscripts[n%10])
	}
	return fmt.Sprintf("%d", n)
}

// GenerateKittyGraphics creates a pixel-based sparkline using Kitty graphics protocol
// This works in Ghostty and Kitty terminals
func GenerateKittyGraphics(values []int, maxVal, width, height int) string {
	if len(values) == 0 {
		return ""
	}

	if maxVal <= 0 {
		maxVal = 12
	}

	barWidth := width / len(values)
	if barWidth < 2 {
		barWidth = 2
	}

	// Create RGBA pixel data
	pixels := make([]byte, width*height*4)

	// Fill with transparent background
	for i := 0; i < len(pixels); i += 4 {
		pixels[i] = 0   // R
		pixels[i+1] = 0 // G
		pixels[i+2] = 0 // B
		pixels[i+3] = 0 // A (transparent)
	}

	// Draw bars
	for i, v := range values {
		barHeight := 0
		if v > 0 {
			barHeight = (v * height) / maxVal
			if barHeight > height {
				barHeight = height
			}
		}

		x0 := i * barWidth
		x1 := x0 + barWidth - 1
		if x1 >= width {
			x1 = width - 1
		}

		// Color: tomato red for work intervals
		r, g, b := byte(255), byte(99), byte(71)

		for y := height - barHeight; y < height; y++ {
			for x := x0; x < x1; x++ {
				idx := (y*width + x) * 4
				if idx+3 < len(pixels) {
					pixels[idx] = r
					pixels[idx+1] = g
					pixels[idx+2] = b
					pixels[idx+3] = 255 // opaque
				}
			}
		}

		// Draw goal line (at maxVal height) - subtle gray
		goalY := 0 // top of the sparkline area
		for x := x0; x < x1; x++ {
			idx := (goalY*width + x) * 4
			if idx+3 < len(pixels) {
				pixels[idx] = 100
				pixels[idx+1] = 100
				pixels[idx+2] = 100
				pixels[idx+3] = 128
			}
		}
	}

	return encodeKittyGraphics(pixels, width, height)
}

func encodeKittyGraphics(pixels []byte, width, height int) string {
	encoded := base64.StdEncoding.EncodeToString(pixels)

	var buf bytes.Buffer
	// Kitty graphics protocol: APC sequence
	// a=T (action=transmit and display)
	// f=32 (format=RGBA)
	// s=width, v=height
	// m=1 means more data follows, m=0 means last chunk
	
	// For simplicity, send in one chunk if small enough
	// Max chunk size is typically 4096 bytes
	chunkSize := 4096
	
	for i := 0; i < len(encoded); i += chunkSize {
		end := i + chunkSize
		if end > len(encoded) {
			end = len(encoded)
		}
		chunk := encoded[i:end]
		
		m := 1
		if end >= len(encoded) {
			m = 0
		}
		
		if i == 0 {
			// First chunk: include all parameters
			buf.WriteString(fmt.Sprintf("\x1b_Ga=T,f=32,s=%d,v=%d,m=%d;%s\x1b\\", width, height, m, chunk))
		} else {
			// Continuation chunk
			buf.WriteString(fmt.Sprintf("\x1b_Gm=%d;%s\x1b\\", m, chunk))
		}
	}

	return buf.String()
}

// CompactStatus generates a compact status line for menu bar / tmux
// Format: "🍅 18:32 ⣀⣤⣶⣿⣷⣄⣿₈"
func CompactStatus(icon, time, sparkline string, todayCount int) string {
	return fmt.Sprintf("%s %s %s%s", icon, time, sparkline, Subscript(todayCount))
}
