package main

import (
	"github.com/bloeys/nmage/engine"
	"github.com/bloeys/nmage/input"
	"github.com/bloeys/nmage/logging"
	nmageimgui "github.com/bloeys/nmage/ui/imgui"
	"github.com/inkyblackness/imgui-go/v4"
	"github.com/veandco/go-sdl2/sdl"
)

type Gopad struct {
	Win       *engine.Window
	ImGUIInfo nmageimgui.ImguiInfo
	Quitting  bool

	mainFont imgui.Font
	buffer   []rune
}

var ()

func main() {

	if err := engine.Init(); err != nil {
		panic(err)
	}

	window, err := engine.CreateOpenGLWindowCentered("nMage", 1280, 720, engine.WindowFlags_RESIZABLE|engine.WindowFlags_ALLOW_HIGHDPI)
	if err != nil {
		logging.ErrLog.Fatalln("Failed to create window. Err: ", err)
	}
	defer window.Destroy()

	g := Gopad{
		Win:       window,
		ImGUIInfo: nmageimgui.NewImGUI(),
		buffer:    make([]rune, 0, 10000),
	}

	engine.Run(&g)
}

func (g *Gopad) Init() {

	g.Win.EventCallbacks = append(g.Win.EventCallbacks, g.handleWindowEvents)

	var fontSize float32 = 16

	fConfig := imgui.NewFontConfig()
	defer fConfig.Delete()

	fConfig.SetOversampleH(2)
	fConfig.SetOversampleV(2)
	g.mainFont = g.ImGUIInfo.AddFontTTF("./res/fonts/courier-prime.regular.ttf", fontSize, &fConfig, nil)
}

func (g *Gopad) handleWindowEvents(event sdl.Event) {

	switch e := event.(type) {

	case *sdl.TextEditingEvent:
	case *sdl.TextInputEvent:
		g.buffer = append(g.buffer, []rune(e.GetText())...)
	}
}

func (g *Gopad) FrameStart() {
}

func (g *Gopad) Update() {

	if input.IsQuitClicked() {
		g.Quitting = true
	}

	if x, y := input.GetMousePos(); x > 0 && y > 0 && input.MouseClicked(sdl.BUTTON_LEFT) {
		println("Start")
		sdl.StartTextInput()
	}

	if input.KeyClicked(sdl.K_ESCAPE) {
		println("End")
		sdl.StopTextInput()
	}

	if input.KeyClicked(sdl.K_BACKSPACE) && len(g.buffer) > 0 {
		g.buffer = g.buffer[:len(g.buffer)-1]
	}

	if input.KeyClicked(sdl.K_RETURN) || input.KeyClicked(sdl.K_RETURN2) {
		g.buffer = append(g.buffer, rune('\n'))
	}
}

func (g *Gopad) Render() {

	open := true
	w, h := g.Win.SDLWin.GetSize()
	sidebarSize := float32(w) * 0.10

	//Global imgui settings
	imgui.PushStyleColor(imgui.StyleColorText, imgui.Vec4{X: 1, Y: 1, Z: 1, W: 1})
	imgui.PushFont(g.mainFont)

	//Sidebar
	imgui.SetNextWindowPos(imgui.Vec2{X: 0, Y: 0})
	imgui.SetNextWindowSize(imgui.Vec2{X: sidebarSize, Y: float32(h)})
	imgui.BeginV("sidebar", &open, imgui.WindowFlagsNoCollapse|imgui.WindowFlagsNoDecoration|imgui.WindowFlagsNoMove)
	imgui.End()

	//Text area
	imgui.SetNextWindowPos(imgui.Vec2{X: sidebarSize, Y: 0})
	imgui.SetNextWindowSize(imgui.Vec2{X: float32(w) - sidebarSize, Y: float32(h)})
	imgui.BeginV("editor", &open, imgui.WindowFlagsNoCollapse|imgui.WindowFlagsNoDecoration|imgui.WindowFlagsNoMove)
	imgui.Text(string(g.buffer) + "|")
	imgui.End()

	imgui.PopFont()
	imgui.PopStyleColor()
}

func (g *Gopad) FrameEnd() {

}

func (g *Gopad) ShouldRun() bool {
	return !g.Quitting
}

func (g *Gopad) GetWindow() *engine.Window {
	return g.Win
}

func (g *Gopad) GetImGUI() nmageimgui.ImguiInfo {
	return g.ImGUIInfo
}

func (g *Gopad) Deinit() {
	g.Win.Destroy()
}
