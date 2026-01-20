package terminal

import (
	"bytes"
	"regexp"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/mattn/go-runewidth"
)

// runeWidthCond is configured for terminal use (East Asian Ambiguous = narrow)
var runeWidthCond = func() *runewidth.Condition {
	c := runewidth.NewCondition()
	c.EastAsianWidth = false // Treat ambiguous as narrow (width 1)
	return c
}()

// VirtualTerminal provides a virtual terminal emulator
// that converts raw PTY output with ANSI escape sequences
// into clean text for agent observation.
//
// This implementation properly handles ANSI CSI sequences for:
// - Cursor movement (CUU, CUD, CUF, CUB, CUP, etc.)
// - Line/screen clearing (ED, EL)
// - Scrolling regions
// - Alternative screen buffer
// - SGR (Select Graphic Rendition) for colors and text attributes
type VirtualTerminal struct {
	mu sync.RWMutex

	cols int
	rows int

	// Screen buffer (current visible content) - runes only for backward compatibility
	screen [][]rune

	// Styled cell buffer - cells with color and attribute information
	cells [][]Cell

	// Cursor position
	cursorX int
	cursorY int

	// Current text style (applied to new characters)
	currentFg             Color
	currentBg             Color
	currentAttrs          CellAttrs
	currentUnderlineStyle UnderlineStyle
	currentUnderlineColor Color

	// Line wrap tracking (true if line is wrapped from previous line)
	isWrapped []bool

	// History buffer (scrolled-off lines) - plain text for backward compatibility
	history    []string
	maxHistory int

	// Styled history buffer (scrolled-off lines with full style information)
	// Each entry is a row of cells, preserving colors and attributes
	historyStyled   [][]Cell
	historyIsWrapped []bool // Wrap flags for styled history lines

	// Flag to track if we've received any data
	hasData bool

	// Escape sequence parsing state
	escState    escapeState
	escBuffer   []byte
	escParams   []int
	escPrivate  byte
	escRawSeq   []byte // Raw sequence for SGR parsing with colons

	// Saved cursor position
	savedCursorX int
	savedCursorY int

	// Alternative screen buffer support
	altScreen       [][]rune
	altCells        [][]Cell
	altCursorX      int
	altCursorY      int
	useAltScreen    bool
	savedMainScreen [][]rune
	savedMainCells  [][]Cell
}

// escapeState represents the current state of escape sequence parsing
type escapeState int

const (
	stateNormal escapeState = iota
	stateEscape             // After ESC
	stateCSI                // After ESC [
	stateOSC                // After ESC ]
	stateDCS                // After ESC P
)

// ANSI escape sequence pattern (for simple stripping)
var ansiPattern = regexp.MustCompile(`\x1b\[[?>=]?[0-9;]*[a-zA-Z]|\x1b\][^\x07]*\x07|\x1b[PX^_][^\x1b]*\x1b\\`)

// NewVirtualTerminal creates a new virtual terminal
func NewVirtualTerminal(cols, rows, maxHistory int) *VirtualTerminal {
	if cols <= 0 {
		cols = 80
	}
	if rows <= 0 {
		rows = 24
	}
	if maxHistory <= 0 {
		maxHistory = 10000
	}

	vt := &VirtualTerminal{
		cols:             cols,
		rows:             rows,
		maxHistory:       maxHistory,
		history:          make([]string, 0),
		historyStyled:    make([][]Cell, 0),
		historyIsWrapped: make([]bool, 0),
	}
	vt.initScreen()
	return vt
}

