---
type: spec-design
---
# kbf-v0-foundation – Design

Architecture authority: `.agents/steering/project-architecture.md` (layout,
meta-model table, tooling packages). This design adds the feature-level detail;
it does not restate steering. Every rule below binds the linter, the content and
the fixtures alike.

## Components

Beyond the packages steering lists:

- `tools/internal/conformance/`: a thin `go test` runner over `conformance/`. Each
  case is `input/<playbook>/...` plus `expect.yaml` (`ok: true`, or `ok: false`
  with `rules: [...]`). Fixtures contain no Go.
- `tools/internal/docextract/`: walks the markdown AST of `spec/**/*.md`, parses
  every `yaml` fence, and checks any block attributed to a repository file
  verbatim against that file (`make spec-examples`).
- `scripts/boundaries.go`: the public-hygiene scan, in the `scripts/` Go module.
- `scripts/embedsync/`: same module; mirrors the embedded sources into
  `tools/internal/embedded/data/` (`make embed-sync`) and verifies the mirror
  (`make embed-freshness`).

## Data & concepts

### The meta-model (authoring shapes)

One YAML document per element, one file per entity plus its owned elements, or
grouped files per kind: the linter accepts both; `kind:` discriminates.

```yaml
kind: entity
name: book            # kebab-case, unique within namespace
meaning: >-           # one sentence, required
  A title the library lends.
identity: [book_id]
resolution: identity-mapping   # free text in v0, named strategies later
tier: structural
synonyms: {en: [title], es: [libro]}
attributes:
  - {name: genre, type: text, slot: catalog.genre}
states: [active, retired]      # optional
```

```yaml
kind: relation
name: contains
from: branch
to: book
cardinality: many-to-many      # enum: one-to-one|one-to-many|many-to-one|many-to-many
join: [branch_id, book_id]
origin: source-synced          # enum: source-synced|client-configured
temporal: false
```

```yaml
kind: metric
name: loan-rate
formula: loans / copies        # v0: opaque expression string
grain: [book, branch, business-date]
additivity: non-additive       # enum: additive|semi-additive|non-additive
unit: ratio
thresholds: {warn-below: 0.60} # glossary tier by definition
tier: structural
```

```yaml
kind: action
name: flag-for-review
"on": book                     # always quoted
approval: automatic            # enum: automatic|required
writes: finding
```

```yaml
kind: competency-question
question: Which books fell below their loan rate this month, per branch?
expects: [loan-rate]           # elements the answer must use
```

```yaml
# manifest.yaml: one flat document, no kind
name: cafe-demo
version: 0.1.0
spec: v0
builds-on: [core-operations]   # a list; [] makes a core playbook a root
layer: vertical                # enum: core|vertical
```

Slot mappings live in `install/slots.yaml`, a bare list of `{slot: catalog.genre,
source: ""}` rows. Static coverage is the share of rows with a non-empty source.
A core playbook's file lists only the slots of its own attributes; a leaf
install's file lists the slots of its whole closure.

`formula`, `resolution`, `unit` and `attributes[].type` are opaque strings in v0:
parsing expressions is query-gate work. The linter checks cross-references
(`expects`, `grain`, `on`, `from`, `to`, slot rows), not expression semantics.

**`Action.on` is always quoted.** A YAML 1.1 parser resolves a bare `on` key to a
boolean and loses the field, and conformance is meant to hold for any parser.
`kbf` itself parses YAML 1.2 and reads either spelling, so `kbf lint` does not
flag a bare `on`.

### Governance fields

The vocabulary is kind-specific; each field answers a different question.

| Kind | Field | Values |
|---|---|---|
| Entity, Metric | `tier` | `structural`, `glossary`, `instance` |
| Relation | `origin` | `source-synced`, `client-configured` |
| Action | `approval` | `automatic`, `required` |
| Competency question, slot mapping | none | `KBF010` does not apply |

Closed vocabularies are exactly those fields plus `cardinality`, `additivity`,
`layer` and `kind`.

Cardinality has four values because `from` and `to` are directional:
child-to-parent is `many-to-one`, and restating it as `one-to-many` would reverse
the relation's meaning.

### Element identity

An entity, metric or action is identified by kind and name. A relation is
identified by `(name, from, to)`: `name` holds a controlled verb meant to recur
across unrelated pairs. A relation whose triple is new in the closure is a new
relation, even when the verb is already in use elsewhere: that is how
`cafe-demo` adds a `client-configured` `belongs-to` between two locations.
Competency questions have no name and take no part in identity matching.

