# Technical & Product Decisions

A running log of notable decisions — why we made them, what alternatives we considered, and what context future-us needs to understand them.

---

## 2026-03-12 — Source of Truth: Bitbucket; Issues: GitHub

**Decision:** Push code to Bitbucket (`git@bitbucket.org:superiortechnologies/ichor-backend.git`) as the primary remote, but keep GitHub (`git@github.com:timmaaaz/ichor.git`) for issue tracking and as a backup.

**Why:**
- Team repo lives on the Superior Technologies Bitbucket workspace
- GitHub Issues is more mature and the existing issue history is there
- GitHub and Bitbucket are independent — you can push to Bitbucket while keeping issues/PRs on GitHub without any conflict

**Alternatives considered:**
- Full migration to Bitbucket (issues + code) — rejected because Bitbucket Issues and its CLI (`bb`) are less capable than GitHub Issues + `gh`
- Full migration to GitHub only — not viable, team workspace is on Bitbucket

**How remotes are set up locally:**
```bash
origin    git@bitbucket.org:superiortechnologies/ichor-backend.git  # primary push target
github    git@github.com:timmaaaz/ichor.git                         # issues + backup
```

**Tooling note:** The `/fix-issue` skill uses the `gh` CLI and continues to work unchanged since GitHub is still the issue tracker.

---
