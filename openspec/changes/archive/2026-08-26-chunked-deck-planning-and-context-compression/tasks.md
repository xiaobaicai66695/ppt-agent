## 1. Chunked Planning Core

- [x] 1.1 Add blueprint, section shard and section metadata types for deck planning.
- [x] 1.2 Implement deterministic section partitioning and merge helpers that fill engineering-managed task fields.
- [x] 1.3 Add blueprint and section planner prompts/agent constructors with section-scoped output instructions.
- [x] 1.4 Route eligible initial generation through blueprint + section shards + merge before the existing Task Reviewer.

## 2. Context Compression Visibility

- [x] 2.1 Add compressor callback plumbing for start/success/fallback compression events.
- [x] 2.2 Emit SSE progress/runtime metadata for context compression without exposing internal summary JSON as assistant chat.
- [x] 2.3 Lower or derive practical compression thresholds so planning calls compress before approaching the model window.
- [x] 2.4 Add frontend live-status wording for the `compressing_context` phase.

## 3. Tests And Documentation

- [x] 3.1 Add focused backend tests for section partitioning, merge validation and deterministic task fields.
- [x] 3.2 Add focused backend tests for compression event emission and hidden internal summaries.
- [x] 3.3 Run focused backend and frontend validation commands.
- [x] 3.4 Update iteration records with implementation and verification evidence.
