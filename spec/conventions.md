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
| Identity keys, `join` keys | snake_case | `product_id`, `parent_location_id` |
| Slot ids | dotted lowercase, `<domain>.<concept>` | `catalog.product-label` |

Identity and join keys read like source-system column names on purpose:
they are usually one. Slot domains (`core`, `sales`, `catalog`, `crm`,
`hr`, `purchasing` in `universal-core`) are a convention, not a closed
list; an extending package may introduce its own domain prefix.

## The controlled verb vocabulary

Relation names come from a small, shared vocabulary, not from whatever
reads best in the moment. The v0 seed, declared in
`packages/universal-core/ontology/relations.yaml`:

| Verb | Meaning |
|---|---|
| `contains` | The `from` entity is composed of, or line-items, the `to` entity. |
| `belongs-to` | The `from` entity is owned or grouped by the `to` entity. |
| `located-at` | The `from` entity happened, or is situated, at the `to` location. |
| `works-at` | The `from` employee is assigned to the `to` location. |
| `supplies` | The `from` supplier provides the `to` entity. |
| `sells` | The `from` location offers the `to` product. |
| `places` | The `from` entity initiates the `to` transaction. |
| `staffed-by` | The `from` shift is worked by the `to` employee. |
| `billed-to` | The `from` transaction is charged to the `to` organization. |
| `derived-from` | The `from` entity originates from, or corrects, the `to` entity. |
| `supersedes` | The `from` entity replaces the `to` entity of the same kind. |
| `responsible-for` | The `from` employee is accountable for the `to` entity, by human assignment rather than a source feed. |

**The vocabulary is exactly what `universal-core` declares, nothing more.**
There is no separate list to keep in sync: the controlled vocabulary *is*
the set of distinct `Relation.name` values already present in a package's
extends-root. Adding a twelfth, thirteenth, or later verb means adding a
relation that uses it to `universal-core` itself, through an RFC (see
`rfcs/README.md`), which then unlocks that verb for every package that
extends it. An extending package cannot introduce a new verb on its own,
by design: the vocabulary stays small only if it is genuinely shared.

A verb is meant to recur across unrelated entity pairs; a small vocabulary
covering dozens of relations cannot work any other way. Relation identity
for uniqueness and fork-detection is the triple `(name, from, to)`, not
`name` alone: see `spec/primitives/relation.md`.

## Synonyms

Every entity may declare `synonyms: {en: [...], es: [...]}`: alternate
words an agent should treat as referring to the same entity ("menu item"
for `product`, "ticket" for `order"). Synonyms exist so that a business's
own vocabulary reaches the ontology without forking it: adding one is a
glossary edit (see below), always available to an extending package even
when nothing else about the entity may change.

## Governance tiers

"Tier" is not one vocabulary; it is kind-specific, and each kind's version
answers a different question.

| Kind | Field | Values | Question it answers |
|---|---|---|---|
| Entity, Metric | `tier` | `structural \| glossary \| instance` | Who owns this element's definition? |
| Relation | `tier` | `source-synced \| client-configured` | Does a source system emit this relation, or does a person configure it? |
| Action | `risk` | `auto \| confirm` | May an agent execute this unattended? |
| Competency question, Slot mapping | none | (n/a) | Neither applies: these are operational elements, not governed content. |

`structural` is the default for anything `universal-core` defines: the
shared, non-negotiable shape of the business. `glossary` and `instance`
exist in the vocabulary for content this spec does not populate in v0 (see
`spec/versioning.md`); every element in `packages/universal-core` and
`examples/cafe-demo` is `structural`.

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
