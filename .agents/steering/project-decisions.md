# kozmo-bf – Project Decisions

Append-only. New entries on top. Project-level decisions only; feature detail
stays in capsules.

## 2026-08-17 - Batteries-included binary: embedded cores/skill/docs with local-override precedence

Owner-approved Batch 9. Supersedes `project-architecture.md`'s Boundaries
entry "tools/ may not embed playbook content; playbooks are always
inputs": evolved, not dropped, by this decision. `project-architecture.md`
carries the full current rule; this entry is the why and the reversal
condition.

Rationale, three things the old rule made needlessly hard:

1. **Closure UX for consumers.** Composing `core-operations` from a fresh
   playbook required cloning this whole repository first, just to get
   three directories of YAML onto disk before `kbf lint` could resolve
   the name. Real friction for exactly the audience v0 exists to serve:
   an authoring agent or a human standing up a new business's ontology.
2. **One-artifact distribution.** `go install
   github.com/tadanahq/kbf/tools/cmd/kbf@latest` should be the whole
   install. It wasn't: the binary alone couldn't do anything useful
   without a matching clone sitting next to it.
3. **The on-prem story.** Kozmo's own on-prem posture (per-client
   install, no shared hosted state) wants a single binary fully
   functional on a machine that has never seen this repository. A binary
   that needs a clone alongside it to be useful undercuts that for KBF
   specifically.

New in this decision:

- `tools/` embeds `playbooks/core-business`, `playbooks/core-operations`,
  `playbooks/core-services`, `skills/kbf-authoring/`, and `spec/*.md` (+
  `spec/primitives/*.md`), synced from the repo root by
  `scripts/embedsync` (`make embed-sync`), verified by `make
  embed-freshness` (wired into `make check` and CI): the same
  copy-and-check shape `kbf schema --check` already established for
  `schema/`.
- Precedence is absolute and one-directional, enforced in code, not by
  convention: a locally-passed path always wins on a name collision;
  the embedded copy is consulted only for a `builds-on` name no local
  path provided at all (`lint.LoadWithEmbedded`'s fallback only runs for
  a name local resolution already failed on).
- `kbf vendor` exists specifically so "embedded" never means "hidden":
  anyone can materialize the exact embedded copy to a real, inspectable,
  editable, replaceable directory on demand.

Reversal: if embedded content and this repository's own source ever ship
out of sync despite `make embed-freshness` (a bug in the freshness check
itself, or a release process that skips `make check`), that is the signal
to revisit, either strengthen the gate (a build-time assertion, not just
CI) or drop back to `kbf vendor`-only (materialize on demand, never
silently resolve). Not foreseeable otherwise: the mechanism is a strict
subset of what `kbf schema --check` already does successfully for
`schema/`.

## 2026-08-17 - Composition (DAG closure) replaces single-parent extends; core-business rename; origin/approval field renames

