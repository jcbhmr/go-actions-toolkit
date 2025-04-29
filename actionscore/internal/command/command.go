package command

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/jcbhmr/go-toolkit/actionscore/internal/utils"
)

var eol = func() string {
	if runtime.GOOS == "windows" {
		return "\r\n"
	} else {
		return "\n"
	}
}()

type CommandProperties = map[string]any

func IssueCommand(command string, properties CommandProperties, message any) error {
	cmd := newCommand(command, properties, message)
	cmdStr, err := cmd.string2()
	if err != nil {
		return err
	}
	fmt.Print(cmdStr + eol)
	return nil
}

func Issue(name string, messageRaw *string) error {
	var message string
	if messageRaw != nil {
		message = *messageRaw
	}
	return IssueCommand(name, CommandProperties{}, message)
}

const cmdString = "::"

type command struct {
	command    string
	message    any
	properties CommandProperties
}

func newCommand(commandVar string, properties CommandProperties, message any) *command {
	if commandVar == "" {
		commandVar = "missing.command"
	}
	return &command{
		command:    commandVar,
		message:    message,
		properties: properties,
	}
}

func (c *command) string2() (string, error) {
	cmdStr := cmdString + c.command
	if len(c.properties) > 0 {
		cmdStr += " "
		first := true
		for k, v := range c.properties {
			if v != nil {
				if first {
					first = false
				} else {
					cmdStr += ","
				}
				ev, err := escapeProperty(v)
				if err != nil {
					return "", err
				}
				cmdStr += k + "=" + ev
			}
		}
		emessage, err := escapeData(c.message)
		if err != nil {
			return "", err
		}
		cmdStr += cmdString + emessage
		return cmdStr, nil
	}
	return "", nil
}

func escapeData(s any) (string, error) {
	str, err := utils.ToCommandValue(s)
	if err != nil {
		return "", err
	}
	return strings.NewReplacer("%", "%25", "\r", "%0D", "\n", "%0A").Replace(str), nil
}

func escapeProperty(s any) (string, error) {
	str, err := utils.ToCommandValue(s)
	if err != nil {
		return "", err
	}
	return strings.NewReplacer("%", "%25", "\r", "%0D", "\n", "%0A", ":", "%3A", ",", "%2C").Replace(str), nil
}
