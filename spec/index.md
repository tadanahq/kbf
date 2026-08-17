---
type: spec-doc
---

# KBF Ontology Spec

Semantic layers describe data for dashboards. The KBF Ontology describes a
business for agents.

## What it is

The KBF Ontology is a versioned, YAML-authored contract that declares what a
business's data means and what an agent may do about it. It is not a schema
for a database and not a BI semantic layer: those describe tables and
columns. The KBF Ontology describes the business itself, in vocabulary an
agent can reason with and a human can read: organizations, customers,
transactions, the metrics computed over them, and the actions an agent may
take when it notices something. A small, truly universal set of those
(what every business shares) sits under layered base packages (what a
whole vertical shares) and the businesses that extend them: see "Where the
primitives live for real" below.

A contract only works if it can be checked. Every KBF package is validated
by the `kbf` CLI (`kbf lint`) against a published meta-schema, so a broken
ontology fails fast with a file, a line, and a fix, the same way a broken
build does.

## The seven primitives, in three groups

The spec has seven primitives. They group into three families by what they
carry:

| Group | Primitives | Answers |
|---|---|---|
| Semantic elements | entity, relation, metric | What does the data mean? |
| Action elements | action | What may an agent do about it? |
| Operational elements | slot mapping, competency question, namespace | How does this install, and how is it proven correct? |

Semantic elements are the nouns and their arithmetic: an `entity` is a thing
the business has (a customer, an offering), a `relation` is a typed
connection between two entities (a customer *places* a transaction), and a
`metric` is a number computed over them (revenue, gross margin), always with
a grain and an additivity so it can be rolled up or joined without lying.

Action elements are the verbs an agent is allowed to perform: flagging a
record for review, requesting a fresh sync, proposing an extension to the
ontology itself. Every action declares a risk tier, so a runtime can decide
what needs a human in the loop.

Operational elements make the contract installable and provable. A slot
mapping declares which source system fills a given attribute. A competency
question is a plain-language question paired with the elements an answer
must use: the acceptance test for the whole ontology. A namespace is the
package boundary itself (`manifest.yaml`): what a package is called, what it
extends, and what version of the spec it targets.

## Reading order

1. This document, for the shape of the whole spec.
2. [`primitives/entity.md`](primitives/entity.md),
   [`primitives/relation.md`](primitives/relation.md),
   [`primitives/metric.md`](primitives/metric.md): the semantic elements.
3. [`primitives/action.md`](primitives/action.md): the action element.
4. [`primitives/slot-mapping.md`](primitives/slot-mapping.md),
   [`primitives/competency-question.md`](primitives/competency-question.md),
   [`primitives/namespace.md`](primitives/namespace.md): the operational
   elements.
5. [`package-format.md`](package-format.md): how the primitives above are
   laid out on disk into a package.
6. [`conventions.md`](conventions.md): naming, the controlled verb
   vocabulary, and the governance tiers that decide what an extending
   package may edit.
7. [`versioning.md`](versioning.md): how the spec itself versions, and what
   v0 leaves opaque on purpose.

## Where the primitives live for real

Reading the primitives in the abstract only goes so far. Five packages in
this repository, in three layers, are the worked answer key:

- [`packages/universal-core`](../packages/universal-core) (`extends:
  null`): the truly universal floor, every primitive in its root form.
  Organization, customer, offering, transaction, employee, supplier,
  purchase: nothing here assumes a physical location or a scoped
  engagement, because not every business has either.
- [`packages/operations-core`](../packages/operations-core) and
  [`packages/services-core`](../packages/services-core) (both `extends:
  universal-core`): the base-package layer, one per shape of business.
  `operations-core` adds location and shift, for businesses that operate
  from a site. `services-core` adds engagement and deliverable, for
  businesses that sell scoped work instead. Neither extends the other;
  both extend `universal-core` directly.
- [`examples/cafe-demo`](../examples/cafe-demo) (`extends:
  operations-core`) and [`examples/studio-demo`](../examples/studio-demo)
  (`extends: services-core`): a fictional cafe and a fictional marketing
  studio, one worked example per base chain. Each exercises every
  extension mechanic the spec allows: a glossary synonym, a
  client-configured relation, a threshold override, and a net-new
  vertical metric. Two examples, not one, so "one shared multi-vertical
  core" is something you can check against two visibly different
  businesses, not just assert.

Every YAML example in this spec is copied from one of these five packages
verbatim, so what you read here and what `kbf lint` enforces never drift
apart.
