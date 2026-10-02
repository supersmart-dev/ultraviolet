package uv

import (
	"fmt"
	"image/color"
	"testing"
)

// refInsertLineArea and refDeleteLineArea are the cell-by-cell moves the
// whole-line fast paths replace.
func refInsertLineArea(b *Buffer, y, n int, c *Cell, area Rectangle) {
	if n <= 0 || y < area.Min.Y || y >= area.Max.Y || y >= b.Height() {
		return
	}
	if y+n > area.Max.Y {
		n = area.Max.Y - y
	}
	for i := area.Max.Y - 1; i >= y+n; i-- {
		for x := area.Min.X; x < area.Max.X; x++ {
			b.Lines[i][x] = b.Lines[i-n][x]
		}
	}
	for i := y; i < y+n; i++ {
		for x := area.Min.X; x < area.Max.X; x++ {
			b.SetCell(x, i, c)
		}
	}
}

func refDeleteLineArea(b *Buffer, y, n int, c *Cell, area Rectangle) {
	if n <= 0 || y < area.Min.Y || y >= area.Max.Y || y >= b.Height() {
		return
	}
	if n > area.Max.Y-y {
		n = area.Max.Y - y
	}
	for dst := y; dst < area.Max.Y-n; dst++ {
		for x := area.Min.X; x < area.Max.X; x++ {
			b.Lines[dst][x] = b.Lines[dst+n][x]
		}
	}
	for i := area.Max.Y - n; i < area.Max.Y; i++ {
		for x := area.Min.X; x < area.Max.X; x++ {
			b.SetCell(x, i, c)
		}
	}
}

func numberedBuffer(w, h int) *Buffer {
	b := NewBuffer(w, h)
	for y := range h {
		for x := range w {
			b.SetCell(x, y, &Cell{Content: fmt.Sprintf("%c", 'a'+(y*w+x)%26), Width: 1})
		}
	}
	return b
}

func sameBuffers(t *testing.T, name string, got, want *Buffer) {
	t.Helper()
	for y := range want.Lines {
		for x := range want.Lines[y] {
			if !got.Lines[y][x].Equal(&want.Lines[y][x]) {
				t.Fatalf("%s: cell %d,%d is %+v, want %+v", name, x, y, got.Lines[y][x], want.Lines[y][x])
			}
		}
	}
}

// Moving whole lines leaves the buffer as moving their cells did, for any
// region, count, and blank cell; and a line moved isn't shared by two rows.
func TestLineAreaMatchesCellMoves(t *testing.T) {
	const w, h = 7, 9
	bg := &Cell{Content: " ", Width: 1, Style: Style{Bg: color.RGBA{1, 2, 3, 255}}}
	areas := []Rectangle{Rect(0, 0, w, h), Rect(0, 2, w, 5), Rect(0, 0, w, 4), Rect(1, 0, w-2, h)}
	for _, area := range areas {
		for y := area.Min.Y - 1; y <= area.Max.Y; y++ {
			for n := 0; n <= h+1; n++ {
				for _, c := range []*Cell{nil, bg} {
					name := fmt.Sprintf("area %v y %d n %d bg %v", area, y, n, c != nil)

					got, want := numberedBuffer(w, h), numberedBuffer(w, h)
					got.DeleteLineArea(y, n, c, area)
					refDeleteLineArea(want, y, n, c, area)
					sameBuffers(t, "delete "+name, got, want)

					got, want = numberedBuffer(w, h), numberedBuffer(w, h)
					got.InsertLineArea(y, n, c, area)
					refInsertLineArea(want, y, n, c, area)
					sameBuffers(t, "insert "+name, got, want)

					for i := range got.Lines {
						got.Lines[i][0] = Cell{Content: fmt.Sprint(i), Width: 1}
					}
					for i := range got.Lines {
						if got.Lines[i][0].Content != fmt.Sprint(i) {
							t.Fatalf("%s: row %d shares its cells with another", name, i)
						}
					}
				}
			}
		}
	}
}

func BenchmarkDeleteLineAreaScroll(b *testing.B) {
	buf := numberedBuffer(160, 48)
	area := buf.Bounds()
	for b.Loop() {
		buf.DeleteLineArea(0, 1, nil, area)
	}
}
