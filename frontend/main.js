'use strict';

// Wails runtime is injected at /wails/ipc.js
// All business logic lives in Go — this file is display-only glue (~60 lines).

let currentTheme = localStorage.getItem('vibemd-theme') || 'dark';

function applyTheme(theme) {
  currentTheme = theme;
  document.body.className = theme;
  document.getElementById('status-theme').textContent = theme.toUpperCase();
  localStorage.setItem('vibemd-theme', theme);
}

function onMarkdownRendered(data) {
  const content = document.getElementById('content');
  content.innerHTML = data.html;
  document.getElementById('titlebar-filename').textContent = data.filename || '';
  document.getElementById('status-filename').textContent = data.filename || '—';
  document.getElementById('status-words').textContent =
    (data.wordCount || 0) + ' words';
  document.title = 'VIBEMD > ' + (data.filename || '');
  window.scrollTo(0, 0);
  assignLineNumbers();
}

function scrollToHeading(text) {
  const headings = document.querySelectorAll('h1,h2,h3,h4,h5,h6');
  for (const h of headings) {
    if (h.textContent.includes(text)) {
      h.scrollIntoView({ behavior: 'smooth' });
      return;
    }
  }
}

// Keyboard shortcuts
document.addEventListener('keydown', (e) => {
  const mod = e.metaKey || e.ctrlKey;
  if (mod && e.key === 't') { e.preventDefault(); toggleTheme(); }
  if (mod && e.key === 'w') { e.preventDefault(); window.go.main.App.GetFilePath().then(() => window.close()); }
});

function toggleTheme() {
  applyTheme(currentTheme === 'dark' ? 'light' : 'dark');
}

document.getElementById('btn-theme').addEventListener('click', toggleTheme);

// ── Line numbers ──────────────────────────────────
let lineNumbersOn = localStorage.getItem('vibemd-linenum') === 'true';

function assignLineNumbers() {
  const content = document.getElementById('content');
  // Remove any existing numbers first
  content.querySelectorAll('.ln-num').forEach(s => s.remove());
  if (!lineNumbersOn) return;

  let n = 0;
  const stamp = (el) => {
    const s = document.createElement('span');
    s.className = 'ln-num';
    s.textContent = ++n;
    el.prepend(s);
  };

  for (const el of content.children) {
    const tag = el.tagName.toLowerCase();
    if (tag === 'hr') continue;                       // structural, not a content line
    if (tag === 'ul' || tag === 'ol') {
      el.querySelectorAll(':scope > li').forEach(stamp); // each bullet = its own number
    } else {
      stamp(el);
    }
  }
}

function applyLineNumbers(on) {
  lineNumbersOn = on;
  document.body.classList.toggle('line-numbers', on);
  document.getElementById('btn-linenum').classList.toggle('active', on);
  localStorage.setItem('vibemd-linenum', on);
  assignLineNumbers();
}

document.getElementById('btn-linenum').addEventListener('click', () => {
  applyLineNumbers(!lineNumbersOn);
});

// Boot
window.addEventListener('load', () => {
  applyTheme(currentTheme);

  // Register event listeners first, then tell Go we're ready.
  // Go may have tried to emit during startup before JS loaded — Ready() flushes that.
  window.runtime.EventsOn('markdown:rendered', onMarkdownRendered);
  window.runtime.EventsOn('theme:changed', applyTheme);

  applyLineNumbers(lineNumbersOn);
  window.go.main.App.Ready();
});
