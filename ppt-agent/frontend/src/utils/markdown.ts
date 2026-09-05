export type MarkdownInline =
  | { type: 'text'; value: string }
  | { type: 'link'; label: string; href: string }
  | { type: 'image'; alt: string; src: string }
  | { type: 'code'; value: string }
  | { type: 'strong'; value: string }
  | { type: 'break' }

export type MarkdownBlock =
  | { type: 'heading'; level: 1 | 2 | 3; content: MarkdownInline[] }
  | { type: 'paragraph'; content: MarkdownInline[] }
  | { type: 'list'; items: MarkdownInline[][] }
  | { type: 'image'; alt: string; src: string }

function safeURL(value: string) {
  try {
    const url = new URL(value)
    return ['http:', 'https:'].includes(url.protocol) ? url.href : ''
  } catch {
    return ''
  }
}

// A previously persisted stream can contain a synthetic space inside a
// Markdown URL (for example `h ttps://…`). Whitespace is never meaningful in
// a Markdown link destination, so recover it before applying URL allowlisting.
function safeMarkdownDestination(value: string) {
	return safeURL(value.replace(/\s+/g, ''))
}

function appendText(target: MarkdownInline[], value: string) {
  if (!value) return
  const previous = target[target.length - 1]
  if (previous?.type === 'text') previous.value += value
  else target.push({ type: 'text', value })
}

export function parseMarkdownInline(source: string): MarkdownInline[] {
  const tokens: MarkdownInline[] = []
  let index = 0
  while (index < source.length) {
    // Keep this scanner linear in the input length. The previous implementation
    // ran several anchored regexes against `source.slice(index)` for every
    // character, making long streamed answers increasingly expensive to render.
    const imageStart = source.startsWith('![', index)
    const linkStart = source[index] === '['
    if (imageStart || linkStart) {
      const labelStart = index + (imageStart ? 2 : 1)
      const labelEnd = source.indexOf(']', labelStart)
      if (labelEnd >= 0 && source[labelEnd + 1] === '(') {
        const destinationEnd = source.indexOf(')', labelEnd + 2)
        if (destinationEnd >= 0) {
          const label = source.slice(labelStart, labelEnd)
          const destination = source.slice(labelEnd + 2, destinationEnd).trim()
          const safeDestination = safeMarkdownDestination(destination)
          if (safeDestination) {
            tokens.push(imageStart
              ? { type: 'image', alt: label || '图片参考', src: safeDestination }
              : { type: 'link', label, href: safeDestination })
            index = destinationEnd + 1
            continue
          }
        }
      }
    }
    if (source[index] === '`') {
      const end = source.indexOf('`', index + 1)
      if (end > index + 1) {
        tokens.push({ type: 'code', value: source.slice(index + 1, end) })
        index = end + 1
        continue
      }
    }
    if (source.startsWith('**', index)) {
      const end = source.indexOf('**', index + 2)
      if (end > index + 2) {
        tokens.push({ type: 'strong', value: source.slice(index + 2, end) })
        index = end + 2
        continue
      }
    }
    // A streamed Markdown media/link token is not valid until its closing
    // parenthesis arrives. Preserve the whole partial token as text instead
    // of turning only the URL suffix into a link and hiding the syntax.
    if (imageStart || linkStart) {
      const labelStart = index + (imageStart ? 2 : 1)
      const labelEnd = source.indexOf(']', labelStart)
      if (labelEnd >= 0 && source[labelEnd + 1] === '(' && source.indexOf(')', labelEnd + 2) < 0) {
        appendText(tokens, source.slice(index))
        break
      }
    }
    if (source.startsWith('http://', index) || source.startsWith('https://', index)) {
      let end = index
      while (end < source.length && !/[\s<>'"]/.test(source[end])) end += 1
      const value = source.slice(index, end)
      const trimmed = value.replace(/[.,;:!?]+$/, '')
      const href = safeURL(trimmed)
      if (href) {
        tokens.push({ type: 'link', label: trimmed, href })
        appendText(tokens, value.slice(trimmed.length))
        index = end
        continue
      }
    }
    appendText(tokens, source[index])
    index += 1
  }
  return tokens
}

function isOnlyImage(content: MarkdownInline[]): content is [{ type: 'image'; alt: string; src: string }] {
  return content.length === 1 && content[0].type === 'image'
}

export function parseMarkdown(source: string): MarkdownBlock[] {
  const blocks: MarkdownBlock[] = []
  const paragraph: string[] = []
  const list: string[] = []
  const flushParagraph = () => {
    if (!paragraph.length) return
    const content: MarkdownInline[] = []
    paragraph.forEach((line, index) => {
      if (index) content.push({ type: 'break' })
      content.push(...parseMarkdownInline(line))
    })
    if (isOnlyImage(content)) blocks.push(content[0])
    else blocks.push({ type: 'paragraph', content })
    paragraph.length = 0
  }
  const flushList = () => {
    if (!list.length) return
    blocks.push({ type: 'list', items: list.map(parseMarkdownInline) })
    list.length = 0
  }

  for (const rawLine of (source || '').replace(/\r/g, '').split('\n')) {
    const line = rawLine.trimEnd()
    if (!line.trim()) {
      flushParagraph()
      flushList()
      continue
    }
    const heading = line.match(/^(#{1,3})\s+(.+)$/)
    if (heading) {
      flushParagraph()
      flushList()
      blocks.push({ type: 'heading', level: heading[1].length as 1 | 2 | 3, content: parseMarkdownInline(heading[2]) })
      continue
    }
    const item = line.match(/^[-*]\s+(.+)$/)
    if (item) {
      flushParagraph()
      list.push(item[1])
      continue
    }
    flushList()
    const content = parseMarkdownInline(line)
    if (isOnlyImage(content)) {
      flushParagraph()
      blocks.push(content[0])
    } else paragraph.push(line)
  }
  flushParagraph()
  flushList()
  return blocks
}
