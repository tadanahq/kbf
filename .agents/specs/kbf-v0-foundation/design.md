---
type: spec-design
---
# kbf-v0-foundation – Design

Architecture authority: `.agents/steering/project-architecture.md` (layout,
meta-model table, tooling packages). This design adds the feature-level detail;
it does not restate steering.

## The meta-model (authoring shapes)

One YAML document per element, one file per entity plus its owned elements, or
grouped files per kind: the linter accepts both; `kind:` discriminates.

```yaml
kind: entity
name: dish            # kebab-case, unique within namespace
meaning: >-           # one sentence, required
  A sellable menu item.
identity: [dish_id]
resolution: identity-mapping   # free text in v0, named strategies later
tier: structural
synonyms: {en: [item], es: [plato]}
attributes:
  - {name: category, type: text, slot: pos.catalog}
states: [active, retired]      # optional
```

```yaml
kind: relation
name: contains
from: dish
to: ingredient
cardinality: many-to-many      # enum: one-to-one|one-to-many|many-to-one|many-to-many
join: [dish_id, ingredient_id]
tier: source-synced            # enum: source-synced|client-configured
temporal: false
```

```yaml
kind: metric
name: dish-margin
formula: (revenue - ingredient-cost) / revenue   # v0: opaque expression string
grain: [dish, location, business-date]
additivity: non-additive       # enum: additive|semi-additive|non-additive
unit: ratio
thresholds: {warn-below: 0.60} # glossary tier by definition
tier: structural
```

```yaml
kind: action
name: flag-for-review
"on": dish                     # always quoted, see clarifications below
risk: auto                     # enum: auto|confirm
writes: finding
```

```yaml
kind: competency-question
question: Which dishes lost margin this month, per location?
expects: [dish-margin]         # elements the answer must use
```

Slot mappings live in `install/slots.yaml`: `{slot: pos.catalog, source: ""}`
rows; static coverage = share of declared slots with a non-empty source.
`manifest.yaml`: `{name, version, spec: v0, extends: universal-core}`
(universal-core itself: `extends: null`).

v0 keeps `formula` and `resolution` as opaque strings deliberately: parsing
expressions is gate-era work; the linter checks presence and cross-references
(`expects`, `grain`, `slot` targets must exist), not expression semantics.

## Lint rule set (rule ids are public API)

`KBF001` unknown field · `KBF002` bad enum value · `KBF003` duplicate name in
namespace · `KBF004` entity without identity · `KBF005` metric without
grain/additivity · `KBF006` relation endpoint not declared · `KBF007` verb
outside controlled vocabulary · `KBF008` fork of a core element (redefinition
in an extending package) · `KBF009` dangling cross-reference · `KBF010`
missing governance tier · `KBF011` manifest missing/invalid · `KBF012` slot
reference without declaration. Human render: lipgloss table grouped by file.
`--format json`: `{rules: [{id, file, line, element, message, fix}]}`.

## Controlled verb vocabulary

No fixed seed list anymore: superseded by the "Implementation
clarifications" entries on chain-wide union and the KBF007 final rule
(below, owner-adjudicated 2026-08-13). `universal-core`'s original nine
(`contains, belongs-to, supplies, places, billed-to, employed-by,
derived-from, supersedes, responsible-for`) plus each base package's own
mint (`operations-core`: `located-at, sells, staffed-by, works-at`) are
what exist today; any package may mint a new one on a pair that touches
an entity it declares itself, RFC discipline still applies for anything
broader (reuse first), see `spec/conventions.md`.

## Package architecture (layered, owner correction 2026-08-13)

Three tiers, not one flat `universal-core` + verticals. See the
"Layered base packages" entry in `.agents/steering/project-decisions.md`
for the full rationale (the original `universal-core` carried an
operating-business bias: location, shift, order, and product are not
universal).

- **`packages/universal-core`** (`extends: null`): the truly universal
  floor. Entities: organization, customer, offering (product or service),
  transaction (the revenue event), employee, supplier, purchase. Metrics:
  revenue, transaction-count, average-transaction-value, gross-margin,
  labor-cost-ratio, purchase-cost, grain stops at `[organization,
  business-date]`. Actions unchanged in spirit (flag-for-review, annotate,
  request-sync, propose-extension), retargeted off `location`/`order`.
  Competency questions: 1+ per entity, business-neutral wording.
