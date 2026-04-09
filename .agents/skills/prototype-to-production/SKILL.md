---
name: prototype-to-production
description: Tutorial for taking prototype code and refactoring it into clean, production ready code.
disable-model-invocation: true
---

Prototype code often implements the user's desired functionality, but has poor architecture, and bad maintainability:
- Poor encapsulation
- bad naming conventions
- bad or lacking documentation
- bloated files with multiple, disjunct concerns

When dealing with prototype code:
- Remember: The functionality might already be implemented; but the code itself is sloppy. Try to simplify the code through refactoring and applying polish. Consider alternative implementation directions entirely.

A production ready codebase looks like:
- Good naming conventions for files, classes, functions
- consistent, pleasant documentation
- files have a unified, (ideally) single concern
- the most contextually appropriate, simplest design patterns are used
- easy for developers to understand, navigate, and modify