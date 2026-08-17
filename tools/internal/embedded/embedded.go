// Copyright 2026 The kbf Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package embedded is what makes kbf batteries-included: the public core
// playbooks, the kbf-authoring skill, and the prose spec, all baked into
// the binary so a consumer never has to clone this repository just to
// compose against core-operations or read spec/onboarding.md
// (project-decisions.md: "Batteries-included binary"). go:embed cannot
// reach outside this module, so data/ is a committed mirror of the real
// files at the repo root (playbooks/core-*, skills/kbf-authoring/,
// spec/*.md, spec/primitives/*.md), produced by scripts/embedsync
// (`make embed-sync`) and kept honest by `make embed-freshness`, the
// same copy-and-check shape `kbf schema --check` already uses for
// schema/. Every file in data/ carries a DO-NOT-EDIT header pointing back
// at its real source; this package only ever reads data/, never writes
// it.
//
// project-architecture.md's Boundaries section used to read "tools/ may
// not embed playbook content; playbooks are always inputs". That rule is
// evolved, not dropped, by the same decision this package exists to
// implement: tools/ may embed the *public* core playbooks specifically,
// strictly as a resolution fallback (internal/lint's PlaybookSource,
// consulted only for a name no local path already provided), never a
// privileged or hidden source of truth. See project-architecture.md's
// current Boundaries section for the full rule.
package embedded

import "embed"

// FS is every embedded file, rooted at data/. Package-level accessors
// (Playbook, Skill, Doc, and their *Names siblings) are the intended
// entry points; FS is exported only for the rare caller that needs to
// walk everything at once.
//
//go:embed all:data
var FS embed.FS
