---
status: accepted
---

# A rule can carry a deterministic pattern; judgement about a UI stays with the client

A rule may have a `Forbid:` line, a Go RE2 regular expression, and a `Fix:` line. The `critique` tool reports each match in a UI source file as a finding, naming the rule, the line, the matched text and the fix. Rules without a `Forbid:` line stay valid and are never checked by `critique`. The server has no model access, so it can only report what a pattern expresses. Judgement that needs the point of view or the user journey stays with the client, which reads the rules through the delivery tools and reasons over them. The trade-off is that a pattern is an approximation of its rule (the bundled spacing pattern flags some values that the rule allows), and the rule author must write and maintain the patterns. This keeps findings reproducible and testable.
