---
name: reconcile-docs
description: A procedure for systematically detecting and reconciling drift in AGENTS.md files, source code files, or any other file that contains codebase documentation. Use when the user asks you to "update documentation", "reconcile docs", or explicitly asks you to ensure existing documentation is up-to-date.
---

Follow this procedure to reconcile drift between documentation files and the actual state of the codebase.

## Identify Scope

Determine which files to reconcile before proceeding.

| Situation            | Scope                                |
| -------------------- | ------------------------------------ |
| User specified files | Those files                          |
| No specification     | Repo's AGENTS.md(s) + root README.md |

> **Note:**: If the README.md contains generic boilerplate for a framework or project starter, skip it. If no AGENTS.md or README.md exists, defer to the user to define scope.

## Procedure

### Step 1 — Check file paths, folders, and named references

For each referenced file path, folder, class, type, or function name:

- **If a file or folder is missing** → check if it exists elsewhere. If found in a new location → update the reference.
- **If a named symbol is missing** → infer whether it was renamed or refactored. If renamed → update the reference. If gone entirely → remove the documentation cleanly, leaving no gaps.

### Step 2 — Verify claims against codebase state

Identify specific claims made in the documentation and verify each one is still true.

A **claim** is any statement that asserts something concrete about the codebase — for example:

- How something works ("X is responsible for Y")
- Relationships between components ("A calls B", "C extends D")
- Behavioral descriptions ("this function returns...", "this module handles...")
- Configuration or setup facts ("the default value is...", "requires X to be set")

For each claim:

- **If still true** → leave it as-is.
- **If partially true or outdated** → update it to reflect the current state.
- **If no longer true and no replacement exists** → remove it cleanly, ensuring no gaps are left behind.

> **Note:** Only change a claim if you can verify its current state from the codebase directly. If you cannot determine whether a claim is still true, do not update it, and cite the claim in your response, asking the user for clarification.

### Step 3 — Final review pass

Re-read every file you changed and correct any:

- Typos
- Formatting issues
- Grammar issues

> **Note:** If your changes leave missing context, fill it in. Do not add new content speculatively — only add what is necessary to keep the documentation coherent.
