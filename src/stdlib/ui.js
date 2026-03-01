function element(tag, props = {}, children = []) {
  return { __airtType: 'element', tag, props, children };
}

function text(value) {
  return { __airtType: 'text', value: String(value) };
}

function renderToHtml(node) {
  if (node == null) return '';
  if (Array.isArray(node)) return node.map(renderToHtml).join('');
  if (typeof node === 'string' || typeof node === 'number' || typeof node === 'boolean') return String(node);
  if (node.__airtType === 'text') return escapeHtml(node.value);
  if (node.__airtType === 'element') {
    const attrs = Object.entries(node.props || {})
      .map(([k, v]) => {
        if (k === 'style' && typeof v === 'object') {
          const style = Object.entries(v).map(([sk, sv]) => `${sk}:${sv}`).join(';');
          return `style="${escapeAttr(style)}"`;
        }
        if (k.startsWith('on')) return '';
        return `${k}="${escapeAttr(v)}"`;
      })
      .filter(Boolean)
      .join(' ');
    const start = attrs ? `<${node.tag} ${attrs}>` : `<${node.tag}>`;
    return `${start}${(node.children || []).map(renderToHtml).join('')}</${node.tag}>`;
  }
  return '';
}

function escapeHtml(v) { return String(v).replace(/[&<>]/g, (m) => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;' }[m])); }
function escapeAttr(v) { return String(v).replace(/[&<>"]/g, (m) => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;' }[m])); }

module.exports = { element, text, renderToHtml };
