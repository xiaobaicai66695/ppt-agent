import type { MessageRoute } from '../api'

export type ComposerMode = 'chat' | 'pptagent'

const deliveryDirectives = {
  pptagent: '（生成PPT）',
  web: '（网络搜索）',
  images: '（图片搜索）',
} as const

// Checkboxes become explicit, durable instructions in the same message that
// RouterAgent and downstream agents consume.
export function appendDeliveryDirectives(message: string, mode: ComposerMode, webSearch: boolean, imageSearch: boolean) {
  const directives = [
    mode === 'pptagent' ? deliveryDirectives.pptagent : '',
    webSearch ? deliveryDirectives.web : '',
    imageSearch ? deliveryDirectives.images : '',
  ].filter(directive => directive && !message.includes(directive))
  return directives.length ? `${message}\n\n${directives.join(' ')}` : message
}

// RouterAgent is authoritative for every route, including an explicit PPT
// generation instruction appended by the composer.
export function shouldStartPPTGeneration(route: Pick<MessageRoute, 'intent' | 'action' | 'needs_confirmation'>) {
  return route.intent === 'create' && route.action === 'prepare_create' && !route.needs_confirmation
}
