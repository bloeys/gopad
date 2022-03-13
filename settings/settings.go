package settings

import "github.com/inkyblackness/imgui-go/v4"

var (
	FontSize           float32    = 16
	TextSelectionColor imgui.Vec4 = imgui.Vec4{X: 84 / 255.0, Y: 153 / 255.0, Z: 199 / 255.0, W: 0.4}
	EditorBgColor      imgui.Vec4 = imgui.Vec4{X: 0.1, Y: 0.1, Z: 0.1, W: 1}

	//NOTE: Imgui hardcodes tab size to 4 in '#define IM_TABSIZE 4'
	TabSize           int        = 4
	ScrollSpeed       float32    = 4
	CursorWidthFactor float32    = 0.15
	CursorColor       imgui.Vec4 = imgui.Vec4{X: 1, Y: 1, Z: 1, W: 1}
)