// initScreen initializes/resets the screen buffer
func (vt *VirtualTerminal) initScreen() {
	vt.screen = make([][]rune, vt.rows)
	vt.cells = make([][]Cell, vt.rows)
	vt.isWrapped = make([]bool, vt.rows)
	for i := range vt.screen {
		vt.screen[i] = make([]rune, vt.cols)
		vt.cells[i] = make([]Cell, vt.cols)
		vt.isWrapped[i] = false
		for j := range vt.screen[i] {
			vt.screen[i][j] = ' '
			vt.cells[i][j] = NewCell(' ')
		}
	}
	vt.cursorX = 0
	vt.cursorY = 0
	vt.currentFg = DefaultColor()
	vt.currentBg = DefaultColor()
	vt.currentAttrs = AttrNone
	vt.currentUnderlineStyle = UnderlineNone
	vt.currentUnderlineColor = DefaultColor()
}

// Feed processes raw PTY data with proper UTF-8 support
func (vt *VirtualTerminal) Feed(data []byte) {
	vt.mu.Lock()
	defer vt.mu.Unlock()

	vt.hasData = true

	// Process data with UTF-8 awareness
	for len(data) > 0 {
		b := data[0]

		// ESC sequence or in escape state: process byte by byte
		if b == 0x1b || vt.escState != stateNormal {
			vt.processByte(b)
			data = data[1:]
			continue
		}

		// Control characters (< 0x20) and DEL (0x7f): process as single byte
		if b < 0x20 || b == 0x7f {
			vt.processByte(b)
			data = data[1:]
			continue
		}

		// Normal characters: decode UTF-8 properly
		r, size := utf8.DecodeRune(data)
		if r == utf8.RuneError && size == 1 {
			// Invalid UTF-8 byte, skip it
			data = data[1:]
			continue
		}
		vt.processChar(r)
		data = data[size:]
	}
}

// processByte processes a single byte through the state machine
func (vt *VirtualTerminal) processByte(b byte) {
	switch vt.escState {
	case stateNormal:
		if b == 0x1b { // ESC
			vt.escState = stateEscape
			vt.escBuffer = nil
			vt.escParams = nil
			vt.escPrivate = 0
			vt.escRawSeq = nil
		} else {
			vt.processChar(rune(b))
		}

	case stateEscape:
		vt.processEscapeByte(b)

	case stateCSI:
		vt.processCSI(b)

	case stateOSC:
		// OSC sequences end with BEL (0x07) or ST (ESC \)
		if b == 0x07 {
			vt.escState = stateNormal
		} else {
			vt.escBuffer = append(vt.escBuffer, b)
		}

	case stateDCS:
		// DCS sequences end with ST (ESC \)
		if b == 0x1b {
			vt.escBuffer = append(vt.escBuffer, b)
		} else if len(vt.escBuffer) > 0 && vt.escBuffer[len(vt.escBuffer)-1] == 0x1b && b == '\\' {
			vt.escState = stateNormal
		} else {
			vt.escBuffer = append(vt.escBuffer, b)
		}
	}
}

// processEscapeByte handles byte after ESC
func (vt *VirtualTerminal) processEscapeByte(b byte) {
	switch b {
	case '[': // CSI
		vt.escState = stateCSI
		vt.escParams = []int{}
	case ']': // OSC
		vt.escState = stateOSC
		vt.escBuffer = nil
	case 'P': // DCS
		vt.escState = stateDCS
		vt.escBuffer = nil
	case '7': // Save cursor (DECSC)
		vt.savedCursorX = vt.cursorX
		vt.savedCursorY = vt.cursorY
		vt.escState = stateNormal
	case '8': // Restore cursor (DECRC)
		vt.cursorX = vt.savedCursorX
		vt.cursorY = vt.savedCursorY
		vt.escState = stateNormal
	case 'c': // Reset (RIS)
		vt.initScreen()
		vt.escState = stateNormal
	case 'D': // Index (IND) - move down
		vt.cursorY++
		if vt.cursorY >= vt.rows {
			vt.scroll()
			vt.cursorY = vt.rows - 1
		}
		vt.escState = stateNormal
	case 'M': // Reverse Index (RI) - move up
		vt.cursorY--
		if vt.cursorY < 0 {
			vt.scrollDown()
			vt.cursorY = 0
		}
		vt.escState = stateNormal
	case 'E': // Next Line (NEL)
		vt.cursorX = 0
		vt.cursorY++
		if vt.cursorY >= vt.rows {
			vt.scroll()
			vt.cursorY = vt.rows - 1
		}
		vt.escState = stateNormal
	default:
		// Unknown escape sequence, return to normal
		vt.escState = stateNormal
	}
}

