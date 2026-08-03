## ADDED Requirements

### Requirement: Creation home displays every available preset
The creation home SHALL load and display all preset templates returned by the template API together with one smart recommendation option and one custom composition option.

#### Scenario: Preset catalog loads
- **WHEN** the template API returns the current preset catalog
- **THEN** every preset appears as a selectable thumbnail card on the same page as the prompt input

#### Scenario: Catalog loading fails
- **WHEN** the template API is temporarily unavailable
- **THEN** the page keeps the prompt usable, shows a bounded retry state, and retains smart recommendation and custom composition choices

### Requirement: Template selection occurs once
The system MUST NOT ask the user to select the full-deck template again after the user selects it on the creation home.

#### Scenario: User selects a preset
- **WHEN** the user enters a prompt, selects a preset card, and submits
- **THEN** the system creates the task with that preset and navigates directly to the task dashboard

#### Scenario: User selects custom composition
- **WHEN** the user selects custom composition and submits the prompt
- **THEN** the system opens the composition workspace with an editable custom outline and no required preset-selection step

### Requirement: Dashboard composer has one responsibility
The Dashboard conversation composer SHALL create free-form tasks or continue existing tasks without exposing a second template checkbox or template dropdown.

#### Scenario: User starts from dashboard
- **WHEN** no task is selected and the user submits a prompt from Dashboard
- **THEN** the system creates a free-form task without showing template selection controls

### Requirement: Template content is completed inside the main Agent run
The task service MUST NOT run a separate model-based template-fill phase before creating a task. The main Agent SHALL accept both blank template fields and user-populated template fields during its normal generation run.

#### Scenario: Preset scaffold contains blank or example content
- **WHEN** a preset or recommended template is selected
- **THEN** the task starts immediately with the template structure and the main Agent rewrites or fills its content around the user prompt before slide execution

#### Scenario: Custom outline contains user text and blanks
- **WHEN** a custom outline contains some non-empty user-authored fields and some empty fields
- **THEN** the main Agent preserves the non-empty user content as constraints and fills only the missing content in the same run

#### Scenario: Template is already fully populated
- **WHEN** a custom outline already contains complete titles, descriptions, and optional content plans
- **THEN** the main Agent proceeds to slide generation without invoking a separate fill endpoint or replacing the supplied content
