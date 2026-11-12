# Agents

Guidance for any coding agent working in this repository. It is the estate's contract-first Go and Vue template: an ogen API, a Vue SPA and Postgres, built into one image.

The repo's agent contract is [CLAUDE.md](CLAUDE.md); the estate-wide conventions are at https://github.com/JorisJonkers-dev/workspace/blob/main/CLAUDE.md.

## Build and test

Tasks are defined in `Taskfile.yml`; `task --list` shows them. Tool versions are pinned in `mise.toml`.

- `task check` runs lint, generation check, tests, build and the secrets scan
- `task e2e` runs the end-to-end suite
- `task gen` regenerates the generated code