// processChar processes a single character
func (vt *VirtualTerminal) processChar(ch rune) {
	switch ch {
	case '\n':
		vt.newLine()
	case '\r':
		vt.cursorX = 0
	case '\b':
		if vt.cursorX > 0 {
			vt.cursorX--
		}
	case '\t':
		// Move to next tab stop (every 8 columns)
		vt.cursorX = ((vt.cursorX / 8) + 1) * 8
		if vt.cursorX >= vt.cols {
			vt.cursorX = vt.cols - 1
		}
	case '\x1b':
		// Start of escape sequence - handled by stripping later
	default:
		if ch >= ' ' && ch != '\x7f' {
			vt.putChar(ch)
		}
	}
}

// putChar puts a character at the current cursor position
func (vt *VirtualTerminal) putChar(ch rune) {
	// Get character width (1 for normal, 2 for CJK wide chars)
	width := runeWidthCond.RuneWidth(ch)
	if width == 0 {
		width = 1 // Control chars and combining chars treated as width 1
	}

	// Handle line wrap when cursor reaches end of line
	// For wide chars, need to check if there's room for both cells
	if vt.cursorX+width > vt.cols {
		// Mark the next line as wrapped (soft wrap)
		if vt.cursorY+1 < vt.rows {
			vt.isWrapped[vt.cursorY+1] = true
		}
		vt.newLine()
	}

	if vt.cursorY >= 0 && vt.cursorY < vt.rows && vt.cursorX >= 0 && vt.cursorX < vt.cols {
		// Handle overwriting wide characters:
		currentCell := vt.cells[vt.cursorY][vt.cursorX]

		// If we're writing on a placeholder (width 0), clear the previous wide char
		if currentCell.Width == 0 && vt.cursorX > 0 {
			vt.screen[vt.cursorY][vt.cursorX-1] = ' '
			vt.cells[vt.cursorY][vt.cursorX-1] = NewCell(' ')
		}

		// If we're overwriting a wide char (width 2), clear its placeholder
		if currentCell.Width == 2 && vt.cursorX+1 < vt.cols {
			vt.screen[vt.cursorY][vt.cursorX+1] = ' '
			vt.cells[vt.cursorY][vt.cursorX+1] = NewCell(' ')
		}

		// If we're writing a wide char and it will overlap with something
		if width == 2 && vt.cursorX+1 < vt.cols {
			nextCell := vt.cells[vt.cursorY][vt.cursorX+1]
			// If next cell is placeholder of a wide char, clear the wide char before it
			if nextCell.Width == 0 && vt.cursorX > 0 {
				// The wide char is at cursorX (which we're overwriting anyway)
			}
			// If next cell is a wide char, clear it and its placeholder
			if nextCell.Width == 2 {
				vt.screen[vt.cursorY][vt.cursorX+1] = ' '
				vt.cells[vt.cursorY][vt.cursorX+1] = NewCell(' ')
				if vt.cursorX+2 < vt.cols && vt.cells[vt.cursorY][vt.cursorX+2].Width == 0 {
					vt.screen[vt.cursorY][vt.cursorX+2] = ' '
					vt.cells[vt.cursorY][vt.cursorX+2] = NewCell(' ')
				}
			}
		}

		vt.screen[vt.cursorY][vt.cursorX] = ch
		// Update styled cell with full style information
		vt.cells[vt.cursorY][vt.cursorX] = NewFullStyledCell(
			ch,
			vt.currentFg,
			vt.currentBg,
			vt.currentAttrs,
			uint8(width),
			vt.currentUnderlineStyle,
			vt.currentUnderlineColor,
		)
		vt.cursorX++

		// For wide characters (CJK), add placeholder cell
		if width == 2 && vt.cursorX < vt.cols {
			vt.screen[vt.cursorY][vt.cursorX] = 0 // Placeholder
			vt.cells[vt.cursorY][vt.cursorX] = NewFullStyledCell(
				0, // No character
				vt.currentFg,
				vt.currentBg,
				vt.currentAttrs,
				0, // Width 0 = placeholder
				vt.currentUnderlineStyle,
				vt.currentUnderlineColor,
			)
			vt.cursorX++
		}
	} else {
		vt.cursorX++
	}
}

