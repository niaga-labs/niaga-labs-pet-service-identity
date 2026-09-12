---
name: project_state
description: The resume point for this repo - current checkpoint (sha, environment, open units table, recommended next unit) at the top, earlier checkpoints below. Read first in every session; rewritten by /recap.
metadata:
  type: project
---

## 2026-08-31 state (resume here)

- **Repo:** `main` @ `36ed5a7` - Merge pull request #1 from Kilat-Pet-Delivery/add-mit-license
- **Environment:** dev-infra stack up (`./dev.ps1 up kilat`). Database `kilat_identity` is migrated and clean.
- **Open units**

| Unit / ticket | State | Blocked on | Note |
|---|---|---|---|
| KPD-2 cmd/migrate and docs | In Review | review | PR #2 |
| KPD-56 referrals and user_referral_codes SQL migration | In Review | review | PR #3, stacked on #2 |
| KPD-63 gofmt / KPD-6 .env.example | In Review | review | PRs #4, #5 |
| KPD-53 CI caller | In Progress | **a credential scope** | committed locally only |

- **Recommended next unit:** merge PR #2 then #3 - #3 is stacked on it.
- **Waiting on Luqman:** merge the open PRs above. Several are stacked, so order matters.

## Earlier checkpoints

(none - this layer was created 2026-08-31 under KPD-51)