- **`packages/operations-core`** (`extends: universal-core`): the base
  layer for site-based businesses. Adds location, shift; verbs
  located-at, staffed-by, works-at (all minted here, `employed-by` stays
  universal-core's location-free HR fact), belongs-to and responsible-for
  reused on new pairs. Metrics: average-ticket and
  location-labor-cost-ratio, both net-new names at `[location,
  business-date]` grain (not overrides: grain isn't glossary-eligible, so
  a same-name redeclaration at a different grain would fork).
- **`packages/services-core`** (`extends: universal-core`): the base
  layer for engagement-based businesses. Adds engagement (synonyms:
  project, retainer), deliverable. Mints no new verbs at all: `places`,
  `contains`, `derived-from`, `responsible-for` all reused on new pairs,
  RFC-reuse-first taken literally. Metrics: effective-rate and
  on-time-delivery-rate; utilization considered and dropped (would need
  employee capacity data universal-core's employee doesn't carry, and
  services-core can't add attributes to an ancestor's entity without
  forking it).
- **`examples/cafe-demo`** (`extends: operations-core`): re-pointed, not
  rebuilt. Same four extension mechanics as before (glossary synonym,
  client-configured relation, threshold override, one vertical metric),
  same fictional business, adjusted only where the restructure forces it:
  the entity fragment now matches `offering` (declared two levels up, at
  universal-core; KBF008 matches the nearest ancestor with the identity,
  so this is unaffected by the extra hop), and the threshold override now
  targets operations-core's `location-labor-cost-ratio` instead of a
  metric that no longer has a location grain.
- **`examples/studio-demo`** (`extends: services-core`, new): the second
  demo, one per base chain. Fictional single-team marketing agency
  ("Demo Studio"), mirrors cafe-demo's teaching role exactly: a glossary
  synonym (`engagement` → "campaign"), a client-configured relation
  (customer-to-customer, parent-account grouping, the same self-relation
  shape as cafe-demo's location grouping), a threshold override
  (on-time-delivery-rate), one net-new metric (repeat-engagement-rate),
  slots mapped to fictional sources (`demopm`, `demobooks`), a couple
  left unmapped on purpose (`purchasing.supplier-*`, a different gap than
  cafe-demo's, proof the pattern isn't hardcoded to one business).

Both demo chains exist specifically so the "one shared multi-vertical
core" claim is checkable, not asserted: `universal-core` is never linted
alone as "the" example, only as the common ancestor of two visibly
different businesses.

## Key flows

- `kbf schema`: model structs → JSON Schema → YAML files in `schema/`;
  `--check` mode diffs against committed files (CI freshness job).
- `kbf lint <path...>`: load manifest → resolve extends chain (v0: single
  parent) → parse files position-aware → structural validation (against model)
  → semantic rules → render. Exit 1 on any rule hit.
- `kbf coverage <path>`: table of slots by entity: declared / mapped / unmapped.
- `kbf compile --to mermaid <path>`: entities as nodes, relations as labeled
  edges, actions as annotations. Deterministic output (sorted), golden-tested.
- Boundaries: `scripts/boundaries.go` walks the tree, fails on blocklist terms;
  blocklist named in-file; self-test plants a violation in a temp tree.

## Edge cases & constraints

- Lint must not panic on arbitrary YAML: fuzz the loader with the invalid
  fixtures; unknown `kind` is KBF002 not a crash.
- Windows paths and CRLF tolerated; output paths always forward-slash.
- Everything deterministic: sorted iteration everywhere (maps never range
  unsorted into output).

## Decisions

- Cross-file resolution happens per package after full load; no incremental
  mode in v0.
- ~~`extends` chain depth 1 in v0 (universal-core only); deeper chains are a
  spec-versioning question, not a linter feature.~~ Superseded, owner
  correction 2026-08-13: the package architecture is layered (verticals →
  base packages → universal-core), so `extends` is fully recursive as of
  the entry below. Depth 1 was a v0 simplification, not a spec-versioning
  boundary; it did not survive contact with the real architecture.
- Mermaid over any richer render: it previews natively in editors and GitHub.

## Implementation clarifications (resolved during Batch 1/2, binding)

Gaps found while building the model/linter. Recorded here per "don't silently
diverge"; binds Batch 3/4 content authoring too.

- **`tier` vocabulary is kind-specific.** Entity/Metric `tier` uses the
  steering governance vocabulary (structural/glossary/instance), inherited
  silently since design.md examples don't restate steering. Relation `tier`
  is a distinct axis (source-synced/client-configured), explicitly annotated
  in its example because it deviates from the default. Action has no `tier`
  field; its governance-equivalent is `risk` (auto/confirm) per
  project-architecture's "risk tier" framing. Competency-question and slot
  mapping (operational elements) carry neither: KBF010 does not apply to them.
- **KBF010 scope**: fires on empty Entity.tier, Relation.tier, Metric.tier, or
  Action.risk. **KBF002 scope**: fires on any non-empty enum field whose value
  is outside its kind-specific allowed set (tier, cardinality, additivity,
  risk, and `kind` itself). Only fields design.md marks `# enum:`, plus tier/
  risk per above, are closed vocabularies; `resolution`, `formula`, `unit`,
  `attributes[].type` stay opaque strings in v0 (matches the documented
  "formula and resolution are opaque strings" intent, generalized).
- **KBF007 vocabulary source**: no separate declaration file, no new
  primitive, no `kind: verb-vocabulary`. (Superseded twice below, kept for
  the "no declaration file" fact, which is still true: first by chain-wide
  union when `extends` went recursive, then by the owner-adjudicated
  own-entity minting rule — see "KBF007 final rule" further down for the
  current, exact test.)
- **KBF008 fork detection**: an extending package redeclaring an element with
  a name+kind that exists in its extends-root is a fork UNLESS every
  non-glossary field on the child's copy is zero-valued, i.e. the child only
  layers glossary-tier fields. Glossary-eligible fields in v0: Entity.synonyms,
  Metric.thresholds (matches "thresholds: glossary tier by definition").
  Relation and Action have no glossary carve-out: any redeclaration forks.
- **`extends` resolution**: `kbf lint` takes one or more package paths as
  positional args and loads all of them into one in-memory set keyed by
  manifest `name`. Each package's `extends` is resolved by name against that
  set (not by filesystem convention: kbf lints arbitrary directories, not just
  this repo's `packages/`). A package whose `extends` isn't among the supplied
  paths fails KBF011 with a fix hint to pass the parent's path too. Linting a
  single root package (`extends: null`) needs only its own path.
- **KBF009 vs KBF006 vs KBF012 split**: KBF006 is relation.from/to only.
  KBF012 is attribute slot references against `install/slots.yaml` only.
  KBF009 (generic dangling cross-reference) covers everything else that names
  another element: metric.grain entries, action.on, competency-question.expects.
  Manifest.extends failing to resolve is KBF011 (manifest invalid), not KBF009.
- **Schema file scope**: exactly two files as tasked. `SlotMapping` (the
  `install/slots.yaml` row shape) is reflected into `manifest.schema.yaml`'s
  `$defs` (installation-config family) but is not itself a file root schema in
  v0, so `install/slots.yaml` doesn't get direct yaml-language-server root
  validation yet. Structural validation of slots still happens via the Go
  struct in the loader; this is a known, deliberate v0 gap, not a silent one.
- **Relation identity for KBF003/KBF008 is (name, from, to), not bare name**
  (found authoring content; binds Batch 2 as much as the entries above).
  Relation.name holds a controlled-vocabulary *verb*, meant to recur across
  unrelated entity pairs (that's the whole point of a 10-20 word vocabulary
  covering dozens of relations: project-standards.md, and KBF007's own
  wording, "the *set of distinct* Relation.name values", presupposes
  repeats feeding that set). So (1) KBF003 duplicate-name-in-namespace, for
  `kind: relation` only, keys on the (name, from, to) triple, not name
  alone: universal-core legitimately declares `contains` and `supplies` and
  `located-at` twice each, over different pairs. (2) KBF008 fork-matching
  for `kind: relation` also keys on (name, from, to): a child package
  declaring a relation whose triple doesn't already exist in the parent is
  a new relation, full stop, not a fork candidate at all, even when the verb
  is already used elsewhere in the parent under a different pair. This has
  to be true or cafe-demo's own required extension (a new client-configured
  `belongs-to` between two locations, reusing a verb the parent already
  spends on location→organization) would be structurally impossible to
  express: no package could ever add a relation without first winning an
  RFC for a brand-new, never-before-used verb. "Relation... any
  redeclaration forks" (above) still holds for the case this note carves
  out from it: same (name, from, to) triple as the parent, re-stated by a
  child, with or without field changes, is always a fork, no glossary
  carve-out. Entity/Metric/Action/competency-question identity stays plain
  name (already globally unique by construction), unaffected by this note.
- **Completeness checks (KBF004/KBF005/KBF010) skip override fragments.**
  Confirmed against real content: cafe-demo's `product` entity fragment is
  `{kind, name, synonyms}` only, and its `labor-cost-ratio` metric fragment is
  `{kind, name, thresholds}` only, matching the glossary-override shape
  exactly (fields elsewhere in this doc). Validated standalone, both would
  fail KBF004 (no identity) and KBF010/KBF005 (no tier / no grain+additivity)
  even though they are correct content. So: an element is classified as an
  override fragment (name+kind match in extends-root; name+from+to match for
  relations) BEFORE structural completeness rules run, and KBF004/005/010 are
  skipped for it: KBF008 already validates that everything it sets is
  glossary-eligible, and completeness is inherited from the parent, not
  re-asserted. A NEW element (no match in extends-root) still needs full
  completeness: cafe-demo's `waste-ratio` metric and `location`-`location`
  relation are both declared in full, and are still checked.
- **KBF009's resolution set spans every kind, matched by bare name.**
  universal-core's competency questions reference relation verbs directly
  (`expects: [sells]`, `expects: [works-at]`, `expects: [places]`), not just
  entity/metric names. So KBF009 resolves a referenced name against entities,
  metrics, actions, and relations together (relations by their `name`/verb,
  same field used for KBF007, not the full (name,from,to) triple: confirming
  *something* by that name exists is KBF009's job, not disambiguating which
  pair, so a recurring verb still resolves).
- **Cardinality is four values, not three: `many-to-one` was a real gap,
  not a content mistake.** The original enum above (`one-to-one|one-to-many
  |many-to-many`) was incomplete. `from`/`to` are fixed and directional, so
  "many locations belong-to one organization" is genuinely many-to-one; it
  cannot be restated as one-to-many without reversing from/to, which would
  also reverse the relation's meaning. `packages/universal-core` uses
  `many-to-one` 7 times (every child-to-parent relation: `belongs-to`,
  `works-at`, `staffed-by`, `located-at`, `billed-to`) and correctly so;
  linting it against the 3-value enum produced 8 false-positive KBF002s
  across both packages before this was caught. Fixed in
  `internal/model.Cardinalities` and regenerated schema.
  `spec/primitives/relation.md`'s field table fixed by the owner directly
  (commit `a1166b1`); the model already matched, nothing more to do there.
- **`extends` is fully recursive: chains, not a single hop (owner
  correction, 2026-08-13, binding).** The package architecture is layered
  (verticals extend base packages, which extend `universal-core`), so a
  package's parent may itself have a parent. `Universe.Chain(pkg)`
  (`internal/lint/universe.go`) walks the whole line and returns it
  nearest-ancestor-first (immediate parent, then grandparent, ..., ending
  at a package with `extends: null`); every rule that used to take a
  single resolved parent now takes this chain instead. Four concrete
  semantics changes, each checked by a fixture:
  - **Cycle detection.** The walk tracks package names already seen; a
    repeat means a cycle. This is `KBF011` (manifest invalid), not a hang
    or a stack overflow: a package whose extends chain loops back on
    itself is exactly as broken as one whose manifest is missing fields.
    Every package on the cycle gets its own finding (each independently
    detects the same loop starting from itself), not just one.
  - **Missing-ancestor degradation is unchanged in kind, just possibly
    farther up.** A package's own `extends` still only needs to resolve
    against the universe *(checkManifest, immediate hop, unchanged)*; the
    "grandparent wasn't passed to `kbf lint`" case surfaces as `Chain`
    returning a shorter-than-expected slice, which every downstream rule
    already treats as "less context available", not a special case to add.
  - **KBF007's vocabulary is the union across the chain, not just the
    root.** A base package that adds a verb (`operations-core` declaring
    `located-at`, say) legalizes it for every package that extends
    *that* package, directly or through further descendants, not for an
    unrelated chain that also happens to extend `universal-core`. A
    package's own newly-declared verbs never count toward its own check
    (unchanged principle, just evaluated against the full union now): only
    ancestors establish vocabulary. Self counts only when the package has
    no ancestors at all (linting a root package alone) — the original "or
    self when linting universal-core itself" carve-out, now phrased
    generally instead of naming one specific package.
  - **KBF008 matches against the nearest ancestor that has the identity,
    not just the immediate parent.** A grandchild package forking an
    element that only its *grandparent* declares (the parent never touched
    it) still gets caught, reported against the grandparent by name. If
    two ancestors both happen to carry the same identity (the parent
    already layered a legitimate glossary override onto something the
    grandparent declared, and a third-generation package tries to touch it
    again), the nearer one is what is being extended or forked; farther
    ancestors are not consulted once a match is found. Glossary carve-outs
    and the "relation/action have none" rule are unchanged, applied at
    whichever ancestor the match resolves to.
  - **Coverage and compile's chain-wide behavior is a generalization, not
    a new decision.** Coverage's attribute-to-entity lookup and slot-usage
    checks now union every package in the chain, not just the immediate
    parent, for the same reason as before: an install configures the whole
    resolved ontology. Coverage's leaf-only *reporting* rule is unaffected
    by chain depth: "not the extends-root of any other loaded package" is
    already a chain-depth-agnostic definition. Compile's graph-building was
    already chain-depth-agnostic too (it unions every loaded package
    unconditionally, root or leaf, and always has): no code change there
    at all, confirmed by re-reading it during this pass, not assumed.
- **`Action.on` must always be quoted in authored YAML (found while
  restructuring content for the layered-packages correction).**
  `packages/universal-core/ontology/actions.yaml` had four unquoted `on:`
  keys. goccy/go-yaml is YAML-1.2-compliant (its own source comments this
  explicitly: `reservedLegacyBoolKeywords` is used only to decide when
  *encoding* needs quotes for other parsers' benefit, never at parse
  time), so `kbf` itself reads a bare `on:` correctly as the string key
  `"on"`, not a boolean. But this repo's own stated goal is third-party,
  language-agnostic conformance (`conformance/`), and a YAML-1.1 parser
  (PyYAML's default `safe_load`, among others) resolves an unquoted
  `on`/`off`/`yes`/`no` map key to a boolean, silently losing the field.
  Fixed at the source: every `on:` in `packages/universal-core`,
  `spec/primitives/action.md`, and this file's own meta-model example is
  now `"on":`. Not a `kbf lint` gap (goccy/go-yaml was never wrong), but
  worth a rule anyway: if `internal/lint` ever gains a style/portability
  pass, "unquoted `on:` key" is a clean candidate.
- **RESOLVED by owner, 2026-08-13 — see the entry below this one for the
  exact adopted rule and what changed in the implementation.** Kept below
  as-is for the reasoning trail: does a "base package" (operations-core,
  services-core) get to introduce a verb itself, or only reuse one
  universal-core already declares? Found dogfooding against the
  in-progress restructure, not resolved by guessing either way, because
  the two readings are mutually exclusive and each breaks something real:
  - **Reading A (implemented): only a true root (`extends: null`) is
    self-included; every other package's vocabulary is strictly the union
    of its ancestors', never its own new declarations.** This is what
    ships right now (`controlledVocabulary`, `internal/lint/semantic.go`).
    It keeps KBF007 reachable at all (see below) and keeps every fixture
    in `internal/lint/testdata/chain/` and this batch's conformance
    fixtures internally consistent, `invalid-relation-bad-verb`
    (Batch 4, still passing) included. Under this reading,
    `packages/operations-core` currently fails KBF007 on `works-at`,
    `staffed-by`, `located-at`, and `sells`, because
    `packages/universal-core`'s own `ontology/relations.yaml` does not
    (yet, as of this note) declare any relation using those verbs.
  - **Reading B (rejected, not just unimplemented): self is always
    included, for every package, root or not.** This is the only reading
    under which `operations-core`'s *current* content is legal as
    written, and matches one plausible parse of "self included when
    linting a root" (root as the trivial illustrative case of a broader
    rule, not the only case). Rejected, not merely deferred, because it is
    provably unworkable, not just a style preference: verb legality is
    checked per relation, and a relation's own `name:` is definitionally
    part of "its own package's declarations", so self-inclusion makes
    every freshly-authored relation legalize itself unconditionally, for
    every package, always. KBF007 could then never fire on any
    non-matched relation, for any input: not this repo's content, not a
    conformance fixture, not anything. Both `invalid-relation-bad-verb`
    (Batch 4) and this batch's own sibling/cousin fixture
    (`chain-cousin` in `internal/lint/testdata/chain/` and
    `conformance/`, once added) would stop failing, and there would be no
    way to construct a fixture that does: a rule that cannot fail on any
    input is not a rule.
  - **What this means for the content in flight**: either
    `packages/universal-core` still needs a relation using each of these
    four verbs (even a minimal, universally-plausible pairing, the same
    role `chain/grandparent`'s `assembled-from: batch -> material` plays
    in this batch's fixtures) before the restructure is done, or Reading A
    itself needs owner correction to something narrower than "always" that
    this note did not find (chain position relative to the root was
    considered and does not work either: `invalid-relation-bad-verb`'s
    failing child is exactly one hop below its root, the same distance
    `operations-core` is below `universal-core`). Not resolved here on
    purpose: this is a judgment call for whoever owns the vocabulary
    policy, not something to silently pick a side on. Full readout in the
    implementer's report for this batch.
  - **Reading C (content-agent proposal, not implemented, owner to
    confirm or reject): self-inclusion is anchored to newly-declared
    entities, not to the package as a whole.** A package's vocabulary
    contribution is the set of verbs used on a relation where `from` or
    `to` is an entity *that package itself declares* (not inherited),
    ancestors' verbs still union in as before. Checked against a
    universally-plausible pairing rejected above as contrived, this
    reading rejects nothing that reading needed to: every one of
    `operations-core`'s four verbs sits on a pair with a new entity on at
    least one side (`located-at`: transaction→**location**; `works-at`:
    employee→**location**; `staffed-by`: **shift**→employee; `sells`:
    **location**→offering), so all four pass without `universal-core`
    inventing a contrived non-location use of a verb whose entire reason
    to exist is location. It is not vacuous the way Reading B is: a
    relation between two *inherited* entities (both sides already
    declared by an ancestor) still gets no self-inclusion at all, only
    ancestor vocabulary, so a made-up verb on, say, `employee` to
    `organization` inside `operations-core` would still fail exactly like
    Reading A says it should — `invalid-relation-bad-verb` and
    `chain-cousin` need checking against this reading specifically
    (not assumed clean), but the shape of the argument survives because
    neither fixture's premise depended on the bad relation touching a
    newly-declared entity. Rationale, not just a fixture-passing hack: a
    new entity is the unit of "new business concept" a package
    introduces, and it makes sense that introducing one carries the
    right to name the relationships it participates in without
    pre-clearance; a relation between two concepts that already existed
    in an ancestor has no comparable justification for inventing a new
    way to connect them instead of reusing what is already expressible.
    Content-side consequence if adopted: no change needed to
    `packages/operations-core` as authored. Content-side consequence if
    rejected in favor of Reading A as-is: `packages/universal-core` needs
    a genuine reason to touch `located-at`/`works-at`/`staffed-by`/`sells`
    itself, which the whole point of this restructure argues it should
    not have — so rejecting this reading without a fourth alternative
    reopens the question, it does not close it.
- **KBF007 final rule (owner adjudication, 2026-08-13, binding): Reading C
  adopted, stated precisely as a per-relation test, not a package-wide
  vocabulary set.** A relation's verb passes if EITHER (a) the verb is
  declared by any ancestor in the package's resolved extends chain
  (unchanged from Reading A: ordinary reuse), OR (b) the relation's `from`
  or `to` entity is declared by the *same package* that declares the
  relation (minting rights come with introducing new entities). Fails
  KBF007 otherwise, fix hint pointing to reuse-or-RFC. Statically
  evaluable, no declaration ordering involved: unlike Reading A, self
  gets no special-case carve-out for "linting a root alone" — a root's
  own relations pass via (b) automatically, because a root necessarily
  declares every entity it references (nothing above it to inherit from).
  Confirmed against real content: `operations-core`'s four minted verbs
  (`located-at`, `works-at`, `staffed-by`, `sells`) each touch `location`
  or `shift`, both of which it declares, so all four pass; `services-core`
  mints nothing, so it passes trivially via (a) alone. Implementation
  notes beyond what Reading C's prose above specified, found while coding
  it precisely:
  - **Evaluated per relation, not per package.** A package minting a verb
    on one pair does not blanket-legalize that verb for a *different*,
    fully-inherited pair elsewhere in the same package: each relation
    independently needs (a) or its own (b). `internal/lint/testdata/chain
    /mint-own-entity` is exactly this shape (two relations, same verb,
    one passes, one fails) and is unit-tested directly
    (`TestChainMintOnOwnEntity`), since conformance's binary ok/not-ok
    verdict can't express "one relation in this fixture should fail and
    another shouldn't" as cleanly as a Go test's per-finding assertions.
  - **"Declared by the same package" means genuinely new, not merely
    textually present.** A package redeclaring an inherited entity's name
    (fork or legitimate glossary override alike, KBF008) does not gain
    minting rights from that redeclaration: nothing new was actually
    introduced. `internal/lint/testdata/semantic/child`'s `invented-verb`
    relation (`workshop -> widget`) is the fixture that forced this
    distinction: `widget` is textually present in the child's own
    `entities.yaml`, but only as a fork of the parent's `widget`, so it
    must not count. `internal/lint/semantic.go`'s `mintedEntities` builds
    the package's own entity names, then subtracts every name that
    already exists in any ancestor, rather than testing raw presence in
    `pkg.Elements`.
  - **Code**: `internal/lint/semantic.go` — `controlledVocabulary(chain)`
    is condition (a) alone now (ancestor-only, the old empty-chain
    self-fallback removed: subsumed automatically by (b) as noted above);
    `mintedEntities(pkg, chain)` is condition (b); `checkGrainAndVocabulary`
    checks `!ancestorVerbs[v.Name] && !minted[v.From] && !minted[v.To]`.
  - **Tests added**: `internal/lint/chain_test.go`'s
    `TestChainMintOnOwnEntity` (valid mint + the per-relation invalid
    reuse, both in `testdata/chain/mint-own-entity`);
    `TestChainVerbNotInheritedBySibling` reworked (`testdata/chain/
    cousin-invalid-verb` no longer declares its own entities — they moved
    to `testdata/chain/other-root` — so the fixture is genuinely "new verb
    over two inherited entities", the shape Reading B would have let
    through by accident before this precision was added). Conformance:
    `chain-mint-own-entity` (new, valid) and `chain-verb-not-inherited-by-
    sibling` (existing, same entity-ownership fix applied to its `input/`
    copy). Real dogfood: `kbf lint` on both full chains
    (`packages/universal-core packages/operations-core examples/cafe-demo`
    and the `services-core`/`studio-demo` equivalent) is green — confirmed
    after this rule landed, was 4 KBF007 findings on `operations-core`
    under Reading A before it.
