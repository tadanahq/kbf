# kozmo-bf – Project Overview

**KBF (Kozmo Business Framework)** is the open framework that organizes a
business for AI agents: the business layer of an agentic operating system.
Its modules, in dependency order: the **KBF Ontology** (the contract), the
**playbook format** (playbooks as runnable units of business capability), the
**findings contract** (the judgment layer), and **Outcomes** (value
measurement). This repo is the home of all of them; the ontology never
stands in for the whole framework.

The **KBF Ontology** is the foundation module and v0's shipped artifact: a
versioned, YAML-authored contract declaring what a business's data *means*
(semantic elements: entities, relations, metrics) and what agents may *do*
about it (action elements), plus the operational elements that make it
installable (slot mappings, competency questions, namespaces).

One line for the framework: **KBF makes a business installable for agents.**
One line for the module: **semantic layers describe data for dashboards; the
KBF Ontology describes a business for agents.**

## Why it exists

Agents acting on business data fail without chosen structure: schemas encode
storage, not meaning; prompts don't enforce anything. The evidence line this
project is built on: LLM accuracy on enterprise questions roughly triples with
a semantic layer, and improves again when generated queries are validated and
repaired against an ontology. The ontology must therefore be a **contract**:
machine-validated, versioned, with acceptance tests: not documentation.

## What ships from this repo

1. **The spec** (`spec/`): prose, seven primitives, authoring conventions.
2. **The meta-schema** (`schema/`): generated, published for editor support.
3. **Reference tooling** (`tools/`): the `kbf` CLI. v0 scope is **config-phase
   only**: help an agent (or human) create and validate an ontology: `kbf lint`,
   `kbf coverage` (static), `kbf compile --to mermaid`. Runtime enforcement (the
   query gate) is deliberately out of v0: see the public roadmap section in
   README when it lands.
4. **Playbooks** (`playbooks/`), layered: `core-business` is the
   truly-universal minimal floor (only what EVERY business shares: the
   revenue event, customers, offerings, people, purchases); **core playbooks**
   (`core-operations` for physical-operations businesses, `core-services` for
   engagement businesses) build on it; verticals build on one or more core
   playbooks (composition, a DAG, not a single parent). Composition, never
   fork, at every layer; root-ness is derived (a core playbook with an
   empty `builds-on`), never a hardcoded package: any organization can
   author its own core.
5. **Teaching content** (`examples/`): `cafe-demo` (operations chain) and
   `studio-demo` (services chain), each small, fictional, fully valid, every
   primitive used at least once. Two demos on two chains is the visible proof
   of the multi-vertical claim.
6. **Conformance** (`conformance/`): language-agnostic fixtures so third-party
   implementations can prove compliance.

## Consumers

- **Authoring agents** (Claude Code, Codex, and similar) drive the CLI through
  skills/plugins to spin up and validate ontologies.
- **Humans** get the same via the editor experience (published schemas) and docs.
- **Runtimes** consume playbooks as data; they live outside this repo.

## v0 definition of done

A stranger can: read README + spec in one sitting, run `kbf lint` on
`examples/cafe-demo` and `playbooks/core-business`, break a file and get a
file:line error that teaches the rule, and render the demo's ontology map with
`kbf compile --to mermaid`.