Owner-approved Batch 7, dispatched mid-Batch-6-report. Supersedes this
file's own entry directly below ("Playbook rename executed; the
core/vertical taxonomy is machine-checked") on every point where the two
disagree: that entry's `layer: root | base | vertical` is now `layer:
core | vertical` (root-ness derived from an empty `builds-on`, never a
declared value); its `extends` (single nullable parent) is now
`builds-on` (a list, resolved as a DAG closure, not a chain); its
`KBF013` cross-check table (root ⇒ null; base/vertical ⇒ must extend a
root-or-base playbook) is replaced by a two-value build-target table
(core ⇒ core only; vertical ⇒ core or vertical, at least one). What that
entry got right and this one keeps: the `^core-` naming rule, and KBF013
as a machine gate rather than a documented-only convention.

New in this decision, not previously logged:

1. Composition is a DAG, not a chain: a playbook can build on more than
   one other playbook at once (the diamond case; `examples/bistro-demo`
   composes both `core-operations` and `core-services` directly),
   resolved transitively and deduped by name.
2. Cross-playbook identity collision (two different closure members
   declaring the same entity/relation/metric/action, neither one forking
   the other) is a new `KBF003` variant: composition has no resolution
   order, so this is always an error, reported symmetrically, never a
   silent first-loaded-wins.
3. `core-universal` renames to `core-business` ("Playbook Zero" nickname
   unchanged).
4. `Relation.tier` renames to `Relation.origin` (values unchanged);
   `Action.risk` renames to `Action.approval` (`auto|confirm` becomes
   `automatic|required`), since `approval` names what the field actually
   gates (whether a human confirms before it runs), which `risk` only
   implied.

Rationale: real content demanded the composition change immediately.
`examples/bistro-demo` (a business that operates from a site and sells
scoped work, needing both `core-operations` and `core-services` at once)
has no expression under single-parent `extends` at all. The taxonomy
simplification and the two field renames rode the same batch because the
schema change was already breaking; a second breaking pass later just to
fix `core-universal`'s and `risk`'s naming would have cost more than
doing it once, in the same pull request, while every manifest was already
being touched.

Reversal: none foreseeable for the composition mechanics, a DAG is a
strict superset of what a chain could express, so no real content shape
this forecloses. The field renames could in principle reverse if
`origin`/`approval` prove confusing in practice, but that would need real
authoring friction to show up, not a preference, to reopen.

## 2026-08-17 - Playbook rename executed; the core/vertical taxonomy is machine-checked

Owner decision, executing the 2026-08-13 "unit is a playbook" entry now that
the layered-cores restructure is in: `packages/` → `playbooks/`,
`universal-core`/`operations-core`/`services-core` →
`core-universal`/`core-operations`/`core-services`. Category vocabulary is
**"core playbooks"** (root and base layer, name matches `^core-`) and
**"vertical playbooks"** (bare industry nouns, name must not); this
supersedes the 2026-08-13 entry's "base playbooks" term, which is now stale
prose, not the live vocabulary.

New, not previously logged: the taxonomy is a manifest field, not just a
naming habit. `manifest.yaml` gains `layer: root | base | vertical`;
`KBF013` cross-checks it against `extends` (root ⇒ null; base/vertical ⇒
must extend a root-or-base playbook) and against `name` (root/base ⇒
`^core-`; vertical ⇒ not). Rationale: a convention nobody enforces drifts;
this repo's own thesis (a broken ontology fails fast, machine-checked, not
documentation) applies to its own package taxonomy the same way it applies
to business content. Reversal: none foreseeable; if a future layer needs a
different extends-target set than {root, base}, that is an addition to
`KBF013`'s rule table, not a reason to drop the check.

## 2026-08-13 - The unit is a playbook; the folder is playbooks/

Owner decision: the shipped unit's name is **playbook** (a runnable package of
business capability), everywhere: the repo folder becomes `playbooks/` (was
`packages/`), the spec doc becomes `spec/playbook-format.md`, prose commits to
one term (bases are "base playbooks"; universal-core stays "Playbook Zero";
"package" survives only as plain English inside the definition). Rationale:
the folder name should teach the product's central concept; "install the
restaurant playbook" is the sentence the architecture exists to make true.
`examples/` keeps its name (demos teach, they don't install). Applies as a
mechanical rename batch after the layered-cores restructure lands.

Owner correction: the first universal-core carried an operating-business bias
(location, shift, order, product labeled "universal"), which contradicts the
framework's core claim (one shared multi-vertical core). New architecture:
`universal-core` (minimal, truly universal: organization, customer, offering,
transaction, employee, supplier, purchase) ← `operations-core` /
`services-core` (base packages) ← verticals. Linter resolves `extends` chains
recursively; KBF007's verb vocabulary is the union along the chain; KBF008
checks forks against all ancestors. Two demos, one per chain (cafe-demo,
studio-demo), keep the claim honest. The extends-root stays a parameter: the
toolchain never privileges our packages. Reversal: none foreseeable; a bias
reappearing in universal-core is a bug, not a preference.

## 2026-08-13 - v0 is config-phase tooling only

The CLI serves an agent creating and validating an ontology: `lint`, static
`coverage`, `compile --to mermaid`. Runtime enforcement (query gate: validating
agent-generated SQL against contract semantics with a repair pass) is
deliberately excluded from v0 and parked on the roadmap as pending discussion.
Rationale: ship the spec and authoring experience first; the gate needs real
install experience to be designed well. Reversal: first production install
demonstrating a concrete enforcement need.

## 2026-08-13 - Go stack, single binary, zero runtime deps

Go (current stable), goccy/go-yaml, cobra + lipgloss, generated JSON Schema
(invopop), conformance validated with santhosh-tekuri/jsonschema, testscript
golden tests, goreleaser. Chosen over Python not for speed but for: single-binary
distribution into arbitrary adopter CI/environments, the on-prem install story,
and a language-neutral subprocess contract for future runtime tooling.
Conformance fixtures stay language-agnostic so other implementations can port.

## 2026-08-13 - Canonical structs, generated schema

Go structs in `tools/internal/model/` are the meta-model. `schema/*.yaml` is a
committed build artifact; hand-editing it is a violation; CI regenerates and
fails on drift. Conformance validates against the published schema files with a
vanilla validator to keep the public interface honest.

## 2026-08-13 - Naming and vocabulary

The spec is the **KBF Ontology Spec**. Element vocabulary: **semantic elements**
(entities, relations, metrics) and **action elements**. "Kinetic" is not used.
Authoring format is YAML; meta-schema is JSON Schema semantics (published as
YAML files). Docs carry OKF-conformant frontmatter for portability.

## 2026-08-13 - Public repo, absolute hygiene

This repo is public from its first commit. No client names, internal project
references, private paths, or prices anywhere, fixtures and comments included.
Machine-enforced by `make boundaries` with a named blocklist.
