export function renderMd(text: string): string {
  let html = text
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;');

  html = html.replace(/^### (.+)$/gm, '<span class="md-h3">$1</span>');
  html = html.replace(/^## (.+)$/gm, '<span class="md-h2">$1</span>');
  html = html.replace(/^# (.+)$/gm, '<span class="md-h1">$1</span>');
  html = html.replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>');
  html = html.replace(/`([^`]+)`/g, '<code class="md-code">$1</code>');
  html = html.replace(/^---$/gm, '<span class="md-hr"></span>');
  html = html.replace(/((?:\/[\w.-]+)+\.\w+)/g, '<code class="md-path">$1</code>');
  return html;
}