// newLine moves to the next line, scrolling if necessary
func (vt *VirtualTerminal) newLine() {
	vt.cursorX = 0
	vt.cursorY++
	if vt.cursorY >= vt.rows {
		vt.scroll()
		vt.cursorY = vt.rows - 1
	}
}

// Resize resizes the terminal
func (vt *VirtualTerminal) Resize(cols, rows int) {
	vt.mu.Lock()
	defer vt.mu.Unlock()

	if cols <= 0 {
		cols = 80
	}
	if rows <= 0 {
		rows = 24
	}

	vt.cols = cols
	vt.rows = rows
	vt.initScreen()
}

// GetDisplay returns the current screen content
func (vt *VirtualTerminal) GetDisplay() string {
	vt.mu.RLock()
	defer vt.mu.RUnlock()

	if !vt.hasData {
		return ""
	}

	var lines []string
	for rowIdx, row := range vt.screen {
		var lineBuilder strings.Builder
		for colIdx, ch := range row {
			// Skip placeholder cells (width 0 after wide chars)
			if vt.cells[rowIdx][colIdx].Width == 0 {
				continue
			}
			lineBuilder.WriteRune(ch)
		}
		line := strings.TrimRight(lineBuilder.String(), " ")
		lines = append(lines, line)
	}

	// Remove trailing empty lines
	for len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	return strings.Join(lines, "\n")
}

// GetOutput returns recent terminal output (history + current screen)
func (vt *VirtualTerminal) GetOutput(lines int) string {
	vt.mu.RLock()
	defer vt.mu.RUnlock()

	if !vt.hasData {
		return ""
	}

	var result []string
	result = append(result, vt.history...)

	for rowIdx, row := range vt.screen {
		var lineBuilder strings.Builder
		for colIdx, ch := range row {
			// Skip placeholder cells (width 0 after wide chars)
			if vt.cells[rowIdx][colIdx].Width == 0 {
				continue
			}
			lineBuilder.WriteRune(ch)
		}
		line := strings.TrimRight(lineBuilder.String(), " ")
		if line != "" {
			result = append(result, line)
		}
	}

	if len(result) > lines {
		result = result[len(result)-lines:]
	}

	return strings.Join(result, "\n")
}

// GetScreenSnapshot returns a snapshot of the current screen
func (vt *VirtualTerminal) GetScreenSnapshot() string {
	return vt.GetDisplay()
}

// Clear clears the terminal and history
func (vt *VirtualTerminal) Clear() {
	vt.mu.Lock()
	defer vt.mu.Unlock()

	vt.initScreen()
	vt.history = make([]string, 0)
	vt.historyStyled = make([][]Cell, 0)
	vt.historyIsWrapped = make([]bool, 0)
	vt.hasData = false
}

// CursorPosition returns the current cursor position
func (vt *VirtualTerminal) CursorPosition() (row, col int) {
	vt.mu.RLock()
	defer vt.mu.RUnlock()
	return vt.cursorY, vt.cursorX
}

// Cols returns the terminal width
func (vt *VirtualTerminal) Cols() int {
	vt.mu.RLock()
	defer vt.mu.RUnlock()
	return vt.cols
}

