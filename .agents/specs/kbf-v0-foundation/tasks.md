---
type: spec-tasks
status: done
---
# kbf-v0-foundation – Tasks

Each batch uses the names in force when it ran; Batches 6, 7 and 10 rename
folders, fields and commands. `overview.md` and `design.md` state the current names.

## Batch 1: Scaffold + model + schema

A Go module, a generated meta-schema and one repository gate exist before any
content does.

- [x] Go module under `tools/` with a cobra `kbf` root and a `schema` command. Done means: `go run ./cmd/kbf --help` lists `schema`.
- [x] Canonical model structs for the seven primitives in `internal/model/`, every field with a why-comment. Done means: compiles, gofmt passes.
- [x] `internal/schemagen/` and `kbf schema [--check]` generate `schema/ontology.schema.yaml` and `schema/manifest.schema.yaml`, with doc comments as `description`. Done means: `--check` exits 1 on a stale schema, 0 on a current one.
- [x] Root Makefile with `check` (fmt, lint, test, dogfood-lint, conformance, schema-freshness, boundaries) and `.github/workflows/ci.yml` running that same target. Done means: CI cannot drift from the local gate.
- [x] Apache-2.0 `LICENSE`, short-form license headers on Go files in `tools/` and `scripts/`, `.gitignore`. Done means: headers stay out of `go doc` output and schema descriptions.

## Batch 2: Linter

`kbf lint` validates a package's structure and semantics and reports every finding
with file, line, rule id and fix.

- [x] Position-aware loader (goccy) with manifest and `extends` resolution by name across the paths given to `kbf lint`. Done means: every unknown field is reported in one pass at its own line; malformed YAML and a missing manifest become findings, never a panic.
- [x] Structural validation against the model (`KBF001` to `KBF004`, `KBF010`, `KBF011`), with `many-to-one` added to `model.Cardinalities`. Done means: `KBF002` fires on bad tier, cardinality, additivity and risk values and on an unknown `kind`.
- [x] Semantic rules (`KBF005` to `KBF009`, `KBF012`). `KBF012` fires on a `slots.yaml` row with no matching attribute; `KBF007` and `KBF008` key relations on `(name, from, to)`. Done means: `universal-core` and `cafe-demo` lint with no findings when passed together.
- [x] Renderers: one lipgloss table per file, and a stable `--format json` with no HTML-escaping. Done means: golden tests pin both; `<placeholder>` fix hints arrive literally.
- [x] Unit tests for all twelve rules and a no-findings baseline, on generic fixtures under `internal/lint/testdata/`. Done means: `go test ./...` green.

## Batch 3: Content

The first ontology content exists, and `kbf coverage` and `kbf compile` report on
it.

- [x] `packages/universal-core` (`extends: null`): nine entities (location, order, product and shift among them), relations over the controlled verbs, six metrics, four actions with `risk`, competency questions, a slot template. Done means: `kbf lint packages/universal-core` reports no findings.
- [x] `examples/cafe-demo` (`extends: universal-core`): a synonym override, a `client-configured` relation, a threshold override, net-new `waste-ratio`, slots mapped to `demopos` and `demobooks`. Done means: it lints with no findings alongside its parent; only `crm.customer-*` stays unmapped.
- [x] `kbf coverage` in `internal/coverage/`, built on `lint.Load` and `lint.Universe`, reporting on leaf packages only. Done means: a golden test shows exactly the `crm.customer-*` rows unmapped on `cafe-demo`.
- [x] `kbf compile --to mermaid` in `internal/compile/`, unioning every loaded package, sorted and deduped by identity. Done means: golden and determinism tests pass; the output renders in Mermaid.

## Batch 4: Conformance + boundaries + docs

A language-agnostic fixture suite, the public-hygiene scan and the prose docs land,
and `make check` enforces every gate.

