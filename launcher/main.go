package main

import (
	"os/user"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"github.com/keinenclue/sasm-docker/launcher/internal/config"
	"github.com/keinenclue/sasm-docker/launcher/internal/gui"
	// "github.com/keinenclue/sasm-docker/launcher/internal/util"
)

func main() {
	u, _ := user.Current()
	config.Setup(u.HomeDir + "/sasm-data")
	a := app.New()

	// Always show dependency window for now to avoid blocking
	if true { // util.HasMissingDependencies() {
		var mainWindow fyne.Window
		depWindow := gui.NewDependencyWindow(a, func() {
			mainWindow = a.NewWindow("Sasm-docker launcher")
			mainWindow.Resize(fyne.NewSize(500, 300))
			mainWindow.SetContent(gui.New(a, mainWindow))
			mainWindow.Show()
		})
		depWindow.ShowAndRun()
	} else {
		w := a.NewWindow("Sasm-docker launcher")
		w.Resize(fyne.NewSize(500, 300))
		w.SetContent(gui.New(a, w))
		w.ShowAndRun()
	}
}
