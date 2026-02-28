---
description: Analyze all branch changes against main for code quality, SOLID, tests, and factorization opportunities
model: eu.anthropic.claude-sonnet-4-5-20250929-v1:0
allowed-tools:
  - Bash(git diff *)
  - Bash(git log *)
  - Bash(git merge-base *)
  - Read
  - Glob
  - Grep
---

# Code Quality Analysis — Branch vs Main

Perform a thorough code quality review of all changes in the current branch compared to `main`. Output a structured markdown report.

## Step 1: Gather the diff

1. Run `git merge-base main HEAD` to find the common ancestor
2. Run `git diff main...HEAD --stat` to get a summary of changed files
3. Run `git diff main...HEAD` to get the full diff
4. For each modified file, read the **full file** (not just the diff) so you can evaluate the change in context

## Step 2: Analyze each changed file

For every file with meaningful changes, evaluate the following dimensions:

### Code Smells
- Long methods or functions (>20 lines of logic)
- Deep nesting (>3 levels)
- God classes or modules doing too many things
- Magic numbers or strings without named constants
- Dead code, unused imports, commented-out code
- Mutable global state
- Overly complex conditionals

### SOLID Principles
- **S** — Single Responsibility: Does each class/module have one reason to change?
- **O** — Open/Closed: Is the code open for extension but closed for modification?
- **L** — Liskov Substitution: Are subtypes properly substitutable?
- **I** — Interface Segregation: Are interfaces lean and focused?
- **D** — Dependency Inversion: Does business logic depend on abstractions, not concrete implementations?

### Clean Architecture
- Are layers properly separated (domain, use cases, infrastructure)?
- Is there leaking of infrastructure concerns into business logic?
- Are dependencies pointing inward (infrastructure depends on domain, never the reverse)?
- Is there proper use of dependency injection?

### Test Quality
Follow the project's testing conventions:
- Are tests marked with `@pytest.mark.unit` or `@pytest.mark.integration`?
- Is `@pytest.mark.parametrize` used where appropriate?
- Do tests follow the Arrange-Act-Assert structure?
- Are services tested with mocks rather than testing mocks directly?
- Are tests testing behavior, not implementation details?
- Is there adequate coverage of edge cases?
- For web UI tests: is the right tool used (BeautifulSoup for HTML structure, Playwright for interactions)?

### Factorization Opportunities
- Duplicated logic across files or within the same file
- Similar patterns that could share a common abstraction
- Copy-pasted code with minor variations
- Utility functions that should be extracted to shared libs
- Repeated error handling or validation patterns

## Step 3: Produce the report

Output a single markdown document with this structure:

```
# Code Quality Analysis Report

**Branch**: `<branch_name>`
**Compared to**: `main`
**Files analyzed**: <count>
**Date**: <today>

## Executive Summary

<2-3 sentences: overall quality assessment and top priorities>

## Findings by Category

### Code Smells
| File | Line(s) | Issue | Severity |
|------|---------|-------|----------|
| ... | ... | ... | ... |

<Details and explanations for each finding>

### SOLID Violations
| File | Principle | Issue | Severity |
|------|-----------|-------|----------|
| ... | ... | ... | ... |

<Details and explanations for each finding>

### Clean Architecture
| File | Issue | Severity |
|------|-------|----------|
| ... | ... | ... |

<Details and explanations for each finding>

### Test Quality
| File | Issue | Severity |
|------|-------|----------|
| ... | ... | ... |

<Details and explanations for each finding>

### Factorization Opportunities
| Files | Pattern | Suggested Refactor |
|-------|---------|-------------------|
| ... | ... | ... |

<Details and explanations for each opportunity>

## Summary

| Category | Issues Found |
|-------------------------|:------:|
| Code Smells | N |
| SOLID Violations | N |
| Clean Architecture | N |
| Test Quality | N |
| Factorization Opportunities | N |

## Top 3 Recommendations

1. **...**: ...
2. **...**: ...
3. **...**: ...
```

## Rules

- **Severity levels**: use `critical`, `warning`, `info`
- **Be specific**: always reference the exact file and line number(s)
- **Be constructive**: for each issue, briefly suggest how to fix it
- **No false positives**: only report genuine issues, not style preferences
- **Respect existing patterns**: if the codebase consistently uses a pattern, don't flag it unless it's genuinely problematic
- **Focus on the diff**: analyze only changed/added code, but use the full file context to understand it
- If a category has zero findings, write "No issues found." and move on
