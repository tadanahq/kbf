---
type: spec-doc
---

# The kbf CLI

Four commands, three altitudes. `lint` and `compile` serve the author,
`coverage` serves the install, `schema` serves the format itself. If you
are authoring or installing playbooks you will use the first three; you
may never run `schema`, and that is by design.

| Command | Question it answers | Cadence |
|---|---|---|
| `kbf lint` | Is this ontology valid? | Every authoring loop |
| `kbf coverage` | Is this ontology connected to real sources? | Every install, ongoing |
| `kbf compile` | What can be made from this ontology? | Whenever an artifact is needed |
| `kbf schema` | What is a valid ontology allowed to look like? | Only when the spec itself changes |

## Passing playbooks: the closure rule

Every command takes one or more playbook directories. `builds-on` is
resolved by manifest `name` across exactly the paths you pass, never by
filesystem convention: a playbook that builds on `core-operations` needs
`playbooks/core-operations` (and everything that playbook builds on, all
the way down) on the same command line. Passing an incomplete closure is
not an error by itself; the linter reports the missing pieces and every
finding that cascades from them, so the fix is always visible.

## kbf lint

```sh
kbf lint examples/bistro-demo playbooks/core-operations playbooks/core-services playbooks/core-business
```

Validates structure and semantics: the full rule set (`KBF001` to
`KBF013`) with every finding carrying file, line, rule id, and a fix
hint. Exit code 0 on a clean run, 1 if any rule fires. Use it as the
inner loop while authoring (edit, lint, fix, repeat) and as the CI gate
for any repository that holds playbooks.

### The JSON interface (for agents)

`kbf lint --format json` emits the stable machine contract:

```json
{
  "rules": [
    {
      "id": "KBF004",
      "file": "playbooks/core-business/ontology/transaction.yaml",
      "line": 1,
      "element": "transaction",
      "message": "entity has no identity key(s)",
      "fix": "add identity: [<key>, ...]"
    }
  ]
}
```

Guarantees an authoring agent can rely on: the shape is
`{rules: [{id, file, line, element, message, fix}]}` and does not change
casually; a clean run renders `{"rules": []}`, never `null`, so no
zero-findings special case is needed; output is not HTML-escaped, so fix
hints containing `<placeholder>` syntax arrive literally. The intended
agent loop: run lint, parse `rules`, apply the `fix` hints, run again,
stop at an empty array.

## kbf coverage

```sh
kbf coverage examples/cafe-demo playbooks/core-operations playbooks/core-business
```

Reports `install/slots.yaml` completeness for the leaf playbook: every
declared slot, its mapped source or its absence, and the totals.
Playbooks passed only as composition context are not reported on (their
slot files are templates by definition). Use it:

- **As the scoping agenda**: right after the entity interview, the
  unmapped list is, row by row, the integration conversation to have with
  the owner ("who has this data, and if nobody does: skip or buy?").
- **As the install's progress number**: mapped-over-declared is the one
  figure an owner understands without explanation, week over week.
- **As the blast-radius report**: remove a retired source system's
  mappings and re-run; the output is the exact list of attributes that
  just went dark, before anything downstream fails silently.

`spec/onboarding.md` step 7 shows where this sits in the raising flow.

## kbf compile

```sh
kbf compile --to mermaid examples/cafe-demo playbooks/core-operations playbooks/core-business > map.mmd
```

Treats the ontology as source and emits a derived artifact per `--to`
target. v0 ships `mermaid`: the ontology map (entities as nodes,
relations as labeled edges, actions as annotations), deterministic output
that renders anywhere Mermaid does, including GitHub. Use it for the
owner sign-off walk (`spec/onboarding.md` step 8) and for embedding the
current map in documentation. Further targets are the designed growth
path for this command; each new target is an emitter, not a new command.

## kbf schema

```sh
kbf schema            # regenerate schema/ from the canonical model
kbf schema --check    # exit non-zero if the committed files are stale
```

Regenerates the published meta-schemas (`schema/ontology.schema.yaml`,
`schema/manifest.schema.yaml`) from the reference implementation's
canonical model. This is the format's supply chain, not an authoring
tool:

- **CI drift gate**: `--check` fails any change to the model that forgot
  to regenerate, so the published schema can never silently disagree with
  what the linter enforces.
- **Editor experience**: the schemas it emits are what
  `yaml-language-server` serves as autocomplete, hover docs, and inline
  validation in any editor configured against this repository.
- **Third-party ground truth**: an independent implementation targets the
  published schemas plus `conformance/`, never this repository's source
  code; `--check` is the guarantee those files are current.

If you author playbooks and never touch the spec, you never run this.

## Reading order

New to the toolchain: `README.md`'s five-minute tour first, then
`spec/onboarding.md` for the full raising flow, then this page as the
reference you return to.
