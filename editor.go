package main

import (
	"fmt"
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
}

func (e *Editor) SetCursorPos(x, y int) {
	e.MouseX = x
	e.MouseY = y
}

func (e *Editor) Render(drawStartPos, winSize *imgui.Vec2) {

	origDrawStartPos := *drawStartPos

	//Draw window
	imgui.SetNextWindowPos(*drawStartPos)
	imgui.SetNextWindowSize(*winSize)
	imgui.BeginV("editorText", nil, imgui.WindowFlagsNoCollapse|imgui.WindowFlagsNoDecoration|imgui.WindowFlagsNoMove)

	imgui.PushStyleColor(imgui.StyleColorFrameBg, settings.EditorBgColor)
	imgui.PushStyleColor(imgui.StyleColorTextSelectedBg, settings.TextSelectionColor)

	//Add padding to text
	drawStartPos.X += textPadding
	drawStartPos.Y += textPadding

	//Draw lines
	lineHeight := imgui.TextLineHeightWithSpacing()
	charWidth := imgui.CalcTextSize("a", false, 1000).X
	linesToDraw := int(winSize.Y / lineHeight)
	// println("Lines to draw:", linesToDraw)

	cx := clampInt(e.MouseX-int(drawStartPos.X), 0, int(winSize.X))
	cy := clampInt(e.MouseY-int(drawStartPos.Y), 0, int(winSize.Y))

	clickedLine := clampInt(cy/int(lineHeight), 0, e.LineCount)
	clickedCol := clampInt(cx/int(charWidth), 0, e.GetLineCharCount(clickedLine))

	dl := imgui.WindowDrawList()
	for i := 0; i < linesToDraw; i++ {
		dl.AddText(*drawStartPos, imgui.PackedColorFromVec4(imgui.Vec4{X: 1, Y: 1, Z: 1, W: 1}), string(e.GetLine(0+i).chars))
		drawStartPos.Y += lineHeight
	}

	fmt.Printf("line,col: %v,%v\n", clickedLine, clickedCol)

	origDrawStartPos.X += textPadding
	origDrawStartPos.Y += textPadding
	lineStart := imgui.Vec2{
		X: origDrawStartPos.X + float32(clickedCol)*charWidth,
		Y: origDrawStartPos.Y + float32(clickedLine)*lineHeight - lineHeight*0.25,
	}
	lineEnd := imgui.Vec2{
		X: origDrawStartPos.X + float32(clickedCol)*charWidth,
		Y: origDrawStartPos.Y + float32(clickedLine)*lineHeight + lineHeight*0.75,
	}
	dl.AddLineV(lineStart, lineEnd, imgui.PackedColorFromVec4(imgui.Vec4{Z: 0.7, W: 1}), 5)
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
