package command_test

import (
	"io"
	"os"
	"runtime"
	"testing"

	"github.com/jcbhmr/go-toolkit/actionscore/internal/command"
)

var originalStdout *os.File

func beforeAll() {
	originalStdout = os.Stdout
}

func TestMain(m *testing.M) {
	beforeAll()
	exitCode := m.Run()
	os.Exit(exitCode)
}

func beforeEach(t *testing.T) {
	t.Helper()
	var err error
	os.Stdout, err = os.CreateTemp("", "command_test.log")
	if err != nil {
		t.Fatal(err)
	}
}

func afterEach() {
	fakeStdout := os.Stdout
	os.Stdout = originalStdout
	err := fakeStdout.Close()
	if err != nil {
		panic(err)
	}
	err = os.Remove(fakeStdout.Name())
	if err != nil {
		panic(err)
	}
}

func assertStdout(t *testing.T, expected string) {
	t.Helper()
	os.Stdout.Sync()
	bytes, err := os.ReadFile(os.Stdout.Name())
	if err != nil {
		t.Fatal(err)
	}
	if string(bytes) != expected {
		t.Errorf("expected %q, got %q", expected, string(bytes))
	}
}

func resetFakeStdout(t *testing.T) {
	_, err := os.Stdout.Seek(0, io.SeekStart)
	if err != nil {
		t.Fatal(err)
	}
	err = os.Stdout.Truncate(0)
	if err != nil {
		t.Fatal(err)
	}
}

var eol = func()string{
	if runtime.GOOS == "windows" {
		return "\r\n"
	}
	return "\n"
}()

func TestCommandOnly(t *testing.T) {
	beforeEach(t)
	t.Cleanup(afterEach)
	
	command.IssueCommand("some-command", command.CommandProperties{}, "")
	assertStdout(t, "::some-command::"+eol)
}

func TestCommandEscapesMessage(t *testing.T) {
	beforeEach(t)
	t.Cleanup(afterEach)
	
	command.IssueCommand("some-command", command.CommandProperties{}, "percent % percent % cr \r cr \r lf \n lf \n")
	assertStdout(t, "::some-command::percent %25 percent %25 cr %0D cr %0D lf %0A lf %0A"+eol)

	resetFakeStdout(t)

	command.IssueCommand("some-command", command.CommandProperties{}, "%25 %25 %0D %0D %0A %0A")
	assertStdout(t, "::some-command::%2525 %2525 %250D %250D %250A %250A"+eol)
}

func TestCommandEscapesProperty(t *testing.T) {
	beforeEach(t)
	t.Cleanup(afterEach)

	command.IssueCommand("some-command", command.CommandProperties{
		"name": "percent % percent % cr \r cr \r lf \n lf \n colon : colon : comma , comma ,",
	}, "")
	assertStdout(t, "::some-command name=percent %25 percent %25 cr %0D cr %0D lf %0A lf %0A colon %3A colon %3A comma %2C comma %2C::"+eol)

	resetFakeStdout(t)

	command.IssueCommand("some-command", command.CommandProperties{}, "%25 %25 %0D %0D %0A %0A %3A %3A %2C %2C")
	assertStdout(t, "::some-command::%2525 %2525 %250D %250D %250A %250A %253A %253A %252C %252C"+eol)
}
