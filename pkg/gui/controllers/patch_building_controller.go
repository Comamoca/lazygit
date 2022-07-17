package controllers

import (
	"github.com/jesseduffield/gocui"
	"github.com/jesseduffield/lazygit/pkg/gui/types"
)

type PatchBuildingController struct {
	baseController
	*controllerCommon
}

var _ types.IController = &PatchBuildingController{}

func NewPatchBuildingController(
	common *controllerCommon,
) *PatchBuildingController {
	return &PatchBuildingController{
		baseController:   baseController{},
		controllerCommon: common,
	}
}

func (self *PatchBuildingController) GetKeybindings(opts types.KeybindingsOpts) []*types.Binding {
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

func (self *PatchBuildingController) Context() types.Context {
	return self.contexts.PatchBuilding
}

func (self *PatchBuildingController) context() types.ILBLContext {
	return self.contexts.PatchBuilding
}

func (self *PatchBuildingController) GetMouseKeybindings(opts types.KeybindingsOpts) []*gocui.ViewMouseBinding {
	return []*gocui.ViewMouseBinding{}
}

func (self *PatchBuildingController) OpenFile() error {
	path := self.contexts.CommitFiles.GetSelectedPath()

	if path == "" {
		return nil
	}

	lineNumber := self.context().GetState().CurrentLineNumber()
	return self.helpers.Files.OpenFileAtLine(path, lineNumber)
}

func (self *PatchBuildingController) EditFile() error {
	path := self.contexts.CommitFiles.GetSelectedPath()

	if path == "" {
		return nil
	}

	lineNumber := self.context().GetState().CurrentLineNumber()
	return self.helpers.Files.EditFileAtLine(path, lineNumber)
}

func (self *PatchBuildingController) Escape() error {
	return self.c.PushContext(self.contexts.Files)
}
