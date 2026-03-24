package ui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/frknikiz/curspace/internal/scanner"
)

var projectTypeColors = map[scanner.ProjectType]string{
	scanner.Go:     "#00ADD8",
	scanner.Node:   "#68A063",
	scanner.Java:   "#F89820",
	scanner.Python: "#FFD43B",
	scanner.Rust:   "#DEA584",
	scanner.DotNet: "#512BD4",
	scanner.PHP:    "#777BB4",
	scanner.Git:    "#F05032",
}

func projectTypeTagStyle(pt scanner.ProjectType) lipgloss.Style {
	color := "#888888"
	if v, ok := projectTypeColors[pt]; ok {
		color = v
	}

	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color(color)).
		Padding(0, 1).
		Bold(true)
}

func renderProjectTypeTag(pt scanner.ProjectType) string {
	return projectTypeTagStyle(pt).Render(string(pt))
}
