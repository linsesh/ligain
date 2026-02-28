---
description: Commit files modified during this conversation
model: eu.anthropic.claude-haiku-4-5-20251001-v1:0
allowed-tools:
  - Bash(git add *)
  - Bash(git commit *)
  - Bash(git log *)
  - Bash(git status *)
---

# Commit Conversation Changes

Create a git commit containing only the files you modified during this conversation.

## Instructions

1. **Identify modified files**: Review your conversation history and list all files you modified using Edit or Write tools during this session. Only include files YOU changed, not files that were already modified before the conversation started.

2. **If no files were modified**: If you haven't modified any files in this conversation, inform the user and do not proceed.

3. **Stage the files**: Run `git add` for each file you modified (use specific file paths, not `git add .` or `git add -A`).

4. **Check recent commit style**: Run `git log --oneline -5` to see the repository's commit message style.

5. **Generate commit message**: Based on the changes you made, write a conventional commit message:
   - Format: `type(scope): concise description`
   - Types: feat, fix, refactor, test, docs, chore, style, perf
   - Keep the title under 72 characters
   - **Keep it simple**: title says what changed and why — no bullet lists, no implementation details, no how
   - Only add a body if the why truly cannot fit in the title

6. **Author** IMPORTANT: do not include claude as a co-author

7. **Create the commit**: Use a HEREDOC for the commit message:
   ```bash
   git commit -m "$(cat <<'EOF'
   type(scope): description

   Optional body explaining the changes.
   EOF
   )"
   ```

8. **If pre-commit hooks fail**: Hooks (e.g. formatters) may modify staged files, which unstages them. If the commit fails due to a hook, re-stage the same files with `git add` and retry the commit. 
VERY IMPORTANT Do NOT use `--no-verify` to skip hooks

9. **Verify**: Run `git status` to confirm the commit was successful.


## Important

- Do NOT stage files you didn't modify in this conversation
- Do NOT use `git add .` or `git add -A`
- Do NOT commit if there are no conversation-modified files
- Do NOT include co-authored by Claude
