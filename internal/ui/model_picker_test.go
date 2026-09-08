package ui

import (
	"testing"

	"github.com/frknikiz/curspace/internal/litellm"
)

func TestFilteredModelIndexes(t *testing.T) {
	m := NewAppModel(AppConfig{})
	m.modelPick.models = []litellm.Model{{ID: "gpt-5"}, {ID: "claude-sonnet"}, {ID: "gpt-5-mini"}}
	m.modelPick.search = "GPT-5"
	got := m.filteredModelIndexes()
	if len(got) != 2 || got[0] != 0 || got[1] != 2 {
		t.Fatalf("filtered indexes = %#v", got)
	}
}
