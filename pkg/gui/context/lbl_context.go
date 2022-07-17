package context

import (
	"github.com/jesseduffield/gocui"
	"github.com/jesseduffield/lazygit/pkg/gui/lbl"
	"github.com/jesseduffield/lazygit/pkg/gui/types"
)

type LBLContext struct {
	*SimpleContext

	lblState               *lbl.State
	viewTrait              *ViewTrait
	getIncludedLineIndices func() []int
	c                      *types.HelperCommon
}

var _ types.ILBLContext = (*LBLContext)(nil)

func NewLBLContext(
	view *gocui.View,
	windowName string,
	key types.ContextKey,

	// TODO: see if we need to pass these
	onFocus func(types.OnFocusOpts) error,
	onFocusLost func(opts types.OnFocusLostOpts) error,
	getIncludedLineIndices func() []int,

	c *types.HelperCommon,
) *LBLContext {
	return &LBLContext{
		lblState:               nil,
		viewTrait:              NewViewTrait(view),
		c:                      c,
		getIncludedLineIndices: getIncludedLineIndices,
		SimpleContext: NewSimpleContext(NewBaseContext(NewBaseContextOpts{
			View:       view,
			WindowName: windowName,
			Key:        key,
			Kind:       types.MAIN_CONTEXT,
			Focusable:  true,
		}), ContextCallbackOpts{
			OnFocus:     onFocus,
			OnFocusLost: onFocusLost,
		}),
	}
}

func (self *LBLContext) GetState() *lbl.State {
	return self.lblState
}

func (self *LBLContext) SetState(state *lbl.State) {
	self.lblState = state
}

func (self *LBLContext) GetViewTrait() types.IViewTrait {
	return self.viewTrait
}

func (self *LBLContext) GetIncludedLineIndices() []int {
	return self.getIncludedLineIndices()
}

func (self *LBLContext) RenderAndFocus() error {
	self.GetView().SetContent(self.GetContentToRender())

	if err := self.focusSelection(); err != nil {
		return err
	}

	self.c.Render()

	return nil
}

func (self *LBLContext) Render() error {
	self.GetView().SetContent(self.GetContentToRender())

	self.c.Render()

	return nil
}

func (self *LBLContext) Focus() error {
	if err := self.focusSelection(); err != nil {
		return err
	}

	self.c.Render()

	return nil
}

func (self *LBLContext) focusSelection() error {
	view := self.GetView()
	state := self.GetState()
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

func (self *LBLContext) GetContentToRender() string {
	return self.GetState().RenderForLineIndices(self.GetIncludedLineIndices())
}

func (self *LBLContext) NavigateTo(selectedLineIdx int) error {
	self.GetState().SetLineSelectMode()
	self.GetState().SelectLine(selectedLineIdx)

	return self.RenderAndFocus()
}
