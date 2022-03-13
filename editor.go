package main

import (
	"math"
	"os"
	"path/filepath"

	"github.com/bloeys/gopad/settings"
	"github.com/inkyblackness/imgui-go/v4"
)

// hello		there my friend
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

func (e *Editor) RoundToGrid(x float32) float32 {
	return float32(math.Round(float64(x/e.CharWidth))) * e.CharWidth
}

func (e *Editor) UpdateAndDraw(drawStartPos, winSize *imgui.Vec2, newRunes []rune) {

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

	//Draw text
	dl := imgui.WindowDrawList()
	linesToDraw := int(winSize.Y / e.LineHeight)
	startLine := clampInt(int(e.StartPos), 0, e.LineCount)
	for i := startLine; i < startLine+linesToDraw; i++ {
		dl.AddText(*drawStartPos, imgui.PackedColorFromVec4(imgui.Vec4{X: 1, Y: 1, Z: 1, W: 1}), string(e.GetLine(0+i).chars))
		drawStartPos.Y += e.LineHeight
	}

	posInfo := e.getPositions(&paddedDrawStartPos)
	tabCount, charsToOffsetBy := getTabs(posInfo.Line, posInfo.GridXEditor)

	textWidth := float32(len(posInfo.Line.chars)-tabCount+tabCount*settings.TabSize) * e.CharWidth
	lineX := clampF32(float32(posInfo.GridXGlobal)+float32(charsToOffsetBy)*e.CharWidth, 0, paddedDrawStartPos.X+textWidth)

	lineStart := imgui.Vec2{
		X: lineX,
		Y: paddedDrawStartPos.Y + float32(posInfo.GridYEditor)*e.LineHeight - e.LineHeight*0.25,
	}
	lineEnd := imgui.Vec2{
		X: lineX,
		Y: paddedDrawStartPos.Y + float32(posInfo.GridYEditor)*e.LineHeight + e.LineHeight*0.75,
	}
	dl.AddLineV(lineStart, lineEnd, imgui.PackedColorFromVec4(settings.CursorColor), settings.CursorWidthFactor*e.CharWidth)

	// charAtCursor := getCharFromCursor(clickedLine, clickedColGridXEditor)
	// println("Chars:", "'"+charAtCursor+"'", ";", clickedColGridXEditor)
}

type mousePosInfo struct {

	//Global represents a grid on the whole window (i.e. 0,0 is the top left of the program window)
	GridXGlobal int
	//Global represents a grid on the whole window (i.e. 0,0 is the top left of the program window)
	GridYGlobal int

	//Editor represents a grid on the editor window (i.e. 0,0 is the top left of the editor window)
	GridXEditor int
	//Editor represents a grid on the editor window (i.e. 0,0 is the top left of the editor window)
	GridYEditor int

	//Line is the currently selected line
	Line    *Line
	LineNum int
}

func (e *Editor) getPositions(paddedDrawStartPos *imgui.Vec2) mousePosInfo {

	//Calculate position of cursor in window and grid coords.
	//Window coords are as reported by SDL, but we correct for padding and snap to the nearest
	//char window pos.
	//
	//Since Gopad only supports fixed-width fonts, we treat the text area as a grid with each
	//cell having identical width and one char inside.
	//
	//'Global' suffix means the position is in window coords.
	//'Editor' suffix means coords are within the text editor coords, where sidebar and tabs have been adjusted for

	roundedMouseX := e.RoundToGrid(float32(e.MouseX))
	gridXGlobal := clampInt(int(roundedMouseX), 0, math.MaxInt)

	roundedMouseY := e.RoundToGrid(float32(e.MouseY))
	gridYGlobal := clampInt(int(roundedMouseY), 0, math.MaxInt)

	gridXEditor := int(
		roundF32(
			(float32(gridXGlobal) - paddedDrawStartPos.X) / e.CharWidth,
		),
	)

	windowYEditor := clampInt(e.MouseY-int(paddedDrawStartPos.Y), 0, math.MaxInt)
	gridYEditor := clampInt(windowYEditor/int(e.LineHeight), 0, e.LineCount)

	startLineIndex := clampInt(int(e.StartPos), 0, e.LineCount)
	return mousePosInfo{

		GridXGlobal: gridXGlobal,
		GridYGlobal: gridYGlobal,

		GridXEditor: gridXEditor,
		GridYEditor: gridYEditor,

		Line:    e.GetLine(startLineIndex + gridYEditor),
		LineNum: startLineIndex + gridYEditor,
	}
}

func getCharFromCursor(l *Line, cursorGridX int) string {
	i := getCharIndexFromCursor(l, cursorGridX)
	if i == -1 {
		return ""
	}
	return string(l.chars[i])
}

func getCharIndexFromCursor(l *Line, cursorGridX int) int {

	if cursorGridX <= 0 || len(l.chars) == 0 {
		return -1
	}

	gridSize := 0
	for i := 0; i < len(l.chars) && i <= cursorGridX; i++ {

		if l.chars[i] == '\t' {
			gridSize += settings.TabSize
		} else {
			gridSize++
		}

		if gridSize < cursorGridX {
			continue
		}

		return i
	}

	if cursorGridX >= gridSize {
		return len(l.chars) - 1
	}

	return -1
}

//TODO: The offset chars must be how many grid cols between cursor col and the nearest non-tab char.
func getTabs(l *Line, gridPosX int) (tabCount, charsToOffsetBy int) {

	charIndex := getCharIndexFromCursor(l, gridPosX)
	if charIndex == -1 {
		return 0, 0
	}

	//gridSize represents the visual grid (e.g. \tHi has a visual grid size of 'tabSize+2')
	gridSize := 0
	for i := charIndex; i >= 0; i-- {
		if l.chars[i] == '\t' {
			tabCount++
			gridSize += settings.TabSize
		} else {
			gridSize++
		}
	}

	if l.chars[charIndex] != '\t' {
		return tabCount, 0
	}

	return tabCount, gridSize - gridPosX
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

	if fileContents[len(fileContents)-1] != '\n' {
		lineCount++
		currNode.Lines[currLine].chars = []rune(fileContents[start:end])
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
