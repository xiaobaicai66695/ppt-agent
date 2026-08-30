# Planner First-Draft Visual Quality

## Purpose

Define the boundary between visual semantic planning by the backend Planner and deterministic asset materialization, together with first-draft regression evidence.

## Requirements

### Requirement: Planner visual planning boundary

The backend Planner SHALL create complete visual asset semantics according to the standalone skill, but SHALL NOT be responsible for downloading assets during first-draft planning.

#### Scenario: Backend image service is configured

- **WHEN** the Planner runs with the backend image provider available
- **THEN** its instruction requires complete visual queries and delegates download to the deterministic materialization stage

#### Scenario: Backend image service is unavailable

- **WHEN** the Planner runs without the backend image provider
- **THEN** it still records valid visual semantics or the explicit no-image policy without advertising an unavailable download tool

### Requirement: Planner first-draft visual quality regression evidence

The project SHALL maintain automated first-draft quality checks that verify the Planner instruction retains the skill visual policy and that representative gold DeckSpecs satisfy deterministic review.

#### Scenario: Prompt contract changes

- **WHEN** the Planner instruction or skill visual workflow changes
- **THEN** focused tests verify that the Planner does not reintroduce conflicting image-download instructions

#### Scenario: Benchmark execution

- **WHEN** the Planner benchmark is run
- **THEN** it reports deterministic review outcomes for the selected first-draft cases and fails on a quality-gate regression
