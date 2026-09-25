# skills

Agent skills, owned rather than installed.

Each directory under `skills/` is one skill: a `SKILL.md` with YAML frontmatter,
plus whatever reference files it needs beside it. Harnesses that support skills
(Claude Code, OpenCode) discover them per-project, which is why consuming repos
vendor a copy rather than pointing at this one.

Each skill has a status, so it is always clear which ones are still someone
else's thinking:

- **upstream**: as imported. Works, but not yet read critically.
- **adapted**: targeted edits on top of upstream; `git log -- skills/<name>`
  says what and why.
- **owned**: rewritten, and every line can be justified.

Moving a skill down that list is the long-term work here, one skill at a time,
driven by [FAILURES.md](FAILURES.md).

## The main flow

| Skill             | Status   | For                                                                 |
| ----------------- | -------- | ------------------------------------------------------------------- |
| `grill-with-docs` | upstream | Interview a plan to resolution, updating `CONTEXT.md` and ADRs      |
| `to-spec`         | adapted  | Synthesise the conversation into a spec on the issue tracker        |
| `to-tickets`      | adapted  | Break a spec into tracer-bullet tickets with blocking edges         |
| `implement`       | upstream | Work a spec or ticket through `tdd` and `code-review`               |
| `code-review`     | adapted  | Two-axis review of a diff: repo standards, and the originating spec |

## Engineering

| Skill                           | Status   | For                                                            |
| ------------------------------- | -------- | -------------------------------------------------------------- |
| `tdd`                           | upstream | The red-green loop, and what makes a test worth keeping        |
| `diagnosing-bugs`               | adapted  | The feedback-loop-first diagnosis discipline                   |
| `domain-modeling`               | adapted  | Build and sharpen `CONTEXT.md` and the ADR record              |
| `codebase-design`               | upstream | Deep-module vocabulary: interface, depth, seam, adapter        |
| `improve-codebase-architecture` | upstream | Find deepening opportunities and grill through one             |
| `prototype`                     | upstream | Throwaway code that answers one design question                |
| `research`                      | upstream | Answer a question from primary sources into a cited file       |
| `resolving-merge-conflicts`     | upstream | Resolve a conflict hunk by hunk, by intent                     |
| `triage`                        | upstream | Move issues through a state machine of triage roles            |
| `wayfinder`                     | upstream | Plan work too big for one session as a map of decision tickets |
| `wizard`                        | upstream | Generate a bash script that walks a human through manual steps |
| `setup-cory-gyarmathy-skills`   | upstream | Configure a repo's issue tracker, labels and doc layout        |
| `ask-cory`                      | upstream | Router: which skill or flow fits the situation                 |

## Productivity

| Skill                | Status   | For                                                          |
| -------------------- | -------- | ------------------------------------------------------------ |
| `grilling`           | upstream | The interview primitive the grill skills are built on        |
| `grill-me`           | upstream | Grilling with no repository to write into                    |
| `writing-for-agents` | upstream | Writing skills, CLAUDE.md, and any doc an agent reads        |
| `handoff`            | upstream | Compact a conversation into a document another agent resumes |
| `to-questionnaire`   | upstream | Turn a decision only someone else can make into questions    |
| `wait-what`          | upstream | Re-explain a message that did not land                       |
| `teach`              | upstream | Multi-session teaching in a stateful workspace               |

## Provenance

These began as copies from [`mattpocock/skills`](https://github.com/mattpocock/skills)
(MIT), taken at `c55ee46`, and are modified from there. See [NOTICE](NOTICE). It
is a fork, not a dependency: there is no upstream sync, and no drift to detect against upstream.
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