- [x] Conformance fixtures under `conformance/` (`input/<package>/` plus `expect.yaml`), one invalid fixture per rule id, on a generic library domain, with a thin `go test` runner. Done means: at least 6 valid and 6 invalid fixtures pass; breaking a fixture fails with the named rule.
- [x] `scripts/boundaries.go` in its own Go module, with `make boundaries` running a self-test and then the scan. It fails on private filesystem paths and prices; names are reviewed by hand. Done means: the self-test catches a planted violation; `TestRealRepoIsClean` passes.
- [x] Spec prose: `spec/index.md`, seven `spec/primitives/*.md`, `spec/package-format.md`, `spec/conventions.md`, `spec/versioning.md`. Done means: each primitive doc carries a valid YAML example.
- [x] Spec-examples test in `tools/internal/docextract/`: walks the markdown AST, parses every `yaml` fence, and checks attributed blocks verbatim against their files. Done means: an edited example fails with a diff naming both sides.
- [x] README (thesis and five-minute tour), `CONTRIBUTING.md`, `rfcs/README.md`. Done means: each tour step produces what the README says it will.
- [x] `make check` runs fmt, lint, test, dogfood-lint, conformance, spec-examples, schema-freshness and boundaries. Done means: every target enforces, none passes on empty.

## Batch 5: Layered package architecture (recursive extends)

`universal-core` narrows to what every business shares, base packages
`operations-core` and `services-core` extend it, verticals extend those, and
`extends` resolves recursively.

- [x] `Universe.Chain` replaces single-hop `ExtendsRoot`: walks the full chain, nearest ancestor first, and stops on a missing link or a repeated name. Done means: a cycle is `KBF011` on every package on it; a missing ancestor still leaves the resolved part as context.
- [x] `KBF007` vocabulary is the union of verbs across the chain. Done means: `chain-verb-not-inherited-by-sibling` fails with `KBF007`.
- [x] `KBF008` checks all ancestors, nearest match wins. Done means: `chain-fork-against-grandparent` is reported against the grandparent by name.
- [x] Coverage's attribute-to-entity lookup unions the full chain; compile already unions every loaded package and needs no change. Done means: both commands run on both chains.
- [x] No depth limit remains in `tools/`; `manifest.yaml` keeps one `extends` name. Done means: `kbf schema --check` passes with no regeneration.
- [x] Chain tests in `internal/lint/chain_test.go`, mirrored as `conformance/chain-*` on a generic supply-chain domain. Done means: valid three-level, sibling verb, grandparent fork and cycle cases all pass.
- [x] `KBF007` own-entity minting: a verb also passes when the relation's `from` or `to` is an entity the same package introduces, tested per relation; redeclaring an inherited name grants nothing. Done means: `TestChainMintOnOwnEntity` and `conformance/chain-mint-own-entity` pass.
- [x] Makefile `dogfood-lint` covers both three-level chains, and the coverage and compile golden tests run table-driven over both. Done means: `make dogfood-lint` and `go test ./...` green.
- [x] `packages/universal-core` narrowed to seven entities: `location` and `shift` move out, `product` becomes `offering`, `order` becomes `transaction`, `employed-by` replaces `works-at`, metric grain stops at `[organization, business-date]`, `on` keys are quoted. Done means: the package lints with no findings.
- [x] `packages/operations-core` (`extends: universal-core`): adds location and shift, mints `located-at`, `works-at`, `staffed-by` and `sells`, adds `average-ticket` and `location-labor-cost-ratio`. Done means: it lints with no findings under own-entity minting.
- [x] `packages/services-core` (`extends: universal-core`): adds engagement and deliverable, mints no verbs, adds `effective-rate` and `on-time-delivery-rate`. Done means: it lints with no findings on ancestor verbs alone.
- [x] `examples/cafe-demo` re-pointed to `operations-core`: the synonym fragment targets `offering`, the threshold override targets `location-labor-cost-ratio`, slots cover the full chain. Done means: `crm.customer-*` is the only unmapped group.
- [x] `examples/studio-demo` (`extends: services-core`): a synonym, a `client-configured` `customer`-to-`customer` relation, a threshold override, net-new `repeat-engagement-rate`. Done means: `purchasing.supplier-*` is the only unmapped group.
- [x] `spec/` updated for the layered architecture: `index.md` walks all five packages, `conventions.md` states chain-wide vocabulary, `package-format.md` states recursive extension rules, `action.md` quotes `"on":`. Done means: `make spec-examples` green.
- [x] `README.md` updated: rows for both base packages and `studio-demo`, tour commands passing the full chain. Done means: the commands match what the CLI needs.
- [x] Proof: both chains lint with no findings; `make check` green.

