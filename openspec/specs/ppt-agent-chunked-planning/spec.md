# ppt-agent-chunked-planning Specification

## Purpose
TBD - created by archiving change chunked-deck-planning-and-context-compression. Update Purpose after archive.
## Requirements
### Requirement: Deck generation uses chunked first-draft planning for larger decks
The system SHALL construct larger initial deck drafts through a blueprint phase, one or more section-planning shards, deterministic merge, and one final Task Reviewer quality gate.

#### Scenario: Larger deck starts planning
- **WHEN** a user creates a deck whose target page count is above the chunking threshold
- **THEN** the system creates a blueprint that fixes page indexes, section boundaries, page titles and content types before detailed page content is generated
- **AND** section planners only fill detailed content for their assigned page ranges

#### Scenario: Short deck starts planning
- **WHEN** a user creates a deck whose target page count is at or below the chunking threshold
- **THEN** the system MAY use a single section shard or the existing monolithic planner path
- **AND** the final output still conforms to the same `tasks.json` contract

### Requirement: Section shard outputs are merged deterministically
The system SHALL merge section-planner outputs into one draft manifest without allowing shards to overwrite unrelated pages or engineering-managed task fields.

#### Scenario: Section shards complete
- **WHEN** all required section shards return page plans
- **THEN** the merger orders pages by blueprint `page_index`
- **AND** the merger assigns stable `task_id`, `output_file`, `status`, `created_at` and capacity counts
- **AND** the merger rejects duplicate or missing page indexes before review

#### Scenario: Shard omits deterministic fields
- **WHEN** a section shard omits task id, output file, status or capacity count
- **THEN** the merger fills those fields from the blueprint and component list

### Requirement: Planning metadata supports granular repair
The system SHALL preserve enough section and evidence metadata in each planned page to support later page-level or section-level fixes.

#### Scenario: User requests one page revision
- **WHEN** a completed deck receives a fix request targeting one page
- **THEN** the system can identify the page's `section_id`, page purpose and evidence references without reading unrelated generated pages

#### Scenario: User requests section revision
- **WHEN** a completed deck receives a fix request targeting a named section
- **THEN** the system can select only pages in that section for semantic repair

### Requirement: Chunked planning remains a single user-facing generation action
The system SHALL present chunked planning as one generation operation rather than a multi-round corrective workflow.

#### Scenario: Chunked planning succeeds
- **WHEN** blueprint planning, section planning and merge complete
- **THEN** the system runs the existing Task Reviewer gate once on the merged draft before rendering
- **AND** the frontend progress stream shows concise planning progress without asking the user to manage shards

