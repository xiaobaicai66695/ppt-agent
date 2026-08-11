## ADDED Requirements

### Requirement: LLM-guided recommended style planning
The system SHALL derive recommended template style, palette, background theme, and page count from one structured LLM intent result, and SHALL let the main Agent dynamically plan slide content for the user's topic.

#### Scenario: User selects intelligent recommendation
- **WHEN** a task is created with template selection mode `recommended`
- **THEN** the task uses a valid LLM-suggested template when available
- **AND** the outline carries style guidance and a bounded suggested page count without copying the preset's default slide list

#### Scenario: Intent classification is unavailable
- **WHEN** the structured intent model cannot return a usable recommendation
- **THEN** the system uses a valid generic style and bounded default page count
- **AND** task creation continues without keyword-scored preset selection

#### Scenario: User explicitly selects a preset
- **WHEN** a task is created with template selection mode `preset`
- **THEN** the system preserves the selected preset's template scaffold behavior

### Requirement: Semantic background coverage planning
The main Agent SHALL select backgrounds from the available background catalog according to page purpose, information density, readability, and whole-deck visual rhythm, targeting background use on 45%-65% of slides when a suitable background theme is available.

#### Scenario: Recommended deck contains visual narrative pages
- **WHEN** the main Agent plans title, section, quote, summary, image-text, image-hero, brand-focus, or spacious content pages
- **THEN** those pages receive the highest priority for valid background references
- **AND** adjacent references from one theme rotate across available images

#### Scenario: Deck contains dense data pages
- **WHEN** the main Agent plans chart, KPI, table-like, or other data-dense pages
- **THEN** it prioritizes a clear information surface and counts those pages as low-priority background candidates

#### Scenario: Planned background coverage is incomplete
- **WHEN** a recommended-style manifest has a valid recommended background but remains below the target after planning
- **THEN** deterministic normalization fills eligible high-priority pages with valid rotating background references

### Requirement: Layout-aware content density
Single-page template contracts SHALL express a useful target density range separately from the maximum renderable capacity, and the planning prompt SHALL request complete information units that fit the selected layout.

#### Scenario: Main Agent plans an information slide
- **WHEN** it writes `content_plan` for content, summary, multi-column, card, quote, image-text, KPI, or chart layouts
- **THEN** the amount and structure of content follow that layout's target density
- **AND** the maximum capacity remains a rendering safety boundary rather than the normal writing target

#### Scenario: A page has little source material
- **WHEN** the topic provides less content than the layout's target density
- **THEN** the planner may choose a more spacious layout or add relevant evidence, interpretation, or visual intent
- **AND** the generated page does not rely on artificially short template example text

### Requirement: Positive planning contract
Agent instructions SHALL define desired outcomes, decision priorities, valid choices, and structured output schemas so planning behavior covers supported scenarios without relying on accumulating narrow negative prompt patches.

#### Scenario: Prompt handles competing layout concerns
- **WHEN** background visibility, text readability, content completeness, and layout capacity all apply
- **THEN** the prompt provides an explicit priority order and selection criteria
- **AND** the Agent can produce one schema-valid decision for the page