## Batch 6: playbook naming + layer taxonomy

`packages/` becomes `playbooks/`, the core playbooks take `core-*` names, and a
manifest `layer` field makes the taxonomy machine-checked by `KBF013`.

- [x] `packages/` renamed to `playbooks/`, with every live reference updated: Makefile, spec, README, `CONTRIBUTING.md`, steering, Go sources and tests, fixtures. Done means: no live file calls the folder `packages/`.
- [x] Playbooks renamed: `universal-core` to `core-universal`, `operations-core` to `core-operations`, `services-core` to `core-services`. "Core playbooks" and "vertical playbooks" replace "base packages". Done means: every manifest's `name` and `extends` matches.
- [x] `spec/package-format.md` renamed to `spec/playbook-format.md`, with one term across `spec/`, defined there as "a runnable package of business capability". Done means: `package` appears nowhere else in `spec/`.
- [x] Manifest gains `layer: root | base | vertical` (`model.Layers`), assigned on all five playbooks. Done means: the schema regenerates with `layer` required; both chains lint with no findings.
- [x] `KBF013` in `internal/lint/taxonomy.go` cross-checks `layer` against `extends` (a root has none; base and vertical extend a root or base) and `name` (root and base match `^core-`, vertical does not). Done means: a unit test per branch, including the cases that must not fire.
- [x] `layer:` added to every fixture that asserts zero findings. Done means: `go test ./...` green.
- [x] Rule documented in `finding.go`, `kbf lint --help`, `spec/primitives/namespace.md` and `spec/conventions.md`, with fixtures `valid-layered-taxonomy` and `invalid-vertical-named-core`. Done means: conformance green.
- [x] Decision logged: "Playbook rename executed; the core/vertical taxonomy is machine-checked". Done means: the entry is in `project-decisions.md`.
- [x] Proof: `make check` green; both chains lint with no findings under the new names; the README tour runs as written.

## Batch 7: composition + meaningful naming

`builds-on` (a list resolved as a DAG closure) replaces `extends`, `layer` collapses
to `core | vertical`, `core-universal` becomes `core-business`, `Relation.tier`
becomes `origin`, `Action.risk` becomes `approval`, and `bistro-demo` exercises the diamond.

- [x] `Manifest.BuildsOn` replaces `Extends`; `Universe.Closure` replaces `Chain`: a walk deduped by name, with path-relative cycle detection and a result sorted by name. Done means: a diamond resolves its shared root to one instance; a genuine cycle is `KBF011`.
- [x] `KBF003` cross-playbook variant: two different closure members declaring one identity is an error on every file involved. Done means: it fires when a playbook composes both, and stays silent when they are only loaded side by side.
- [x] `KBF008` matches any closure member, `KBF007`'s vocabulary is closure-wide, `KBF011` reports one finding per unresolved `builds-on` entry. Done means: `conformance/chain-*` and `go test ./...` green.
- [x] `model.Layers` becomes `core`, `vertical`; `checkBuildsOnLayer` replaces `checkExtendsLayer` with the two-value build-target table; root-ness derives from an empty `builds-on`. Done means: taxonomy tests cover core-on-core, vertical-on-vertical, hybrid and every invalid shape.
- [x] Every `manifest.yaml` migrated to `builds-on` and the two-value `layer`, with taxonomy fixtures reworked for the new rules. Done means: `make check` green.
- [x] Field renames: `Relation.Tier` to `Origin`, `Action.Risk` to `Approval` (`auto|confirm` to `automatic|required`), across model, lint, compile, YAML content and primitive docs. Done means: the schema regenerates; entity and metric `tier` are untouched.
- [x] `playbooks/core-universal` renamed to `playbooks/core-business`; the "Playbook Zero" nickname stays; "extends", "chain" and "ancestor" give way to "builds on", "composition" and "closure" in live prose. Done means: `spec/` and README describe composition in those terms.
- [x] `examples/bistro-demo` (Demo Bistro & Events), `builds-on: [core-operations, core-services]`: the diamond and the hybrid, with the same four composition mechanics as the other demos. Done means: `core-business` dedupes to one instance; it lints with no findings; `dogfood-lint` and the golden tests include it.
- [x] Docs: `spec/playbook-format.md`'s composition rules, `namespace.md`, `conventions.md`, `versioning.md`, `index.md`, README, and comments in playbook files and fixtures. Done means: `make spec-examples` green.
- [x] Conformance: `diamond-valid`, `cross-playbook-collision-invalid`, `hybrid-valid`, `vertical-builds-on-vertical-valid`, `core-builds-on-vertical-invalid`. Done means: conformance green.
- [x] `design.md` states the closure rules: DAG mechanics, path-relative cycle detection, closure-wide `KBF007` and `KBF008`, the `KBF003` cross-playbook variant, the two-value layer table. Done means: each rule matches `internal/lint`.
- [x] `spec/onboarding.md` ("Raising an ontology for a new business"): eight steps in authoring order, with competency questions before modeling and the business owner interviewed throughout. Done means: `type: spec-doc` frontmatter, no YAML fences to verify.
- [x] `spec/index.md` gains a "For DDD practitioners" section and lists `onboarding.md` in its reading order. Done means: the section covers published language, bounded context, anticorruption layer, ubiquitous language, context integration and subdomain triage.
- [x] `spec/playbook-format.md`'s opening paragraph links `spec/onboarding.md`. Done means: one link, nothing restated on either side.
- [x] Decision logged: "Composition (DAG closure) replaces single-parent extends; core-business rename; origin/approval field renames". Done means: the entry is in `project-decisions.md` with rationale and a reversal condition.
- [x] Proof: `make check` green; all three demos lint with no findings over their full closures; the README tour runs as written.

