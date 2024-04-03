package core

import (
	"os"
	"runtime"
	"testing"
)

func TestSummary(t *testing.T) {
	err := os.RemoveAll(".out/core_test")
	if err != nil {
		t.Error(err)
	}
	err = os.MkdirAll(".out/core_test", 0755)
	if err != nil {
		t.Error(err)
	}
	t.Setenv("GITHUB_OUTPUT", ".out/core_test/GITHUB_OUTPUT.txt")
	t.Setenv("GITHUB_PATH", ".out/core_test/GITHUB_PATH.txt")
	t.Setenv("GITHUB_ENV", ".out/core_test/GITHUB_ENV.txt")
	t.Setenv("GITHUB_STEP_SUMMARY", ".out/core_test/GITHUB_STEP_SUMMARY.txt")
	t.Logf("Summary={buffer=%v,path=%v}", Summary.buffer, Summary.path)
	Summary.AddLink("GitHub link", "https://github.com")
	_, err = Summary.Write(nil)
	if err != nil {
		t.Error(err)
	}
	githubStepSummary, err := os.ReadFile(os.Getenv("GITHUB_STEP_SUMMARY"))
	if err != nil {
		t.Error(err)
	}
	githubStepSummaryString := string(githubStepSummary)
	if githubStepSummaryString != "<a href=\"https://github.com\">GitHub link</a>\n" {
		t.Error(err)
	}
}

func TestPlatform(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("not linux")
	}
	details, err := Platform.GetDetails()
	if err != nil {
		t.Error(err)
	}
	if !details.IsLinux {
		t.Error("not linux")
	}
}
