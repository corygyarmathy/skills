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
The send-back is yours: nothing should draft it for you, an interactive session
included.

Two send-backs are refused, each with a reply saying so, and nothing else is done:

- one written while a revision is in flight, after it started and before its reply or hand-back. Wait for the reply or the hand-back, then send the points again;
- a review, or a line comment, begun on a head the PR has since left: started before a revision pushed, submitted after. The reply names the commit it was written on. Start the review again on the current head.

**Close.** One line saying why.

## The second sitting

When the revision comes back, read the delta since your last review (the compare
link in its reply) and check that each point you sent back is addressed. Re-read
the whole PR only if the delta changed the design. Start that review after the
revision's reply: one begun before its push is refused (see **Send-back**).
