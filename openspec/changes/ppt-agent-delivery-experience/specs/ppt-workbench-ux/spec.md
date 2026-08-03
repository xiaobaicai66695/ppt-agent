## ADDED Requirements

### Requirement: Real preset previews
The template library SHALL display a real, locally available preview image for each preset template.

#### Scenario: Preset library loads
- **WHEN** the template API returns a preset thumbnail URL
- **THEN** that URL resolves to an actual generated template image
- **AND** the UI does not substitute an emoji or decorative gradient for the template

#### Scenario: Preview asset is unavailable
- **WHEN** a preset image cannot be loaded
- **THEN** the UI shows an explicit missing-preview state without implying that a decorative placeholder is the real template

### Requirement: Complete layout metadata presentation
The Compose page SHALL derive layout labels from the layout API and SHALL expose the selected layout's required fields and capacity guidance.

#### Scenario: Newly added layout is returned by the backend
- **WHEN** a layout exists in the templates API but not in a frontend hardcoded label map
- **THEN** Compose displays its backend `display_name`
- **AND** the selected layout shows its required fields without a frontend code change

### Requirement: Responsive Dashboard workbench
The Dashboard SHALL remain usable without horizontal scrolling at 375px, 768px, 1024px, and desktop widths.

#### Scenario: Dashboard opens on a narrow viewport
- **WHEN** the viewport is below the desktop navigation breakpoint
- **THEN** task navigation is available through an explicit drawer control
- **AND** progress, logs, previews, and chat use a single-column reading order

### Requirement: Responsive Compose workbench
The Compose page SHALL preserve template selection, slide ordering, editing, and generation actions on tablet and phone widths.

#### Scenario: Compose opens on a phone
- **WHEN** the viewport is 375px wide
- **THEN** the template library and slide list stack vertically
- **AND** the slide editor opens as a usable overlay or drawer
- **AND** primary generation actions remain reachable

### Requirement: Accessible workbench interactions
Interactive controls in the task, preview, modal, navigation, and editor workflows SHALL be keyboard reachable and have accessible names.

#### Scenario: User navigates with a keyboard
- **WHEN** the user tabs through task history, previews, modal actions, and chat controls
- **THEN** each action receives a visible focus state
- **AND** clickable non-semantic containers are replaced or given equivalent keyboard behavior

#### Scenario: User operates on a touch screen
- **WHEN** the user taps primary icon controls
- **THEN** controls provide a practical touch target and do not depend on hover-only affordances

### Requirement: Product-consistent user copy
Default user-facing copy SHALL describe the enabled PPT workflow and SHALL not advertise disabled QA/reviewer steps.

#### Scenario: QA is disabled by default
- **WHEN** a user opens the Dashboard welcome or progress areas
- **THEN** the copy describes planning, generation, preview, and delivery without promising online visual QA
