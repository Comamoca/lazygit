package gui

import (
	"github.com/jesseduffield/lazygit/pkg/gui/lbl"
	"github.com/jesseduffield/lazygit/pkg/gui/types"
)

// Currently there are two 'pseudo-panels' that make use of this 'pseudo-panel'.
// One is the staging panel where we stage files line-by-line, the other is the
// patch building panel where we add lines of an old commit's file to a patch.
// This file contains the logic around selecting lines and displaying the diffs
// staging_panel.go and patch_building_panel.go have functions specific to their
// use cases

func (gui *Gui) withLBLActiveCheck(context types.ILBLContext, f func(*lbl.State) error) error {
	gui.Mutexes.LineByLinePanelMutex.Lock()
	defer gui.Mutexes.LineByLinePanelMutex.Unlock()

	state := context.GetState()
	if state == nil {
		return nil
	}

	return f(state)
}
