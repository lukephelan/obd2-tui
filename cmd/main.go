package main

import (
	"log"
	"os"

	"github.com/jroimartin/gocui"
	"github.com/lukephelan/obd2-tui/internal/state"
	"github.com/lukephelan/obd2-tui/internal/ui"
)

var logFile *os.File

func init() {
	var err error
	logFile, err = os.OpenFile("debug.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		log.Fatal("Failed to open log file:", err)
	}
	log.SetOutput(logFile)
	log.Println("===== Application Started =====")
}

func enterMenu(g *gocui.Gui) error {
	item := state.CurrentMenu[state.SelectedIndex]

	if item.SubMenu != nil {
		// Save current menu state before navigating deeper
		state.MenuHistory = append(state.MenuHistory, state.CurrentMenu)
		state.IndexHistory = append(state.IndexHistory, state.SelectedIndex)

		// Enter submenu
		state.CurrentMenu = item.SubMenu
		state.SelectedIndex = 0 // Reset selection

		// Ensure we don't land on a heading
		for state.SelectedIndex < len(state.CurrentMenu) && state.CurrentMenu[state.SelectedIndex].IsHeading {
			state.SelectedIndex++ // Move to first non-heading item
		}

		state.ShowLiveData = false

	} else {
		state.ShowLiveData = true
	}
	ui.RenderMenu(g)
	ui.UpdateDataView(g)
	return nil
}

func exitMenu(g *gocui.Gui) error {
	if len(state.MenuHistory) > 0 {
		state.CurrentMenu = state.MenuHistory[len(state.MenuHistory)-1]
		state.SelectedIndex = state.IndexHistory[len(state.IndexHistory)-1]

		state.MenuHistory = state.MenuHistory[:len(state.MenuHistory)-1]
		state.IndexHistory = state.IndexHistory[:len(state.IndexHistory)-1]

		state.ShowLiveData = false // Restore controls view when going back
	}

	ui.RenderMenu(g)
	ui.UpdateDataView(g)
	return nil
}

func keybindings(g *gocui.Gui) error {
	if err := g.SetKeybinding("menu", gocui.KeyArrowUp, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		return ui.MoveSelection(g, -1)
	}); err != nil {
		return err
	}

	if err := g.SetKeybinding("menu", gocui.KeyArrowDown, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		return ui.MoveSelection(g, 1)
	}); err != nil {
		return err
	}

	if err := g.SetKeybinding("menu", gocui.KeyEnter, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		return enterMenu(g)
	}); err != nil {
		return err
	}

	if err := g.SetKeybinding("menu", gocui.KeyArrowRight, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		return enterMenu(g)
	}); err != nil {
		return err
	}

	if err := g.SetKeybinding("menu", gocui.KeyArrowLeft, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		return exitMenu(g)
	}); err != nil {
		return err
	}

	if err := g.SetKeybinding("menu", gocui.KeyEsc, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		return exitMenu(g)
	}); err != nil {
		return err
	}

	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		return gocui.ErrQuit
	}); err != nil {
		return err
	}

	return nil
}

func main() {
	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	g.SetManagerFunc(ui.Layout)

	if err := keybindings(g); err != nil {
		log.Panicln(err)
	}

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}
}
