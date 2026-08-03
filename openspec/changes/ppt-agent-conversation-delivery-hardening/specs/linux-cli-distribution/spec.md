## ADDED Requirements

### Requirement: Linux Shell Execution
The project CLI SHALL execute Agent shell commands through `/bin/sh -c` and SHALL not include a Windows shell execution branch.

#### Scenario: Agent executes a shell command
- **WHEN** the Linux CLI invokes the command operator
- **THEN** the command is executed by `/bin/sh`

### Requirement: Linux Python Default
The project CLI SHALL use a Linux Python executable default and SHALL allow `PYTHON_BIN` to override it.

#### Scenario: Python binary is not configured
- **WHEN** `PYTHON_BIN` is empty
- **THEN** the configured Linux virtual-environment Python path is returned

#### Scenario: Python binary is configured
- **WHEN** `PYTHON_BIN` is set
- **THEN** its value is used unchanged

### Requirement: No Windows CLI Distribution Contract
The project SHALL not publish or document `.exe`, `.bat`, `.cmd`, PowerShell, or Windows-specific CLI execution as supported deliverables.

#### Scenario: CLI distribution files are inspected
- **WHEN** release and project files are scanned
- **THEN** only Linux CLI entry and path conventions are documented as supported

