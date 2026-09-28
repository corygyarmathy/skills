---
name: reviewing-changes
description: 'Review the changes since a fixed point (commit, branch, tag, or merge-base) on four axes: Standards, Spec, Correctness, and Approach, each in its own sub-agent, reported side by side. Use when the user wants a branch, a PR, or work in progress reviewed, or asks to "review since X".'
---

Four-axis review of the diff between `HEAD` and a fixed point:

- **Standards**: does the code conform to this repo's documented coding standards?
- **Spec**: does the code faithfully implement the originating issue / spec?
- **Correctness**: where does the code produce wrong output, crash, or break a caller?
- **Approach**: is this the right way to do it, and what does it do to the rest of the codebase?

Each axis runs as a **parallel sub-agent** with a fresh context, then this skill aggregates their findings. Run this skill in a session that did not write the change: the reviewer that shares the author's reasoning accepts the author's justifications.

The review takes a **severity floor** as input: `should-fix` by default, and an interactive caller may ask for `consider`. Nothing below the floor is produced, so every finding in the report is worth reading.

This repo's issue tracker is described in `docs/agents/issue-tracker.md`. If that file is missing and a spec has to be fetched from a tracker, say so rather than guessing at a `gh` invocation - a review that cites the wrong tracker is worse than one that admits it has no spec.

## Severities

Each finding carries one severity, and the floor cuts between them in the same place every time:

- **blocker**: produces wrong output, loses data, breaks a caller, or fails the spec. Merging it as it stands would be a mistake.
- **should-fix**: works today, but a concrete cost follows if it merges: a missed spec case, an untested failure path, a documented standard broken, a design that makes the next change harder. The finding names that cost; a finding that can't name one is `consider`.
- **consider**: a matter of taste or a possible improvement. Merging without it costs nothing concrete.

## Process

### 1. Pin the fixed point

Use the fixed point the caller gave (a commit SHA, branch name, tag, `main`, `HEAD~5`, etc.). If none was given, use the merge-base with the default branch, and state that assumption in the report.

Capture the diff command once: `git diff <fixed-point>...HEAD` (three-dot, so the comparison is against the merge-base). This covers committed work only; if the working tree is dirty, say so in the report. Also note the list of commits via `git log <fixed-point>..HEAD --oneline`.

Before going further, confirm the fixed point resolves (`git rev-parse <fixed-point>`) and the diff is non-empty. A bad ref or empty diff should fail here, not inside four parallel sub-agents.

### 2. Identify the spec source

Look for the originating spec, in this order:

1. A path or issue the caller passed.
2. Issue references in the commit messages (`#123`, `Closes #45`, GitLab `!67`, etc.), fetched via the workflow in `docs/agents/issue-tracker.md`.
3. A spec file under `docs/`, `specs/`, or `.scratch/` matching the branch name or feature.

If nothing is found and a user is in the session, ask where the spec is. Otherwise, or if they say there isn't one, the **Spec** sub-agent is skipped and the report says "no spec available".

### 3. Identify the standards sources

Anything in the repo that documents how code should be written, such as `CODING_STANDARDS.md`, `CONTRIBUTING.md`, or `AGENTS.md`.

At a `consider` floor, the Standards axis also carries the **smell baseline** below: a fixed set of Fowler code smells (_Refactoring_, ch.3) that applies even when a repo documents nothing. Its smells are only ever `consider`, so at a `should-fix` floor it stays out of the brief. Two rules bind it:

- **The repo overrides.** A documented repo standard always wins; where it endorses something the baseline would flag, suppress the smell.
- **Always a judgement call.** Each smell is a labelled heuristic ("possible Feature Envy"), never a hard violation. Like any standard here, skip anything tooling already enforces.

Each smell reads _what it is_ → _how to fix_; match it against the diff:

- **Mysterious Name**: a function, variable, or type whose name doesn't reveal what it does or holds. → rename it; if no honest name comes, the design's murky.
- **Duplicated Code**: the same logic shape appears in more than one hunk or file in the change. → extract the shared shape, call it from both.
- **Feature Envy**: a method that reaches into another object's data more than its own. → move the method onto the data it envies.
- **Data Clumps**: the same few fields or params keep travelling together (a type wanting to be born). → bundle them into one type, pass that.
- **Primitive Obsession**: a primitive or string standing in for a domain concept that deserves its own type. → give the concept its own small type.
- **Repeated Switches**: the same `switch`/`if`-cascade on the same type recurs across the change. → replace with polymorphism, or one map both sites share.
- **Shotgun Surgery**: one logical change forces scattered edits across many files in the diff. → gather what changes together into one module.
- **Divergent Change**: one file or module is edited for several unrelated reasons. → split so each module changes for one reason.
- **Speculative Generality**: abstraction, parameters, or hooks added for needs the spec doesn't have. → delete it; inline back until a real need shows.
- **Message Chains**: long `a.b().c().d()` navigation the caller shouldn't depend on. → hide the walk behind one method on the first object.
- **Middle Man**: a class or function that mostly just delegates onward. → cut it, call the real target direct.
- **Refused Bequest**: a subclass or implementer that ignores or overrides most of what it inherits. → drop the inheritance, use composition.

