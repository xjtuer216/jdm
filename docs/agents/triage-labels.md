# Triage Labels

The skills speak in terms of five canonical triage roles. This file maps those
roles to the actual label strings used in this repo's issue tracker (Gitee
Issues in `xjtuer216/jdm`).

| Label in mattpocock/skills | Label in our tracker | Meaning                                  |
| -------------------------- | -------------------- | ---------------------------------------- |
| `needs-triage`             | `needs-triage`       | Maintainer needs to evaluate this issue  |
| `needs-info`               | `needs-info`         | Waiting on reporter for more information |
| `ready-for-agent`          | `ready-for-agent`    | Fully specified, ready for an AFK agent  |
| `ready-for-human`          | `ready-for-human`    | Requires human implementation            |
| `wontfix`                  | `wontfix`            | Will not be actioned                     |

All five exist on the Gitee repo already.

## Category roles

`/triage` also applies exactly one *category* role per issue, on top of the
state role. This repo's Gitee labels already carry the canonical names:

| Category role | Gitee label   |
| ------------- | ------------- |
| `bug`         | `bug`         |
| `enhancement` | `enhancement` |

Other pre-existing Gitee labels — `feature`, `question`, `duplicate`,
`invalid` — are not part of the triage vocabulary. Use `bug` or
`enhancement` as the category; use `feature` only as extra context.

When a skill mentions a role (e.g. "apply the AFK-ready triage label"), use
the corresponding label string from these tables.
