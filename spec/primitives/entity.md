---
type: spec-doc
---

# Entity

An entity is a thing the business has: an organization, a location, an
order. It carries what the thing means, how it is identified, what it is
called in conversation, and what data it has.

## Fields

| Field | Required | Shape | Notes |
|---|---|---|---|
| `kind` | yes | `entity` | Discriminates this document from other primitives. |
| `name` | yes | kebab-case string | Unique within the namespace. What every other element refers to it by. |
| `meaning` | yes (root only) | one sentence | Plain-language definition. Omitted on a glossary-only override fragment, see below. |
| `identity` | yes (root only) | list of snake_case keys | The key(s) that identify one instance. Never empty on a root definition: an entity without an identity key fails `KBF004`. |
| `resolution` | yes (root only) | free text | How cross-source identity resolution works. Opaque in v0: named strategies land later (`spec/versioning.md`). |
| `tier` | yes (root only) | `structural \| glossary \| instance` | Governance tier. Empty fails `KBF010`. |
| `synonyms` | no | `{en: [...], es: [...]}` | Alternate names an agent should recognize. The one field an extending package may set without forking the entity. |
| `attributes` | yes (root only) | list of `{name, type, slot}` | Typed fields, each pointing at a slot filled by an install. `type` is a free-form string in v0 (common values: `text`, `number`, `boolean`, `date`, `timestamp`, `currency`); it is not yet a closed vocabulary. |
| `states` | no | list of kebab-case strings | The lifecycle, only where a real one exists. Many entities have none. |

"Root only" means these fields are required when an entity is being defined
for the first time. They are the opposite of the one field, `synonyms`,
that an extending package is allowed to touch: see "Common mistakes" below.

## Example

Copied from `packages/universal-core/ontology/organization.yaml`:

```yaml
kind: entity
name: organization
meaning: >-
  The legal or commercial entity that owns and operates the business.
identity: [organization_id]
resolution: identity-mapping
tier: structural
synonyms: {en: [company, business], es: [empresa, negocio]}
attributes:
  - {name: legal-name, type: text, slot: core.organization-legal-name}
  - {name: display-name, type: text, slot: core.organization-display-name}
  - {name: tax-id, type: text, slot: core.organization-tax-id}
states: [active, inactive]
```

## Common mistakes

- **Unknown field.** A typo like `synonym:` instead of `synonyms:` fails
  `KBF001`, naming the file, line, and the field the linter expected.
- **Missing identity keys.** `identity: []` or an omitted `identity` fails
  `KBF004`. Every root entity needs at least one.
- **Missing governance tier.** An empty `tier` fails `KBF010`.
- **Forking instead of extending.** An extending package that wants to add
  a synonym should declare a *fragment*, not a copy of the whole entity.
  Repeating `meaning`, `identity`, `resolution`, `tier`, `attributes`, or
  `states`, even with values identical to the parent, is a fork and fails
  `KBF008`: the linter has no way to tell "unchanged" from "silently
  different" without diffing every field, so it does not try. Leave every
  field except `synonyms` out entirely:

  ```yaml
  # Only the glossary-eligible field is set. See
  # examples/cafe-demo/ontology/entities.yaml for the real file.
  kind: entity
  name: product
  synonyms: {en: [item, sku, menu-item], es: [producto, articulo]}
  ```

- **Assuming `type` on an attribute is validated.** It is a free-form string
  in v0 (see `spec/versioning.md`); picking a value outside the common set
  above will not fail `kbf lint`, but it will confuse the next author. Stay
  inside the common set unless you have a real reason not to.
