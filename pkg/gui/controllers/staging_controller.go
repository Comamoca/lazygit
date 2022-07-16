package controllers

import (
	"github.com/jesseduffield/gocui"
	"github.com/jesseduffield/lazygit/pkg/gui/types"
)

type StagingController struct {
	baseController
	*controllerCommon

	context      types.Context
	otherContext types.Context
}

var _ types.IController = &StagingController{}

func NewStagingController(
	common *controllerCommon,
	context types.Context,
	otherContext types.Context,
) *StagingController {
	return &StagingController{
		baseController:   baseController{},
		controllerCommon: common,
		context:          context,
		otherContext:     otherContext,
	}
}

func (self *StagingController) GetKeybindings(opts types.KeybindingsOpts) []*types.Binding {
	return []*types.Binding{}
}

func (self *StagingController) Context() types.Context {
	return self.context
}

func (self *StagingController) GetMouseKeybindings(opts types.KeybindingsOpts) []*gocui.ViewMouseBinding {
	return []*gocui.ViewMouseBinding{
		{
			ViewName: self.context.GetViewName(),
			Key:      gocui.MouseLeft,
			Handler: func(opts gocui.ViewMouseBindingOpts) error {
				return self.c.PushContext(self.context, types.OnFocusOpts{
					ClickedWindowName:  self.context.GetWindowName(),
					ClickedViewLineIdx: opts.Y,
				})
			},
		},
	}
}
