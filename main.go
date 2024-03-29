package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/bloeys/gopad/settings"
	"github.com/bloeys/nmage/engine"
	"github.com/bloeys/nmage/input"
	"github.com/bloeys/nmage/logging"
	"github.com/bloeys/nmage/renderer/rend3dgl"
	nmageimgui "github.com/bloeys/nmage/ui/imgui"
	"github.com/inkyblackness/imgui-go/v4"
	"github.com/veandco/go-sdl2/sdl"
)

//TODO: Cache os.ReadDir so we don't have to use lots of disk

type Gopad struct {
	Win       *engine.Window
	mainFont  imgui.Font
	ImGUIInfo nmageimgui.ImguiInfo

	mainMenuBarHeight  float32
	sidebarWidthFactor float32
	sidebarWidthPx     float32

	CurrDir         string
	CurrDirContents []fs.DirEntry

	editors          []Editor
	editorToClose    int
	activeEditor     int
	lastActiveEditor int
	newRunes         []rune

	//Errors
	haveErr bool
	errMsg  string

	//Cache window size
	winWidth  float32
	winHeight float32
}

func main() {

	chdirErr := os.Chdir(filepath.Dir(os.Args[0]))
	if chdirErr != nil {
		panic(chdirErr.Error())
	}

	if err := engine.Init(); err != nil {
		panic(err)
	}

	window, err := engine.CreateOpenGLWindowCentered("nMage", 1280, 720, engine.WindowFlags_RESIZABLE|engine.WindowFlags_ALLOW_HIGHDPI, rend3dgl.NewRend3DGL())
	if err != nil {
		logging.ErrLog.Fatalln("Failed to create window. Err: ", err)
	}
	defer window.Destroy()

	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	g := Gopad{
		Win:                window,
		ImGUIInfo:          nmageimgui.NewImGUI(),
		CurrDir:            dir,
		editors:            []Editor{*NewScratchEditor()},
		editorToClose:      -1,
		sidebarWidthFactor: 0.15,
		newRunes:           []rune{},
	}

	// Init runs within an imgui frame, but imgui frames do NOT allow adding fonts,
	// so we do it here
	g.LoadFonts()

	// engine.SetVSync(true)
	engine.Run(&g, g.Win, g.ImGUIInfo)
}

func (g *Gopad) LoadFonts() {

	fConfig := imgui.NewFontConfig()
	fConfig.SetOversampleH(2)
	fConfig.SetOversampleV(2)

	g.mainFont = g.ImGUIInfo.AddFontTTF("./res/fonts/courier-prime.regular.ttf", settings.FontSize, &fConfig, nil)

	fConfig.Delete()
}

func (g *Gopad) Init() {

	g.Win.SDLWin.SetTitle("Gopad")
	g.Win.EventCallbacks = append(g.Win.EventCallbacks, g.handleWindowEvents)

	//Sidebar
	g.CurrDirContents = getDirContents(g.CurrDir)

	w, h := g.Win.SDLWin.GetSize()
	g.winWidth = float32(w)
	g.winHeight = float32(h)
	g.sidebarWidthPx = g.winWidth * g.sidebarWidthFactor

	//Read os.Args
	for i := 1; i < len(os.Args); i++ {

		e := *NewEditor(os.Args[i])
		g.editors = append(g.editors, e)
		g.activeEditor = len(g.editors) - 1
	}

	g.activeEditor = len(g.editors) - 1

	// Prepare editors
	imgui.PushFont(g.mainFont)
	for i := 0; i < len(g.editors); i++ {
		e := &g.editors[i]
		e.RefreshFontSettings()
	}
	imgui.PopFont()
}

func (g *Gopad) handleWindowEvents(event sdl.Event) {

	switch e := event.(type) {

	case *sdl.TextEditingEvent:
	case *sdl.TextInputEvent:
		g.newRunes = append(g.newRunes, []rune(e.GetText())...)
	case *sdl.WindowEvent:
		if e.Event == sdl.WINDOWEVENT_SIZE_CHANGED {
			w, h := g.Win.SDLWin.GetSize()
			g.winWidth = float32(w)
			g.winHeight = float32(h)
			g.sidebarWidthPx = g.winWidth * g.sidebarWidthFactor
		}
	}
}

