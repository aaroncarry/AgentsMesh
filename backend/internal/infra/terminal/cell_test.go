package terminal

import (
	"strings"
	"testing"
)

// Test Cell structure
func TestCell_Methods(t *testing.T) {
	// Default cell
	c := NewCell('A')
	if c.Char != 'A' || !c.Fg.IsDefault() || !c.Bg.IsDefault() {
		t.Error("NewCell should create default styled cell")
	}
	if !c.IsAttributeDefault() {
		t.Error("Default cell should have default attributes")
	}

	// Styled cell
	c2 := NewStyledCell('B', PaletteColor(1), PaletteColor(2), AttrBold|AttrItalic)
	if c2.Fg.Index() != 1 || c2.Bg.Index() != 2 {
		t.Error("StyledCell colors incorrect")
	}
	if !c2.Attrs.Has(AttrBold) || !c2.Attrs.Has(AttrItalic) {
		t.Error("StyledCell attributes incorrect")
	}

	// Full styled cell
	c3 := NewFullStyledCell('C', RGBColor(255, 0, 0), DefaultColor(), AttrUnderline, 2, UnderlineCurly, PaletteColor(5))
	if c3.Width != 2 {
		t.Errorf("FullStyledCell width incorrect: got %d", c3.Width)
	}
	if c3.UnderlineStyle != UnderlineCurly {
		t.Error("FullStyledCell underline style incorrect")
	}
}

// Test Color methods
func TestColor_Methods(t *testing.T) {
	// Default color
	dc := DefaultColor()
	if !dc.IsDefault() || dc.IsPalette() || dc.IsRGB() {
		t.Error("DefaultColor type check failed")
	}

	// Palette color
	pc := PaletteColor(196)
	if pc.IsDefault() || !pc.IsPalette() || pc.IsRGB() {
		t.Error("PaletteColor type check failed")
	}
	if pc.Index() != 196 {
		t.Errorf("PaletteColor index: got %d, want 196", pc.Index())
	}

	// RGB color
	rc := RGBColor(255, 128, 64)
	if rc.IsDefault() || rc.IsPalette() || !rc.IsRGB() {
		t.Error("RGBColor type check failed")
	}
	r, g, b := rc.RGB()
	if r != 255 || g != 128 || b != 64 {
		t.Errorf("RGBColor values: got (%d,%d,%d), want (255,128,64)", r, g, b)
	}

	// Equality
	if !PaletteColor(5).Equals(PaletteColor(5)) {
		t.Error("Same palette colors should be equal")
	}
	if PaletteColor(5).Equals(PaletteColor(6)) {
		t.Error("Different palette colors should not be equal")
	}
	if !RGBColor(1, 2, 3).Equals(RGBColor(1, 2, 3)) {
		t.Error("Same RGB colors should be equal")
	}

	// Different types should not be equal
	if DefaultColor().Equals(PaletteColor(0)) {
		t.Error("Default and Palette colors should not be equal")
	}
	if PaletteColor(0).Equals(RGBColor(0, 0, 0)) {
		t.Error("Palette and RGB colors should not be equal")
	}

	// Test invalid colorType (edge case for default branch)
	// This tests the fallback behavior for invalid color types
	invalidColor := Color{colorType: 99} // Invalid type
	if invalidColor.Equals(invalidColor) {
		t.Error("Invalid color type should return false from default case")
	}
	// Two invalid colors with same invalid type should still return false
	invalidColor2 := Color{colorType: 99}
	if invalidColor.Equals(invalidColor2) {
		t.Error("Two invalid colors should not be equal")
	}
}

// Test Cell IsEmpty and StyleEquals methods
func TestCell_IsEmptyAndStyleEquals(t *testing.T) {
	// Empty cell
	empty := NewCell(' ')
	if !empty.IsEmpty() {
		t.Error("Space with default style should be empty")
	}

	// Non-empty cells
	nonEmpty1 := NewCell('A')
	if nonEmpty1.IsEmpty() {
		t.Error("Non-space char should not be empty")
	}

	nonEmpty2 := NewStyledCell(' ', PaletteColor(1), DefaultColor(), AttrNone)
	if nonEmpty2.IsEmpty() {
		t.Error("Space with colored fg should not be empty")
	}

	nonEmpty3 := NewStyledCell(' ', DefaultColor(), PaletteColor(1), AttrNone)
	if nonEmpty3.IsEmpty() {
		t.Error("Space with colored bg should not be empty")
	}

	nonEmpty4 := NewStyledCell(' ', DefaultColor(), DefaultColor(), AttrBold)
	if nonEmpty4.IsEmpty() {
		t.Error("Space with attributes should not be empty")
	}

	// StyleEquals
	cell1 := NewStyledCell('A', PaletteColor(1), PaletteColor(2), AttrBold)
	cell2 := NewStyledCell('B', PaletteColor(1), PaletteColor(2), AttrBold)
	if !cell1.StyleEquals(cell2) {
		t.Error("Same style with different chars should have equal style")
	}

	cell3 := NewStyledCell('A', PaletteColor(2), PaletteColor(2), AttrBold)
	if cell1.StyleEquals(cell3) {
		t.Error("Different fg colors should not have equal style")
	}

	// Full styled cell with underline
	cell4 := NewFullStyledCell('A', PaletteColor(1), PaletteColor(2), AttrBold, 1, UnderlineCurly, PaletteColor(5))
	cell5 := NewFullStyledCell('B', PaletteColor(1), PaletteColor(2), AttrBold, 1, UnderlineCurly, PaletteColor(5))
	if !cell4.StyleEquals(cell5) {
		t.Error("Same full style should be equal")
	}

	cell6 := NewFullStyledCell('A', PaletteColor(1), PaletteColor(2), AttrBold, 1, UnderlineDouble, PaletteColor(5))
	if cell4.StyleEquals(cell6) {
		t.Error("Different underline style should not be equal")
	}
}

// Test Cell GetWidth
func TestCell_GetWidth(t *testing.T) {
	c1 := NewCell('A')
	if c1.GetWidth() != 1 {
		t.Errorf("Normal cell width: got %d, want 1", c1.GetWidth())
	}

	c2 := NewFullStyledCell(0, DefaultColor(), DefaultColor(), AttrNone, 0, UnderlineNone, DefaultColor())
	if c2.GetWidth() != 0 {
		t.Errorf("Placeholder cell width: got %d, want 0", c2.GetWidth())
	}

	c3 := NewFullStyledCell('中', DefaultColor(), DefaultColor(), AttrNone, 2, UnderlineNone, DefaultColor())
	if c3.GetWidth() != 2 {
		t.Errorf("Wide cell width: got %d, want 2", c3.GetWidth())
	}
}

// Test GetScreenSnapshot
func TestGetScreenSnapshot(t *testing.T) {
	vt := NewVirtualTerminal(80, 24, 1000)
	vt.Feed([]byte("Hello, World!"))

	snapshot := vt.GetScreenSnapshot()
	display := vt.GetDisplay()

	if snapshot != display {
		t.Errorf("GetScreenSnapshot should equal GetDisplay: got %q, want %q", snapshot, display)
	}
}

// Test StripANSIBytes
func TestStripANSIBytes(t *testing.T) {
	input := []byte("\x1b[31mRed\x1b[0m Normal")
	result := StripANSIBytes(input)

	// Should strip ESC sequences
	if string(result) == string(input) {
		t.Error("StripANSIBytes should modify input with ANSI sequences")
	}

	// Should contain the text
	if !strings.Contains(string(result), "Red") || !strings.Contains(string(result), "Normal") {
		t.Errorf("StripANSIBytes should preserve text: %q", result)
	}
}