### Composition

`builds-on` is a list, resolved by manifest `name` into a DAG closure.
`Universe.Closure` walks every entry transitively, dedupes by name and sorts the
result; a playbook is never in its own closure. The diamond (two playbooks
composing one root, a third composing both) is the normal shape.

- **Diamond versus cycle is path-relative.** A root reached through two paths is
  one instance. Only a playbook repeated on the same path is a cycle, so the walk
  tracks `onPath` (per branch) apart from `seen` (permanent, for dedupe).
- **A missing link degrades.** The closure keeps whatever resolved before the
  gap, and every rule treats that as less context; `KBF011` reports the gap.
- **Forks match the closure as a set.** There is no nearest-member tie-break:
  composition forbids two members declaring one identity, so at most one matches
  in the well-formed case.
- **Coverage and compile.** Coverage's attribute-to-entity lookup unions the
  closure, because an install configures the whole resolved ontology, and it
  reports only on leaves (a playbook no other loaded playbook builds on). Compile
  unions every loaded playbook, deduped by identity.

### Layer taxonomy

| `layer` | May build on | Name |
|---|---|---|
| `core` | other `core` playbooks only (`[]` allowed: that is a root) | must match `^core-` |
| `vertical` | `core` or `vertical` playbooks, at least one | must not match `^core-` |

Root-ness is derived from an empty `builds-on`: there is no `root: true` field,
and nothing privileges `core-business` by name over any other core playbook with
an empty `builds-on`.

### Controlled verb vocabulary

There is no fixed seed list, no declaration file and no `kind: verb-vocabulary`:
the vocabulary is derived from relations. `core-business` declares nine verbs
(`contains`, `belongs-to`, `supplies`, `places`, `billed-to`, `employed-by`,
`derived-from`, `supersedes`, `responsible-for`); `core-operations` mints four
(`located-at`, `sells`, `staffed-by`, `works-at`); `core-services` mints none.
RFC discipline applies to anything broader, reuse first: see
`spec/conventions.md`.

A relation's verb passes if either condition holds, and fails `KBF007` otherwise:

1. A closure member declares the verb (ordinary reuse). A playbook composing two
   core playbooks sees both vocabularies; a verb one of them mints stays out of
   reach for anything that does not compose it.
2. The relation's `from` or `to` is an entity the same playbook introduces, since
   a new entity is the unit of new business concept and carries the right to name
   its relationships.

The test is per relation, not per playbook: minting a verb on one pair does not
legalize it on a fully inherited pair in the same playbook. "Introduces" means
genuinely new: redeclaring a composed entity's name, as a fork or a glossary
fragment, grants nothing. The test is static, with no declaration order. A root
passes through condition 2, because it declares every entity it references.

### Playbook architecture

The layering keeps location, shift and order out of what is claimed as universal:
see `project-decisions.md`'s "The unit is a playbook; the folder is playbooks/".

- **`playbooks/core-business`** (`builds-on: []`): the truly universal floor.
  Entities: organization, customer, offering (product or service), transaction
  (the revenue event), employee, supplier, purchase. Metrics: revenue,
  transaction-count, average-transaction-value, gross-margin, labor-cost-ratio,
  purchase-cost, with grain stopping at `[organization, business-date]`. Actions:
  flag-for-review, annotate, request-sync, propose-extension. Competency
  questions: at least one per entity, in business-neutral wording.
- **`playbooks/core-operations`** (`builds-on: [core-business]`): for site-based
  businesses. Adds location and shift. Mints `located-at`, `staffed-by`,
  `works-at` and `sells`; `employed-by` stays `core-business`'s location-free HR
  fact; `belongs-to` and `responsible-for` are reused on new pairs. Metrics:
  average-ticket and location-labor-cost-ratio, net-new names at `[location,
  business-date]`, because grain is not glossary-eligible and a same-name
  redeclaration at another grain would fork.
- **`playbooks/core-services`** (`builds-on: [core-business]`): for
  engagement-based businesses. Adds engagement (synonyms: project, retainer) and
  deliverable. Mints no verbs: `places`, `contains`, `derived-from` and
  `responsible-for` are reused on new pairs. Metrics: effective-rate and
  on-time-delivery-rate. Utilization stays out: it needs capacity data
  `core-business`'s employee does not carry, and adding an attribute to a composed
  entity forks it.
