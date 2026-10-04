# The operator's review

How to read a pull request the agent opened, and how to end that reading. There is
one procedure for every PR. It is orientation, not a checklist: the same way in,
whatever the change.

The **operator's review** is the only review that decides anything. The
**advisory review** the agent posts is advice.

## Before the code

Read the ticket, then the PR description, then the diff. The first decision comes
before any code: should this change exist, and is it the ticket's concern?

Read the diff core first: the change the description names under **Start here**,
then its tests, then the rest. Never GitHub's alphabetical order: the files read
last get the least scrutiny.

## One reading, three questions

- **Design.** Does it fit? Is it the simplest thing for _this_ ticket, and is it one concern?
- **Behaviour.** Does it do what the ticket asked? What are the edge cases and failure paths? Would the tests fail if the code were wrong?
- **Understanding.** Can you explain every line that isn't generated? A line you can't explain is a finding against the PR.

Style, formatting and vet are left to CI.

Running the change is recommended when the diff can't show the behaviour:
concurrency, CLI output, a NixOS option.

A PR that touches a **sensitive path** names it under **Sensitive:** in its
description. The matched files are read line by line, even on a recipe PR, where
the rest may be sampled. Running the change is recommended rather than optional.
Reading order stays core first.

A **recipe PR** is deliberately large and one concern. Check the recipe, the
exceptions and a sample.

## When to stop

A design objection, or "this shouldn't exist", stops the reading early and sends
the PR back. Otherwise stop once every line has been read, or on a recipe PR, the
recipe, the exceptions and the sample. Then decide against one bar: does it
definitely improve code health, even if it isn't perfect? There is no timer.

## Your own reading first

Open the advisory review only after you have decided. Afterwards, looking at it
for what you missed or what it got wrong is encouraged; nothing is required and
nothing is recorded. No finding is owed an answer.

## How it ends

Exactly one of three. Only written words count: a reaction, an approval or a
resolved thread decides nothing.

**Merge.** Nothing is written, and no "Approve" is needed. For points that can
wait, merge and open a follow-up issue in your own words that names the PR it came
from ("Reference in new issue" on a PR comment adds the link; replace the quoted
text with your own).

**Send-back.** The points in your own words, in either of two forms:

- a PR comment whose first line is `/revise`, followed by the points;
- a submitted review whose body starts `/revise`. Its line comments are part of the send-back, and its state (comment, approve, request changes) decides nothing.

An advisory finding enters only when you cite it, one at a time: "advisory 3:
agreed, and cover the empty case too". A citation means the latest advisory review.
After a revision, that is the review of the revision's delta, numbered from 1
again: "advisory 3" is finding 3 of the delta's review. A finding from an earlier
review, on code the revision didn't touch, is not in the latest review: quote it
in your own words rather than cite it by number, which would name a different
finding.
The send-back is yours: nothing should draft it for you, an interactive session
included.

Three send-backs you may meet are refused, each with a reply saying so, and
nothing else is done:

- one written while a revision is in flight, after it started and before its reply or hand-back. Wait for the reply or the hand-back, then send the points again. A revision that parks posts neither, so nothing is refused: the next send-back starts a new revision from wherever the branch was left, and carries only its own points;
- a review written on a commit that is no longer the PR's head, or with a line comment written on one: begun before a push to the branch, the revision's or anyone's, and submitted after. The whole review is refused, and the reply names that commit. Start the review again on the current head;
- one with no points, such as a bare `/revise` with the points in a comment of their own. Write the points after `/revise` in the same comment, or in the same review.

**Close.** One line saying why.

## The second sitting

When the revision comes back, read the delta since your last review (the compare
link in its reply) and check that each point you sent back is addressed. Re-read
the whole PR only if the delta changed the design. Start that review after the
revision's reply or hand-back: one begun before its push is refused (see
**Send-back**).

The revision has its own advisory review, posted after its reply. Its summary
names a range ("Advisory review of `abc1234..def5678`") rather than one head, and
it covers the same delta you are reading. **Your own reading first** applies to
it unchanged: open it after you have decided.
