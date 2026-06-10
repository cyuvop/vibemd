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
  reapplySearch();
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
  if (mod && e.key === 'f') { e.preventDefault(); openSearch(); }
  if (mod && e.key === 't') { e.preventDefault(); toggleTheme(); }
  if (mod && e.key === 'w') { e.preventDefault(); window.go.main.App.GetFilePath().then(() => window.close()); }
});

function toggleTheme() {
  applyTheme(currentTheme === 'dark' ? 'light' : 'dark');
}

document.getElementById('btn-theme').addEventListener('click', toggleTheme);

// ── Title bar drag ────────────────────────────────
// Both -webkit-app-region:drag and WindowStartDragging() have a WKWebView
// bug where they stop working after the first drag. Manual drag via
// WindowGetPosition / WindowSetPosition is the only reliable approach.
{
  let dragging = false;
  let startScreenX = 0, startScreenY = 0;
  let winX = 0, winY = 0;

  document.querySelector('.titlebar').addEventListener('mousedown', async (e) => {
    if (e.button !== 0 || e.target.closest('button')) return;
    e.preventDefault();
    const pos = await window.runtime.WindowGetPosition();
    winX = pos.x;
    winY = pos.y;
    startScreenX = e.screenX;
    startScreenY = e.screenY;
    dragging = true;
  });

  window.addEventListener('mousemove', (e) => {
    if (!dragging) return;
    window.runtime.WindowSetPosition(
      winX + (e.screenX - startScreenX),
      winY + (e.screenY - startScreenY)
    );
  });

  window.addEventListener('mouseup', () => { dragging = false; });
}


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

// ── Search ────────────────────────────────────────
let searchMatches = [];
let searchIndex  = -1;
let lastSearchTerm = '';
let searchDebounce = null;

function openSearch() {
  document.getElementById('search-bar').removeAttribute('hidden');
  const input = document.getElementById('search-input');
  input.focus();
  input.select();
}

function closeSearch() {
  document.getElementById('search-bar').setAttribute('hidden', '');
  clearMarks();
  searchMatches = [];
  searchIndex = -1;
  lastSearchTerm = '';
  document.getElementById('search-input').value = '';
  document.getElementById('search-count').textContent = '';
}

function clearMarks() {
  document.querySelectorAll('.search-match').forEach(m => {
    const parent = m.parentNode;
    if (parent) {
      parent.replaceChild(document.createTextNode(m.textContent), m);
      parent.normalize();
    }
  });
}

function escapeRegex(s) {
  return s.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
}

function runSearch(term) {
  clearMarks();
  searchMatches = [];
  searchIndex = -1;
  lastSearchTerm = term;

  if (!term) { updateSearchCount(); return; }

  const content = document.getElementById('content');
  const re = new RegExp(escapeRegex(term), 'gi');

  // Collect text nodes first (can't mutate DOM while walking)
  const textNodes = [];
  const walker = document.createTreeWalker(content, NodeFilter.SHOW_TEXT);
  let node;
  while ((node = walker.nextNode())) {
    if (node.textContent.toLowerCase().includes(term.toLowerCase())) {
      textNodes.push(node);
    }
  }

  textNodes.forEach(textNode => {
    const text = textNode.textContent;
    const fragment = document.createDocumentFragment();
    let last = 0, match;
    re.lastIndex = 0;

    while ((match = re.exec(text)) !== null) {
      if (match.index > last)
        fragment.appendChild(document.createTextNode(text.slice(last, match.index)));
      const mark = document.createElement('mark');
      mark.className = 'search-match';
      mark.textContent = match[0];
      fragment.appendChild(mark);
      searchMatches.push(mark);
      last = match.index + match[0].length;
    }
    if (last < text.length)
      fragment.appendChild(document.createTextNode(text.slice(last)));

    textNode.parentNode.replaceChild(fragment, textNode);
  });

  if (searchMatches.length > 0) { searchIndex = 0; activateCurrent(); }
  updateSearchCount();
}

function activateCurrent() {
  searchMatches.forEach((m, i) => m.classList.toggle('current', i === searchIndex));
  if (searchMatches[searchIndex])
    searchMatches[searchIndex].scrollIntoView({ behavior: 'smooth', block: 'center' });
}

function searchNext() {
  if (!searchMatches.length) return;
  searchIndex = (searchIndex + 1) % searchMatches.length;
  activateCurrent();
  updateSearchCount();
}

function searchPrev() {
  if (!searchMatches.length) return;
  searchIndex = (searchIndex - 1 + searchMatches.length) % searchMatches.length;
  activateCurrent();
  updateSearchCount();
}

function updateSearchCount() {
  const el = document.getElementById('search-count');
  if (!lastSearchTerm)          { el.textContent = ''; return; }
  if (!searchMatches.length)    { el.textContent = 'NO MATCH'; return; }
  el.textContent = `${searchIndex + 1} / ${searchMatches.length}`;
}

// Re-apply search after content is re-rendered
function reapplySearch() {
  if (lastSearchTerm && !document.getElementById('search-bar').hidden) {
    runSearch(lastSearchTerm);
  }
}

document.getElementById('search-input').addEventListener('input', (e) => {
  clearTimeout(searchDebounce);
  searchDebounce = setTimeout(() => runSearch(e.target.value.trim()), 120);
});

document.getElementById('search-input').addEventListener('keydown', (e) => {
  if (e.key === 'Enter')  { e.preventDefault(); e.shiftKey ? searchPrev() : searchNext(); }
  if (e.key === 'Escape') { closeSearch(); }
});

document.getElementById('search-next').addEventListener('click', searchNext);
document.getElementById('search-prev').addEventListener('click', searchPrev);
document.getElementById('search-close').addEventListener('click', closeSearch);

// ── Help dialog ───────────────────────────────────
document.getElementById('btn-help').addEventListener('click', () => {
  document.getElementById('help-overlay').removeAttribute('hidden');
  window.go.main.App.GetVersion().then(v => {
    document.getElementById('help-version').textContent = 'v' + v;
  });
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

  // Check for updates in the background — never blocks startup
  window.go.main.App.CheckForUpdate().then(info => {
    if (!info.hasUpdate) return;
    const btn = document.getElementById('btn-update');
    btn.title = `v${info.version} available — click to download`;
    btn.textContent = `▲ v${info.version}`;
    btn.removeAttribute('hidden');
    btn.addEventListener('click', () => {
      window.runtime.BrowserOpenURL(info.url);
    });
  }).catch(() => {}); // silently ignore network errors
});
