## ADDED Requirements

### Requirement: Prompt-first creation home
The default product entry SHALL expose a usable PPT creation action before marketing or explanatory content.

#### Scenario: User opens the home route
- **WHEN** the application home loads
- **THEN** the first viewport provides a presentation requirement input and clear path to compose or generate
- **AND** available real template previews or recent work are used as supporting visual content

### Requirement: Editor-style composition workspace
Compose SHALL organize template discovery, slide ordering, slide editing, and generation as one spatially coherent editor workflow.

#### Scenario: User edits an outline on desktop
- **WHEN** the viewport has room for the full editor
- **THEN** template resources, slide sequence, and selected-slide properties occupy stable adjacent regions
- **AND** the primary generation action remains reachable without scrolling to a separate page section

#### Scenario: User edits an outline on phone
- **WHEN** the viewport is 375 pixels wide
- **THEN** the slide sequence remains the primary region and properties open in a dismissible full-screen sheet
- **AND** template selection, reordering, editing, deletion, and generation remain available

### Requirement: Delivery-first task workspace
Dashboard SHALL visually prioritize user-understandable progress and ready slide previews over raw event logs and diagnostics.

#### Scenario: Slides become available during generation
- **WHEN** file and thumbnail lifecycle events arrive
- **THEN** stable 16:9 preview positions appear and update progressively
- **AND** ready previews can be inspected or downloaded without waiting for the full task

#### Scenario: Task activity is verbose
- **WHEN** a task emits many logs or runtime metadata events
- **THEN** progress and previews remain ahead of logs in the reading order
- **AND** developer diagnostics remain collapsed or secondary unless a warning requires attention

### Requirement: Complete product states
Home, Compose, Dashboard, Auth, and Admin SHALL provide distinct loading, empty, error, disabled, and success feedback appropriate to their workflows.

#### Scenario: Data cannot be loaded
- **WHEN** a template, task, profile, or administrator request fails
- **THEN** the affected region displays a concise error and recovery action
- **AND** unrelated navigation and available user actions remain usable

### Requirement: Authentic visual content
Template and slide surfaces SHALL use real locally available previews when present and SHALL identify missing media explicitly.

#### Scenario: Preview media exists or fails
- **WHEN** a template or slide has a valid thumbnail
- **THEN** the thumbnail is displayed with a stable aspect ratio and lazy loading
- **AND** when it fails, the UI shows a missing or failed preview state rather than decorative fake artwork
