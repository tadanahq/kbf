---
type: spec-doc
---

<!-- DO NOT EDIT: generated from spec/onboarding.md by scripts/embedsync. Edit the source, then run `make embed-sync`. -->

# Raising an ontology for a new business

Every other document in `spec/` describes a shape: what a valid entity,
relation, or playbook looks like once it exists. This one describes the
order those shapes actually get authored in, for a business that has none
of them yet. It is written for the pair doing that work together, an
agent authoring YAML and the owner who knows the business, not for `kbf
lint`: nothing in this document is enforced, only sequenced. The
playbooks in this repository (`core-business`, `core-operations`,
`core-services`, `cafe-demo`, `studio-demo`, `bistro-demo`) are the
finished answer key; this document is the walk that produces one, before
it is finished.

The order matters more than any single step's shape. Each step produces
an artifact the next step needs, and skipping ahead (modeling entities
before the competency questions exist to justify them, say) tends to
produce an ontology that mirrors a source system instead of the business
behind it.

## 1. Start from the closure

Before writing a single entity, decide what this business already shares
with. Pick every core playbook whose shape actually fits (a site-based
business needs `core-operations`; a project-based one needs
`core-services`; a business that does both, the way `examples/bistro-demo`
does, needs both) and set `builds-on` to that list. `layer: vertical`,
always, for a new business: whether any of what it needs deserves to
become its own core playbook is not a decision to make this early, and
minting one prematurely is how a `core-` playbook ends up narrower than
its name promises (`spec/conventions.md`). Naming the wrong closure here
is cheap to fix, edit `builds-on`, re-run `kbf lint`; naming it right the
first time just saves the churn. Everything downstream, the vocabulary
available in step 2, the entities already sitting in step 4, is a
function of this list, so get it close before going deeper, not perfect
before moving on.

## 2. Competency questions first