- **`examples/cafe-demo`** (`builds-on: [core-operations]`, Demo Cafe): a glossary
  synonym on `offering` (declared in `core-business`, reached through
  `core-operations`), a `client-configured` `belongs-to` between locations, a
  threshold override on `location-labor-cost-ratio`, net-new `waste-ratio`. Slots
  map to `demopos` and `demobooks`; `crm.customer-*` stays unmapped on purpose.
- **`examples/studio-demo`** (`builds-on: [core-services]`, Demo Studio, a
  fictional single-team marketing agency): a glossary synonym on `engagement`
  ("campaign"), a `client-configured` `belongs-to` between customers
  (parent-account grouping), a threshold override on `on-time-delivery-rate`,
  net-new `repeat-engagement-rate`. Slots map to `demopm` and `demobooks`;
  `purchasing.supplier-*` stays unmapped, a different gap on purpose.
- **`examples/bistro-demo`** (`builds-on: [core-operations, core-services]`, Demo
  Bistro & Events): the diamond and the hybrid in one playbook. A glossary synonym
  on `engagement`, a `client-configured` `located-at` from engagement to location
  (a pair neither core playbook has both entities to declare), a threshold
  override on `core-business`'s `gross-margin`, net-new `event-booking-rate`.
  `delivery.deliverable-*` stays unmapped, a third gap.

The demos make the "one shared multi-vertical core" claim checkable:
`core-business` is the common closure member of visibly different businesses.

## Lint rule set

Rule ids are public API. A human render is one lipgloss table per file. `--format
json` renders `{rules: [{id, file, line, element, message, fix}]}`: a clean run is
`{"rules": []}`, never `null`, and output is not HTML-escaped.

| Rule | Fires on |
|---|---|
| `KBF001` | Unknown field in an ontology or evals document; malformed YAML; a document with no `kind`; an `install/slots.yaml` that is not a bare list |
| `KBF002` | A non-empty closed-vocabulary value outside its allowed set, an unknown `kind` included |
| `KBF003` | Duplicate identity within a playbook, or the same identity declared by two closure members |
| `KBF004` | Entity without identity |
| `KBF005` | Metric without grain or additivity |
| `KBF006` | `relation.from` or `relation.to` not a declared entity |
| `KBF007` | Verb outside the controlled vocabulary |
| `KBF008` | Fork of an element in the composition closure |
| `KBF009` | Any other dangling reference: `metric.grain`, `action.on`, `competency-question.expects` |
| `KBF010` | Empty `Entity.tier`, `Metric.tier`, `Relation.origin`, `Action.approval` or `Manifest.layer` |
| `KBF011` | Manifest missing or invalid: empty file, missing `name`, `version` or `spec`, unresolved `builds-on` entry, cycle |
| `KBF012` | An `install/slots.yaml` row with no matching attribute |
| `KBF013` | `layer` inconsistent with `builds-on` or `name` |

- **`KBF003`, across playbooks.** Two different closure members declaring one
  identity have no correct winner, because composition has no resolution order. It
  is always an error, reported against every file involved, and fires only when a
  playbook composes both.
- **`KBF008`.** A playbook restating an identity found in its closure forks,
  unless every non-glossary field on its copy is zero-valued. Glossary-eligible
  fields: `Entity.synonyms`, `Metric.thresholds`. Repeating another field with an
  identical value still forks. Relation and action have no carve-out: restating
  one forks, changed or not.
- **Override fragments skip completeness.** An element whose identity matches a
  closure member is classified before completeness runs, so `KBF004`, `KBF005`
  and `KBF010` skip it: `KBF008` validates what it sets, and completeness is
  inherited. A new element needs full completeness.
- **`KBF009` resolution.** `grain` entries resolve to entities or `business-date`.
  `on` resolves to entities.
  `expects` resolves by bare name across entities, metrics, actions and relations
  (the verb, not the triple), so a recurring verb still resolves.
- **`KBF011`.** Each unresolved `builds-on` entry gets its own finding, with a fix
  hint to pass that playbook's path. Each playbook whose walk reaches a cycle gets
  its own finding.
- **`KBF012`.** The check runs in one direction: an attribute slot with no row is
  not a lint error in v0.