func (g *Gopad) closeEditor(eIndex int) {

	g.editors = append(g.editors[:eIndex], g.editors[eIndex+1:]...)

	if g.activeEditor >= len(g.editors) {
		g.activeEditor = len(g.editors) - 1
	}

	g.lastActiveEditor = g.activeEditor
}

func (g *Gopad) Update() {

	if input.IsQuitClicked() {
		engine.Quit()
		return
	}

	if g.haveErr {
		g.showErrorPopup()
	}

	if input.MouseClicked(sdl.BUTTON_LEFT) {
		x, y := input.GetMousePos()
		g.getActiveEditor().SetCursorPos(int(x), int(y))
	}

	if yMove := input.GetMouseWheelYNorm(); yMove != 0 {
		g.getActiveEditor().SetStartPos(yMove)
	}

	//Close editor if needed
	if input.KeyDown(sdl.K_LCTRL) && input.KeyClicked(sdl.K_w) {
		g.closeEditor(g.activeEditor)
		g.editorToClose = -1
	}

	e := g.getActiveEditor()

	//Save if needed
	if !e.IsModified {
		return
	}

	if !input.KeyDown(sdl.K_LCTRL) || !input.KeyClicked(sdl.K_s) {
		return
	}

	g.saveEditor(e)
}

func (g *Gopad) saveEditor(e *Editor) {

	if e.FileName == "**scratch**" {
		e.IsModified = false
		return
	}

	err := os.WriteFile(e.FilePath, []byte(e.FileContents), os.ModePerm)
	if err != nil {
		g.triggerError("Failed to save file. Error: " + err.Error())
		return
	}

	e.IsModified = false
}

func (g *Gopad) triggerError(errMsg string) {
	imgui.OpenPopup("err")
	g.haveErr = true
	g.errMsg = errMsg
}

func (g *Gopad) showErrorPopup() {

	w, h := g.Win.SDLWin.GetSize()
	imgui.SetNextWindowPos(imgui.Vec2{X: float32(w) * 0.5, Y: float32(h) * 0.5})

	shouldEnd := imgui.BeginPopup("err")

	imgui.Text(g.errMsg)
	if imgui.Button("OK") {
		g.haveErr = false
		imgui.CloseCurrentPopup()
	}

	if shouldEnd {
		imgui.EndPopup()
	} else {
		g.haveErr = false
	}
}

func (g *Gopad) Render() {

	//Global imgui settings
	imgui.PushFont(g.mainFont)

	g.drawMenubar()
	g.drawSidebar()
	g.drawEditors()

	imgui.PopFont()
}

func (g *Gopad) drawMenubar() {

	shouldCloseMenuBar := imgui.BeginMainMenuBar()

	if imgui.BeginMenu("File") {

		if imgui.MenuItem("Save") {
			g.saveEditor(g.getActiveEditor())
		}

		imgui.EndMenu()
	}

	g.mainMenuBarHeight = imgui.WindowHeight()
	if shouldCloseMenuBar {
		imgui.EndMainMenuBar()
	}
}

func (g *Gopad) drawSidebar() {

	imgui.SetNextWindowPos(imgui.Vec2{X: 0, Y: g.mainMenuBarHeight})
	imgui.SetNextWindowSize(imgui.Vec2{X: g.sidebarWidthPx, Y: g.winHeight - g.mainMenuBarHeight})
	imgui.BeginV("sidebar", nil, imgui.WindowFlagsNoCollapse|imgui.WindowFlagsNoDecoration|imgui.WindowFlagsNoMove)

	imgui.PushStyleColor(imgui.StyleColorButton, imgui.Vec4{W: 0})
	for i := 0; i < len(g.CurrDirContents); i++ {

		c := g.CurrDirContents[i]
		if c.IsDir() {
			g.drawDir(c, g.CurrDir+"/"+c.Name()+"/")
		} else {
			g.drawFile(c, g.CurrDir+"/"+c.Name())
		}
	}

	imgui.PopStyleColor()
	imgui.End()
}

