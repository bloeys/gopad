package main

import (
	"math"
	"os"
	"path/filepath"

	"github.com/bloeys/gopad/settings"
	"github.com/inkyblackness/imgui-go/v4"
)

const (
	linesPerNode = 100
	textPadding  = 10
)

type Line struct {
	chars []rune
}

type LinesNode struct {
	Lines [linesPerNode]Line
	Next  *LinesNode
}

type Editor struct {
	FileName     string
	FilePath     string
	FileContents string

	MouseX int
	MouseY int

	SelectionStart  int
	SelectionLength int

	IsModified bool

	LinesHead *LinesNode
	LineCount int

	LineHeight float32
	CharWidth  float32

	StartPos float32
}

func (e *Editor) SetCursorPos(x, y int) {
	e.MouseX = x
	e.MouseY = y
}

func (e *Editor) SetStartPos(mouseDeltaNorm int32) {
	e.StartPos = clampF32(e.StartPos+float32(-mouseDeltaNorm)*settings.ScrollSpeed, 0, float32(e.LineCount))
}

func (e *Editor) RefreshFontSettings() {
	e.LineHeight = imgui.TextLineHeightWithSpacing()

	//NOTE: Because of 'https://github.com/ocornut/imgui/issues/792', CalcTextSize returns slightly incorrect width
	//values for sentences than to be expected with singleCharWidth*sentenceCharCount. For example, with 3 chars at width 10
	//we expect width of 30 (for a fixed-width font), but instead we might get 29.
	// This is fixed in the newer releases, but imgui-go hasn't updated yet.
	//
	//That's why instead of getting width of one char, we get the average width from the width of a sentence, which helps us position
	//cursors properly for now
	e.CharWidth = imgui.CalcTextSize("abcdefghijklmnopqrstuvwxyz", false, 1000).X / 26
}

func (e *Editor) RoundToNearestChar(x float32) float32 {
	return float32(math.Round(float64(x/e.CharWidth))) * e.CharWidth
}

func (e *Editor) Render(drawStartPos, winSize *imgui.Vec2) {

	//Draw window
	imgui.SetNextWindowPos(*drawStartPos)
	imgui.SetNextWindowSize(*winSize)
	imgui.BeginV("editorText", nil, imgui.WindowFlagsNoCollapse|imgui.WindowFlagsNoDecoration|imgui.WindowFlagsNoMove)

	imgui.PushStyleColor(imgui.StyleColorFrameBg, settings.EditorBgColor)
	imgui.PushStyleColor(imgui.StyleColorTextSelectedBg, settings.TextSelectionColor)

	//Add padding to text
	drawStartPos.X += textPadding
	drawStartPos.Y += textPadding
	paddedDrawStartPos := *drawStartPos

	//Draw lines
	linesToDraw := int(winSize.Y / e.LineHeight)
	// println("Lines to draw:", linesToDraw)

	dl := imgui.WindowDrawList()
	startLine := clampInt(int(e.StartPos), 0, e.LineCount)
	for i := startLine; i < startLine+linesToDraw; i++ {
		dl.AddText(*drawStartPos, imgui.PackedColorFromVec4(imgui.Vec4{X: 1, Y: 1, Z: 1, W: 1}), string(e.GetLine(0+i).chars))
		drawStartPos.Y += e.LineHeight
	}

	//Calculate position of cursor in window and grid coords.
	//Window coords are as reported by SDL, but we correct for padding and snap to the nearest
	//char window pos.
	//
	//Since gopad only supports fixed-width fonts, we treat the text area as a grid with each
	//cell having identical width and one char.
	clickedColWindowY := clampInt(e.MouseY-int(paddedDrawStartPos.Y), 0, math.MaxInt)
	clickedColGridY := clampInt(clickedColWindowY/int(e.LineHeight), 0, e.LineCount)

	clickedColWindowX := clampInt(int(e.RoundToNearestChar(float32(e.MouseX))), 0, math.MaxInt)
	clickedColGridX := clickedColWindowX / int(e.CharWidth)

	clickedLine := e.GetLine(startLine + clickedColGridY)
	tabCount, _ := getTabs(clickedLine, clickedColGridX)

	textWidth := float32(len(clickedLine.chars)-tabCount+tabCount*settings.TabSize) * e.CharWidth
	lineX := clampF32(float32(clickedColWindowX), 0, paddedDrawStartPos.X+textWidth)
	lineStart := imgui.Vec2{
		X: lineX,
		Y: paddedDrawStartPos.Y + float32(clickedColGridY)*e.LineHeight - e.LineHeight*0.25,
	}
	lineEnd := imgui.Vec2{
		X: lineX,
		Y: paddedDrawStartPos.Y + float32(clickedColGridY)*e.LineHeight + e.LineHeight*0.75,
	}
	dl.AddLineV(lineStart, lineEnd, imgui.PackedColorFromVec4(imgui.Vec4{Z: 0.7, W: 1}), settings.CursorWidthFactor*e.CharWidth)
}