### 4. Size the review

Count the changed lines, leaving out tests, lockfiles and generated code. Under about 50, skip the Approach sub-agent and give its **blast radius** brief to the Correctness sub-agent instead; the report says Approach was folded in.

### 5. Spawn the sub-agents in parallel

Every sub-agent prompt carries the same inputs and nothing else:

- The diff command and commit list.
- The spec, pasted verbatim or as a path to it, never as your summary of it.
- The severity floor, with the severity definitions above.
- The inputs specific to its axis, below.

Leave out any explanation of why the change was made the way it was. Each reviewer judges the code cold; the reasons it needs are in the spec and the code. Tell each one that commit messages are the author's claims, to be checked against the code rather than taken as reasons.

Every brief ends with the same output rules: "The floor is `<floor>`: report nothing below it. Rank findings most severe first. Each finding is at most about 3 lines: its severity, `file:line`, what's wrong, and a one-line fix, with a short code suggestion when one fits. A should-fix names the cost of merging it. Read beyond the diff wherever judging a hunk needs it. Under 400 words; if the cap bites, drop the least severe findings."

Standards is checklist matching: run it on a smaller, cheaper model where the harness lets you choose. The other axes need the strongest model available.

**Standards** also gets the standards-source files from step 3. Brief: "Report every place the diff violates a documented standard: cite the standard (file + the rule). Skip anything tooling enforces."

At a `consider` floor, paste the smell baseline from step 3 in full as well (the sub-agent has no other access to it), and add to the brief: "Also report any baseline smell you spot: name it and quote the hunk. Baseline smells are always **consider**, and a documented repo standard overrides the baseline."

**Spec** brief: "Report (a) requirements the spec asked for that are missing or partial; (b) behaviour in the diff that wasn't asked for (scope creep); (c) requirements that look implemented but where the implementation looks wrong. Quote the spec line for each finding." Skipped when there is no spec.

**Correctness** brief: "Find the ways this change produces wrong output, crashes, loses data, or breaks an existing caller: edge cases, error paths, concurrency, resource cleanup, and the invariants the surrounding code relies on. Every finding states a **failure scenario**: the inputs and state, and the wrong result they produce. A suspicion you cannot turn into a failure scenario is listed separately as a question."

**Approach** brief: "Judge the approach the change takes, not its details. Work in this order:

1. **Design it twice.** Before reading the diff, read the spec and the code it touches, and sketch in a few lines how you would have done it. Then read the diff and compare. Report where your sketch is simpler, or where the diff's approach is better and why.
2. **Blast radius.** List every interface, type, schema, config key, or behaviour the diff changes, find their callers and dependents across the repo, and report any the change affects that it didn't update. Report existing code the change duplicates instead of reusing.
3. **Pre-mortem.** Assume that six months from now this change caused a problem. Name the most likely one: migration, reversibility, performance at scale, operability, security, or a future change it makes harder.
4. **Spec pushback.** Report where the code shows the spec itself was wrong, unnecessary, or missing a case. With no spec, ask whether the change should exist at all, and whether a smaller change solves the same problem."

### 6. Aggregate

Present the reports under `## Standards`, `## Spec`, `## Correctness` and `## Approach` headings, lightly cleaned. Keep the axes separate: do not merge or rerank findings across them (see _Why separate axes_). An axis with no findings collapses to one line under its heading.

Number the findings `1…n` continuously across the whole report, so any one can be cited alone ("advisory 3"). Each keeps the about-3-line shape from the output rules; a Correctness finding's "what's wrong" is its failure scenario. After the findings, list the questions (suspicions with no failure scenario) as `Q1…Qn`.

End with a one-line summary: total findings per axis, and the worst issue _within each axis_ (if any). Don't pick a single winner across axes: that's the reranking the separation exists to prevent. A review with nothing at or above the floor still reports, and says so, naming the floor.

The review is advisory.

## Why separate axes

A change can pass one axis and fail another:

- Code that follows every standard but implements the wrong thing → **Standards pass, Spec fail.**
- Code that does exactly what the issue asked but breaks the project's conventions → **Spec pass, Standards fail.**
- Code that meets the spec and the standards but fails on an empty input → **Correctness fail.**
- Code that is correct and on-spec but rebuilds a helper the repo already has, or bolts a special case onto a module that should have been reshaped → **Approach fail.**

Reporting them separately stops one axis from masking another.
