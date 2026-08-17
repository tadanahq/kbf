# Skills

Agent skills that ship with KBF. A skill is a plain-markdown instruction
package (`SKILL.md` with YAML frontmatter) that agent runtimes such as
Claude Code load by its `description` trigger; the format is portable to
any runtime that reads markdown instructions.

| Skill | For |
|---|---|
| [`kbf-authoring/`](kbf-authoring/) | Raising or extending a playbook with the CLI: the interview conduct rules and the lint-driven authoring loop. Thin by design: canonical knowledge stays in `spec/`, the skill routes to it. |

## Install

Copy (or symlink) the skill folder into the project where the agent
works:

```sh
cp -r skills/kbf-authoring /path/to/your-project/.claude/skills/
```

Claude Code picks it up from `.claude/skills/` automatically. Keep the
kbf repository cloned alongside your project: the skill drives its
binary, core playbooks, and spec.
