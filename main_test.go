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

func Test_DeletingBranchesWhenDryRunOptionIsFalse(t *testing.T) {
	onlyCI(t)

	results := captureOutput(func() { runMain(Merged, false, false) })

	expected := fmt.Sprintf("%s %s", green("✔"), "Deleting branches...")
	assert.Contains(t, results, expected)
}

func Test_DoNotDeleteBranchesWhenDryRunOptionIsTrue(t *testing.T) {
	onlyCI(t)

	results := captureOutput(func() { runMain(Merged, true, false) })

	expected := fmt.Sprintf("%s %s", hiBlack("-"), "Deleting branches...")
	assert.Contains(t, results, expected)
}

func Test_ProtectAndUnprotect(t *testing.T) {
	onlyCI(t)

	runProtect([]string{"main"}, false)
	protectResults := captureOutput(func() { runMain(Merged, true, false) })
	expected := fmt.Sprintf("main %s", hiBlack("[protected]"))
	assert.Contains(t, protectResults, expected)

	runUnprotect([]string{"main"}, false)
	unprotectResults := captureOutput(func() { runMain(Merged, true, false) })
	assert.NotContains(t, unprotectResults, expected)
}

func onlyCI(t *testing.T) {
	if os.Getenv("CI") == "" {
		t.Skip("skipping test in local")
	}

	os.Chdir("ci-test")
}

func captureOutput(f func()) string {
	org := os.Stdout
	defer func() {
		os.Stdout = org
	}()

	r, w, _ := os.Pipe()
	os.Stdout = w
	color.Output = w

	f()

	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)

	return buf.String()
}