- **`KBF013`.** It runs only on a valid `layer` and compares layers only for a
  `builds-on` entry that resolves, so a dangling entry is reported once, as
  `KBF011`.

## Key flows

- `kbf schema [--check]`: model structs to JSON Schema to YAML files in `schema/`;
  `--check` diffs against the committed files (the freshness gate).
- `kbf lint <path...>`: load every path into one set keyed by manifest `name`;
  resolve `builds-on` by name against that set, never by filesystem convention,
  then the embedded fallback; parse position-aware; run structural validation,
  then semantic rules; render. Exit 1 on any finding. A root, or a playbook built
  only on embedded core playbooks, needs only its own path.
- `kbf coverage <path...>`: for each leaf playbook, the rows of its own
  `install/slots.yaml` with their entity and source, plus declared and mapped
  totals.
- `kbf compile --to mermaid <path...>`: entities as nodes, relations as labeled
  edges, actions as annotations. Deterministic output (sorted), golden-tested.
- `kbf init <name> [--builds-on a,b] [--layer core|vertical]`: scaffolds the
  manifest, empty `ontology/` and `evals/`, and a template `install/slots.yaml`.
  `--layer` defaults to `vertical`, which requires `--builds-on`. It never
  overwrites an existing directory, because an initialized playbook can hold work
  that nothing can regenerate.
- `kbf playbooks pin [--to <dir>] [--force]`: writes the three embedded core
  playbooks to real directories, default `builds-on/`, so tools that are not `kbf`
  read the full closure from disk. It refuses to overwrite without `--force`,
  naming every conflict up front. `kbf init` does not create `builds-on/`; pin does.
- `kbf skill install [--force]`: writes the `kbf-authoring` skill to
  `.claude/skills/kbf-authoring/`, with the same overwrite guard.
- `kbf docs [name]`: lists embedded doc names, or prints one to stdout. An unknown
  name is an error listing the valid names.
- Embedded resolution: `lint`, `coverage` and `compile` load through
  `lint.LoadWithEmbedded`. A local path always wins; the embedded copy fills only
  a `builds-on` name no path provides, transitively, to a fixed point. Disk and
  embedded share one `fs.FS` loader. `lint` and `coverage` add a human-render
  footer naming what resolved from embedded; `compile` adds none (the output stays
  valid mermaid); `--format json` does not change. Conformance and `internal/lint`
  tests use `lint.Run`, which never falls back.
- Embedded content: `playbooks/core-*`, `skills/kbf-authoring/`, `spec/*.md` and
  `spec/primitives/*.md`, each mirrored with a DO-NOT-EDIT header. `make
  embed-freshness` fails on stale content and on orphaned files. `schema/` and
  `examples/` are not embedded.
- Boundaries: `scripts/boundaries.go` walks the tree and fails on private
  filesystem paths and prices. A self-test plants a violation in a temp tree.

## Edge cases & constraints

- Lint does not panic on arbitrary YAML, and one bad file never aborts the run:
  malformed YAML is `KBF001`, an unknown `kind` is `KBF002`, a missing manifest is
  `KBF011`. The invalid conformance fixtures exercise the loader.
- A missing `evals/` or `install/slots.yaml` is not an error. A comment-only
  `install/slots.yaml`, which is what `kbf init` writes, is zero rows.
- CRLF line endings parse. Finding paths use the host's path separator; a playbook
  resolved from the embedded fallback reports as `embedded:<name>`.
- Everything is deterministic: sorted iteration everywhere (maps never range
  unsorted into output), findings sorted by file, line and rule id.
- Schema scope is exactly two files. `SlotMapping` is reflected into
  `manifest.schema.yaml`'s `$defs` and is not a root schema, so `install/slots.yaml`
  has no direct yaml-language-server validation in v0; the loader's struct still
  validates it.

## Decisions

- Cross-file resolution happens per playbook after full load; no incremental mode
  in v0.
- Mermaid over any richer render: it previews natively in editors and code hosts.
- `project-decisions.md` holds the project-level decisions this capsule implements:
  "Composition (DAG closure) replaces single-parent extends; core-business rename;
  origin/approval field renames", "Playbook rename executed; the core/vertical
  taxonomy is machine-checked", "Batteries-included binary: embedded cores/skill/docs
  with local-override precedence", and "The CLI speaks KBF vocabulary: `kbf vendor`
  renamed to `kbf playbooks pin`".
