import { describe, expect, it } from 'vitest'
import { parseMarkdown, parseMarkdownInline } from './markdown'

describe('Markdown rendering model', () => {
  it('keeps two image references as separate media blocks', () => {
    const blocks = parseMarkdown('### 图片参考\n![第一张](https://images.unsplash.com/first?ixid=1)\n摄影：[甲](https://unsplash.com/@a?utm_source=ppt_agent) · [Unsplash](https://unsplash.com/?utm_source=ppt_agent&utm_medium=referral)\n![第二张](https://images.unsplash.com/second?ixid=2)\n摄影：[乙](https://unsplash.com/@b?utm_source=ppt_agent)')
    expect(blocks.filter(block => block.type === 'image')).toHaveLength(2)
    const links = blocks.flatMap(block => block.type === 'paragraph' ? block.content : []).filter(token => token.type === 'link')
    expect(links).toHaveLength(3)
    expect(links.map(link => link.href)).toContain('https://unsplash.com/?utm_source=ppt_agent&utm_medium=referral')
  })

  it('does not join the following image markdown into a preceding URL', () => {
    const content = parseMarkdownInline('[Unsplash](https://unsplash.com/?utm_source=ppt_agent&utm_medium=referral)\n![第二张](https://images.unsplash.com/second)')
    expect(content.filter(token => token.type === 'link')).toHaveLength(1)
    expect(content.filter(token => token.type === 'image')).toHaveLength(1)
  })

  it('leaves an incomplete streamed image token as readable text until it is complete', () => {
    const partial = parseMarkdownInline('![第二张](https://images.unsplash.com/sec')
    expect(partial.some(token => token.type === 'image')).toBe(false)
    expect(partial.map(token => token.type === 'text' ? token.value : '').join('')).toContain('https://images.unsplash.com/sec')
  })

  it('repairs a legacy whitespace split inside Markdown destinations', () => {
    const content = parseMarkdown('![第一张](h ttps://images.unsplash.com/first?ixid=1)\n![第二张](h ttp://images.unsplash.com/second?ixid=2)\n[资料](h ttps://baijiahao.baidu.com/s?id=1&wfr=spider)')
    expect(content.filter(block => block.type === 'image')).toHaveLength(2)
    const paragraph = content.find(block => block.type === 'paragraph')
    expect(paragraph?.type).toBe('paragraph')
    if (paragraph?.type === 'paragraph') {
      expect(paragraph.content).toContainEqual({ type: 'link', label: '资料', href: 'https://baijiahao.baidu.com/s?id=1&wfr=spider' })
    }
  })
})
