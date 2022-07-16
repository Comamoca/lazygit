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
