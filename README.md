# KBF: the Kozmo Business Framework

KBF organizes a business for AI agents. It is the business layer of an
agentic operating system: the framework that declares what a company's
data means, what its agents may do about it, and how a whole vertical's
operating knowledge ships as an installable package.

## The framework

| Module | What it declares | Status |
|---|---|---|
| **KBF Ontology** | The contract: semantic elements (entities, relations, metrics) + action elements + operational elements | **v0, ships today** (spec + tooling in this repo) |
| **Package format** | Playbooks as runnable packages: ontology + integrations + workflows + agents + evals + install | Partial in v0 (manifest + ontology + evals + install linted; the rest reserved) |
| **Findings contract** | The judgment layer: claims with evidence, confidence, temporal validity | Roadmap |
| **Outcomes** | Value measurement over metrics and findings | Roadmap |

The ontology is the foundation module: everything else in the framework
builds on it, which is why it ships first.

## The KBF Ontology (what ships today)

Semantic layers describe data for dashboards. The KBF Ontology describes a
business for agents.

A semantic layer maps tables and columns to friendly names, so a BI tool
can build a chart without an analyst writing SQL by hand. That is not
enough for an agent: an agent needs to know what a location *is*, what
relations it can join through, which metrics are safe to add up, and what
it is and is not allowed to do when it notices something wrong. The KBF
Ontology is a versioned, YAML-authored contract that answers those
questions, checked the same way a broken build is checked: `kbf lint`
fails fast with a file, a line, a rule id, and the fix.

## Why

LLM accuracy on enterprise questions roughly triples with a semantic
layer, and improves again when the queries an agent generates are
validated and repaired against an ontology before they run. A dashboard
semantic layer gets an agent halfway there. The rest, what an order *is*
versus what a purchase *is*, which relations actually exist, which
metrics can be summed across locations, has to be declared somewhere an
agent can read it and a linter can check it. That declaration is the KBF
Ontology.

## What's here

| Path | What it is |
|---|---|
| [`spec/`](spec/) | The prose spec: start at [`spec/index.md`](spec/index.md). |
| [`schema/`](schema/) | Generated JSON Schema for `spec/`'s shapes, for editor support (`yaml-language-server`). Never hand-edited. |
| [`packages/universal-core/`](packages/universal-core/) | Playbook Zero: the base ontology every business shares. |
| [`examples/cafe-demo/`](examples/cafe-demo/) | A fictional single-location cafe extending `universal-core`, exercising every extension mechanic the spec allows. |
| [`tools/`](tools/) | The `kbf` CLI: `lint`, `coverage`, `compile --to mermaid`, `schema`. |
| [`conformance/`](conformance/) | Language-agnostic fixtures (YAML in, expected outcome out), so an implementation other than `kbf` can prove it matches the spec. |
| [`rfcs/`](rfcs/) | How the spec itself changes once it is public. |

## Status: v0

v0 ships the framework's foundation module (the KBF Ontology Spec, its
packages, and its tooling) and is **config-phase tooling only**: `kbf`
helps an agent or a human author and validate an ontology. It does not
run queries against one.
Runtime enforcement (validating an agent-generated query against the
contract before it executes) is on the roadmap, not in this repository
yet. If you are looking for that, it does not exist here.

## Five-minute tour

Prerequisite: Go 1.26 or later.

```sh
git clone https://github.com/tadanahq/kozmo-bf.git
cd kozmo-bf
```

**1. Read the thesis.** You just did (above). For the full shape of the
spec, read [`spec/index.md`](spec/index.md); it is short and links to
everything else.

**2. Build the CLI.**

```sh
cd tools && go build -o ../bin/kbf ./cmd/kbf && cd ..
```

**3. Lint the base ontology.**

```sh
./bin/kbf lint packages/universal-core
```

**4. Lint the demo.** `cafe-demo` extends `universal-core`, so pass both
paths: `kbf` resolves `extends` against whatever packages you give it, not
against this repository's folder layout.

```sh
./bin/kbf lint examples/cafe-demo packages/universal-core
```

**5. See what's mapped.** Demo Cafe has no CRM yet; `coverage` should show
the three `crm.customer-*` slots unmapped and everything else mapped to
`demopos` or `demobooks`.

```sh
./bin/kbf coverage examples/cafe-demo packages/universal-core
```

**6. Render the ontology map.**

```sh
./bin/kbf compile --to mermaid examples/cafe-demo packages/universal-core > cafe-demo.mmd
```

Open `cafe-demo.mmd` in an editor with Mermaid preview, or paste it into a
GitHub Markdown file: entities become nodes, relations become labeled
edges.

**7. Break something, on purpose.** Open
`packages/universal-core/ontology/order.yaml`, delete the `identity:`
line, and re-run step 3. The error names the file, the line, the rule
(`KBF004`), and the fix, the same shape as every other `kbf lint` error.
Put the line back when you're done.

That is the whole loop: author YAML against the primitives in
[`spec/primitives/`](spec/primitives/), lint it, fix what it flags, ship.

## Contributing

See [`CONTRIBUTING.md`](CONTRIBUTING.md). Spec changes beyond typos go
through [`rfcs/`](rfcs/).

## License

Apache-2.0. See [`LICENSE`](LICENSE).
