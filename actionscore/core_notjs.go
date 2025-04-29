//go:build !js

package core

import "github.com/jcbhmr/go-toolkit/actionscore/internal/command"

func exportVariable(name string, val any) error {
	panic("not implemented")
}

func setSecret(secret string) error {
	return command.IssueCommand("add-mask", command.CommandProperties{}, secret)
}

