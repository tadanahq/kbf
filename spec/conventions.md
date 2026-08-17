---
type: spec-doc
---

# Conventions

Naming, the controlled verb vocabulary, synonyms, and the governance tiers
that decide what a composing playbook may edit without forking. These
conventions are load-bearing: the linter's semantic rules assume them.

## Naming

| What | Case | Example |
|---|---|---|
| Element `name` (entity, relation, metric, action, playbook) | kebab-case | `labor-cost-ratio`, `flag-for-review` |
| Attribute `name` | kebab-case | `unit-price`, `hired-at` |
| Identity keys, `join` keys | snake_case | `offering_id`, `parent_location_id` |
| Slot ids | dotted lowercase, `<domain>.<concept>` | `catalog.offering-label` |

Identity and join keys read like source-system column names on purpose:
they are usually one. Slot domains (`core`, `sales`, `catalog`, `crm`,
`hr`, `purchasing` in `core-business`; `delivery` in `core-services`)
are a convention, not a closed list; a core playbook or a composing
playbook may introduce its own domain prefix.

Playbook names carry one more rule, on top of kebab-case: a `layer: core`
playbook ("core playbooks", the foundation every vertical builds on:
`spec/primitives/namespace.md`'s `layer` field) is named with a `core-`
prefix; a `layer: vertical` playbook is not. `kbf lint` enforces this
(`KBF013`), not just documents it. Sorted alphabetically, this repo's six
playbooks read `bistro-demo`, `cafe-demo`, `core-business`,
`core-operations`, `core-services`, `studio-demo`: the three core
playbooks sort together as one contiguous block, the three vertical
leaves bookending them, exactly the point of the prefix: scan a sorted
listing and the foundation layer is the block that shares a name, not
six names you have to already know the architecture to sort by hand.

## The controlled verb vocabulary

Relation names come from a small, shared vocabulary, not from whatever
reads best in the moment. **The vocabulary a playbook sees is the union of
every `Relation.name` already declared anywhere in its composition
closure**, not a single fixed list: `core-business` seeds nine verbs
available to everything; a core playbook may mint its own on top,
available to everything that composes that core playbook (and only
those), the same way `core-operations` mints four more for anything that
builds on it.

`core-business`'s nine (`playbooks/core-business/ontology/relations.yaml`):

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

`core-operations` mints four more, available to anything that builds on it
(`playbooks/core-operations/ontology/relations.yaml`), all needing
`location` or `shift`, which is exactly why they aren't universal:
`located-at`, `works-at`, `staffed-by`, `sells`. `core-services` mints
none: every relation it adds reuses one of `core-business`'s nine on a
new pair (`places` for customer-to-engagement, `contains` for
engagement-to-deliverable, and so on), the RFC-reuse-first principle
below taken as far as it goes. `examples/bistro-demo` composes both
`core-operations` and `core-services` directly, so it sees both of their
vocabularies at once plus `core-business`'s: exactly what lets it mint
`located-at: engagement -> location`, a pairing neither core playbook
alone has both entities to declare.

Adding a genuinely new verb means adding a relation that uses it to the
playbook meant to own it, through an RFC (see `rfcs/README.md`), which
then unlocks that verb for everything that composes that playbook, never
retroactively for an unrelated closure that happens to share a distant
root. Prefer reusing an existing verb on a new entity pair before
proposing a new one: a verb is meant to recur across unrelated pairs, and
a small vocabulary covering dozens of relations cannot work any other way.
Relation identity for uniqueness and fork-detection is the triple `(name,
from, to)`, not `name` alone: see `spec/primitives/relation.md`.

## Synonyms

Every entity may declare `synonyms: {en: [...], es: [...]}`: alternate
words an agent should treat as referring to the same entity ("menu item"
for `offering`, "invoice line" for `transaction`). Synonyms exist so that
a business's own vocabulary reaches the ontology without forking it:
adding one is a glossary edit (see below), always available to a
composing playbook even when nothing else about the entity may change,
and layerable across more than one level: `core-operations` could add
"product" to `offering`'s synonyms, and `examples/cafe-demo` (composing
`core-operations`) could add "menu item" on top of that, without either
layer touching what the one before it set.

## Governance tiers

"Tier" is not one vocabulary; it is kind-specific, and each kind's version
answers a different question.

| Kind | Field | Values | Question it answers |
|---|---|---|---|
| Entity, Metric | `tier` | `structural \| glossary \| instance` | Who owns this element's definition? |
| Relation | `origin` | `source-synced \| client-configured` | Does a source system emit this relation, or does a person configure it? |
| Action | `approval` | `automatic \| required` | May an agent execute this unattended? |
| Competency question, Slot mapping | none | (n/a) | Neither applies: these are operational elements, not governed content. |

`structural` is the default for anything a core playbook defines: the
shared, non-negotiable shape of the business at that layer. `glossary` and
`instance` exist in the vocabulary for content this spec does not populate
in v0 (see `spec/versioning.md`); every entity and metric across
`playbooks/core-business`, `playbooks/core-operations`,
`playbooks/core-services`, and all three teaching examples is `structural`.

The tier that matters most for authoring is narrower than the table above:
which *fields*, not which *elements*, a composing playbook may set without
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
takes, and "Composition rules" in `playbook-format.md` for how this fits
the larger composition-not-fork rule.
