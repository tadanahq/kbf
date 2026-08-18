---
type: spec-doc
---

<!-- DO NOT EDIT: generated from spec/cli.md by scripts/embedsync. Edit the source, then run `make embed-sync`. -->

# The kbf CLI

Eight commands. `kbf init`, `kbf lint`, and `kbf compile` serve the
author; `kbf coverage` serves the install; `kbf schema` serves the format
itself; `kbf playbooks pin`, `kbf skill install`, and `kbf docs` are what make
the other five usable with nothing but the binary, no clone of this
repository required for any of it (see "Embedded content" below). If you
are authoring or installing playbooks you will use `init`, `lint`,
`coverage`, and `compile`; you may never run `schema`, and that is by
design.

| Command | Question it answers | Cadence |
|---|---|---|
| `kbf init` | How do I start a new playbook? | Once, at the beginning |
| `kbf lint` | Is this ontology valid? | Every authoring loop |
| `kbf coverage` | Is this ontology connected to real sources? | Every install, ongoing |
| `kbf compile` | What can be made from this ontology? | Whenever an artifact is needed |
| `kbf schema` | What is a valid ontology allowed to look like? | Only when the spec itself changes |
| `kbf playbooks pin` | Where can I see the embedded cores for real? | Rarely, on demand |
| `kbf skill install` | How does my agent learn this workflow? | Once per project |
| `kbf docs` | Where's the spec, without a clone? | As needed |

## Passing playbooks: the closure rule

Every command takes one or more playbook directories. `builds-on` is
resolved by manifest `name` across exactly the paths you pass, plus kbf's
embedded core playbooks as a fallback for a name none of your paths
provide (see "Embedded content"): a playbook that builds on
`core-operations` needs nothing extra on the command line for that name
specifically, since `core-operations` is embedded, but a playbook that
builds on some other, non-core playbook needs that playbook's path
passed explicitly, the same as always. Passing an incomplete closure is
not an error by itself; the linter reports the missing pieces and every
finding that cascades from them, so the fix is always visible.

## kbf init

```sh
kbf init my-business --builds-on core-operations
```

Scaffolds `my-business/`: a `manifest.yaml` (`spec: v0`, `version: 0.1.0`,
your `--builds-on` list, `--layer` defaulting to `vertical`), empty
`ontology/` and `evals/` directories, and an `install/slots.yaml`
template. `--layer vertical` (the default) requires `--builds-on`: a
vertical must build on at least one playbook, so an empty closure is
rejected with the embedded core names as the fix, not silently defaulted.
Pass `--layer core` (with `--builds-on` empty or naming other core
playbooks) to start a new foundation playbook instead. Refuses to run if
the target directory already exists: init never overwrites.

`spec/onboarding.md` step 1 is this command; step 2 (competency questions
with the owner, before any modeling) is what comes right after.

## kbf lint

```sh
kbf lint examples/bistro-demo playbooks/core-operations playbooks/core-services playbooks/core-business
```

