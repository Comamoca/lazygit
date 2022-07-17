package gui

import (
	"github.com/jesseduffield/lazygit/pkg/commands/patch"
	"github.com/jesseduffield/lazygit/pkg/gui/lbl"
)

// Currently there are two 'pseudo-panels' that make use of this 'pseudo-panel'.
// One is the staging panel where we stage files line-by-line, the other is the
// patch building panel where we add lines of an old commit's file to a patch.
// This file contains the logic around selecting lines and displaying the diffs
// staging_panel.go and patch_building_panel.go have functions specific to their
// use cases

// returns whether the patch is empty so caller can escape if necessary
// both diffs should be non-coloured because we'll parse them and colour them here
func (gui *Gui) refreshLineByLinePanel(diff string, secondaryDiff string, title string, secondaryTitle string, secondaryFocused bool, selectedLineIdx int) (bool, error) {
	gui.splitMainPanel(true)

	oldState := gui.State.Contexts.Staging.GetState()

	if secondaryFocused {
		diff, secondaryDiff = secondaryDiff, diff
	}

	state := lbl.NewState(diff, selectedLineIdx, oldState, gui.Log)
	gui.State.Contexts.Staging.SetState(state)
	if state == nil {
		return true, nil
	}

	gui.State.Contexts.Staging.SetState(state)

	focusedContent, err := gui.getMainDiffForLbl(gui.State.Contexts.Staging.GetState())
	if err != nil {
		return false, err
	}

	// TODO: see if this should happen AFTER setting content.
	if err := gui.focusSelection(gui.State.Contexts.Staging.GetState()); err != nil {
		return false, err
	}

	secondaryPatchParser := patch.NewPatchParser(gui.Log, secondaryDiff)

	pair := gui.currentLblMainPair()
	gui.moveMainContextPairToTop(pair)
	unfocusedContent := secondaryPatchParser.Render(-1, -1, nil)

	var mainContent, secondaryContent string
	if secondaryFocused {
		mainContent, secondaryContent = unfocusedContent, focusedContent
	} else {
		mainContent, secondaryContent = focusedContent, unfocusedContent
	}

	return false, gui.refreshMainViews(refreshMainOpts{
		pair: gui.currentLblMainPair(),
		main: &viewUpdateOpts{
			task:  NewRenderStringWithoutScrollTask(mainContent),
			title: title,
		},
		secondary: &viewUpdateOpts{
			task:  NewRenderStringWithoutScrollTask(secondaryContent),
			title: secondaryTitle,
		},
	})
}

func (gui *Gui) refreshAndFocusLblPanel(state *lbl.State) error {
	if err := gui.refreshMainViewForLineByLine(state); err != nil {
		return err
	}

	return gui.focusSelection(state)
}

func (gui *Gui) refreshMainViewForLineByLine(state *lbl.State) error {
	diff, err := gui.getMainDiffForLbl(state)
	if err != nil {
		return err
	}

	mainView := gui.currentLblMainPair().main.GetView()

	gui.setViewContent(mainView, diff)

	return nil
}

func (gui *Gui) getMainDiffForLbl(state *lbl.State) (string, error) {
	var includedLineIndices []int
	// I'd prefer not to have knowledge of contexts using this file but I'm not sure
	// how to get around this
	if gui.currentContext().GetKey() == gui.State.Contexts.PatchBuilding.GetKey() {
		filename := gui.getSelectedCommitFileName()
		var err error
		includedLineIndices, err = gui.git.Patch.PatchManager.GetFileIncLineIndices(filename)
		if err != nil {
			return "", err
		}
	}
	colorDiff := state.RenderForLineIndices(includedLineIndices)

	return colorDiff, nil
}

// I'd prefer not to have knowledge of contexts using this file but I'm not sure
// how to get around this
func (gui *Gui) currentLblMainPair() MainContextPair {
	// because we're not using the lock it's important this is only called
	if gui.currentStaticContextWithoutLock().GetKey() == gui.State.Contexts.PatchBuilding.GetKey() {
		return gui.patchBuildingMainContextPair()
	} else {
		return gui.stagingMainContextPair()
	}
}

// focusSelection works out the best focus for the staging panel given the
// selected line and size of the hunk
func (gui *Gui) focusSelection(state *lbl.State) error {
	view := gui.currentLblMainPair().main.GetView()

	_, viewHeight := view.Size()
	bufferHeight := viewHeight - 1
	_, origin := view.Origin()

	selectedLineIdx := state.GetSelectedLineIdx()

	newOrigin := state.CalculateOrigin(origin, bufferHeight)

	if err := view.SetOriginY(newOrigin); err != nil {
		return err
	}

	return view.SetCursor(0, selectedLineIdx-newOrigin)
}

func (gui *Gui) escapeLineByLinePanel() {
	gui.State.Contexts.Staging.SetState(nil)
}

// TODO: fix this
func (gui *Gui) handlelineByLineNavigateTo(selectedLineIdx int) error {
	return gui.withLBLActiveCheck(func(state *lbl.State) error {
		state.SetLineSelectMode()
		state.SelectLine(selectedLineIdx)

		return gui.refreshAndFocusLblPanel(state)
	})
}

func (gui *Gui) withLBLActiveCheck(f func(*lbl.State) error) error {
	gui.Mutexes.LineByLinePanelMutex.Lock()
	defer gui.Mutexes.LineByLinePanelMutex.Unlock()

	state := gui.State.Contexts.Staging.GetState()
	if state == nil {
		return nil
	}

	return f(state)
}
