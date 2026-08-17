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
(what every business shares) sits in a root core playbook; other core
playbooks build on it for what a whole shape of business shares, and the
vertical playbooks that compose them add what is actually specific to one
business: see "Where the primitives live for real" below.

A contract only works if it can be checked. Every KBF playbook is validated
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
ontology itself. Every action declares whether it may run automatically or
needs a human to approve it first, so a runtime can decide what needs a
human in the loop.

Operational elements make the contract installable and provable. A slot
mapping declares which source system fills a given attribute. A competency
question is a plain-language question paired with the elements an answer
must use: the acceptance test for the whole ontology. A namespace is the
playbook boundary itself (`manifest.yaml`): what a playbook is called, what
it builds on, and what version of the spec it targets.

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
5. [`playbook-format.md`](playbook-format.md): how the primitives above are
   laid out on disk into a playbook, and how composition resolves.
6. [`conventions.md`](conventions.md): naming, the controlled verb
   vocabulary, and the governance tiers that decide what a composing
   playbook may edit.
7. [`versioning.md`](versioning.md): how the spec itself versions, and what
   v0 leaves opaque on purpose.
8. [`onboarding.md`](onboarding.md): the methodology for raising a new
   ontology end to end, everything above applied in the order it actually
   gets authored.

## Where the primitives live for real

Reading the primitives in the abstract only goes so far. Six playbooks in
this repository are the worked answer key:

- [`playbooks/core-business`](../playbooks/core-business) (`builds-on:
  []`): the truly universal floor, a root by derivation, not declaration.
  Organization, customer, offering, transaction, employee, supplier,
  purchase: nothing here assumes a physical location or a scoped
  engagement, because not every business has either.
- [`playbooks/core-operations`](../playbooks/core-operations) and
  [`playbooks/core-services`](../playbooks/core-services) (both
  `builds-on: [core-business]`): one core playbook per shape of business,
  siblings, not a hierarchy. `core-operations` adds location and shift, for
  businesses that operate from a site. `core-services` adds engagement and
  deliverable, for businesses that sell scoped work instead. Neither
  builds on the other; both build on `core-business` directly.
- [`examples/cafe-demo`](../examples/cafe-demo) (`builds-on:
  [core-operations]`) and [`examples/studio-demo`](../examples/studio-demo)
  (`builds-on: [core-services]`): a fictional cafe and a fictional
  marketing studio, one worked example per core playbook. Each exercises
  every composition mechanic the spec allows on a single path: a glossary
  synonym, a client-configured relation, a threshold override, and a
  net-new vertical metric.
- [`examples/bistro-demo`](../examples/bistro-demo) (`builds-on:
  [core-operations, core-services]`): a fictional cafe that also runs
  events, composing both core playbooks at once. It is the diamond
  (both paths reach `core-business` and dedup to one instance) and the
  hybrid case (it sees both `core-operations`'s and `core-services`'s
  vocabulary, and mints a relation, `located-at: engagement -> location`,
  that needs one entity from each side) in a single worked example, not
  just asserted.

Every YAML example in this spec is copied from one of these six playbooks
verbatim, so what you read here and what `kbf lint` enforces never drift
apart.

## For DDD practitioners

The vocabulary above maps onto strategic domain-driven design, if that
lens is useful. The ontology is the published language a business's data
and its agents share: not a diagram a team agrees to keep updated, but a
contract `kbf lint` enforces, so an agent deciding what to do next reads
the same definitions a human would. Each source system feeding a slot
mapping is its own bounded context; the slot mapping is the
anticorruption layer, translating that system's field names into the
ontology's identity keys without leaking them upward. A playbook is a
bounded context's ubiquitous language made literal, with enforcement: the
lint rules are what would otherwise live in a wiki nobody keeps current.
Composition is context integration, and where two contexts define the
same thing differently, the linter surfaces that as a visible `KBF003`
collision rather than silently picking a winner. What DDD calls subdomain
triage, sorting a capability into generic, supporting, or core domain, is
not a primitive here; it happens before any of these fields get filled
in, and belongs to `onboarding.md`.
