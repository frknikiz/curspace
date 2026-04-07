## Scope
These instructions apply to the entire repository unless a deeper `AGENTS.md` overrides them.

## Working style
- Keep changes tightly scoped to the user's request.
- Prefer existing patterns and libraries over introducing new abstractions.
- Fix root causes when practical; avoid cosmetic churn and unrelated refactors.
- Preserve user work in progress and call out assumptions or blockers early.

## File and code changes
- Read the relevant files before editing and keep context limited to what the task needs.
- Follow local naming, structure, and formatting conventions already present in the codebase.
- Update nearby documentation when behavior, setup, or developer workflow changes.
- Add comments only when the code would otherwise be hard to understand.

## Validation
- Run the narrowest relevant check first, then broader validation only if needed.
- If you cannot run validation, state that clearly in the handoff.
- Do not fix unrelated failing tests or tooling unless the user asks.

## Safety
- Do not overwrite or revert changes you did not make.
- Avoid destructive commands unless the user explicitly requests them.
- Ask before making changes with wide impact, such as dependency upgrades or file moves.

## Handoff
- Summarize the files changed, what was verified, and any remaining risks or follow-up work.