func getTabs(l *Line, col int) (tabCount, charsToOffsetBy int) {

	for i := 0; i < len(l.chars) && i < col; i++ {
		if l.chars[i] == '\t' {
			tabCount++
		}
	}

	charsToOffsetBy = tabCount * settings.TabSize
	if tabCount == 0 {
		return tabCount, charsToOffsetBy
	}

	charsToRemove := col / charsToOffsetBy * settings.TabSize
	charsToRemove += col % charsToOffsetBy
	return tabCount, clampInt(charsToOffsetBy-charsToRemove, 0, charsToOffsetBy)
}

func (e *Editor) GetLine(lineNum int) *Line {

	if lineNum > e.LineCount {
		return &Line{
			chars: []rune{},
		}
	}

	curr := e.LinesHead

	//Advance to correct node
	nodeNum := lineNum / linesPerNode
	lineNum -= nodeNum * linesPerNode
	for nodeNum > 0 {
		curr = curr.Next
		nodeNum--
	}

	return &curr.Lines[lineNum]
}

func (e *Editor) GetLineCharCount(lineNum int) int {

	curr := e.LinesHead

	//Advance to correct node
	nodeNum := lineNum / linesPerNode
	lineNum -= nodeNum * linesPerNode
	for nodeNum > 0 {
		curr = curr.Next
		nodeNum--
	}

	return len(curr.Lines[lineNum].chars)
}

func ParseLines(fileContents string) (*LinesNode, int) {

	head := NewLineNode()

	lineCount := 0
	start := 0
	end := 0
	currLine := 0
	currNode := head
	for i := 0; i < len(fileContents); i++ {

		if fileContents[i] != '\n' {
			end++
			continue
		}

		lineCount++
		currNode.Lines[currLine].chars = []rune(fileContents[start:end])

		end++
		start = end
		currLine++
		if currLine == linesPerNode {
			currLine = 0
			currNode.Next = NewLineNode()
			currNode = currNode.Next
		}
	}

	return head, lineCount
}

func NewLineNode() *LinesNode {

	n := LinesNode{}
	for i := 0; i < len(n.Lines); i++ {
		n.Lines[i].chars = []rune{}
	}

	return &n
}

func clampF32(x, min, max float32) float32 {

	if x > max {
		return max
	}

	if x < min {
		return min
	}

	return x
}

func clampInt(x, min, max int) int {

	if x > max {
		return max
	}

	if x < min {
		return min
	}

	return x
}

func roundF32(x float32) float32 {
	return float32(math.Round(float64(x)))
}

func NewScratchEditor() *Editor {

	e := &Editor{
		FileName:  "**scratch**",
		LinesHead: NewLineNode(),
	}

	return e
}

func NewEditor(fPath string) *Editor {

	b, err := os.ReadFile(fPath)
	if err != nil {
		panic(err)
	}

	e := &Editor{
		FileName:     filepath.Base(fPath),
		FilePath:     fPath,
		FileContents: string(b),
	}
	e.LinesHead, e.LineCount = ParseLines(e.FileContents)
	return e
}
