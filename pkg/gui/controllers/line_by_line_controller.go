package controllers

import (
	"github.com/jesseduffield/gocui"
	"github.com/jesseduffield/lazygit/pkg/gui/types"
)

type LBLControllerFactory struct {
	*controllerCommon
}

func NewLBLControllerFactory(c *controllerCommon) *LBLControllerFactory {
	return &LBLControllerFactory{
		controllerCommon: c,
	}
}

func (self *LBLControllerFactory) Create(context types.ILBLContext) *LBLController {
	return &LBLController{
		baseController:   baseController{},
		controllerCommon: self.controllerCommon,
		context:          context,
	}
}

type LBLController struct {
	baseController
	*controllerCommon

	context types.ILBLContext
}

func (self *LBLController) Context() types.Context {
	return self.context
}

func (self *LBLController) HandlePrevLine() error {
	self.context.GetState().CycleSelection(false)

	return nil
}

func (self *LBLController) HandleNextLine() error {
	self.context.GetState().CycleSelection(true)

	return nil
}

func (self *LBLController) HandlePrevHunk() error {
	self.context.GetState().CycleHunk(false)

	return nil
}

func (self *LBLController) HandleNextHunk() error {
	self.context.GetState().CycleHunk(true)

	return nil
}

func (self *LBLController) HandleToggleSelectRange() error {
	self.context.GetState().ToggleSelectRange()

	return nil
}

func (self *LBLController) HandleToggleSelectHunk() error {
	self.context.GetState().ToggleSelectHunk()

	return nil
}

func (self *LBLController) HandleScrollLeft() error {
	self.context.GetViewTrait().ScrollLeft()

	return nil
}

func (self *LBLController) HandleScrollRight() error {
	self.context.GetViewTrait().ScrollRight()

	return nil
}

func (self *LBLController) HandleScrollUp() error {
	self.context.GetViewTrait().ScrollUp(self.c.UserConfig.Gui.ScrollHeight)

	return self.render()
}

func (self *LBLController) HandleScrollDown() error {
	self.context.GetViewTrait().ScrollDown(self.c.UserConfig.Gui.ScrollHeight)

	return self.render()
}

func (self *LBLController) isSelectedLineInViewPort() bool {
	selectedLineIdx := self.context.GetState().GetSelectedLineIdx()
	startIdx, length := self.context.GetViewTrait().ViewPortYBounds()
	return selectedLineIdx >= startIdx && selectedLineIdx < startIdx+length
}

func (self *LBLController) getContentToRender() string {
	return self.context.GetState().RenderForLineIndices(self.context.GetIncludedLineIndices())
}

func (self *LBLController) HandlePrevPage() error {
	self.context.GetState().SetLineSelectMode()
	self.context.GetState().AdjustSelectedLineIdx(-self.context.GetViewTrait().PageDelta())

	return nil
}

func (self *LBLController) HandleNextPage() error {
	self.context.GetState().SetLineSelectMode()
	self.context.GetState().AdjustSelectedLineIdx(self.context.GetViewTrait().PageDelta())

	return nil
}

func (self *LBLController) HandleGotoTop() error {
	self.context.GetState().SelectTop()

	return nil
}

func (self *LBLController) HandleGotoBottom() error {
	self.context.GetState().SelectBottom()

	return nil
}

func (self *LBLController) HandleMouseDown() error {
	self.context.GetState().SelectNewLineForRange(self.context.GetViewTrait().SelectedLineIdx())

	return nil
}

func (self *LBLController) HandleMouseDrag() error {
	self.context.GetState().SelectLine(self.context.GetViewTrait().SelectedLineIdx())

	return nil
}

func (self *LBLController) CopySelectedToClipboard() error {
	selected := self.context.GetState().PlainRenderSelected()

	self.c.LogAction(self.c.Tr.Actions.CopySelectedTextToClipboard)
	if err := self.os.CopyToClipboard(selected); err != nil {
		return self.c.Error(err)
	}

	return nil
}

// TODO: use or delete
func (self *LBLController) pushContextIfNotFocused() error {
	if !self.isFocused() {
		if err := self.c.PushContext(self.context); err != nil {
			return err
		}
	}

	return nil
}

func (self *LBLController) isFocused() bool {
	return self.c.CurrentContext().GetKey() == self.context.GetKey()
}

func (self *LBLController) renderAndFocus() error {
	self.context.GetView().SetContent(self.getContentToRender())

	if err := self.focusSelection(); err != nil {
		return err
	}

	self.c.Render()

	return nil
}

func (self *LBLController) render() error {
	self.context.GetView().SetContent(self.getContentToRender())

	self.c.Render()

	return nil
}

