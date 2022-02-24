package main

import (
	"os"
	"strings"

	"github.com/inkyblackness/imgui-go/v4"
)

func selectFolder(startDir string, winWidth float32, winHeight float32) (path string, done bool) {

	if strings.TrimSpace(startDir) == "" {
		var err error
		startDir, err = os.UserHomeDir()
		if err != nil {
			panic(err.Error())
		}
	}

	imgui.OpenPopup("selectFolder")

	imgui.SetNextWindowPos(imgui.Vec2{X: float32(winWidth) * 0.5, Y: float32(winHeight) * 0.5})
	shouldEnd := imgui.BeginPopupModalV("selectFolder", nil, imgui.WindowFlagsNoCollapse)

	drawDir(startDir, true)

	if shouldEnd {
		imgui.EndPopup()
	} else {
		done = true
	}

	return path, done
}

func drawDir(fPath string, foldersOnly bool) {

	// contents, err := os.ReadDir(fPath)
	// if err != nil {
	// 	panic(err)
	// }

	// for _, c := range contents {

	// 	if !c.IsDir() {
	// 		continue
	// 	}

	// 	isEnabled := imgui.TreeNodeV( dir.Name(), imgui.TreeNodeFlagsSpanAvailWidth)
	// 	if !isEnabled {
	// 		return
	// 	}

	// 	imgui.bu
	// }

}