Before any entity gets modeled, sit down with the owner and write
competency questions, the plain-language questions
`spec/primitives/competency-question.md` describes, phrased the way the
owner would actually ask an agent something ("how many bookings this
month came through referral, not the website"). Writing these first, not
after the ontology exists, is deliberate: modeling first tempts the
author into declaring whatever the source systems happen to expose and
calling that done. Competency questions invert it, they fence what the
ontology has to be able to answer, so entities, relations, and metrics
get modeled to satisfy a real question, not to mirror a schema. Each
question becomes its own acceptance test (`expects: [...]`) once the
elements it needs exist; a question that cannot be phrased yet against
anything real is a sign a needed entity or relation is still missing,
useful information this early, not a failure.

## 3. Subdomain triage

Not every part of a business deserves the same modeling depth. Before
going entity by entity, sort what the business does into three buckets,
the strategic domain-driven design move: **generic** (payroll, tax
filing: every business does this, nobody competes on it, a bought tool is
fine), **supporting** (inventory counts, shift scheduling: the business
needs it to run, but it is not why a customer chooses them), and **core
domain** (whatever the owner answers when asked "what do you do
differently from every competitor selling roughly the same thing?"). A
multi-location food business's core domain might be how it schedules
staff across sites to hit a labor-cost target, not the point-of-sale
transaction itself, which every competitor also has. The core domain gets
the deepest modeling, its own entities, its own metrics, its own
competency questions; generic and supporting get just enough shape to be
correct, not elaborated. Getting this triage wrong in either direction
costs real time: over-modeling the generic parts is effort spent proving
something nobody disputes, under-modeling the core domain leaves out the
one thing the business actually needed an agent to reason about.

## 4. The entity interview

Now the entities, across the composed closure: the playbook's own new
ones plus whatever the core playbooks in step 1 already declare. For each
thing the business has, interview the owner with a specific question
shape: what does your team actually call this, versus what does the
source system call it? A booking tool's internal field for something and
the owner's own spoken word for the same thing are usually one entity,
different only by vocabulary: the source-system name becomes, or
confirms, the `identity` key (`spec/conventions.md`'s naming rule is
explicit that identity keys read like source-system column names on
purpose), and the owner's word becomes a `synonyms` entry, never a second
entity. Capture both the vendor's field name and the owner's word for the
same thing before picking a `name`; recording only one loses the other
unless someone remembers to ask again later. Where the entity has a real
lifecycle, capture `states` the same way, ask what the actual stages are
called in the owner's own words ("booked", "confirmed", "walked"), not
whatever internal status enum the source tool happens to use, since
`states` is what an agent reasons over, not the source system's own
representation.

## 5. The homonym hunt

Once entities exist, hunt deliberately for homonyms: two things the
business calls by the same word that are not, in fact, the same entity.
"Booking" at a multi-location food business that also caters events can
mean a table reservation (no revenue commitment, cancels for free) or an
event booking (a deposit, a contract, its own cancellation policy): one
word, two entities, and modeling them as one is exactly the blurred
definition an agent cannot safely reason over, because the two things are
safe to sum, join, and act on in different ways. A homonym is resolved by
giving each meaning its own entity name, never by adding a qualifier to
`synonyms` and hoping context disambiguates it later; if the business
genuinely only ever means one of the two, that is not a homonym, just a
name to confirm. This hunt is deliberately its own step, separate from
the entity interview, because an owner rarely surfaces a homonym
unprompted: they use the overloaded word fluently every day and do not
notice it is overloaded until asked directly, "when you say that, do you
always mean the same thing?"

## 6. Event-storming-lite (optional)

For a business whose core domain (step 3) is process-heavy rather than
record-heavy, walking a day in sticky-note form, the classic
event-storming style, sequencing what actually happens ("order placed",
"payment captured", "item marked out for delivery"), surfaces relations
and actions a static entity interview misses, because it forces the
sequence to be named, not just the nouns in it. This step is optional: a
business whose core domain is simple enough to fall out of steps 3
through 5 directly does not need it, and running it on every engagement
regardless is ceremony for its own sake. Reach for it when the entity
list from step 4 feels right but the relations between those entities
still feel arbitrary.

## 7. The slot matrix

Before any integration code gets written, the slot matrix
(`install/slots.yaml`, `spec/primitives/slot-mapping.md`) gets filled in
as a conversation with the owner, not as a side effect of building an
adapter. For every attribute slot in the composed closure, ask: what
system actually has this, and does the business even have that system?
`kbf coverage`, run against the still-empty `slots.yaml`, is the artifact
that conversation produces: a declared-versus-mapped count and a named
list of gaps, the same shape `examples/bistro-demo`'s two deliberately
unmapped `delivery.deliverable-due`/`delivery.deliverable-status` slots
demonstrate. A gap is not automatically a problem to fix before shipping;
it is a skip-or-buy decision to make explicitly with the owner: skip, the
business genuinely has no source for this yet, leave `source: ""` and
move on; or buy, the business needs a new tool or process to fill it,
which is the owner's call, not something the ontology should paper over
silently.

## 8. Lint to green, map, sign, tag

Run `kbf lint` against the full composition closure until it is clean;
every finding names a file, a line, and a fix, so this is a loop, not a
one-shot check. Once it lints clean, render the map (`kbf compile --to
mermaid`; `spec/index.md`'s "Where the primitives live for real" shows
what the output looks like for this repository's own playbooks) and walk
it with the owner. This is the moment the owner is looking at their
business, translated, and confirming it is recognizable, not a step to
skip because the linter is already satisfied: a lint-clean ontology the
owner does not recognize is still wrong, just wrong in a way `kbf lint`
cannot catch. Once the owner signs off, this is the playbook's first
stable cut: tag it `v0.1.0`, the same semver its own `version` field
already carries (`spec/primitives/namespace.md`), and the number every
later change bumps from.

## The install repo

A playbook raised for a real business lives in its own repository, one
playbook per repo, rooted at a directory conventionally named
`playbook/`:

```
playbook/
  manifest.yaml
  ontology/
  evals/
  install/
  builds-on/     # the pinned composition closure
```

`kbf init` creates everything except `builds-on/`; `kbf playbooks pin`,
run from inside `playbook/`, adds it (the command's default `--to` is
`builds-on` for exactly this layout; from the repo root, pass `--to
playbook/builds-on`). Set the layout up right after step 1's `kbf init`;
nothing later depends on when the closure gets pinned. Pinning is what
makes the closure readable to tools that are not `kbf`: `kbf` itself
resolves `builds-on` names against local paths and its embedded cores,
but a downstream consumer, an agent runtime loading the contract from
disk, say, just walks `builds-on/` and sees the full closure, with no
composition resolution reimplemented. Inspectability and version
pinning ride along: the pinned copies are real files, diffable, and
frozen until deliberately re-pinned.

A brownfield install (a business with systems already running, the
normal case) tends to carry three more files at the playbook root:
`MAPPING.md` (contract-to-current-system bindings), `DRIFT.md` (where
the running system disagrees with the contract), `PENDING.md` (open
questions for the owner). They are working install artifacts, not spec
primitives: nothing lints them, and a greenfield install may never
create them.

What comes after this point (wiring real adapters, running the ontology
against live queries, the runtime enforcement `README.md`'s "Status: v0"
section marks as roadmap) is out of this spec's scope by design: v0 is
config-phase tooling, and onboarding a business ends where authoring its
contract ends, not where running it begins.
