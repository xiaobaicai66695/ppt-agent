## ADDED Requirements

### Requirement: Benchmark resolves versioned imported agent datasets
The benchmark command SHALL resolve a named local imported dataset for Planner, Reviewer, and Fixer without changing fixed repository fixture paths. Router SHALL continue to resolve only its existing fixture datasets.

#### Scenario: Imported Planner dataset is selected
- **WHEN** a developer runs `pptbench` with a valid imported Planner dataset name
- **THEN** the command loads the bundle's Planner cases and records the dataset version in run metadata
- **AND** fixed `test` and `validation` fixtures remain unchanged

#### Scenario: Imported Router dataset is requested
- **WHEN** a developer requests an imported dataset for the Router suite
- **THEN** the command rejects the request with an explicit Router fixture-only error
