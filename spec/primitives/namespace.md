---
type: spec-doc
---

# Namespace

A namespace is a package's identity and its place in the extension chain,
declared once per package in `manifest.yaml`. Every other primitive in the
package belongs to the namespace `manifest.yaml` names.

## Fields

| Field | Required | Shape | Notes |
|---|---|---|---|
| `name` | yes | kebab-case string | The package's namespace. What an extending package's `extends` refers to. |
| `version` | yes | semver string | The package's own version, independent of the spec version it targets. |
| `spec` | yes | `v0` (v0 scope) | Which version of the KBF Ontology Spec this package targets. See `spec/versioning.md`. |
| `extends` | yes | package name, or `null` | `null` only for the root package (`universal-core`). Every other package names its parent. An unresolved `extends` fails `KBF011`. |

`manifest.yaml` is a single flat document, not a `kind:`-discriminated one
like entity, relation, metric, action, or competency-question: a package
has exactly one namespace, so there is nothing to discriminate.

## Example

The root package, `packages/universal-core/manifest.yaml`:

```yaml
name: universal-core
version: 0.1.0
spec: v0
extends: null
```

An extending package, `examples/cafe-demo/manifest.yaml`:

```yaml
name: cafe-demo
version: 0.1.0
spec: v0
extends: universal-core
```

## Common mistakes

- **Missing or unresolvable `extends`.** `kbf lint` loads every package
  path passed on the command line into one set, keyed by `name`, and
  resolves each package's `extends` against that set. A package whose
  parent was not also passed on the command line fails `KBF011` with a fix
  hint to pass the parent's path too; it is not a filesystem lookup against
  this repo's own `packages/` folder.
- **A non-null `extends` on the root package.** `universal-core` is the
  one package with no parent. Giving it a real `extends` value creates a
  cycle the linter has no v0 story for.
- **Extending an extension.** v0 supports one level of `extends`:
  `universal-core` plus one child. A package whose parent itself has a
  non-null `extends` is a deeper chain than v0 resolves; that is a
  spec-versioning question, not something to work around in a single
  package (see `spec/versioning.md`).
- **Reusing a `name` across packages.** Namespace collision breaks
  extends-resolution silently (two packages both claiming to be
  `universal-core`, say). Package names are a flat space; keep them
  distinct.
