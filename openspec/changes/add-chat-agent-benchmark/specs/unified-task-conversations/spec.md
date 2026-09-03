## ADDED Requirements

### Requirement: Chat replies integrate material and degrade honestly
The system SHALL answer a routed chat message directly. When verified web or image material is available, it SHALL synthesize the material and render only valid, deduplicated source references; when a requested capability is unavailable, it SHALL state that limitation without claiming a search or preview succeeded.

#### Scenario: User requests material for an existing topic
- **WHEN** a conversation about a travel project is followed by “补充些材料和图片参考” and usable web evidence is available
- **THEN** the chat reply summarizes topic-relevant material before listing valid web sources, and it only renders image references that were actually returned

#### Scenario: Generic route fallback with retrieval evidence
- **WHEN** the intent router supplies a generic chat fallback but usable retrieval evidence exists
- **THEN** the final chat reply uses evidence-backed content rather than returning only the generic fallback

#### Scenario: Invalid source from an augmentation
- **WHEN** an augmentation contains a blank, non-HTTP(S), malformed, or duplicate source URL
- **THEN** the final chat reply omits that source link while retaining any other valid source references
