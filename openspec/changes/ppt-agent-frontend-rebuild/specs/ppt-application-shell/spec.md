## ADDED Requirements

### Requirement: Unified application shell
Authenticated product pages SHALL use a consistent application shell with stable primary navigation, route context, and primary action placement.

#### Scenario: User moves between product pages
- **WHEN** the user navigates among home, compose, dashboard, and administrator routes
- **THEN** primary navigation remains in a predictable location with the current route visibly identified
- **AND** page-specific actions appear in the shared context header rather than creating unrelated navigation patterns

### Requirement: Adaptive navigation
The application shell SHALL remain fully operable at 375, 768, 1024, and 1440 pixel viewport widths without horizontal page scrolling.

#### Scenario: User opens a product page on a narrow viewport
- **WHEN** the viewport cannot accommodate the desktop rail and page content
- **THEN** navigation is available through a labeled drawer or compact control
- **AND** opening or closing navigation does not resize the main content unpredictably

### Requirement: Accessible shared controls
Shared navigation and context controls SHALL use semantic interactive elements, visible focus states, accessible names, and practical touch targets.

#### Scenario: User operates the shell with keyboard or touch
- **WHEN** the user traverses navigation and header actions without a pointer or on a touch device
- **THEN** every action is reachable, named, and visibly focused or pressed
- **AND** icon controls provide at least a 44 by 44 pixel interaction target

### Requirement: Consistent visual language
Product pages SHALL use shared semantic tokens, one icon family, and restrained surface/elevation rules.

#### Scenario: A new page surface is rendered
- **WHEN** a page displays navigation, controls, status, or repeated content
- **THEN** it uses shared surface, text, border, action, and semantic state tokens
- **AND** it does not introduce decorative gradient or emoji substitutes for product UI
