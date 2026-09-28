# Decisions

## 2026-05-20: PR gate comments are best effort

Context: The PR gate must execute pull request code to run local Makefile targets, and reviewers should get a PR-visible status summary when possible.

Decision: Run local checks on the `pull_request` event with read-only repository permissions. Post the PR summary from a separate job that does not check out or execute pull request code, using an upserted issue comment marked with `<!-- capoci-pr-gate-report -->`.

Consequences: Same-repository PRs can receive an updated summary comment. Fork PRs may only get the workflow check and job summary if GitHub restricts the token to read-only permissions. The workflow avoids `pull_request_target` for executing untrusted pull request code.