func (g *Gopad) drawEditors() {

	//Draw editor area window
	imgui.SetNextWindowPos(imgui.Vec2{X: g.sidebarWidthPx, Y: g.mainMenuBarHeight})
	imgui.SetNextWindowSize(imgui.Vec2{X: g.winWidth - g.sidebarWidthPx})
	imgui.BeginV("editor", nil, imgui.WindowFlagsNoCollapse|imgui.WindowFlagsNoDecoration|imgui.WindowFlagsNoMove)

	//Draw tabs
	isEditorsEnabled := imgui.BeginTabBarV("editorTabs", 0)
	shouldForceSwitch := g.activeEditor != g.lastActiveEditor

	prevActiveEditor := g.activeEditor
	for i := 0; i < len(g.editors); i++ {

		e := &g.editors[i]

		flags := imgui.TabItemFlagsNone
		if shouldForceSwitch && g.activeEditor == i {
			flags = imgui.TabItemFlagsSetSelected
		}

		if e.IsModified {
			flags |= imgui.TabItemFlagsUnsavedDocument
		}

		open := true
		if !imgui.BeginTabItemV(e.FileName, &open, flags) {

			if !open {
				g.editorToClose = i
			}
			continue
		}
		if !open {
			g.editorToClose = i
		}

		//If these two aren't equal it means we programmatically changed the active editor (instead of a mouse click),
		//and so we shouldn't change based on what imgui is telling us
		if g.activeEditor == g.lastActiveEditor {
			g.activeEditor = i
			g.lastActiveEditor = i
		}
		imgui.EndTabItem()
	}
	g.lastActiveEditor = g.activeEditor

	if isEditorsEnabled {
		imgui.EndTabBar()
	}

	tabsHeight := imgui.WindowHeight()
	imgui.End()

	g.getActiveEditor().UpdateAndDraw(
		&imgui.Vec2{X: g.sidebarWidthPx, Y: g.mainMenuBarHeight + tabsHeight},
		&imgui.Vec2{X: g.winWidth - g.sidebarWidthPx, Y: g.winHeight - g.mainMenuBarHeight - tabsHeight},
		g.newRunes,
	)

	if shouldForceSwitch || prevActiveEditor != g.activeEditor {
		imgui.SetKeyboardFocusHereV(-1)
	}

	imgui.PopStyleColor()
	imgui.PopStyleColor()
	imgui.End()
}

func (g *Gopad) getActiveEditor() *Editor {
	return g.getEditor(g.activeEditor)
}

func (g *Gopad) getEditor(index int) *Editor {

	if len(g.editors) == 0 {
		e := *NewScratchEditor()
		e.RefreshFontSettings()
		g.editors = append(g.editors, e)
		g.activeEditor = 0
		return &e
	}

	if index >= 0 && index < len(g.editors) {
		return &g.editors[index]
	}

	panic(fmt.Sprint("Invalid editor index: ", index))
}

func (g *Gopad) drawDir(dir fs.DirEntry, path string) {

	isEnabled := imgui.TreeNodeV(dir.Name(), imgui.TreeNodeFlagsSpanAvailWidth)
	if !isEnabled {
		return
	}

	contents := getDirContents(path)
	for _, c := range contents {
		if c.IsDir() {
			g.drawDir(c, path+c.Name()+"/")
		} else {
			g.drawFile(c, path+c.Name())
		}
	}

	imgui.TreePop()
}

func (g *Gopad) drawFile(f fs.DirEntry, path string) {

	if imgui.Button(f.Name()) {
		g.handleFileClick(path)
	}
}

func (g *Gopad) handleFileClick(fPath string) {

	//Check if we already have the file open
	editorIndex := -1
	for i := 0; i < len(g.editors); i++ {

		e := &g.editors[i]
		if e.FilePath == fPath {
			editorIndex = i
			break
		}
	}

	//If already found switch to it
	if editorIndex >= 0 {
		g.activeEditor = editorIndex
		return
	}

	//Read new file and switch to it
	e := *NewEditor(fPath)
	e.RefreshFontSettings()
	g.editors = append(g.editors, e)
	g.activeEditor = len(g.editors) - 1
}

func (g *Gopad) FrameEnd() {
	g.newRunes = []rune{}

	// Close editors if needed
	if g.editorToClose > -1 {
		g.closeEditor(g.editorToClose)
		g.editorToClose = -1
	}
}

func (g *Gopad) DeInit() {
	g.Win.Destroy()
}

func getDirContents(dir string) []fs.DirEntry {

	contents, err := os.ReadDir(dir)
	if err != nil {
		panic(err)
	}

	return contents
}
