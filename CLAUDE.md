# iofxml

## Local validation

GitHub Actions CI is off to save paid minutes, so nothing checks a pull request on GitHub. Run `scripts/check.sh` and make it pass before you push. It runs every check CI ran, so there is no `--full` mode. The rules for a change are in CONTRIBUTING.md.

The pre-push hook (`.githooks/pre-push`) runs `scripts/check.sh`. Claude Code sessions enable it through the SessionStart hook in `.claude/settings.json`; humans enable it once per clone with `git config core.hooksPath .githooks`. `git push --no-verify` skips it.
