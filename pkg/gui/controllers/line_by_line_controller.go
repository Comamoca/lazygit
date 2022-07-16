package controllers

import (
	"github.com/jesseduffield/gocui"
	"github.com/jesseduffield/lazygit/pkg/gui/types"
)

type LBLControllerFactory struct {
	c *types.HelperCommon
}

func NewLBLControllerFactory(c *types.HelperCommon) *LBLControllerFactory {
	return &LBLControllerFactory{
		c: c,
	}
}

func (self *LBLControllerFactory) Create(context types.ILBLContext) *LBLController {
	return &LBLController{
		baseController: baseController{},
		c:              self.c,
		context:        context,
	}
}

type LBLController struct {
	baseController
	c *types.HelperCommon

	context types.ILBLContext
}

func (self *LBLController) Context() types.Context {
	return self.context
}

func (self *LBLController) HandlePrevLine() error {
	self.context.GetState().CycleSelection(false)

	return self.refreshAndFocus()
}

func (self *LBLController) HandleNextLine() error {
	self.context.GetState().CycleSelection(true)

	return self.refreshAndFocus()
}

func (self *LBLController) HandleScrollLeft() error {
	return self.scrollHorizontal(self.context.GetViewTrait().ScrollLeft)
}

func (self *LBLController) HandleScrollRight() error {
	return self.scrollHorizontal(self.context.GetViewTrait().ScrollRight)
}

func (self *LBLController) HandleScrollUp() error {
	self.context.GetViewTrait().ScrollUp()

	return self.refreshAndFocus()
}

func (self *LBLController) HandleScrollDown() error {
	self.context.GetViewTrait().ScrollDown()

	return self.refreshAndFocus()
}

func (self *LBLController) scrollHorizontal(scrollFunc func()) error {
	scrollFunc()

	return self.refreshAndFocus()
}

func (self *LBLController) refreshAndFocus() error {
	self.context.GetView().SetContent(self.getContentToRender())
	self.c.Render()

	return nil
}

func (self *LBLController) getContentToRender() string {
	return self.context.GetState().RenderForLineIndices(self.context.GetIncludedLineIndices())
}

func (self *LBLController) HandlePrevPage() error {
	self.context.GetState().SetLineSelectMode()
	self.context.GetState().AdjustSelectedLineIdx(-self.context.GetViewTrait().PageDelta())

	return self.refreshAndFocus()
}

func (self *LBLController) HandleNextPage() error {
	self.context.GetState().SetLineSelectMode()
	self.context.GetState().AdjustSelectedLineIdx(self.context.GetViewTrait().PageDelta())

	return self.refreshAndFocus()
}

func (self *LBLController) HandleGotoTop() error {
	self.context.GetState().SelectTop()

	return self.refreshAndFocus()
}

func (self *LBLController) HandleGotoBottom() error {
	self.context.GetState().SelectBottom()

	return self.refreshAndFocus()
}

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

func (self *LBLController) GetKeybindings(opts types.KeybindingsOpts) []*types.Binding {
	return []*types.Binding{
		{Tag: "navigation", Key: opts.GetKey(opts.Config.Universal.PrevItemAlt), Handler: self.HandlePrevLine},
		{Tag: "navigation", Key: opts.GetKey(opts.Config.Universal.PrevItem), Handler: self.HandlePrevLine},
		{Tag: "navigation", Key: opts.GetKey(opts.Config.Universal.NextItemAlt), Handler: self.HandleNextLine},
		{Tag: "navigation", Key: opts.GetKey(opts.Config.Universal.NextItem), Handler: self.HandleNextLine},
		{Tag: "navigation", Key: opts.GetKey(opts.Config.Universal.PrevPage), Handler: self.HandlePrevPage, Description: self.c.Tr.LcPrevPage},
		{Tag: "navigation", Key: opts.GetKey(opts.Config.Universal.NextPage), Handler: self.HandleNextPage, Description: self.c.Tr.LcNextPage},
		{Tag: "navigation", Key: opts.GetKey(opts.Config.Universal.GotoTop), Handler: self.HandleGotoTop, Description: self.c.Tr.LcGotoTop},
		{Tag: "navigation", Key: opts.GetKey(opts.Config.Universal.GotoBottom), Description: self.c.Tr.LcGotoBottom, Handler: self.HandleGotoBottom},
		{Tag: "navigation", Key: opts.GetKey(opts.Config.Universal.ScrollLeft), Handler: self.HandleScrollLeft},
		{Tag: "navigation", Key: opts.GetKey(opts.Config.Universal.ScrollRight), Handler: self.HandleScrollRight},
		{
			Key:         opts.GetKey(opts.Config.Universal.StartSearch),
			Handler:     func() error { self.c.OpenSearch(); return nil },
			Description: self.c.Tr.LcStartSearch,
			Tag:         "navigation",
		},
	}
}

func (self *LBLController) GetMouseKeybindings(opts types.KeybindingsOpts) []*gocui.ViewMouseBinding {
	return []*gocui.ViewMouseBinding{
		{
			ViewName: self.context.GetViewName(),
			Key:      gocui.MouseWheelUp,
			Handler:  func(gocui.ViewMouseBindingOpts) error { return self.HandleScrollUp() },
		},
		{
			ViewName: self.context.GetViewName(),
			Key:      gocui.MouseWheelDown,
			Handler:  func(gocui.ViewMouseBindingOpts) error { return self.HandleScrollDown() },
		},
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
