# Standalone PPT Visual Workflow

## Purpose

Define the provider-neutral visual policy, asset materialization, and rendering contract for the standalone ppt skill.

## Requirements

### Requirement: Default visual asset policy

The standalone ppt skill SHALL require a declared visual policy for new visual ppt plans. Unless the user explicitly requests a text-only ppt, the plan MUST require materialized local visual assets and state the minimum image-page coverage.

#### Scenario: Standard visual ppt

- **WHEN** an Agent plans a ppt without a user request to omit images
- **THEN** the plan contains a required visual policy and queryable visual intents for every non-exempt visual page

#### Scenario: Explicit text-only ppt

- **WHEN** the user explicitly requests no images
- **THEN** the plan declares the no-image policy and the standalone workflow does not require asset acquisition

### Requirement: Complete standalone asset materialization

The standalone asset resolver SHALL materialize both page-level visual intents and foreground `image` components before rendering, and SHALL write a readable local path plus provenance metadata back to the manifest.

#### Scenario: Background and foreground assets are planned

- **WHEN** a manifest contains a background visual intent and an `image` component with asset queries
- **THEN** the resolver downloads or reuses both assets and records local path, source URL, and attribution for each

#### Scenario: Planned asset cannot be resolved

- **WHEN** an asset query cannot be resolved or downloaded
- **THEN** the resolver fails without writing a partial manifest update

### Requirement: Rendering enforces visual materialization

The standalone validation and full-ppt rendering entry points SHALL reject a required visual plan containing unmaterialized assets before generating a PPTX.

#### Scenario: Caller skips acquisition

- **WHEN** a caller invokes validation or full-ppt rendering with a query-only required visual asset
- **THEN** the command exits unsuccessfully and identifies the unresolved asset

#### Scenario: Caller completes acquisition

- **WHEN** every required visual asset has a readable local file
- **THEN** validation and full-ppt rendering proceed normally
