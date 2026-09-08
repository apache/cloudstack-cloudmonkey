<!--
Licensed to the Apache Software Foundation (ASF) under one
or more contributor license agreements.  See the NOTICE file
distributed with this work for additional information
regarding copyright ownership.  The ASF licenses this file
to you under the Apache License, Version 2.0 (the
"License"); you may not use this file except in compliance
with the License.  You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing,
software distributed under the License is distributed on an
"AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
KIND, either express or implied.  See the License for the
specific language governing permissions and limitations
under the License.
-->

# Local patches to vendored dependencies

Some files under `vendor/` carry local changes that are **not** present in the
upstream release. Running `go mod vendor` reverts them silently — the build
still succeeds, but the resulting binary loses the fixes. Every local change
must therefore be recorded here.

After any `go mod vendor`, re-apply the patches:

    make vendor-patch

which is equivalent to:

    go mod vendor
    git apply patches/*.patch

## `ergochat-readline-v0.1.3.patch`

Applies to `github.com/ergochat/readline` v0.1.3. Three changes:

### 1. `complete.go` — insert the value, display the detail

CloudMonkey builds completion candidates as `<value> (<detail>)` so the user
can see what a UUID refers to (see `cli/completer.go`). Upstream's
`AutoCompleter.Do` returns a single slice that is used both for display *and*
for insertion, so without this patch the ` (<detail>)` suffix is written into
the command line.

`writeRunes` strips everything from the ` … (` boundary before inserting, and
`truncateBufferAfterLastEqual` drops the partial text after the last `=` when
the candidate matched on its detail rather than on its value.

Originally added for #133 and #196 against `chzyer/readline`; forward-ported
when the library was switched to `ergochat/readline`.

### 2. `complete.go` — redraw the prompt in `CompleteRefresh`

The `\033[J` erase issued while rendering the candidate grid can clear the
prompt line, leaving the prompt blank until the next keystroke. The patch
redraws the prompt and buffer contents after the erase.

This is an upstream rendering bug, not a CloudMonkey-specific need.

### 3. `operation.go` — bound the cursor position query

`Runes()` calls `getAndSetOffset(nil)` before printing every prompt, which
sends a DSR cursor position request (`ESC[6n`) and blocks on the reply with a
nil deadline — i.e. forever. Under a pty whose emulator never answers
(`expect`, `pexpect`, Ansible with a pty, some CI runners) `cmk` prints its
banner and then hangs with no prompt.

The patch passes a deadline of `dsrTimeout` (250ms, the constant upstream
already uses in `waitForDSR`). On expiry `GetCursorPosition` returns an error,
`getAndSetOffset` leaves the offset unchanged and the prompt is printed anyway.
A late CPR reply is still recognised and discarded by `consumeANSIEscape`, so
it is never mistaken for user input.

The timeout would otherwise be paid before *every* prompt, so the result is
latched: on `deadlineExceeded` the `dsrUnsupported` flag is set and no further
queries are made for the life of the session. A terminal that answers — even
one that takes 150ms to do so — never sets the flag and keeps full offset
tracking on every prompt. The tradeoff is that a terminal which misses one
query and then recovers stays latched off for the session; the consequence is
cosmetic (the prompt may overwrite pre-existing text on the line) and real
emulators answer CPR synchronously, so a one-off miss is not expected.

## Upstream status

None of these are fixed in `ergochat/readline` v0.1.3, which has been the
latest release since September 2024. Changes 2 and 3 are general bugs and
should be reported upstream so this patch can shrink; change 1 needs an
upstream API that separates a candidate's display text from its insert text.
