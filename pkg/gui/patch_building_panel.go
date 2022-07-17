package gui

import (
	"github.com/jesseduffield/lazygit/pkg/commands/patch"
	"github.com/jesseduffield/lazygit/pkg/gui/lbl"
	"github.com/jesseduffield/lazygit/pkg/gui/types"
	"github.com/samber/lo"
)

func (gui *Gui) refreshPatchBuildingPanel(selectedLineIdx int) error {
	if !gui.git.Patch.PatchManager.Active() {
		return gui.handleEscapePatchBuildingPanel()
	}

	// get diff from commit file that's currently selected
	node := gui.State.Contexts.CommitFiles.GetSelected()
	if node == nil {
		return nil
	}

	ref := gui.State.Contexts.CommitFiles.CommitFileTreeViewModel.GetRef()
	to := ref.RefName()
	from, reverse := gui.State.Modes.Diffing.GetFromAndReverseArgsForDiff(ref.ParentRefName())
	diff, err := gui.git.WorkingTree.ShowFileDiff(from, to, reverse, node.GetPath(), true)
	if err != nil {
		return err
	}

	secondaryDiff := gui.git.Patch.PatchManager.RenderPatchForFile(node.GetPath(), true, false, true)
	if err != nil {
		return err
	}

	empty, err := gui.refreshLBLPatchBuildingPanel(diff, secondaryDiff, selectedLineIdx)
	if err != nil {
		return err
	}

	if empty {
		return gui.handleEscapePatchBuildingPanel()
	}

	return nil
}

// returns whether the patch is empty so caller can escape if necessary
// both diffs should be non-coloured because we'll parse them and colour them here
func (gui *Gui) refreshLBLPatchBuildingPanel(diff string, secondaryDiff string, selectedLineIdx int) (bool, error) {
	context := gui.State.Contexts.PatchBuilding

	oldState := context.GetState()

	state := lbl.NewState(diff, selectedLineIdx, oldState, gui.Log)
	context.SetState(state)
	if state == nil {
		return true, nil
	}

	mainContent := context.GetContentToRender()
	secondaryPatchParser := patch.NewPatchParser(gui.Log, secondaryDiff)
	secondaryContent := secondaryPatchParser.Render(-1, -1, nil)

	// TODO: see if this should happen AFTER setting content.
	if err := context.Focus(); err != nil {
		return false, err
	}

	return false, gui.refreshMainViews(refreshMainOpts{
		pair: gui.patchBuildingMainContextPair(),
		main: &viewUpdateOpts{
			task:  NewRenderStringWithoutScrollTask(mainContent),
			title: gui.Tr.Patch,
		},
		secondary: &viewUpdateOpts{
			task:  NewRenderStringWithoutScrollTask(secondaryContent),
			title: gui.Tr.CustomPatch,
		},
	})
}

func (gui *Gui) handleRefreshPatchBuildingPanel(selectedLineIdx int) error {
	gui.Mutexes.LineByLinePanelMutex.Lock()
	defer gui.Mutexes.LineByLinePanelMutex.Unlock()

	return gui.refreshPatchBuildingPanel(selectedLineIdx)
}

func (gui *Gui) onPatchBuildingFocus(selectedLineIdx int) error {
	gui.Mutexes.LineByLinePanelMutex.Lock()
	defer gui.Mutexes.LineByLinePanelMutex.Unlock()

	// TODO: switch batck to patch building state
	if gui.State.Contexts.Staging.GetState() == nil || selectedLineIdx != -1 {
		return gui.refreshPatchBuildingPanel(selectedLineIdx)
	}

	return nil
}

func (gui *Gui) handleToggleSelectionForPatch() error {
	err := gui.withLBLActiveCheck(gui.State.Contexts.PatchBuilding, func(state *lbl.State) error {
		toggleFunc := gui.git.Patch.PatchManager.AddFileLineRange
		filename := gui.getSelectedCommitFileName()
		includedLineIndices, err := gui.git.Patch.PatchManager.GetFileIncLineIndices(filename)
		if err != nil {
			return err
		}
		currentLineIsStaged := lo.Contains(includedLineIndices, state.GetSelectedLineIdx())
		if currentLineIsStaged {
			toggleFunc = gui.git.Patch.PatchManager.RemoveFileLineRange
		}

		// add range of lines to those set for the file
		node := gui.State.Contexts.CommitFiles.GetSelected()
		if node == nil {
			return nil
		}

		firstLineIdx, lastLineIdx := state.SelectedRange()

		if err := toggleFunc(node.GetPath(), firstLineIdx, lastLineIdx); err != nil {
			// might actually want to return an error here
			gui.c.Log.Error(err)
		}

		return nil
	})
	if err != nil {
		return err
	}

	if err := gui.handleRefreshPatchBuildingPanel(-1); err != nil {
		return err
	}

	if err := gui.refreshCommitFilesContext(); err != nil {
		return err
	}

	return nil
}

func (gui *Gui) handleEscapePatchBuildingPanel() error {
	if gui.git.Patch.PatchManager.IsEmpty() {
		gui.git.Patch.PatchManager.Reset()
	}

	if gui.currentContext().GetKey() == gui.State.Contexts.PatchBuilding.GetKey() {
		return gui.c.PushContext(gui.State.Contexts.CommitFiles)
	} else {
		// need to re-focus in case the secondary view should now be hidden
		return gui.currentContext().HandleFocus(types.OnFocusOpts{})
	}
}

func (gui *Gui) secondaryPatchPanelUpdateOpts() *viewUpdateOpts {
	if gui.git.Patch.PatchManager.Active() {
		patch := gui.git.Patch.PatchManager.RenderAggregatedPatchColored(false)

		return &viewUpdateOpts{
			task:  NewRenderStringWithoutScrollTask(patch),
			title: gui.Tr.CustomPatch,
		}
	}

	return nil
}
