import type { MessageMode, MessageRoute } from '../api'

// The backend classifier is authoritative for ordinary chat. Manual PPT mode
// remains a deliberate override, so its submit path cannot be downgraded by a
// malformed or unavailable classification response.
export function shouldStartPPTGeneration(route: Pick<MessageRoute, 'intent' | 'action' | 'needs_confirmation'>, manualMode: MessageMode) {
  if (manualMode === 'pptagent') return true
  return route.intent === 'create' && route.action === 'prepare_create' && !route.needs_confirmation
}
