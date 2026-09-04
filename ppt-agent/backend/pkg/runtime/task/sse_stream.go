package task

// normalizeAnswerChunk is the single boundary for provider answer chunks.
// Keeping the adapter here makes stream behavior easy to test without pulling
// in task persistence or lifecycle code.
func normalizeAnswerChunk(currentTurn, chunk string) string {
	return normalizeAnswerChunkLegacy(currentTurn, chunk)
}
