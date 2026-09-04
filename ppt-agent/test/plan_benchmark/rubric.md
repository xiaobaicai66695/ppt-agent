# PPT Planner Judge Rubric

Judge only the semantic quality of `tasks.json`. Deterministic Go checks
already validate JSON shape, legal `content_type`, unique page ids, required
fields, component capacity, background plan, and forbidden render fields.

Score each dimension from 0 to 10:

| Dimension | What Good Looks Like |
| --- | --- |
| intent_coverage | The deck addresses the user's explicit topic, audience, constraints, must-have items, and requested page count or structure. |
| narrative | Slides form a clear story with setup, evidence, implications, and conclusion instead of independent bullet dumps. |
| content_specificity | Pages include named entities, concrete facts, scoped scenarios, numbers, examples, or decisions where relevant. |
| layout_fit | `content_type` choices match the content, and the sequence varies layouts appropriately. |
| data_and_sources | Data, charts, KPI, quotes, and factual claims include usable source notes or provenance where available. |
| visual_planning | Visual intent, background, image, map, or diagram planning supports the page purpose without forcing irrelevant images. Repeated `content_type` pages must plan one shared background query instead of page-by-page rotation. |
| capacity_control | Component count and text density fit the selected layouts; overloaded content is split or summarized. |

Pass threshold is normally `score >= 7.0` with no severe issue. Severe issues
include missing a user must-have, fabricated or anonymous critical data,
unsupported narrative flow, or obvious slide overload.
