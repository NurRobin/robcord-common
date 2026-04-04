/** @type {Record<string, string>} */
const colors = {
  'bg-darkest': 'var(--bg-darkest)',
  'bg-dark': 'var(--bg-dark)',
  'bg-main': 'var(--bg-main)',
  'bg-elevated': 'var(--bg-elevated)',
  'bg-floating': 'var(--bg-floating)',

  'text-primary': 'var(--text-primary)',
  'text-secondary': 'var(--text-secondary)',
  'text-muted': 'var(--text-muted)',
  'text-link': 'var(--text-link)',

  accent: 'var(--accent)',
  'accent-hover': 'var(--accent-hover)',
  'accent-glow': 'var(--accent-glow)',

  green: 'var(--green)',
  yellow: 'var(--yellow)',
  red: 'var(--red)',
  'red-hover': 'var(--red-hover)',

  'border-subtle': 'var(--border-subtle)',
  'border-medium': 'var(--border-medium)',
  'border-strong': 'var(--border-strong)',

  'overlay-dim': 'var(--overlay-dim)',
  'overlay-heavy': 'var(--overlay-heavy)',

  // Semantic aliases
  panel: 'var(--bg-elevated)',
  'accent-soft': 'var(--accent)',
  'text-strong': 'var(--text-primary)',
};

module.exports = { colors };
