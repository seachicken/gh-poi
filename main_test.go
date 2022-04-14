package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/fatih/color"
	"github.com/stretchr/testify/assert"
)

func Test_DeletingBranchesWhenTheDryRunOptionIsFalse(t *testing.T) {
	onlyCI(t)

	actual := captureOutput(func() { runMain(false) })

	expected := fmt.Sprintf("%s %s", green("✔"), "Deleting branches...")
	assert.Contains(t, actual, expected)
}

func Test_DoNotDeleteBranchesWhenTheDryRunOptionIsTrue(t *testing.T) {
	onlyCI(t)

	actual := captureOutput(func() { runMain(true) })

	expected := fmt.Sprintf("%s %s", hiBlack("-"), "Deleting branches...")
	assert.Contains(t, actual, expected)
}

func onlyCI(t *testing.T) {
	if os.Getenv("CI") == "" {
		t.Skip("skipping test in local")
	}

	os.Chdir("ci-test")
}

func captureOutput(f func()) string {
	r, w, _ := os.Pipe()
	os.Stdout = w
	color.Output = w

	f()

	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)

	return buf.String()
}
