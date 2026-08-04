## ADDED Requirements

### Requirement: Manifest updates are structured and atomic
The DeepAgent SHALL update `tasks.json` through a dedicated structured tool that validates the complete in-memory manifest before atomically replacing the file.

#### Scenario: Batch content planning succeeds
- **WHEN** the Agent submits valid structured updates for multiple slides
- **THEN** all updates are applied in one atomic write and the tool returns a machine-readable summary

#### Scenario: Invalid content plan is submitted
- **WHEN** a structured update cannot be decoded or violates required task identity fields
- **THEN** the tool returns an error and leaves the existing `tasks.json` unchanged

### Requirement: Slide completion follows output files
The task coordinator SHALL reconcile slide status from declared output files instead of requiring the main Agent to rewrite the manifest after every successful SlideExecutor call.

#### Scenario: Declared output appears
- **WHEN** a generated PPTX matching a pending task's `output_file` appears in the work directory
- **THEN** the coordinator marks that task done and publishes updated progress without an Agent file edit

### Requirement: Main Agent cannot overwrite the manifest with a generic file tool
The production DeepAgent SHALL NOT expose generic whole-file editing as the supported way to modify `tasks.json`.

#### Scenario: Agent plans or updates pages
- **WHEN** the main Agent needs to initialize or change page content or task state
- **THEN** its instruction and tool set direct it to the dedicated manifest tool
