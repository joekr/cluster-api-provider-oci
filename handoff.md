# PR Gate Workflow Handoff

Date: 2026-05-22
Branch: `add-pr-workflow`
Current commit: `705fcaba6b96`

## What Changed

Added `.github/workflows/pr-gate.yaml`.

The workflow runs on pull request events:

- `opened`
- `synchronize`
- `reopened`
- `ready_for_review`

It runs the requested local-only checks:

- `make build`
- `make test`
- `make generate`
- `make manifests`
- `make generate-e2e-templates`

It also checks generated artifact freshness after generation:

- `git diff --exit-code -- api exp/api config/crd config/rbac config/webhook config/default`
- `git diff --exit-code -- test/e2e/data/infrastructure-oci`

The workflow uses pinned actions:

- `actions/checkout` pinned to `de0fac2e4500dabe0009e67214ff5f5447ce83dd` (`v6.0.2`)
- `actions/setup-go` pinned to `4a3601121dd01d1626a1e23e37211e3254c1c06c` (`v6.4.0`)

Checkout uses `fetch-depth: 0` because `hack/version.sh` depends on git tags for version metadata.

## PR Output

The workflow writes a concise pass/fail table to the GitHub Actions job summary.

It also tries to upsert a PR comment marked with:

```text
<!-- capoci-pr-gate-report -->
```

The comment job is best effort. It does not checkout or execute pull request code. It only reads the check job output and calls the GitHub API.

For fork PRs, the checks should still run after maintainer approval, but the PR comment may not post because GitHub usually downgrades fork PR tokens to read-only. In that case, the job summary and workflow logs are the reliable output.

## Security Model

The workflow uses `pull_request`, not `pull_request_target`.

That means pull request code is executed with read-only repository permissions. This is intentionally safer for fork PRs.

The desired repo or org setting is to require maintainer approval before external fork workflows run. The approval policy should be:

```text
all_external_contributors
```

GitHub-side command if the maintainer has permission:

```bash
gh api \
  --method PUT \
  repos/oracle/cluster-api-provider-oci/actions/permissions/fork-pr-contributor-approval \
  -H "Accept: application/vnd.github+json" \
  -H "X-GitHub-Api-Version: 2026-03-10" \
  -f approval_policy=all_external_contributors
```

Maintainers should review the fork PR diff first, then click **Approve workflows to run** in the PR checks area or on the waiting Actions run.

## Documented But Not Implemented

These additional local-only checks were identified but intentionally left out of the first workflow version:

- `make lint`
- `go mod tidy` followed by `git diff --exit-code -- go.mod go.sum`
- `make build-book` for docs/book changes

They are documented as comments in `.github/workflows/pr-gate.yaml`.

## Validation Done

Local validation completed:

- YAML parsed successfully with Ruby's YAML parser.
- `git diff --check` passed.
- Dry-runs of the requested Make targets expanded successfully:
  - `make -n build`
  - `make -n test`
  - `make -n generate`
  - `make -n manifests`
  - `make -n generate-e2e-templates`

Full `make build`, `make test`, `make generate`, `make manifests`, and `make generate-e2e-templates` were not executed locally.

## Files Added

- `.github/workflows/pr-gate.yaml`
- `decision.md`

`decision.md` records why PR comments are best effort and why the workflow avoids `pull_request_target`.

## Suggested Next Test

Push the branch to a fork and open a PR from a branch in that fork back to the fork's `main` branch. That should test the workflow mechanics, Make steps, job summary, and possibly the PR comment.

Testing the maintainer approval UX requires a fork PR into the upstream repository with the upstream approval policy enabled.
