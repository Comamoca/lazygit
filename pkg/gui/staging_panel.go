package gui

import (
	"strings"

	"github.com/jesseduffield/lazygit/pkg/commands/patch"
	"github.com/jesseduffield/lazygit/pkg/gui/lbl"
	"github.com/jesseduffield/lazygit/pkg/gui/types"
)

func (gui *Gui) refreshStagingPanel(selectedLineIdx int) error {
	gui.splitMainPanel(true)

	file := gui.getSelectedFile()
	if file == nil || (!file.HasUnstagedChanges && !file.HasStagedChanges) {
		return gui.handleStagingEscape()
	}

	// note for custom diffs, we'll need to send a flag here saying not to use the custom diff
	diff := gui.git.WorkingTree.WorktreeFileDiff(file, true, false, false)
	secondaryDiff := gui.git.WorkingTree.WorktreeFileDiff(file, true, true, false)

	// if we have e.g. a deleted file with nothing else to the diff will have only
	// 4-5 lines in which case we'll swap panels
	if len(strings.Split(diff, "\n")) < 5 {
		if len(strings.Split(secondaryDiff, "\n")) < 5 {
			return gui.handleStagingEscape()
		}
		// TODO: change focus
		diff, secondaryDiff = secondaryDiff, diff
	}

	empty, err := gui.refreshLineByLinePanel(diff, secondaryDiff, gui.Tr.UnstagedChanges, gui.Tr.StagedChanges, gui.secondaryStagingFocused(), selectedLineIdx)
	if err != nil {
		return err
	}

	if empty {
		return gui.handleStagingEscape()
	}

	return nil
}

func (gui *Gui) handleRefreshStagingPanel(selectedLineIdx int) error {
	gui.Mutexes.LineByLinePanelMutex.Lock()
	defer gui.Mutexes.LineByLinePanelMutex.Unlock()

	return gui.refreshStagingPanel(selectedLineIdx)
}

func (gui *Gui) onStagingFocus(forceSecondaryFocused bool, selectedLineIdx int) error {
	gui.Mutexes.LineByLinePanelMutex.Lock()
	defer gui.Mutexes.LineByLinePanelMutex.Unlock()

	if gui.State.Contexts.Staging.GetState() == nil || selectedLineIdx != -1 {
		return gui.refreshStagingPanel(selectedLineIdx)
	}

	return nil
}

func (gui *Gui) handleTogglePanel() error {
	gui.Mutexes.LineByLinePanelMutex.Lock()
	gui.escapeLineByLinePanel()
	gui.Mutexes.LineByLinePanelMutex.Unlock()

	if gui.secondaryStagingFocused() {
		if err := gui.c.PushContext(gui.State.Contexts.Staging); err != nil {
			return err
		}
	} else {
		if err := gui.c.PushContext(gui.State.Contexts.StagingSecondary); err != nil {
			return err
		}
	}

	return nil
}

func (gui *Gui) handleStagingEscape() error {
	return gui.c.PushContext(gui.State.Contexts.Files)
}

func (gui *Gui) handleToggleStagedSelection() error {
	return gui.withLBLActiveCheck(func(state *lbl.State) error {
		return gui.applySelection(gui.secondaryStagingFocused(), state)
	})
}

func (gui *Gui) handleResetSelection() error {
	return gui.withLBLActiveCheck(func(state *lbl.State) error {
		if gui.secondaryStagingFocused() {
			// for backwards compatibility
			return gui.applySelection(true, state)
		}

		if !gui.c.UserConfig.Gui.SkipUnstageLineWarning {
			return gui.c.Confirm(types.ConfirmOpts{
				Title:  gui.c.Tr.UnstageLinesTitle,
				Prompt: gui.c.Tr.UnstageLinesPrompt,
				HandleConfirm: func() error {
					return gui.withLBLActiveCheck(func(state *lbl.State) error {
						return gui.applySelection(true, state)
					})
				},
			})
		} else {
			return gui.applySelection(true, state)
		}
	})
}

func (gui *Gui) applySelection(reverse bool, state *lbl.State) error {
	file := gui.getSelectedFile()
	if file == nil {
		return nil
	}

	firstLineIdx, lastLineIdx := state.SelectedRange()
	patch := patch.ModifiedPatchForRange(gui.Log, file.Name, state.GetDiff(), firstLineIdx, lastLineIdx, reverse, false)

	if patch == "" {
		return nil
	}

	// apply the patch then refresh this panel
	// create a new temp file with the patch, then call git apply with that patch
	applyFlags := []string{}
	if !reverse || gui.secondaryStagingFocused() {
		applyFlags = append(applyFlags, "cached")
	}
	gui.c.LogAction(gui.c.Tr.Actions.ApplyPatch)
	err := gui.git.WorkingTree.ApplyPatch(patch, applyFlags...)
	if err != nil {
		return gui.c.Error(err)
	}

	if state.SelectingRange() {
		state.SetLineSelectMode()
	}

	if err := gui.c.Refresh(types.RefreshOptions{Scope: []types.RefreshableView{types.FILES}}); err != nil {
		return err
	}
	if err := gui.refreshStagingPanel(-1); err != nil {
		return err
	}
	return nil
}

func (gui *Gui) handleEditHunk() error {
	return gui.withLBLActiveCheck(func(state *lbl.State) error {
		return gui.editHunk(gui.secondaryStagingFocused(), state)
	})
}

func (gui *Gui) editHunk(reverse bool, state *lbl.State) error {
	file := gui.getSelectedFile()
	if file == nil {
		return nil
	}

	hunk := state.CurrentHunk()
	patchText := patch.ModifiedPatchForRange(gui.Log, file.Name, state.GetDiff(), hunk.FirstLineIdx, hunk.LastLineIdx(), reverse, false)
	patchFilepath, err := gui.git.WorkingTree.SaveTemporaryPatch(patchText)
	if err != nil {
		return err
	}

	lineOffset := 3
	lineIdxInHunk := state.GetSelectedLineIdx() - hunk.FirstLineIdx
	if err := gui.helpers.Files.EditFileAtLine(patchFilepath, lineIdxInHunk+lineOffset); err != nil {
		return err
	}

	editedPatchText, err := gui.git.File.Cat(patchFilepath)
	if err != nil {
		return err
	}

	applyFlags := []string{}
	if !reverse || gui.secondaryStagingFocused() {
		applyFlags = append(applyFlags, "cached")
	}
	gui.c.LogAction(gui.c.Tr.Actions.ApplyPatch)

	lineCount := strings.Count(editedPatchText, "\n") + 1
	newPatchText := patch.ModifiedPatchForRange(gui.Log, file.Name, editedPatchText, 0, lineCount, false, false)
	if err := gui.git.WorkingTree.ApplyPatch(newPatchText, applyFlags...); err != nil {
		return gui.c.Error(err)
	}

	if err := gui.c.Refresh(types.RefreshOptions{Scope: []types.RefreshableView{types.FILES}}); err != nil {
		return err
	}
	if err := gui.refreshStagingPanel(-1); err != nil {
		return err
	}
	return nil
}

func (gui *Gui) secondaryStagingFocused() bool {
	return gui.currentStaticContext().GetKey() == gui.State.Contexts.StagingSecondary.GetKey()
}