func (self *LBLController) focusSelection() error {
	view := self.context.GetView()
	state := self.context.GetState()
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

func (self *LBLController) withRenderAndFocus(f func() error) func() error {
	return func() error {
		if err := f(); err != nil {
			return err
		}

		return self.renderAndFocus()
	}
}

func (self *LBLController) GetKeybindings(opts types.KeybindingsOpts) []*types.Binding {
	return []*types.Binding{
		{
			Tag:     "navigation",
			Key:     opts.GetKey(opts.Config.Universal.PrevItemAlt),
			Handler: self.withRenderAndFocus(self.HandlePrevLine),
		},
		{
			Tag:     "navigation",
			Key:     opts.GetKey(opts.Config.Universal.PrevItem),
			Handler: self.withRenderAndFocus(self.HandlePrevLine),
		},
		{
			Tag:     "navigation",
			Key:     opts.GetKey(opts.Config.Universal.NextItemAlt),
			Handler: self.withRenderAndFocus(self.HandleNextLine),
		},
		{
			Tag:     "navigation",
			Key:     opts.GetKey(opts.Config.Universal.NextItem),
			Handler: self.withRenderAndFocus(self.HandleNextLine),
		},
		{
			Tag:         "navigation",
			Key:         opts.GetKey(opts.Config.Universal.PrevBlock),
			Handler:     self.withRenderAndFocus(self.HandlePrevHunk),
			Description: self.c.Tr.PrevHunk,
		},
		{
			Tag:     "navigation",
			Key:     opts.GetKey(opts.Config.Universal.PrevBlockAlt),
			Handler: self.withRenderAndFocus(self.HandlePrevHunk),
		},
		{
			Tag:         "navigation",
			Key:         opts.GetKey(opts.Config.Universal.NextBlock),
			Handler:     self.withRenderAndFocus(self.HandleNextHunk),
			Description: self.c.Tr.NextHunk,
		},
		{
			Tag:     "navigation",
			Key:     opts.GetKey(opts.Config.Universal.NextBlockAlt),
			Handler: self.withRenderAndFocus(self.HandleNextHunk),
		},
		{
			Tag:         "navigation",
			Key:         opts.GetKey(opts.Config.Main.ToggleDragSelect),
			Handler:     self.withRenderAndFocus(self.HandleToggleSelectRange),
			Description: self.c.Tr.ToggleDragSelect,
		},
		{
			Tag:         "navigation",
			Key:         opts.GetKey(opts.Config.Main.ToggleDragSelectAlt),
			Handler:     self.withRenderAndFocus(self.HandleToggleSelectRange),
			Description: self.c.Tr.ToggleDragSelect,
		},
		{
			Tag:         "navigation",
			Key:         opts.GetKey(opts.Config.Main.ToggleSelectHunk),
			Handler:     self.withRenderAndFocus(self.HandleToggleSelectHunk),
			Description: self.c.Tr.ToggleSelectHunk,
		},
		{
			Tag:         "navigation",
			Key:         opts.GetKey(opts.Config.Universal.PrevPage),
			Handler:     self.withRenderAndFocus(self.HandlePrevPage),
			Description: self.c.Tr.LcPrevPage,
		},
		{
			Tag:         "navigation",
			Key:         opts.GetKey(opts.Config.Universal.NextPage),
			Handler:     self.withRenderAndFocus(self.HandleNextPage),
			Description: self.c.Tr.LcNextPage,
		},
		{
			Tag:         "navigation",
			Key:         opts.GetKey(opts.Config.Universal.GotoTop),
			Handler:     self.withRenderAndFocus(self.HandleGotoTop),
			Description: self.c.Tr.LcGotoTop,
		},
		{
			Tag:         "navigation",
			Key:         opts.GetKey(opts.Config.Universal.GotoBottom),
			Description: self.c.Tr.LcGotoBottom,
			Handler:     self.withRenderAndFocus(self.HandleGotoBottom),
		},
		{
			Tag:     "navigation",
			Key:     opts.GetKey(opts.Config.Universal.ScrollLeft),
			Handler: self.withRenderAndFocus(self.HandleScrollLeft),
		},
		{
			Tag:     "navigation",
			Key:     opts.GetKey(opts.Config.Universal.ScrollRight),
			Handler: self.withRenderAndFocus(self.HandleScrollRight),
		},
		{
			Key:         opts.GetKey(opts.Config.Universal.StartSearch),
			Handler:     func() error { self.c.OpenSearch(); return nil },
			Description: self.c.Tr.LcStartSearch,
		},
		{
			Key:         opts.GetKey(opts.Config.Universal.CopyToClipboard),
			Handler:     self.CopySelectedToClipboard,
			Description: self.c.Tr.LcCopySelectedTexToClipboard,
		},
	}
}

func (self *LBLController) GetMouseKeybindings(opts types.KeybindingsOpts) []*gocui.ViewMouseBinding {
	return []*gocui.ViewMouseBinding{
		{
			ViewName: self.context.GetViewName(),
			Key:      gocui.MouseWheelUp,
			Handler: func(gocui.ViewMouseBindingOpts) error {
				return self.HandleScrollUp()
			},
		},
		{
			ViewName: self.context.GetViewName(),
			Key:      gocui.MouseWheelDown,
			Handler: func(gocui.ViewMouseBindingOpts) error {
				return self.HandleScrollDown()
			},
		},
		{
			ViewName: self.context.GetViewName(),
			Key:      gocui.MouseLeft,
			Handler: func(opts gocui.ViewMouseBindingOpts) error {
				if self.isFocused() {
					return self.withRenderAndFocus(self.HandleMouseDown)()
				}

				return self.c.PushContext(self.context, types.OnFocusOpts{
					ClickedWindowName:  self.context.GetWindowName(),
					ClickedViewLineIdx: opts.Y,
				})
			},
		},
		{
			ViewName: self.context.GetViewName(),
			Key:      gocui.MouseLeft,
			Modifier: gocui.ModMotion,
			Handler: func(gocui.ViewMouseBindingOpts) error {
				return self.withRenderAndFocus(self.HandleMouseDrag)()
			},
		},
	}
}
