package gui

import (
	"fmt"
	"runtime"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/keinenclue/sasm-docker/launcher/internal/util"
)

type DependencyWindow struct {
	window       fyne.Window
	app          fyne.App
	onComplete   func()
	dependencies util.Dependencies

	homebrewStatus *widget.Label
	dockerStatus   *widget.Label
	xquartzStatus  *widget.Label

	homebrewButton *widget.Button
	dockerButton   *widget.Button
	xquartzButton  *widget.Button
	configButton   *widget.Button

	continueButton *widget.Button
	refreshButton  *widget.Button

	content *fyne.Container
}

func NewDependencyWindow(app fyne.App, onComplete func()) *DependencyWindow {
	w := app.NewWindow("Dependency Check - SASM Docker Launcher")
	w.Resize(fyne.NewSize(600, 450))
	w.SetFixedSize(true)

	depWin := &DependencyWindow{
		window:     w,
		app:        app,
		onComplete: onComplete,
	}

	depWin.createUI()
	depWin.refreshDependencies()

	return depWin
}

func (dw *DependencyWindow) createUI() {
	title := widget.NewLabelWithStyle("SASM Docker Launcher - Dependency Check",
		fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	var osInfo string
	if runtime.GOOS == "darwin" {
		osInfo = "macOS detected - checking for required dependencies..."
	} else {
		osInfo = "Non-macOS system detected - some features may not work as expected"
	}

	osLabel := widget.NewLabel(osInfo)

	dw.homebrewStatus = widget.NewLabel("Checking Homebrew...")
	dw.dockerStatus = widget.NewLabel("Checking Docker...")
	dw.xquartzStatus = widget.NewLabel("Checking XQuartz...")

	dw.homebrewButton = widget.NewButton("Install Homebrew", dw.installHomebrew)
	dw.dockerButton = widget.NewButton("Install Docker", dw.installDocker)
	dw.xquartzButton = widget.NewButton("Install XQuartz", dw.installXQuartz)
	dw.configButton = widget.NewButton("Configure XQuartz", dw.configureXQuartz)

	dw.refreshButton = widget.NewButton("Refresh", dw.refreshDependencies)
	dw.continueButton = widget.NewButton("Continue to Launcher", dw.continueToLauncher)

	homebrewRow := container.NewHBox(
		dw.homebrewStatus,
		widget.NewSeparator(),
		dw.homebrewButton,
	)

	dockerRow := container.NewHBox(
		dw.dockerStatus,
		widget.NewSeparator(),
		dw.dockerButton,
	)

	xquartzRow := container.NewHBox(
		dw.xquartzStatus,
		widget.NewSeparator(),
		dw.xquartzButton,
		dw.configButton,
	)

	buttonRow := container.NewHBox(
		dw.refreshButton,
		widget.NewSeparator(),
		dw.continueButton,
	)

	dw.content = container.NewVBox(
		title,
		widget.NewSeparator(),
		osLabel,
		widget.NewSeparator(),
		widget.NewLabelWithStyle("Dependencies Status:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		homebrewRow,
		dockerRow,
		xquartzRow,
		widget.NewSeparator(),
		widget.NewLabelWithStyle("Actions:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		buttonRow,
		widget.NewSeparator(),
		widget.NewLabel("Note: Installation may take several minutes and will require admin privileges."),
	)

	dw.window.SetContent(container.NewPadded(dw.content))
}

func (dw *DependencyWindow) refreshDependencies() {
	dw.dependencies = util.CheckAllDependencies()

	dw.updateHomebrewStatus()
	dw.updateDockerStatus()
	dw.updateXQuartzStatus()
	dw.updateButtons()
}

func (dw *DependencyWindow) updateHomebrewStatus() {
	if dw.dependencies.Homebrew.Installed {
		dw.homebrewStatus.SetText("[OK] Homebrew: Installed")
	} else {
		dw.homebrewStatus.SetText("[MISSING] Homebrew: Not Found")
	}
}

func (dw *DependencyWindow) updateDockerStatus() {
	if dw.dependencies.Docker.Installed {
		dw.dockerStatus.SetText("[OK] Docker: Installed")
	} else {
		dw.dockerStatus.SetText("[MISSING] Docker: Not Found")
	}
}

func (dw *DependencyWindow) updateXQuartzStatus() {
	if dw.dependencies.XQuartz.Installed {
		if util.CheckAllDependencies().XQuartz.Message == "XQuartz is installed and configured" {
			dw.xquartzStatus.SetText("[OK] XQuartz: Installed & Configured")
		} else {
			dw.xquartzStatus.SetText("[WARNING] XQuartz: Installed (Needs Configuration)")
		}
	} else {
		dw.xquartzStatus.SetText("[MISSING] XQuartz: Not Found")
	}
}

func (dw *DependencyWindow) updateButtons() {
	dw.homebrewButton.Hidden = dw.dependencies.Homebrew.Installed
	dw.dockerButton.Hidden = dw.dependencies.Docker.Installed
	dw.xquartzButton.Hidden = dw.dependencies.XQuartz.Installed

	needsConfig := dw.dependencies.XQuartz.Installed &&
		util.CheckAllDependencies().XQuartz.Message != "XQuartz is installed and configured"
	dw.configButton.Hidden = !needsConfig

	allDependenciesMet := dw.dependencies.Homebrew.Installed &&
		dw.dependencies.Docker.Installed &&
		dw.dependencies.XQuartz.Installed &&
		!needsConfig

	if allDependenciesMet {
		dw.continueButton.SetText("Continue to Launcher")
	} else {
		dw.continueButton.SetText("Continue Anyway (May Not Work)")
	}

	dw.content.Refresh()
}

func (dw *DependencyWindow) installHomebrew() {
	dw.homebrewButton.SetText("Installing...")
	dw.homebrewButton.Disable()

	go func() {
		err := util.InstallHomebrew()

		if err != nil {
			dialog.ShowError(fmt.Errorf("Failed to install Homebrew: %v", err), dw.window)
		} else {
			dialog.ShowInformation("Success", "Homebrew installed successfully!", dw.window)
		}

		dw.homebrewButton.SetText("Install Homebrew")
		dw.homebrewButton.Enable()
		dw.refreshDependencies()
	}()
}

func (dw *DependencyWindow) installDocker() {
	if !dw.dependencies.Homebrew.Installed {
		dialog.ShowError(fmt.Errorf("Homebrew is required to install Docker"), dw.window)
		return
	}

	dw.dockerButton.SetText("Installing...")
	dw.dockerButton.Disable()

	go func() {
		err := util.InstallDocker()

		if err != nil {
			dialog.ShowError(fmt.Errorf("Failed to install Docker: %v", err), dw.window)
		} else {
			dialog.ShowInformation("Success", "Docker installed successfully! You may need to start Docker Desktop manually.", dw.window)
		}

		dw.dockerButton.SetText("Install Docker")
		dw.dockerButton.Enable()
		dw.refreshDependencies()
	}()
}

func (dw *DependencyWindow) installXQuartz() {
	if !dw.dependencies.Homebrew.Installed {
		dialog.ShowError(fmt.Errorf("Homebrew is required to install XQuartz"), dw.window)
		return
	}

	dw.xquartzButton.SetText("Installing...")
	dw.xquartzButton.Disable()

	go func() {
		err := util.InstallXQuartz()

		if err != nil {
			dialog.ShowError(fmt.Errorf("Failed to install XQuartz: %v", err), dw.window)
		} else {
			dialog.ShowInformation("Success", "XQuartz installed successfully! Configuration is still needed.", dw.window)
		}

		dw.xquartzButton.SetText("Install XQuartz")
		dw.xquartzButton.Enable()
		dw.refreshDependencies()
	}()
}

func (dw *DependencyWindow) configureXQuartz() {
	dw.configButton.SetText("Configuring...")
	dw.configButton.Disable()

	go func() {
		err := util.ConfigureXQuartz()

		if err != nil {
			dialog.ShowError(fmt.Errorf("Failed to configure XQuartz: %v", err), dw.window)
		} else {
			dialog.ShowInformation("Success", "XQuartz configured successfully! XQuartz will be restarted.", dw.window)
		}

		dw.configButton.SetText("Configure XQuartz")
		dw.configButton.Enable()
		dw.refreshDependencies()
	}()
}

func (dw *DependencyWindow) continueToLauncher() {
	dw.window.Hide()
	if dw.onComplete != nil {
		dw.onComplete()
	}
}

func (dw *DependencyWindow) Show() {
	dw.window.Show()
}

func (dw *DependencyWindow) ShowAndRun() {
	dw.window.ShowAndRun()
}
