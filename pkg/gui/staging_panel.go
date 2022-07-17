package gui

import (
	"strings"

	"github.com/jesseduffield/lazygit/pkg/commands/patch"
	"github.com/jesseduffield/lazygit/pkg/gui/lbl"
	"github.com/jesseduffield/lazygit/pkg/gui/types"
)

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

func (gui *Gui) handleStagingEscape() error {
	return gui.c.PushContext(gui.State.Contexts.Files)
}

func (gui *Gui) handleEditHunk() error {
	return gui.withLBLActiveCheck(gui.State.Contexts.Staging, func(state *lbl.State) error {
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

	gui.c.LogAction(gui.c.Tr.Actions.ApplyPatch)

	lineCount := strings.Count(editedPatchText, "\n") + 1
	newPatchText := patch.ModifiedPatchForRange(gui.Log, file.Name, editedPatchText, 0, lineCount, false, false)
	if err := gui.git.WorkingTree.ApplyPatch(newPatchText, "cached"); err != nil {
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
