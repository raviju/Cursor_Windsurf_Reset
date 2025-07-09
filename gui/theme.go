package gui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

type DarkChineseTheme struct{}

func (t *DarkChineseTheme) Font(style fyne.TextStyle) fyne.Resource {
	return ChineseFontResource
}

func (t *DarkChineseTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {

	if name == theme.ColorNameBackground {
		return color.NRGBA{R: 30, G: 30, B: 30, A: 255}
	} else if name == theme.ColorNameButton {
		return color.NRGBA{R: 45, G: 45, B: 50, A: 255}
	} else if name == theme.ColorNameDisabled {
		return color.NRGBA{R: 80, G: 80, B: 80, A: 255}
	} else if name == theme.ColorNameForeground {
		return color.NRGBA{R: 200, G: 200, B: 200, A: 255}
	} else if name == theme.ColorNameHover {
		return color.NRGBA{R: 60, G: 90, B: 110, A: 255}
	} else if name == theme.ColorNamePlaceHolder {
		return color.NRGBA{R: 120, G: 120, B: 120, A: 255}
	} else if name == theme.ColorNamePrimary {
		return color.NRGBA{R: 65, G: 132, B: 209, A: 255}
	} else if name == theme.ColorNameScrollBar {
		return color.NRGBA{R: 80, G: 80, B: 80, A: 255}
	} else if name == theme.ColorNameShadow {
		return color.NRGBA{R: 0, G: 0, B: 0, A: 100}
	}

	return theme.DefaultTheme().Color(name, theme.VariantDark)
}

func (t *DarkChineseTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (t *DarkChineseTheme) Size(name fyne.ThemeSizeName) float32 {

	if name == theme.SizeNamePadding {
		return theme.DefaultTheme().Size(name) * 1.1
	} else if name == theme.SizeNameText {
		return theme.DefaultTheme().Size(name) * 1.05
	}
	return theme.DefaultTheme().Size(name)
}

func NewDarkChineseTheme() fyne.Theme {
	return &DarkChineseTheme{}
}

type ModernDarkTheme struct{}

func (t *ModernDarkTheme) Font(style fyne.TextStyle) fyne.Resource {
	return ChineseFontResource
}

func (t *ModernDarkTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {

	if name == theme.ColorNameBackground {
		return color.NRGBA{R: 22, G: 25, B: 33, A: 255}
	} else if name == theme.ColorNameButton {
		return color.NRGBA{R: 50, G: 55, B: 65, A: 255}
	} else if name == theme.ColorNameDisabled {
		return color.NRGBA{R: 100, G: 105, B: 115, A: 200}
	} else if name == theme.ColorNameForeground {
		return color.NRGBA{R: 240, G: 245, B: 250, A: 255}
	} else if name == theme.ColorNameHover {
		return color.NRGBA{R: 75, G: 105, B: 145, A: 255}
	} else if name == theme.ColorNamePlaceHolder {
		return color.NRGBA{R: 160, G: 165, B: 175, A: 255}
	} else if name == theme.ColorNamePrimary {
		return color.NRGBA{R: 80, G: 145, B: 235, A: 255}
	} else if name == theme.ColorNameScrollBar {
		return color.NRGBA{R: 60, G: 65, B: 75, A: 180}
	} else if name == theme.ColorNameShadow {
		return color.NRGBA{R: 0, G: 0, B: 0, A: 80}
	} else if name == theme.ColorNameInputBackground {
		return color.NRGBA{R: 15, G: 17, B: 22, A: 255} // 非常深的输入框背景，最大化对比度
	} else if name == theme.ColorNameSelection {
		return color.NRGBA{R: 55, G: 115, B: 200, A: 100} // 选中区域
	} else if name == theme.ColorNameInputBorder {
		return color.NRGBA{R: 60, G: 65, B: 75, A: 255} // 输入框边框
	} else if name == theme.ColorNameMenuBackground {
		return color.NRGBA{R: 30, G: 33, B: 41, A: 255} // 菜单背景
	} else if name == theme.ColorNameOverlayBackground {
		return color.NRGBA{R: 15, G: 18, B: 26, A: 200} // 覆盖层背景
	}

	// 其他颜色使用默认暗色主题
	return theme.DefaultTheme().Color(name, theme.VariantDark)
}

// Icon 返回指定名称的图标资源
func (t *ModernDarkTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

// Size 返回指定名称的大小
func (t *ModernDarkTheme) Size(name fyne.ThemeSizeName) float32 {
	// 调整尺寸使界面更紧凑
	switch name {
	case theme.SizeNamePadding:
		return theme.DefaultTheme().Size(name) * 0.85 // 减小内边距
	case theme.SizeNameText:
		return theme.DefaultTheme().Size(name) * 1.0 // 保持标准文本大小以提高可读性
	case theme.SizeNameHeadingText:
		return theme.DefaultTheme().Size(name) * 0.95 // 稍微减小标题大小
	case theme.SizeNameInputBorder:
		return 1.5 // 更细的边框
	case theme.SizeNameScrollBar:
		return 4.0 // 更窄的滚动条
	case theme.SizeNameInlineIcon:
		return 20.0 // 稍小的图标
	case theme.SizeNameSeparatorThickness:
		return 1.0 // 保持细分隔线
	case theme.SizeNameInnerPadding:
		return 2.0 // 减少内部元素间距
	}
	return theme.DefaultTheme().Size(name)
}

// NewModernDarkTheme 创建一个更现代的暗色主题
func NewModernDarkTheme() fyne.Theme {
	return &ModernDarkTheme{}
}
