---
description: Make a suggestion of squash by comparing current branch and main
model: eu.anthropic.claude-haiku-4-5-20251001-v1:0
---

# Squash Commit Summary

Analyze all commits on the current branch compared to main and propose a summary commit message for squashing.

## Instructions

1. Run `git log main..HEAD --oneline --reverse` to list all commits on the branch
2. Run `git log main..HEAD --reverse --format="=== %h ===\n%B\n--- Stats ---\n" --stat` to get detailed commit messages and file statistics
3. Run `git diff main..HEAD --stat | tail -5` to get total diff statistics

## Analysis

For each commit, identify:
- The type (feat, fix, refactor, test, docs, chore, etc.)
- The scope (which component/module it affects)
- The key changes made

## Output

Provide:
1. A brief table summarizing each commit (hash, type, short description)
2. Total impact (files changed, lines added/removed)
3. A proposed squash commit message following conventional commits format:
   - Title: `type(scope): concise description`
   - Body: explain what the feature/change does and why
   - List major components or sections if applicable
   - **Do NOT include Co-Authored-By lines**

The commit message should:
- Compress the development history into a single logical narrative
- Focus on WHAT was added/changed, not the development journey
- Be suitable for the main branch changelog
- Use present tense ("add" not "added")
