# Issue tracker: Gitee Issues

Issues and specs for this repo live as Gitee issues in `xjtuer216/jdm`
(https://gitee.com/xjtuer216/jdm/issues).

There is no `gh`/`glab` CLI here and no Gitee CLI — every operation goes
through the Gitee REST API v5 with `curl`.

## Auth

All write calls need `access_token=<personal access token>` as a query
parameter. The token lives in Windows Credential Manager; fetch it with:

```bash
CRED=$(printf "protocol=https\nhost=gitee.com\n\n" | git credential fill | sed -n 's/^password=//p')
```

Read calls (GET) on this public repo work without a token.

## Gitee API quirks (learned the hard way — read before scripting)

- **Issue numbers are alphanumeric strings**, e.g. `IKE2K2`, `IKRKAU` — not
  integers. Always quote them in URLs.
- **Create and update take the repo as a form field, not a path segment.**
  `POST /repos/{owner}/{repo}/issues` fails with
  `{"message":"project or enterprise"}`; use `POST /repos/{owner}/issues`
  with `repo=jdm` in the body.
- **`state` accepts only `open` and `closed`.** `close`, `reopen`, and
  `opened` are rejected with `state does not have a valid value`.
- **Label colors take no `#` prefix** (`color=d93f0b`, not `color=#d93f0b`).
- **Issues cannot be deleted.** `DELETE /issues/{number}` returns
  `405 Not Allowed`. Closing is the only way to retire one.
- **No native blocking or sub-issue API.** `parent_url` stays `null` on
  create, and `/related_issue` is 404. Use the body-text conventions below.
- **`PATCH` on an issue replaces its labels rather than adding to them**, and
  the PATCH response body comes back with an empty `labels` array. Read the
  issue back with `GET` when you need to confirm the label set.

## Conventions

Base URL: `https://gitee.com/api/v5/repos/xjtuer216/jdm`

- **Create an issue**:
  ```bash
  curl -s -X POST "https://gitee.com/api/v5/repos/xjtuer216/issues?access_token=$CRED" \
    --form "repo=jdm" --form "title=..." --form "body=..." --form "labels=needs-triage"
  ```
  `labels` is a comma-separated list of label names.
- **Read an issue**:
  `curl -s "$BASE/issues/$NUMBER"` → `number`, `title`, `body`, `state`,
  `labels[].name`, `comments` (count), `html_url`.
- **List issues**:
  `curl -s "$BASE/issues?state=open&labels=needs-triage&per_page=100&page=1"`.
  `state` accepts `open`, `closed`, `all`. The response headers carry
  `total_count` / `total_page` / `open_issues` / `close_issues`.
- **Comment on an issue**:
  `curl -s -X POST "$BASE/issues/$NUMBER/comments?access_token=$CRED" --form "body=..."`
- **Delete a comment** (for cleaning up your own mistakes):
  `curl -s -X DELETE "$BASE/issues/comments/$COMMENT_ID?access_token=$CRED"`
- **Add labels**:
  `curl -s -X POST "$BASE/issues/$NUMBER/labels?access_token=$CRED" \
   -H "Content-Type: application/json" -d '["ready-for-agent"]'`
- **Remove one label**:
  `curl -s -X DELETE "$BASE/issues/$NUMBER/labels/$LABEL?access_token=$CRED"`
- **Remove all labels**:
  `curl -s -X DELETE "$BASE/issues/$NUMBER/labels?access_token=$CRED"`
- **Replace all labels**:
  `curl -s -X PUT "$BASE/issues/$NUMBER/labels?access_token=$CRED" \
   -H "Content-Type: application/json" -d '["bug","needs-triage"]'`
- **Update title / body / state / labels**:
  ```bash
  curl -s -X PATCH "https://gitee.com/api/v5/repos/xjtuer216/issues/$NUMBER?access_token=$CRED" \
    --form "repo=jdm" --form "state=closed"
  ```
  `labels` here is the complete replacement set, and `assignee` takes a
  username.
- **Close an issue**: post the explanation comment first, then `PATCH` with
  `state=closed`. Do not reverse the order — the closing comment is the
  record of why.
- **List repo labels**: `curl -s "$BASE/labels?per_page=100"`
- **Create a repo label**:
  `curl -s -X POST "$BASE/labels?access_token=$CRED" --form "name=..." --form "color=..."`

## Blocking

Gitee has no native issue dependencies. Fall back to a body-text line at
the top of the child issue:

```
Blocked by: IKE2K2
```

A ticket is unblocked when every issue it lists is `closed`.

## Sub-issues / parent-child

Gitee has no sub-issue API. Put the parent at the top of the child body:

```
Part of: IKE2K2
```

## Pull requests as a triage surface

**PRs as a request surface: no.** _(Set to `yes` if this repo treats external
PRs as feature requests; `/triage` reads this flag.)_

When set to `yes`, PRs run through the same labels and states as issues,
using the `/pulls` endpoints:

- **List PRs**: `curl -s "$BASE/pulls?state=open"`
- **Read a PR**: `curl -s "$BASE/pulls/$NUMBER"` and
  `curl -s "$BASE/pulls/$NUMBER/files"` for the diff.
- **Comment / label / close**: `curl -s -X POST "$BASE/pulls/$NUMBER/comments"`,
  `curl -s -X POST "$BASE/pulls/$NUMBER/labels"`,
  `curl -s -X PATCH "$BASE/pulls/$NUMBER" --form "state=closed"`.

Issue ids are alphanumeric and PR ids are numeric, so a bare `#42` is
ambiguous — always name the surface.

## When a skill says "publish to the issue tracker"

Create a Gitee issue (see **Create an issue** above).

## When a skill says "fetch the relevant ticket"

`curl -s "$BASE/issues/$NUMBER"` — the user will normally pass the issue
number directly.

## Wayfinding operations

Used by `/wayfinder`.

- **Map**: a single issue holding the Notes / Decisions-so-far / Fog body.
  Create it with the label `wayfinder:map`.
- **Child ticket**: an issue whose body starts with `Part of: <map-id>`.
  Labels: `wayfinder:<type>` (`research`/`prototype`/`grilling`/`task`).
- **Blocking**: `Blocked by: <id>, <id>` line near the top of the body.
- **Frontier query**: list the map's open children, drop any with an open
  blocker or an assignee; first in map order wins.
- **Claim**: `PATCH` the ticket with `assignee=xjtuer216` — the session's
  first write.
- **Resolve**: comment with the answer, `PATCH state=closed`, then append a
  context pointer to the map's Decisions-so-far.
