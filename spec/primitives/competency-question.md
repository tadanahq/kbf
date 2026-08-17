---
type: spec-doc
---

# Competency question

A competency question is a plain-language question paired with the
elements an answer must use. Together, the competency questions in a
package are its acceptance suite: if an ontology cannot answer the
questions it claims to, it is not done, no matter how complete the
`ontology/` files look.

## Fields

| Field | Required | Shape | Notes |
|---|---|---|---|
| `kind` | yes | `competency-question` | Discriminates this document from other primitives. |
| `question` | yes | plain-language sentence | Specific enough that there is one right family of answer, not a vague prompt. |
| `expects` | yes | list of element names | The entities, relations, or metrics the answer must exercise. Every name must resolve, or fails `KBF009`. |

Competency question carries neither `tier` nor `risk`: `KBF010` does not
apply to it. At least one per entity is the v0 floor
(`packages/universal-core/evals/competency-questions.yaml` has seven, one
per entity; a base package or a leaf install adds more for what it adds,
`packages/operations-core/evals/competency-questions.yaml` and
`examples/cafe-demo/evals/competency-questions.yaml` among them).

## Example

Copied from `packages/universal-core/evals/competency-questions.yaml`:

```yaml
kind: competency-question
question: What is the labor cost ratio for the organization this month?
expects: [labor-cost-ratio]
```

## Common mistakes

- **A dangling `expects` entry.** A metric, entity, or relation name that
  does not resolve fails `KBF009`. Rename or delete an element and its
  competency questions go stale in the same pass.
- **Referencing a relation verb that recurs.** `expects: [supplies]` is
  ambiguous once a package declares `supplies` for more than one `(from,
  to)` pair (see `spec/primitives/relation.md`): nothing in the question
  says which one. Reference the metric or entity the question is really
  about instead, the way
  `packages/universal-core/evals/competency-questions.yaml` asks about
  suppliers with `expects: [supplier, purchase-cost]`, not
  `expects: [supplies]`.
- **A question with no single right answer.** "How is the business doing?"
  cannot be checked against `expects`. A competency question is an
  acceptance test, not a prompt; write it the way you would write a test
  name.
- **Skipping an entity.** Every entity needs at least one competency
  question that exercises it, directly or through a relation or metric at
  its grain. An entity with zero coverage has no proof an agent can
  actually answer anything about it.
