---
type: spec-doc
---

# Relation

A relation is a typed connection between two entities: a customer *places*
a transaction, an employee *works at* a location. The verb is never
invented ad hoc; it comes from the controlled vocabulary in
`conventions.md`.

## Fields

| Field | Required | Shape | Notes |
|---|---|---|---|
| `kind` | yes | `relation` | Discriminates this document from other primitives. |
| `name` | yes | a verb from the controlled vocabulary | Not unique by itself: see "Relation identity" below. |
| `from` | yes | entity name | Must resolve to a declared entity, or fails `KBF006`. |
| `to` | yes | entity name | Must resolve to a declared entity, or fails `KBF006`. |
| `cardinality` | yes | `one-to-one \| one-to-many \| many-to-one \| many-to-many` | Read as "how many `from` per `to`". A bad value fails `KBF002`. |
| `join` | yes | list of snake_case keys | The keys used to join `from` to `to`, ordered `[from_key, to_key]`. For a self-relation, the second key is role-qualified (`parent_location_id`, not a second `location_id`). |
| `origin` | yes | `source-synced \| client-configured` | A distinct axis from entity/metric tier: whether a source system emits this relation, or a person configures it by hand. Empty fails `KBF010`. |
| `temporal` | yes | boolean | Whether the relation's validity should be tracked over time (an employee's location assignment changes; a shift's staffing does not). |

## Relation identity

A relation's identity for uniqueness and fork-detection purposes is the
triple `(name, from, to)`, not `name` alone. The controlled vocabulary is
deliberately small (10 to 20 verbs, see `conventions.md`), meant to recur
across many unrelated entity pairs, both within one playbook and across a
playbook's composition closure: `playbooks/core-business` legitimately
declares `supplies` twice (`supplier` to `offering`, and `supplier` to
`purchase`), and `contains` shows up again, on a different pair, in
`playbooks/core-services` (`engagement` to `deliverable`) without
conflicting with core-business's own `transaction` to `offering` use of
the same verb. Two relations sharing a verb are only in conflict if they
also share the same `from` and `to`.

## Example

Copied from `playbooks/core-business/ontology/relations.yaml`:

```yaml
kind: relation
name: employed-by
from: employee
to: organization
cardinality: many-to-one
join: [employee_id, organization_id]
origin: source-synced
temporal: true
```

## Common mistakes

- **Verb outside the controlled vocabulary.** `name: assigned-to` fails
  `KBF007` if `assigned-to` is not a verb already declared somewhere in
  the playbook's composition closure (the playbooks it builds on, not the
  whole spec). Propose new verbs through `rfcs/`, not by using them first
  and asking later; prefer reusing an existing verb on a new pair before
  proposing one (see `playbooks/core-services`, which mints none of its
  own).
- **Undeclared endpoint.** A typo in `from` or `to` (`transactionn`,
  `custmer`) fails `KBF006`, since it does not resolve to any declared
  entity.
- **Bad cardinality value.** Anything outside `one-to-one`, `one-to-many`,
  `many-to-one`, `many-to-many` fails `KBF002`. `many-to-one` exists
  because child-to-parent is the most common declared direction
  (`employed-by`, `belongs-to`): write the relation in the direction the
  business says it, not inverted to avoid the value.
- **Redeclaring a relation to change one field.** Unlike entity and metric,
  relation has no glossary carve-out: if a `(name, from, to)` triple already
  exists anywhere in the playbook's composition closure, restating it for
  any reason, even to change only `origin` or `temporal`, is a fork and
  fails `KBF008`, wherever in the closure the match is found. A different
  pairing (same verb, different `from`/`to`) is a new relation, not a
  redeclaration, and is exactly how a composing playbook is expected to
  add relations: see the client-configured `belongs-to` between two
  locations in `examples/cafe-demo/ontology/relations.yaml`.
