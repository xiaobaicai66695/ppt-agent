## 1. Deck Style Contract

- [x] 1.1 Add explicit outline background policy fields to prompt template data and DeepAgent instruction construction
- [x] 1.2 Update master planning prompt to preserve the backend-selected theme and background policy
- [x] 1.3 Update initial and continuation SlideExecutor prompts to use manifest theme, preserve empty backgrounds, and save exact output filenames
- [x] 1.4 Add focused prompt rendering tests for palette and background invariants

## 2. Artifact Reconciliation

- [x] 2.1 Normalize a unique page-prefixed PPTX to the manifest output filename during reconciliation
- [x] 2.2 Rename an existing same-stem JPG alongside the normalized PPTX and reject ambiguous candidates
- [x] 2.3 Add filesystem tests for exact, unique drift, thumbnail, and ambiguous candidate cases

## 3. Legacy Delivery Compatibility

- [x] 3.1 Use the actual ready artifact filename when frontend delivery merging uniquely matches a manifest page
- [x] 3.2 Add frontend unit tests for whitespace-drifted and ambiguous filenames

## 4. Verification And Delivery

- [x] 4.1 Run focused backend and frontend tests, builds, and strict OpenSpec validation
- [x] 4.2 Deploy changed files, normalize the reported task artifacts, restart the service, and verify online previews
- [x] 4.3 Record the iteration outcome and mark TODO/OpenSpec tasks complete
