---
type: spec-doc
---

# Slot mapping

A slot mapping declares which source system fills one attribute's data. It
is the operational bridge between the ontology (what data means) and an
install (where that data actually comes from).

## Fields

| Field | Required | Shape | Notes |
|---|---|---|---|
| `slot` | yes | dotted lowercase id, `<domain>.<concept>` | Matches the `slot:` value on exactly one attribute somewhere in `ontology/*.yaml`. A slot with no matching attribute fails `KBF012`. |
| `source` | yes | string, may be empty | The name of the system that fills this slot. Empty means unmapped, not invalid: a template row is a valid row. |

Slot mapping rows live in one file, `install/slots.yaml`, as a flat YAML
list, not as `kind:`-discriminated documents like the other primitives:
there is exactly one row per declared slot, and the file's only job is to
be a checklist. Unlike the other primitives, it carries neither `tier` nor
`risk`: `KBF010` does not apply to it.

A playbook's `install/slots.yaml` covers only the slots declared by
attributes in that playbook's own `ontology/`, not its ancestors': a core
playbook (`playbooks/core-operations`, `playbooks/core-services`) templates
only what it adds on top of `core-universal`, exactly as `core-universal`
templates its own. A leaf install (a teaching playbook like
`examples/cafe-demo`, or a real deployment) is different: it covers the
*full resolved chain*, because an install is mapping one business's whole
ontology to its real systems, not documenting what one layer contributed.
`examples/cafe-demo/install/slots.yaml` has all 27 slots across
`core-universal` and `playbooks/core-operations` combined, even though
`cafe-demo` itself declares no new attributes.

## Example

The template form, one row from `playbooks/core-universal/install/slots.yaml`:

```yaml
- slot: catalog.offering-label
  source: ""
```

The same row, filled in by a leaf install, from
`examples/cafe-demo/install/slots.yaml`:

```yaml
- slot: catalog.offering-label
  source: demopos
```

## Common mistakes

- **A slot with no declaring attribute.** Every row in `install/slots.yaml`
  must match a `slot:` value used by some entity's `attributes:` list
  somewhere in the resolved playbook. A typo or a leftover row from a
  deleted attribute fails `KBF012`.
- **An attribute slot with no row.** The reverse gap does not fail the
  linter in v0 (static coverage reporting, not enforcement, is what
  surfaces it: `kbf coverage`), but it means that attribute can never be
  filled. Keep `install/slots.yaml` in lockstep with `ontology/*.yaml` by
  hand until tooling enforces it.
- **Leaving every source empty forever.** An empty `source` is valid syntax
  or a genuine "not yet mapped" state (see the `crm.customer-*` rows in
  `examples/cafe-demo/install/slots.yaml`, left unmapped on purpose because
  Demo Cafe has no CRM). It should never be the final state for a slot a
  business actually has data for.
- **Inventing a wrapper key.** The file is a bare YAML list of `{slot,
  source}` rows, not `slots: [...]` or `mappings: [...]`. A wrapper key
  will not parse the way the loader expects.
