package main

import (
	"github.com/innoai-tech/infra/pkg/cli"
	"github.com/octohelm/gengo/pkg/agentskill"
)

func init() {
	cli.AddTo(App, &SkillsInstall{})
}

type SkillsInstall struct {
	cli.C `name:"skills-install"`

	agentskill.SkillsInstaller
}
