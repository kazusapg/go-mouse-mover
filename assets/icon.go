package assets

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/color"
	"image/png"
)

func generateIconPNG(c color.Color) []byte {
	img := image.NewRGBA(image.Rect(0, 0, 16, 16))
	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			img.Set(x, y, c)
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		panic("failed to encode icon PNG: " + err.Error())
	}
	return buf.Bytes()
}

func wrapPNGAsICO(pngData []byte) []byte {
	var buf bytes.Buffer
	// ICONDIR (6 bytes)
	buf.Write([]byte{0, 0})   // Reserved
	buf.Write([]byte{1, 0})   // Type: 1 = ICON
	buf.Write([]byte{1, 0})   // Count: 1 image

	// ICONDIRENTRY (16 bytes)
	buf.Write([]byte{16})          // Width
	buf.Write([]byte{16})          // Height
	buf.Write([]byte{0})          // Colors (0 = >256)
	buf.Write([]byte{0})          // Reserved
	buf.Write([]byte{1, 0})       // Planes
	buf.Write([]byte{32, 0})      // BitCount
	binary.Write(&buf, binary.LittleEndian, uint32(len(pngData)))
	binary.Write(&buf, binary.LittleEndian, uint32(22)) // offset = 6 + 16
	buf.Write(pngData)
	return buf.Bytes()
}

// DefaultIcon returns a gray icon for idle state.
func DefaultIcon() []byte {
	return wrapPNGAsICO(generateIconPNG(color.RGBA{128, 128, 128, 255}))
}

// ActiveIcon returns a red icon for moving state.
func ActiveIcon() []byte {
	return wrapPNGAsICO(generateIconPNG(color.RGBA{255, 0, 0, 255}))
}