Validates structure and semantics: the full rule set (`KBF001` to
`KBF013`) with every finding carrying file, line, rule id, and a fix
hint. Exit code 0 on a clean run, 1 if any rule fires. Use it as the
inner loop while authoring (edit, lint, fix, repeat) and as the CI gate
for any repository that holds playbooks. A playbook that only builds on
embedded core playbooks lints clean with just its own path, e.g. `kbf
lint my-business` for the `kbf init` example above.

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
hints containing `<placeholder>` syntax arrive literally; embedded
fallback having resolved a name is never part of this shape, on purpose
(a run's provenance is not a finding), so this contract does not change
because a build-on name happened to come from the embedded cores instead
of a local path. The human-readable render (`--format human`, the
default) is the only place that surfaces as a footer line, e.g. `resolved
from embedded: core-business, core-operations`. The intended agent loop:
run lint, parse `rules`, apply the `fix` hints, run again, stop at an
empty array.

## kbf coverage

```sh
kbf coverage examples/cafe-demo playbooks/core-operations playbooks/core-business
```

Reports `install/slots.yaml` completeness for the leaf playbook: every
declared slot, its mapped source or its absence, and the totals.
Playbooks passed only as composition context are not reported on (their
slot files are templates by definition). Resolves builds-on the same way
`lint` does, embedded fallback included, and its human-readable render
gets the same footer line when fallback was used. Use it:

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
that renders anywhere Mermaid does, including GitHub. Resolves builds-on
with embedded fallback too, but never appends a footer line the way
`lint`/`coverage` do: doing so would stop the output being valid mermaid
as-is. Use it for the owner sign-off walk (`spec/onboarding.md` step 8)
and for embedding the current map in documentation. Further targets are
the designed growth path for this command; each new target is an
emitter, not a new command.

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
Unlike the rest of this page, `schema/` is not embedded in the binary:
reading it needs a clone or GitHub either way, since it exists to be
diffed against real repository state, not carried around inside one.

## Embedded content

`go:embed` cannot reach outside the `tools/` Go module, so the public
core playbooks (`playbooks/core-business`, `playbooks/core-operations`,
`playbooks/core-services`), the `kbf-authoring` skill, and the prose spec
(`spec/*.md`, `spec/primitives/*.md`) are mirrored into
`tools/internal/embedded/data/` and compiled into the binary
(`scripts/embedsync`, `make embed-sync`/`make embed-freshness`). Nothing
else is embedded: no vertical, no example, no client content, ever.

**Precedence is absolute and one-directional.** A locally-passed path
always overrides the embedded copy of the same manifest name; the
embedded copy is consulted only for a `builds-on` name none of your paths
provide at all. A local checkout of `core-operations` mid-edit, not yet
matching the published embedded copy, is always what resolves, never the
embedded fallback sitting behind it. This is enforced in the resolution
code itself, not a convention: see `project-decisions.md`'s
"Batteries-included binary" entry for the full rationale, and
`project-architecture.md`'s Boundaries section for the rule as it stands
today. Embedded content is a convenience default, never a privileged or
hidden source of truth: `kbf playbooks pin` exists specifically so you can
always get the real files.

### kbf playbooks pin

```sh
kbf playbooks pin       # writes ./playbooks/core-{business,operations,services}
kbf playbooks pin --to third_party/kbf
```

Pins the three embedded core playbooks to real, inspectable, editable,
replaceable directories on disk, DO-NOT-EDIT header and all (same as
embedded, since a pinned copy is meant to be read as exactly what
fallback would have used; keep editing it locally from here and it's
yours from that point on, no different from any other local path). The
default `--to playbooks` mirrors this repository's own layout, so a
pinned tree lands where every example on this page already points.
Refuses to overwrite an existing directory; pass `--force` to replace
it.

### kbf skill install

```sh
kbf skill install
```

Writes the embedded `kbf-authoring` skill to `.claude/skills/kbf-authoring/`,
where Claude Code (and any agent runtime reading the same convention)
picks it up automatically. Refuses to overwrite an existing install; pass
`--force` to replace it. Prints what it wrote and a one-line next step.

### kbf docs

```sh
kbf docs                    # list every embedded doc's name
kbf docs onboarding         # print spec/onboarding.md to stdout
kbf docs primitives/entity  # print spec/primitives/entity.md to stdout
```

Serves the prose spec straight out of the binary. A name is a doc's path
under `spec/`, without the `.md` extension. This is how an agent (or a
human) reads any spec document with no clone: pipe the output to a
pager, a file, or another tool's stdin.

## Reading order

New to the toolchain and starting from nothing: `go install
github.com/tadanahq/kbf/tools/cmd/kbf@latest`, `kbf docs onboarding` for
the full raising flow, `kbf docs cli` (this page) as the reference you
return to; `kbf skill install` if an agent is doing the authoring.
Cloned this repository instead: `README.md`'s five-minute tour covers the
same ground against the real files on disk.
