package controllers

import (
	"github.com/jesseduffield/gocui"
	"github.com/jesseduffield/lazygit/pkg/gui/types"
)

type StagingController struct {
	baseController
	*controllerCommon

	context      types.ILBLContext
	otherContext types.Context
}

var _ types.IController = &StagingController{}

func NewStagingController(
	common *controllerCommon,
	context types.ILBLContext,
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
	return []*types.Binding{
		{
			Key:         opts.GetKey(opts.Config.Universal.OpenFile),
			Handler:     self.OpenFile,
			Description: self.c.Tr.LcOpenFile,
		},
		{
			Key:         opts.GetKey(opts.Config.Universal.Edit),
			Handler:     self.EditFile,
			Description: self.c.Tr.LcEditFile,
		},
		{
			Key:         opts.GetKey(opts.Config.Universal.Return),
			Handler:     self.Escape,
			Description: self.c.Tr.ReturnToFilesPanel,
		},
	}
}

func (self *StagingController) Context() types.Context {
	return self.context
}

func (self *StagingController) GetMouseKeybindings(opts types.KeybindingsOpts) []*gocui.ViewMouseBinding {
	return []*gocui.ViewMouseBinding{}
}

func (self *StagingController) OpenFile() error {
	path := self.contexts.Files.GetSelectedPath()

	if path == "" {
		return nil
	}

	lineNumber := self.context.GetState().CurrentLineNumber()
	return self.helpers.Files.OpenFileAtLine(path, lineNumber)
}

func (self *StagingController) EditFile() error {
	path := self.contexts.Files.GetSelectedPath()

	if path == "" {
		return nil
	}

	lineNumber := self.context.GetState().CurrentLineNumber()
	return self.helpers.Files.EditFileAtLine(path, lineNumber)
}

func (self *StagingController) Escape() error {
	return self.c.PushContext(self.contexts.Files)
}
