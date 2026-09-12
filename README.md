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
repository still has them. `bin/vendor-skills` maintains that copy against a
pinned commit recorded in the consumer's `skills-lock.json`:

```bash
scripts/sync-skills.sh verify        # do the vendored copies match the lock?
scripts/sync-skills.sh pull          # re-vendor at the pinned rev
scripts/sync-skills.sh pull --rev master
scripts/sync-skills.sh pull --add diagnosing-bugs
```

Copy `bin/vendor-skills` into the consumer as `scripts/sync-skills.sh`. It
carries its own hash and will tell you when this repository's copy has moved
ahead of it; it will not overwrite itself behind your back.

## Editing a skill

Edit it here, commit, then `pull` in each consumer. A consumer that edits its
vendored copy in place will be told about it by `verify` on the next run - that
is what the per-skill hash in the lock file is for. Divergence is allowed; it
just isn't allowed to be silent.
