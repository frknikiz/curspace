package codex

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/frknikiz/curspace/internal/terminal"
)

// Open launches Codex CLI in the primary directory with additional workspace
// directories passed via --add-dir, using a selected OpenAI API token or Codex's existing login.
func Open(primaryPath string, extraPaths []string, terminalName string, tokenName string, options ...string) error {
	if _, err := exec.LookPath("codex"); err != nil {
		return fmt.Errorf("codex command not found. Install Codex CLI and ensure 'codex' is in your PATH")
	}
	return terminal.Open(buildShellCommand(primaryPath, extraPaths, tokenName, options...), terminalName)
}

func buildShellCommand(primaryPath string, extraPaths []string, tokenName string, options ...string) string {
	var b strings.Builder
	b.WriteString("cd ")
	b.WriteString(shellQuote(primaryPath))
	b.WriteString(" && ")
	hasToken := strings.TrimSpace(tokenName) != ""
	model, baseURL := optionValues(options)
	hasProvider := hasToken || baseURL != ""
	if hasProvider {
		b.WriteString("(")
		if hasToken {
			b.WriteString("CURSPACE_CODEX_API_KEY=\"$(")
			b.WriteString(shellQuote(curspaceExecutable()))
			b.WriteString(" codex token print ")
			b.WriteString(shellQuote(tokenName))
			b.WriteString(")\" && export CURSPACE_CODEX_API_KEY && ")
		}
	}
	b.WriteString("codex")
	if len(extraPaths) > 0 {
		// Additional writable roots require a compatible sandbox policy.
		b.WriteString(" --sandbox workspace-write")
	}
	if model != "" {
		b.WriteString(" -m ")
		b.WriteString(shellQuote(model))
	}
	if hasProvider {
		// An invocation-only provider makes the API key effective in interactive
		// Codex without replacing the user's persisted ChatGPT login.
		b.WriteString(" -c ")
		providerName := "curspace_litellm"
		if baseURL == "" {
			providerName = "curspace_token"
		}
		b.WriteString(shellQuote(`model_provider="` + providerName + `"`))
		b.WriteString(" -c ")
		if baseURL == "" {
			baseURL = "https://api.openai.com/v1"
			providerName = "curspace_token"
		}
		provider := fmt.Sprintf(`model_providers.%s={name="curspace",base_url=%s,wire_api="responses"`, providerName, strconv.Quote(baseURL))
		if hasToken {
			provider += `,env_key="CURSPACE_CODEX_API_KEY"`
		}
		provider += "}"
		b.WriteString(shellQuote(provider))
	}
	for _, p := range extraPaths {
		b.WriteString(" --add-dir ")
		b.WriteString(shellQuote(p))
	}
	if hasProvider {
		b.WriteString(")")
	}
	return b.String()
}

func optionValues(options []string) (model, baseURL string) {
	if len(options) > 0 {
		model = strings.TrimSpace(options[0])
	}
	if len(options) > 1 {
		baseURL = strings.TrimSpace(options[1])
	}
	return model, baseURL
}

func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

func curspaceExecutable() string {
	exe, err := os.Executable()
	if err != nil || exe == "" {
		return "curspace"
	}
	return exe
}
