package gui

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

func (gui *Gui) secondaryStagingFocused() bool {
	return gui.currentStaticContext().GetKey() == gui.State.Contexts.StagingSecondary.GetKey()
}
