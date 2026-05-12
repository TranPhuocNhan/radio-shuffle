Plan: Project Overview and Functional Spec Docs
Create two human- and agent-friendly docs under docs/ by synthesizing the existing architectural, module, workflow, and README descriptions, plus DB schema for domain terms. The first doc gives a concise end-to-end overview (business terms, actors, flows). The second enumerates functional requirements and use cases per module, including API surfaces and non-functional expectations, while explicitly marking stubs or pending work.
Steps
Extract business terms and actors from README.md, docs/architecture.md, and docs/modules.md.
Build a concise system overview and key flows for docs/basic-design.md.
Enumerate module-specific use cases and requirements in docs/functional-specification.md.
Cross-check against db/schema.sql and db/migrations/ for domain entities and constraints.
Add “status” notes for stub modules and explicitly list non-goals or future work.
Further Considerations
Confirm filenames with spaces and casing: keep docs/Basic Design.md and docs/functional specification.md, or prefer consistent casing like docs/functional-specification.md?
Should the functional spec include API endpoint tables or stay at business-level use cases only?
Any specific frontend or stakeholder audience to tailor terminology (e.g., PM vs developer)?

