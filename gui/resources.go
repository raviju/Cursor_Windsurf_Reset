package gui

import (
	_ "embed"
	"fyne.io/fyne/v2"
)

//go:embed NotoSansSC-Regular.ttf
var fontData []byte

var ChineseFontResource = &fyne.StaticResource{
	StaticName:    "NotoSansSC-Regular.ttf",
	StaticContent: fontData,
}
