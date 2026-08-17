# RFCs

The KBF Ontology Spec (`spec/`, `schema/`) is public API once this
repository is public: other implementations, not just this repository's
own `kbf` CLI, depend on its shapes staying stable and its changes being
explainable. An RFC (request for comments) is how a change to the spec
itself gets proposed, discussed, and decided in the open, instead of
landing as a surprise in a pull request.

## When you need one

- Adding, removing, or changing a field on any primitive in
  `spec/primitives/`.
- Adding a verb to the controlled relation vocabulary
  (`spec/conventions.md`).
- Adding, removing, or changing what a `kbf lint` rule id checks.
- Any change to `playbook-format.md` or `versioning.md`'s policies.

A typo fix, a clarifying sentence, or a new example that does not change
what's valid does not need an RFC: open a pull request directly.

## Process

This is deliberately lightweight while the spec is pre-1.0:

1. Open an issue titled `RFC: <short description>`, covering what changes,
   why, and what breaks for an existing package if anything does.
2. Discussion happens on the issue. A maintainer either accepts, rejects,
   or asks for a revised proposal.
3. Once accepted, the implementation (spec prose, schema, linter, or all
   three) ships as a normal pull request that references the RFC issue.
   A breaking change ships with a migration note, per
   `spec/versioning.md`.

There is no separate RFC document format or numbering scheme yet: the
issue itself is the record. That may change once the spec reaches 1.0 and
the volume of proposals justifies more process; until then, more ceremony
would just slow down a spec that is still finding its shape.
