---
name: kbf-authoring
description: Use when raising, creating, or extending a KBF ontology or playbook for a business: onboarding a new business onto KBF, authoring playbook YAML, or driving the kbf CLI (lint, coverage, compile). Covers the interview flow with the business owner and the lint-driven authoring loop.
---

# kbf-authoring

You are authoring a KBF playbook: the versioned contract that declares what
a business's data means and what agents may do about it. This is
config-phase work: you produce YAML and get it to lint clean; you do not
build integrations or run queries. You work as a pair: you author, the
business owner answers. The owner's answers are the source of truth about
the business; the spec is the source of truth about the format. You are
the translator, never the inventor.

## Ground truth first, always

Before authoring anything, read from the kbf repository, in this order:

1. `spec/onboarding.md`: the eight-step raising methodology. It is the
   flow you will run; this skill only adds conduct rules to it.
2. `spec/cli.md`: the command reference and the JSON interface you will
   parse.

Field-level questions (what may an entity carry, which enums exist) are
answered by `schema/ontology.schema.yaml` and `schema/manifest.schema.yaml`,
or by the primitive docs in `spec/primitives/`. Never answer a format
question from memory: the spec moves, your memory doesn't.

## Prerequisites check

Confirm before starting, and fix or ask if missing:

- The `kbf` binary runs (`kbf --help`). If not: clone
  `https://github.com/tadanahq/kbf`, then
  `cd tools && go build -o ../bin/kbf ./cmd/kbf`.
- The core playbooks are reachable on disk (the `playbooks/` directory of
  that clone). Every command takes the full composition closure as paths,
  so you need them locally: `builds-on: [core-operations]` means passing
  `playbooks/core-operations playbooks/core-business` on every invocation.

## Conduct rules for the interview steps

Run the eight steps of `spec/onboarding.md` in order. These rules govern
HOW you run them with a human:

- **One question at a time.** Never send the owner a questionnaire. Ask,
  wait, record, follow up.
- **Competency questions before any modeling** (step 2). If you notice
  yourself writing an entity before the questions exist, stop: you are
  mirroring a source system, not the business.
- **The triage question verbatim** (step 3): "what do you do differently
  from every competitor selling roughly the same thing?" The answer gets
  the deepest modeling; do not paraphrase the question into something
  softer.
- **Capture both words before naming** (step 4): the source system's field
  name and the owner's spoken word for the same thing. The owner's word
  becomes a synonym; do not pick the canonical `name` until you have both.
- **Probe for homonyms explicitly** (step 5): "when you say X, do you
  always mean the same thing?" Owners never surface homonyms unprompted.
- **Unmapped slots are decisions, not defects** (step 7): present each gap
  as skip-or-buy and record the owner's choice; never silently map a slot
  to a source that does not really carry it.

## The authoring loop

After every YAML change:

```sh
kbf lint <playbook> <closure paths...> --format json
```

Parse `rules[]`. For each finding, apply the `fix` hint at `file:line`.
Re-run. The loop ends at `{"rules": []}`, never earlier. Do not suppress,
work around, or reinterpret a finding: if a rule seems wrong for the
situation, stop and tell the owner; the spec's RFC process exists for
that.

Then, in order: `kbf coverage` (walk the unmapped list with the owner,
record skip-or-buy per gap), `kbf compile --to mermaid` (the owner walks
the map and confirms the business is recognizable; a lint-clean ontology
the owner does not recognize is still wrong), tag `v0.1.0`.

## Prohibitions

- **Never fork**: if an element already exists in the closure, you may add
  glossary-tier fields only (synonyms, thresholds). A redeclaration that
  changes anything else fails KBF008; the fix is to extend or to propose
  upstream, never to copy.
- **Reuse verbs before minting**: a new relation verb is legal only on a
  pair touching an entity this playbook declares (KBF007); prefer an
  existing verb on a new pair even then.
- **Never edit `schema/`** in the kbf repository: generated files.
- **Client playbooks are private**: they never enter the public kbf
  repository, and nothing from the client (names, prices, internals) ever
  flows back into public playbooks or upstream proposals without the
  owner's explicit say.

## Exit

Done means: lint clean over the full closure, coverage gaps each carry an
owner decision, the owner has signed the map, the playbook is tagged
`v0.1.0`. What comes next (adapters, deriving storage, runtime
enforcement) is outside this skill: say so and stop.
