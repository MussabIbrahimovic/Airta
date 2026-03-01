package tests

import (
	"airt/internal/airt/runtime"
	"os"
	"path/filepath"
	"testing"
)

func TestRuntimeLoadsExample(t *testing.T) {
	rt := runtime.New(nil)
	m, err := rt.RunFile(filepath.Join("..", "examples", "app", "main.aa"))
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := m["start"]; !ok {
		t.Fatal("missing start export")
	}
}

func TestParseFailure(t *testing.T) {
	rt := runtime.New(nil)
	f := filepath.Join("..", "tests", "broken.aa")
	_ = os.WriteFile(f, []byte("let x ="), 0644)
	defer os.Remove(f)
	_, err := rt.RunFile(f)
	if err == nil {
		t.Fatal("expected parse failure")
	}
}
