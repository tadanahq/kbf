# kozmo-bf – Project Decisions

Append-only. New entries on top. Project-level decisions only; feature detail
stays in capsules.

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
