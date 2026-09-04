import type { MessageMode, MessageRoute } from '../api'

// The backend classifier is authoritative for ordinary chat. Manual PPT mode
// remains a deliberate override, so its submit path cannot be downgraded by a
// malformed or unavailable classification response.
export function shouldStartPPTGeneration(route: Pick<MessageRoute, 'intent'>, manualMode: MessageMode) {
  return manualMode === 'pptagent' || route.intent === 'create'
}
