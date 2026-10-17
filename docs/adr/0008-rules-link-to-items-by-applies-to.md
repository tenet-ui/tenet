# A rule names the items and token categories it applies to, in its own Applies to line

Delivery tools return an item together with its linked rules, so the system needs a way to say which rule belongs to which component, template, flow or token category. The link lives in `rules.md`, in an optional `Applies to:` line of the rule (for example `Applies to: tokens:color, component:button`), not in the item files. The rule owns its scope, and the item files keep the format from slice 2. Rules that name no target are valid but are never delivered by an item or token tool.

Links are checked when the pack loads: a reference to a missing item or token category is a problem reported at the `Applies to:` line, and the rule itself stays valid. A rule whose entry has any parse problem is not valid and is not delivered, so its references have no effect. The tools read the pack fresh on every call, so edits show without restarting the server.

The alternative was a `Rules:` line in each item file. That was rejected because an item would then have to know its rules, and a rule's scope would be split across files that critique (slice 4) has to read together.
