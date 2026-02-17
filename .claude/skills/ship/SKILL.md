# Ship Skill

Ship the current changes: build, test, commit, and push.

## Steps

1. Run `go build ./...` and fix any compile errors
2. Run `go test ./...` and fix any test failures
3. Run `make lint` and fix any lint issues
4. Stage all changed files with `git add -A`
5. Write a conventional commit message describing the changes
6. Push to the current branch
