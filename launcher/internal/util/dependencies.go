package util

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type DependencyStatus struct {
	Name      string
	Installed bool
	Path      string
	Message   string
}

type Dependencies struct {
	Homebrew DependencyStatus
	Docker   DependencyStatus
	XQuartz  DependencyStatus
}

func CheckAllDependencies() Dependencies {
	return Dependencies{
		Homebrew: checkHomebrew(),
		Docker:   checkDocker(),
		XQuartz:  checkXQuartz(),
	}
}

func checkHomebrew() DependencyStatus {
	cmd := exec.Command("/bin/bash", "-c", "which brew")
	err := cmd.Run()

	if err != nil {
		return DependencyStatus{
			Name:      "Homebrew",
			Installed: false,
			Message:   "Homebrew not found - required for installing other dependencies",
		}
	}

	brewPath, _ := exec.Command("/bin/bash", "-c", "which brew").Output()
	return DependencyStatus{
		Name:      "Homebrew",
		Installed: true,
		Path:      strings.TrimSpace(string(brewPath)),
		Message:   "Homebrew is installed",
	}
}

func checkDocker() DependencyStatus {
	dockerAppPath := "/Applications/Docker.app"
	dockerBinaryPath := "/usr/local/bin/docker"

	if _, err := os.Stat(dockerAppPath); err == nil {
		return DependencyStatus{
			Name:      "Docker",
			Installed: true,
			Path:      dockerAppPath,
			Message:   "Docker Desktop is installed",
		}
	}

	if cmd := exec.Command("/bin/bash", "-c", "which docker"); cmd.Run() == nil {
		if dockerPath, err := exec.Command("/bin/bash", "-c", "which docker").Output(); err == nil {
			return DependencyStatus{
				Name:      "Docker",
				Installed: true,
				Path:      strings.TrimSpace(string(dockerPath)),
				Message:   "Docker is installed",
			}
		}
	}

	if _, err := os.Stat(dockerBinaryPath); err == nil {
		return DependencyStatus{
			Name:      "Docker",
			Installed: true,
			Path:      dockerBinaryPath,
			Message:   "Docker is installed",
		}
	}

	return DependencyStatus{
		Name:      "Docker",
		Installed: false,
		Message:   "Docker Desktop not found - required for running containers",
	}
}

func checkXQuartz() DependencyStatus {
	xquartzAppPath := "/Applications/Utilities/XQuartz.app"

	if _, err := os.Stat(xquartzAppPath); err == nil {
		xquartzConfigured := checkXQuartzConfiguration()
		if xquartzConfigured {
			return DependencyStatus{
				Name:      "XQuartz",
				Installed: true,
				Path:      xquartzAppPath,
				Message:   "XQuartz is installed and configured",
			}
		} else {
			return DependencyStatus{
				Name:      "XQuartz",
				Installed: true,
				Path:      xquartzAppPath,
				Message:   "XQuartz is installed but needs configuration",
			}
		}
	}

	return DependencyStatus{
		Name:      "XQuartz",
		Installed: false,
		Message:   "XQuartz not found - required for X11 forwarding",
	}
}

func checkXQuartzConfiguration() bool {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return false
	}

	plistFile := filepath.Join(homeDir, "Library/Preferences/org.macosforge.xquartz.X11.plist")

	if _, err := os.Stat(plistFile); err != nil {
		return false
	}

	cmd := exec.Command("defaults", "read", "org.macosforge.xquartz.X11", "nolisten_tcp")
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	return strings.TrimSpace(string(output)) == "0"
}

func InstallHomebrew() error {
	cmd := exec.Command("/bin/bash", "-c", `$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)`)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func InstallDocker() error {
	cmd := exec.Command("brew", "install", "--cask", "docker")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func InstallXQuartz() error {
	cmd := exec.Command("brew", "install", "--cask", "xquartz")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func ConfigureXQuartz() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %v", err)
	}

	plistFile := filepath.Join(homeDir, "Library/Preferences/org.macosforge.xquartz.X11.plist")
	tempFile := "/tmp/org.macosforge.xquartz.X11.plist"

	exec.Command("launchctl", "stop", "com.apple.cfprefsd.xpc.agent").Run()

	if _, err := os.Stat(plistFile); err != nil {
		plistContent := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
</dict>
</plist>`

		if err := os.WriteFile(tempFile, []byte(plistContent), 0644); err != nil {
			return fmt.Errorf("failed to create temp plist: %v", err)
		}
	} else {
		if err := exec.Command("cp", plistFile, tempFile).Run(); err != nil {
			return fmt.Errorf("failed to copy existing plist: %v", err)
		}
	}

	if err := exec.Command("plutil", "-convert", "xml1", tempFile).Run(); err != nil {
		return fmt.Errorf("failed to convert plist to xml: %v", err)
	}

	perlCmd := `perl -i -pe 'BEGIN{undef $/;} s/(.*)(<\/dict>)/$1        <key>nolisten_tcp<\/key>\n        <false\/>\n        $2/s'`
	if err := exec.Command("bash", "-c", perlCmd+" "+tempFile).Run(); err != nil {
		return fmt.Errorf("failed to modify plist: %v", err)
	}

	if err := exec.Command("cp", tempFile, plistFile).Run(); err != nil {
		return fmt.Errorf("failed to copy plist back: %v", err)
	}

	exec.Command("launchctl", "start", "com.apple.cfprefsd.xpc.agent").Run()
	exec.Command("rm", "-f", tempFile).Run()

	exec.Command("killall", "XQuartz").Run()

	return nil
}

func HasMissingDependencies() bool {
	deps := CheckAllDependencies()
	return !deps.Homebrew.Installed || !deps.Docker.Installed || !deps.XQuartz.Installed || !checkXQuartzConfiguration()
}