## Batch 8: kbf-authoring skill

An agent raises a playbook through an installable skill that routes to the spec
and restates none of it.

- [x] `skills/kbf-authoring/SKILL.md`: reading order, prerequisites check, conduct rules over the eight onboarding steps, the lint-JSON authoring loop, prohibitions, exit criteria. Done means: the skill is a thin router into `spec/`.
- [x] `skills/README.md`: what a skill is, how to install one, a portability note. Done means: a reader can install the skill from it.

## Batch 9: batteries-included distribution

`go install` is the whole install: composing against the core playbooks, running
the `kbf-authoring` skill and reading the spec need no clone. Embedded content is
a fallback; a local path always wins.

- [x] `project-architecture.md`'s Boundaries entry allows embedding the core playbooks, the skill and the prose spec, strictly as a resolution fallback. Done means: `project-decisions.md` carries "Batteries-included binary: embedded cores/skill/docs with local-override precedence".
- [x] `scripts/embedsync` mirrors `playbooks/core-*`, `skills/kbf-authoring/` and the spec docs into `tools/internal/embedded/data/`; `make embed-sync` writes, `make embed-freshness` verifies inside `make check`. Done means: freshness fails on stale content and on orphaned files.
- [x] Resolution precedence: disk and embedded share one `fs.FS` loader; `lint.LoadWithEmbedded` fills only names no path provides, transitively, for `lint`, `coverage` and `compile`. Done means: a broken local override reports the local file; `--format json` is unchanged.
- [x] `kbf vendor [--to <dir>] [--force]` writes the three embedded core playbooks to real directories, default `./playbooks`. Done means: without `--force`, a conflict is refused up front, naming every conflicting directory.
- [x] `kbf skill install [--force]` writes the skill to `.claude/skills/kbf-authoring/`. Done means: same overwrite guard; it prints what it wrote and a next step that assumes no clone.
- [x] `kbf docs [name]` lists embedded doc names or prints one. Done means: an unknown name is an error listing the valid names.
- [x] `kbf init <name> [--builds-on a,b] [--layer core|vertical]` scaffolds a playbook; `--layer` defaults to `vertical`, which requires `--builds-on`. Done means: a fresh scaffold lints with no findings, a comment-only `slots.yaml` counts as zero rows, a second `init` is refused.
- [x] Docs: `SKILL.md` prerequisites, `skills/README.md`, `spec/cli.md` (eight commands, `kbf init`, "Embedded content"), README quickstart with no clone. Done means: `make embed-freshness` green.
- [x] Tests: precedence on a synthetic `fstest.MapFS` fallback, embed-freshness in both directions, CLI tests for `init`, `skill`, `vendor` and `docs`. Done means: a freshly initialized playbook lints with no findings through embedded resolution alone.
- [x] Proof: `make check` green; a binary built from a fresh clone runs `kbf init` and `kbf lint` from an empty directory with no checkout in reach; the README tour runs as written.
- [x] README "What's here" row for `skills/`. Done means: the row links the skill.

