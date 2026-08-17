---
type: spec-doc
---

# Conventions

Naming, the controlled verb vocabulary, synonyms, and the governance tiers
that decide what an extending package may edit without forking. These
conventions are load-bearing: the linter's semantic rules assume them.

## Naming

| What | Case | Example |
|---|---|---|
| Element `name` (entity, relation, metric, action, package) | kebab-case | `labor-cost-ratio`, `flag-for-review` |
| Attribute `name` | kebab-case | `unit-price`, `hired-at` |
| Identity keys, `join` keys | snake_case | `offering_id`, `parent_location_id` |
| Slot ids | dotted lowercase, `<domain>.<concept>` | `catalog.offering-label` |

Identity and join keys read like source-system column names on purpose:
they are usually one. Slot domains (`core`, `sales`, `catalog`, `crm`,
`hr`, `purchasing` in `universal-core`; `delivery` in `services-core`)
are a convention, not a closed list; a base package or an extending
package may introduce its own domain prefix.

## The controlled verb vocabulary

Relation names come from a small, shared vocabulary, not from whatever
reads best in the moment. **The vocabulary a package sees is the union of
every `Relation.name` already declared anywhere in its extends chain**,
not a single fixed list: `universal-core` seeds nine verbs available to
everything; a base package may mint its own on top, available to that
base package's own descendants (and only those), the same way
`operations-core` mints four more for anything that extends it.

`universal-core`'s nine (`packages/universal-core/ontology/relations.yaml`):

| Verb | Meaning |
|---|---|
| `contains` | The `from` entity is composed of, or line-items, the `to` entity. |
| `places` | The `from` entity initiates the `to` transaction (or engagement). |
| `supplies` | The `from` supplier provides the `to` entity. |
| `billed-to` | The `from` transaction is charged to the `to` organization. |
| `employed-by` | The `from` employee is employed by the `to` organization. |
| `belongs-to` | The `from` entity is owned or grouped by the `to` entity. |
| `derived-from` | The `from` entity originates from, or corrects, the `to` entity. |
| `supersedes` | The `from` entity replaces the `to` entity of the same kind. |
| `responsible-for` | The `from` employee is accountable for the `to` entity, by human assignment rather than a source feed. |

`operations-core` mints four more, available to anything that extends it
(`packages/operations-core/ontology/relations.yaml`), all needing
`location` or `shift`, which is exactly why they aren't universal:
`located-at`, `works-at`, `staffed-by`, `sells`. `services-core` mints
none: every relation it adds reuses one of `universal-core`'s nine on a
new pair (`places` for customer-to-engagement, `contains` for
engagement-to-deliverable, and so on), the RFC-reuse-first principle
below taken as far as it goes.

Adding a genuinely new verb means adding a relation that uses it to the
package meant to own it, through an RFC (see `rfcs/README.md`), which
then unlocks that verb for that package's own descendants, never
retroactively for an unrelated chain that happens to share a distant
ancestor. Prefer reusing an existing verb on a new entity pair before
proposing a new one: a verb is meant to recur across unrelated pairs, and
a small vocabulary covering dozens of relations cannot work any other way.
Relation identity for uniqueness and fork-detection is the triple `(name,
from, to)`, not `name` alone: see `spec/primitives/relation.md`.

## Synonyms

Every entity may declare `synonyms: {en: [...], es: [...]}`: alternate
words an agent should treat as referring to the same entity ("menu item"
for `offering`, "invoice line" for `transaction`). Synonyms exist so that
a business's own vocabulary reaches the ontology without forking it:
adding one is a glossary edit (see below), always available to an
extending package even when nothing else about the entity may change,
and layerable across more than one hop: `operations-core` could add
"product" to `offering`'s synonyms, and `examples/cafe-demo` (extending
`operations-core`) could add "menu item" on top of that, without either
layer touching what the one before it set.

## Governance tiers

"Tier" is not one vocabulary; it is kind-specific, and each kind's version
answers a different question.

| Kind | Field | Values | Question it answers |
|---|---|---|---|
| Entity, Metric | `tier` | `structural \| glossary \| instance` | Who owns this element's definition? |
| Relation | `tier` | `source-synced \| client-configured` | Does a source system emit this relation, or does a person configure it? |
| Action | `risk` | `auto \| confirm` | May an agent execute this unattended? |
| Competency question, Slot mapping | none | (n/a) | Neither applies: these are operational elements, not governed content. |

`structural` is the default for anything a base package defines: the
shared, non-negotiable shape of the business at that layer. `glossary` and
`instance` exist in the vocabulary for content this spec does not populate
in v0 (see `spec/versioning.md`); every entity and metric across
`packages/universal-core`, `packages/operations-core`,
`packages/services-core`, and both teaching examples is `structural`.

The tier that matters most for authoring is narrower than the table above:
which *fields*, not which *elements*, an extending package may set without
forking. In v0, exactly two:

| Kind | Glossary-eligible field |
|---|---|
| Entity | `synonyms` |
| Metric | `thresholds` |

Relation and action have no glossary-eligible field at all: any
redeclaration of an existing relation or action, for any reason, is a
fork. This is stricter than entity and metric on purpose. An entity's
meaning does not change when its nickname does, and a metric's formula
does not change when its warning threshold does, but a relation or an
action *is* its shape: there is no edit to one that is not really a
redefinition. See "Common mistakes" in `spec/primitives/entity.md` and
`spec/primitives/metric.md` for the exact fragment shape a glossary edit
takes, and "Extension rules" in `package-format.md` for how this fits the
larger extension-not-fork rule.
