# visual-designer-layout-balance Spec

## ADDED Requirements

### Requirement: Dynamic Text Fitting

Visual Designer generators SHALL adjust visible text size and line spacing based on text length and available box size.

#### Scenario: Sparse text

- GIVEN a title, quote, or short bullet set with low text density
- WHEN the slide is generated
- THEN the text SHALL use larger readable sizes and be positioned as an intentional focal element instead of sitting in a small fixed region.

#### Scenario: Dense text

- GIVEN a generator receives dense but in-contract text
- WHEN the slide is generated
- THEN text SHALL fit inside its intended box without crossing slide bounds or overlapping unrelated content.

### Requirement: Content Band Balance

Generators SHALL place grouped content using vertical centering or balanced bands when the region has extra space.

#### Scenario: Items fewer than maximum

- GIVEN a grid/list/timeline has fewer items than the maximum layout supports
- WHEN the slide is generated
- THEN the group SHALL not appear pinned to the top or bottom of the available region.

### Requirement: All-Template Smoke Test

The implementation SHALL provide a repeatable local smoke test that generates and renders every single-page template.

#### Scenario: Full template run

- GIVEN the repository has `templates/single-page/*.json`
- WHEN the smoke script runs
- THEN it SHALL produce one PPTX and one PNG per template, a contact sheet, and a JSON report with generation/render warnings.
