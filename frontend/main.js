'use strict';

// Wails runtime is injected at /wails/ipc.js
// All business logic lives in Go — this file is display-only glue.

// ── Busy overlay ──────────────────────────────────
const SPINNER_FRAMES = ['⣾','⣽','⣻','⢿','⡿','⣟','⣯','⣷'];
let spinnerInterval = null;

function showBusy() {
  const overlay = document.getElementById('busy-overlay');
  overlay.classList.add('visible');
  let f = 0;
  spinnerInterval = setInterval(() => {
    document.getElementById('busy-spinner').textContent = SPINNER_FRAMES[f++ % SPINNER_FRAMES.length];
  }, 80);
}

function hideBusy() {
  clearInterval(spinnerInterval);
  spinnerInterval = null;
  document.getElementById('busy-overlay').classList.remove('visible');
}

let currentTheme = localStorage.getItem('vibemd-theme') || 'dark';

function applyTheme(theme) {
  currentTheme = theme;
  document.body.className = theme;
  const btn = document.getElementById('btn-theme');
  btn.textContent = theme.toUpperCase();
  btn.classList.toggle('active', theme === 'light');
  localStorage.setItem('vibemd-theme', theme);
}

function onMarkdownRendered(data) {
  hideBusy();
  const content = document.getElementById('content');
  content.innerHTML = data.html;

  const path = data.path || '';
  const displayPath = path.length > 50 ? path.slice(0, 50) + '…' : path;
  const titleLabel = data.filename
    ? `${data.filename} (${displayPath})`
    : '';
  document.getElementById('titlebar-filename').textContent = titleLabel;
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
  const btn = document.getElementById('btn-linenum');
  btn.textContent = on ? 'LN ON' : '# LN';
  btn.classList.toggle('active', on);
  localStorage.setItem('vibemd-linenum', on);
  assignLineNumbers();
}

document.getElementById('btn-linenum').addEventListener('click', () => {
  applyLineNumbers(!lineNumbersOn);
});

// ── Help dialog ───────────────────────────────────
document.getElementById('btn-help').addEventListener('click', () => {
  document.getElementById('help-overlay').removeAttribute('hidden');
});

document.getElementById('btn-help-close').addEventListener('click', () => {
  document.getElementById('help-overlay').setAttribute('hidden', '');
});

document.getElementById('help-overlay').addEventListener('click', (e) => {
  if (e.target === e.currentTarget) {
    e.currentTarget.setAttribute('hidden', '');
  }
});

document.getElementById('help-github').addEventListener('click', (e) => {
  e.preventDefault();
  window.runtime.BrowserOpenURL('https://github.com/cyuvop/vibemd');
});

// ── Manual refresh ────────────────────────────────
document.getElementById('btn-refresh').addEventListener('click', () => {
  showBusy();
  window.go.main.App.Refresh().catch(hideBusy);
});

// Boot
window.addEventListener('load', () => {
  applyTheme(currentTheme);

  // Register event listeners first, then tell Go we're ready.
  // Go may have tried to emit during startup before JS loaded — Ready() flushes that.
  window.runtime.EventsOn('render:start', showBusy);
  window.runtime.EventsOn('markdown:rendered', onMarkdownRendered);
  window.runtime.EventsOn('theme:changed', applyTheme);

  applyLineNumbers(lineNumbersOn);
  window.go.main.App.Ready();
});
