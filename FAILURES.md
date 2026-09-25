# Failures

Every time a skill leads an agent somewhere bad, log it here: what happened,
which skill, and a guess at why. Thirty seconds per entry is enough.

This is the evidence the skills get rewritten from. A line added to a skill
should trace back to an entry here, or to a deliberate preference; a skill due
for an `upstream` to `owned` rewrite starts by reading its entries. Later, the
entries become the cases an evaluation is built from.

Newest first. Strike an entry through, and name the commit, once a change to a
skill addresses it.

<!--
## YYYY-MM-DD `skill-name`

- **What happened:** what the agent did, and what it should have done.
- **Why (guess):** the line in the skill, or the missing one, that let it.
-->
