# skills

Agent skills, owned rather than installed.

Each directory under `skills/` is one skill: a `SKILL.md` with YAML frontmatter,
plus whatever reference files it needs beside it. Harnesses that support skills
(Claude Code, OpenCode) discover them per-project, which is why consuming repos
vendor a copy rather than pointing at this one.

| Skill              | For                                                                 |
| ------------------ | ------------------------------------------------------------------- |
| `tdd`              | The red-green loop, and what makes a test worth keeping             |
| `code-review`      | Two-axis review of a diff: repo standards, and the originating spec |
| `implement`        | Work a spec or ticket through `tdd` and `code-review`               |
| `domain-modeling`  | Build and sharpen `CONTEXT.md` and the ADR record                   |
| `codebase-design`  | Deep-module vocabulary: interface, depth, seam, adapter             |
| `diagnosing-bugs`  | The feedback-loop-first diagnosis discipline                        |

## Provenance

These began as copies from [`mattpocock/skills`](https://github.com/mattpocock/skills)
(MIT) and are modified from there. See [NOTICE](NOTICE). It is a fork, not a
dependency: there is no upstream sync, and no drift to detect against upstream.
An improvement made there is adopted by reading it and deciding.

## Consuming this repo

A consumer keeps a committed copy of the skills it uses, so that a fresh clone,
a CI runner or an unattended agent on a machine that has never seen this
repository still has them. `cmd/vendor-skills` maintains that copy against a
commit pinned in the consumer's `skills-lock.json`, and is run straight from the
module, so a consumer carries no copy of the tool itself:

```bash
go run github.com/corygyarmathy/skills/cmd/vendor-skills@latest verify
go run github.com/corygyarmathy/skills/cmd/vendor-skills@latest pull
go run github.com/corygyarmathy/skills/cmd/vendor-skills@latest pull --rev master
go run github.com/corygyarmathy/skills/cmd/vendor-skills@latest pull --add diagnosing-bugs
```

`verify` exits non-zero on drift, so it is usable as a CI check without parsing
its output.

The tool is deliberately not pinned by the lock file. What has to be
reproducible is the vendored content, and that is pinned by commit and checked
by hash - so a tool that wrote the wrong thing is caught by the next `verify`
rather than trusted.

## Editing a skill

Edit it here, commit, then `pull` in each consumer. A consumer that edits its
vendored copy in place will be told about it by `verify` on the next run - that
is what the per-skill hash in the lock file is for. Divergence is allowed; it
just isn't allowed to be silent.
