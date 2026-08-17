# Contributing to kozmo-bf

Thanks for reading this before opening a pull request. This repository is
the reference implementation of the KBF Ontology Spec, so changes here set
precedent for anyone else building against it: the bar is a little higher
than "it works."

## How work happens here

Every change starts as a spec capsule under `.agents/specs/<name>/`: an
`overview.md` (what and why), a `design.md` (how), and a `tasks.md` (the
plan, kept live as work lands). This repository's [`AGENTS.md`](AGENTS.md)
is the full, canonical description of that workflow, along with the
architecture and standards every change must comply with
(`.agents/steering/`); this file only summarizes the parts a first-time
contributor needs.

In short: propose the change as a capsule before writing code or content,
build against the approved capsule, keep `tasks.md` current as each task
lands, and make sure `make check` (gofmt, lint, tests, `kbf lint` over
`playbooks/` and `examples/`, conformance, schema-freshness, boundaries) is
green before calling anything done. `make check` is the gate; a pull
request that does not pass it is not ready for review.

## License headers

This project is licensed Apache-2.0. Go source files under `tools/` carry
the standard Apache-2.0 header comment at the top of the file. YAML content
(`playbooks/`, `examples/`, `conformance/`) and Markdown documentation
(`spec/`, this file, `README.md`) do not carry a per-file header: the
repository-level [`LICENSE`](LICENSE) file covers them, following the same
convention most open-source projects use for data and prose files that
are not compiled.

## Changing the spec

The spec (`spec/`, `schema/`) is public API: other implementations depend
on it, not just this repository's own `kbf` CLI. A typo fix can land
directly; anything that changes a primitive's shape, adds or removes a
lint rule, or changes what a controlled vocabulary allows goes through the
RFC process described in [`rfcs/README.md`](rfcs/README.md) first. If
you're unsure whether your change needs an RFC, open one: it is cheaper to
close an RFC as unnecessary than to revert a merged breaking change.

## Questions

Open an issue. If you're proposing something larger than a fix, a short
issue describing the problem before the pull request saves everyone a
rewrite.