## Batch 10: `kbf vendor` renamed to `kbf playbooks pin` + vocabulary sweep

The CLI speaks KBF vocabulary and every mutation is a verb under a noun namespace:
`kbf vendor` becomes `kbf playbooks pin`, with no compatibility alias before 1.0.

- [x] `tools/cmd/kbf/vendor.go` becomes `playbooks.go`: a help-only `playbooks` parent with a `pin` subcommand, and identifiers, error prefixes and help text in pin vocabulary. Done means: `kbf playbooks pin` works; `kbf vendor` does not exist.
- [x] `--to` keeps its `playbooks` default, matching every documented layout at this point; `spec/cli.md` states why. Done means: the doc examples and the default agree.
- [x] `vendor_test.go` becomes `playbooks_test.go`, with comments updated in `init_test.go`, `skill.go` and `testhelpers_test.go`. Done means: `go test ./...` green.
- [x] Docs: `spec/cli.md`, README, `SKILL.md` prerequisites, `project-architecture.md`. Domain uses of "vendor" and dev-internal tooling stay. Done means: `make embed-freshness` green.
- [x] Decision logged: "The CLI speaks KBF vocabulary: `kbf vendor` renamed to `kbf playbooks pin`". Done means: the entry is in `project-decisions.md`.
- [x] Proof: `make check` green; `kbf playbooks pin` writes all three cores, a re-run refuses without `--force`, and `--to` is honored.

## Batch 11: install-repo layout documented, pin default flipped to `builds-on`

The install-repo layout is documented in the spec, and the default `--to` of `kbf
playbooks pin` flips to `builds-on`, where tools that are not `kbf` read the closure.

- [x] `spec/onboarding.md` gains "The install repo": one playbook per repo, rooted at `playbook/`, the pinned closure in `builds-on/`. Done means: `make spec-examples` green.
- [x] `playbooks pin` default `--to` flipped from `playbooks` to `builds-on`, with help text, `spec/cli.md` and tests following. Done means: pin run from inside `playbook/` lands the three cores in `builds-on/`.
- [x] The "The CLI speaks KBF vocabulary" decision entry records the flip. Done means: the entry names the spec section that documents the layout.
- [x] `kbf init` stays without `builds-on/`: pin creates the directory on demand, and `spec/onboarding.md` says so. Done means: doc and CLI agree.
- [x] Proof: `make check` green; `kbf init playbook --builds-on core-operations`, pin from inside `playbook/`, then `kbf lint playbook` from the repo root reports no findings.

## Batch 12: spec, tooling and docs state one set of rules

Each rule reads the same in the spec, the linter, the fixtures, the doc comments
and the steering docs, and every doc states what the CLI does.

- [x] A dangling action `on` reports `KBF009`, as `spec/primitives/action.md` states. Done means: a conformance case asserts the rule id.
- [x] Spec prose matches the CLI: eight commands, embedded core playbooks resolve without a path, naming examples pass `KBF013`, fields the linter does not check are named as such. Done means: each claim run against `kbf`; `make spec-examples` green.
- [x] Error messages, help text and fixture names say "playbook" and `builds-on`. Done means: goldens updated, `go test ./...` green, JSON keys unchanged.
- [x] Doc comments state each rule where it applies; schema regenerated. Done means: `make schema-freshness` green.
- [x] Fixtures state the `KBF007` rule and quote `on`. Done means: conformance green.
- [x] `make boundaries` scans for private paths and prices, with a self-test. Done means: `make boundaries` green.
- [x] Standards name the toolchain as built. Done means: entry in `project-decisions.md`.
- [x] Proof: `make check` green.
