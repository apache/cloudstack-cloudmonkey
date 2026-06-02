<!--
SPDX-License-Identifier: Apache-2.0

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
-->

# Security Policy

## Reporting a Vulnerability

`apache/cloudstack-cloudmonkey` follows the [Apache Software Foundation security process](https://www.apache.org/security/).
Please report suspected vulnerabilities privately to `security@apache.org`; do not
open public GitHub issues or pull requests for security reports.

## Threat Model

`apache/cloudstack-cloudmonkey` is part of the Apache CloudStack project and is covered by the
**project-wide CloudStack threat model** rather than a per-repository copy. What the
project treats as in scope and out of scope, the security properties it provides and
disclaims, the adversary model, and how findings are triaged are documented in that
model: <https://github.com/apache/cloudstack/blob/main/THREAT_MODEL.md>.

(That link resolves once the project-wide model lands on `apache/cloudstack`'s
`main` branch — see apache/cloudstack#13293. A thin `cloudstack-cloudstack-cloudmonkey`-specific
addendum can be added here later if this component needs one.)