// Rows returns the terminal height
func (vt *VirtualTerminal) Rows() int {
	vt.mu.RLock()
	defer vt.mu.RUnlock()
	return vt.rows
}

// IsEmpty returns true if the terminal has no content (no history and screen is blank)
func (vt *VirtualTerminal) IsEmpty() bool {
	vt.mu.RLock()
	defer vt.mu.RUnlock()

	// Check if there's any history
	if len(vt.history) > 0 {
		return false
	}

	// Check if any cell on the screen has content
	// vt.screen stores runes directly, not Cell structs
	for y := 0; y < vt.rows; y++ {
		for x := 0; x < vt.cols; x++ {
			ch := vt.screen[y][x]
			if ch != 0 && ch != ' ' {
				return false
			}
		}
	}
	return true
}

// StripANSI removes ANSI escape sequences from text
func StripANSI(text string) string {
	return ansiPattern.ReplaceAllString(text, "")
}

// StripANSIBytes removes ANSI escape sequences from bytes
func StripANSIBytes(data []byte) []byte {
	return bytes.ReplaceAll(
		bytes.ReplaceAll(data, []byte("\x1b["), []byte("")),
		[]byte("\x1b"), []byte(""),
	)
}

// GetCellsRow returns a copy of the cells for a given row
// Used by serializer to access styled cell data
func (vt *VirtualTerminal) GetCellsRow(row int) []Cell {
	vt.mu.RLock()
	defer vt.mu.RUnlock()

	if row < 0 || row >= len(vt.cells) {
		return nil
	}
	result := make([]Cell, len(vt.cells[row]))
	copy(result, vt.cells[row])
	return result
}

// IsLineWrapped returns true if the given line is wrapped from the previous line
func (vt *VirtualTerminal) IsLineWrapped(row int) bool {
	vt.mu.RLock()
	defer vt.mu.RUnlock()

	if row < 0 || row >= len(vt.isWrapped) {
		return false
	}
	return vt.isWrapped[row]
}

// GetCurrentStyle returns the current text style (used for cursor style serialization)
func (vt *VirtualTerminal) GetCurrentStyle() (fg, bg Color, attrs CellAttrs, ulStyle UnderlineStyle, ulColor Color) {
	vt.mu.RLock()
	defer vt.mu.RUnlock()
	return vt.currentFg, vt.currentBg, vt.currentAttrs, vt.currentUnderlineStyle, vt.currentUnderlineColor
}

// getCurrentStyleNoLock returns the current text style without locking (caller must hold lock)
func (vt *VirtualTerminal) getCurrentStyleNoLock() (fg, bg Color, attrs CellAttrs, ulStyle UnderlineStyle, ulColor Color) {
	return vt.currentFg, vt.currentBg, vt.currentAttrs, vt.currentUnderlineStyle, vt.currentUnderlineColor
}

// GetHistoryStyledRow returns a copy of styled history cells for a given history index
// Index is relative to history start (0 = oldest history line)
func (vt *VirtualTerminal) GetHistoryStyledRow(index int) []Cell {
	vt.mu.RLock()
	defer vt.mu.RUnlock()

	if index < 0 || index >= len(vt.historyStyled) {
		return nil
	}
	result := make([]Cell, len(vt.historyStyled[index]))
	copy(result, vt.historyStyled[index])
	return result
}

// GetHistoryStyledLength returns the number of styled history lines
func (vt *VirtualTerminal) GetHistoryStyledLength() int {
	vt.mu.RLock()
	defer vt.mu.RUnlock()
	return len(vt.historyStyled)
}

// IsHistoryLineWrapped returns true if the given history line was wrapped
func (vt *VirtualTerminal) IsHistoryLineWrapped(index int) bool {
	vt.mu.RLock()
	defer vt.mu.RUnlock()

	if index < 0 || index >= len(vt.historyIsWrapped) {
		return false
	}
	return vt.historyIsWrapped[index]
}
