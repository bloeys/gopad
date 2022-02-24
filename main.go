package main

import (
	"io/fs"
	"os"

	"github.com/bloeys/nmage/engine"
	"github.com/bloeys/nmage/input"
	"github.com/bloeys/nmage/logging"
	nmageimgui "github.com/bloeys/nmage/ui/imgui"
	"github.com/inkyblackness/imgui-go/v4"
	"github.com/veandco/go-sdl2/sdl"
)

//TODO: Cache os.ReadDir so we don't have to use lots of disk

type Gopad struct {
	Win       *engine.Window
	mainFont  imgui.Font
	ImGUIInfo nmageimgui.ImguiInfo
	Quitting  bool

	sidebarSize float32

	CurrDir         string
	CurrDirContents []fs.DirEntry

	buffer []rune
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

	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	g := Gopad{
		Win:       window,
		ImGUIInfo: nmageimgui.NewImGUI(),
		buffer:    make([]rune, 0, 10000),
		CurrDir:   dir,
	}

	engine.Run(&g)
}

func (g *Gopad) Init() {

	g.Win.EventCallbacks = append(g.Win.EventCallbacks, g.handleWindowEvents)

	//Setup font
	var fontSize float32 = 16
	fConfig := imgui.NewFontConfig()
	defer fConfig.Delete()

	fConfig.SetOversampleH(2)
	fConfig.SetOversampleV(2)
	g.mainFont = g.ImGUIInfo.AddFontTTF("./res/fonts/courier-prime.regular.ttf", fontSize, &fConfig, nil)

	//Sidebar
	g.CurrDirContents = getDirContents(g.CurrDir)

	w, _ := g.Win.SDLWin.GetSize()
	g.sidebarSize = float32(w) * 0.10
}

func (g *Gopad) handleWindowEvents(event sdl.Event) {

	switch e := event.(type) {

	case *sdl.TextEditingEvent:
	case *sdl.TextInputEvent:
		g.buffer = append(g.buffer, []rune(e.GetText())...)

	case *sdl.WindowEvent:
		if e.Event == sdl.WINDOWEVENT_SIZE_CHANGED {
			w, _ := g.Win.SDLWin.GetSize()
			g.sidebarSize = float32(w) * 0.10
		}
	}
}

func (g *Gopad) FrameStart() {
}

func (g *Gopad) Update() {

	if input.IsQuitClicked() {
		g.Quitting = true
	}

	if x, _ := input.GetMousePos(); x > int32(g.sidebarSize) && input.MouseClicked(sdl.BUTTON_LEFT) {
		sdl.StartTextInput()
	}

	if input.KeyClicked(sdl.K_ESCAPE) {
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

	//Global imgui settings
	imgui.PushStyleColor(imgui.StyleColorText, imgui.Vec4{X: 1, Y: 1, Z: 1, W: 1})
	imgui.PushFont(g.mainFont)

	//Sidebar
	imgui.SetNextWindowPos(imgui.Vec2{X: 0, Y: 0})
	imgui.SetNextWindowSize(imgui.Vec2{X: g.sidebarSize, Y: float32(h)})
	imgui.BeginV("sidebar", &open, imgui.WindowFlagsNoCollapse|imgui.WindowFlagsNoDecoration|imgui.WindowFlagsNoMove)

	g.refreshSidebar()

	imgui.End()

	//Text area
	imgui.SetNextWindowPos(imgui.Vec2{X: g.sidebarSize, Y: 0})
	imgui.SetNextWindowSize(imgui.Vec2{X: float32(w) - g.sidebarSize, Y: float32(h)})
	imgui.BeginV("editor", &open, imgui.WindowFlagsNoCollapse|imgui.WindowFlagsNoDecoration|imgui.WindowFlagsNoMove)
	imgui.Text(string(g.buffer) + "|")
	imgui.End()

	imgui.PopFont()
	imgui.PopStyleColor()
}

func (g *Gopad) refreshSidebar() {

	for i := 0; i < len(g.CurrDirContents); i++ {

		c := g.CurrDirContents[i]
		if c.IsDir() {
			drawDir(c, g.CurrDir+"/"+c.Name()+"/")
		} else {
			imgui.Button(c.Name())
		}
	}
}

func drawDir(dir fs.DirEntry, path string) {

	isEnabled := imgui.TreeNode(dir.Name())
	if !isEnabled {
		return
	}

	contents := getDirContents(path)
	for _, c := range contents {
		if c.IsDir() {
			drawDir(c, path+c.Name()+"/")
		} else {
			imgui.Button(c.Name())
		}
	}

	imgui.TreePop()
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

func getDirContents(dir string) []fs.DirEntry {

	contents, err := os.ReadDir(dir)
	if err != nil {
		panic(err)
	}

	return contents
}
