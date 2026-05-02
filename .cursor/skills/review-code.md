---
name: review-code
description: Reviews code against Go/Gin/SQLC best practices, security, module boundaries, and correctness. Outputs a structured report with [ISSUE], [SUGGESTION], and [OK] findings. Use when the user says "review code", "review this file", or "review [filename]".
---

# Skill: Review Code

## Trigger

User says: `review code`, `review this file`, or `review [filename]`

## Steps

### 1. Activate Review Agent

Adopt the **Senior Code Reviewer & Security Engineer** persona from `agent-review.mdc`.

### 2. Identify Target

- If the user named a specific file, read that file.
- If no file was named, read the currently open or most recently edited file.
- If multiple files were changed together (e.g. handler + service + repository within a module), review all of them as a unit.

### 3. Run the Checklists

Work through all checklists from `agent-review.mdc` in order:
1. Security checklist
2. API contract checklist (response envelope, error codes, pagination)
3. Code quality checklist
4. Module boundary checklist
5. Gin-specific checklist

### 4. Output the Report

Use this exact format:

```
## Code Review: {filename}

### Findings

[ISSUE] {line or function}: {description of problem and why it matters}
[ISSUE] ...

[SUGGESTION] {line or function}: {description and recommended fix}
[SUGGESTION] ...

[OK] {what was done well}
[OK] ...

### Summary

Verdict: REQUEST CHANGES / APPROVE / NEEDS DISCUSSION

Issues to fix before merging:
1. {issue 1}
2. {issue 2}
```

### 5. Offer to Fix

After the report, ask:
> "Want me to fix the `[ISSUE]` items now?"

If yes, switch to the **backend-agent** persona and apply fixes surgically.
