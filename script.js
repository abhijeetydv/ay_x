// ============================================
// Abhijeet — Backend Engineer Portfolio
// Interactive JavaScript
// ============================================

document.addEventListener('DOMContentLoaded', () => {
    initClock();
    initSmoothScroll();
    initNavHighlight();
    initNavPanel();
    initCareerTabs();
    initCodeCopy();
    initTocHighlight();
    initGitHubData();
});

const CONFIG_TIMEZONE = ((document.body && document.body.dataset.timezone) || '').trim();

// --- Live Clock ---
function initClock() {
    const el = document.getElementById('current-time');
    if (!el || !CONFIG_TIMEZONE) return;

    function updateTime() {
        try {
            const now = new Date();
            const ist = new Date(now.toLocaleString('en-US', { timeZone: CONFIG_TIMEZONE }));
            const hours = String(ist.getHours()).padStart(2, '0');
            const minutes = String(ist.getMinutes()).padStart(2, '0');
            const seconds = String(ist.getSeconds()).padStart(2, '0');
            el.textContent = `${hours}:${minutes}:${seconds}`;
        } catch {
            // Keep existing text if timezone is invalid
        }
    }

    updateTime();
    setInterval(updateTime, 1000);
}

// --- Footer Time-on-Site Timer ---
const SESSION_START_KEY = 'portfolio:sessionStart';
const SESSION_MAX_AGE_MS = 24 * 60 * 60 * 1000;

function formatElapsed(ms) {
    const totalSeconds = Math.max(0, Math.floor(ms / 1000));
    const hours = Math.floor(totalSeconds / 3600);
    const minutes = Math.floor((totalSeconds % 3600) / 60);
    const seconds = totalSeconds % 60;
    const mm = String(minutes).padStart(2, '0');
    const ss = String(seconds).padStart(2, '0');
    if (hours > 0) {
        return `${hours}:${mm}:${ss}`;
    }
    return `${mm}:${ss}`;
}

function getOrInitSessionStart(nowMs, storage) {
    if (storage) {
        try {
            const raw = storage.getItem(SESSION_START_KEY);
            if (raw !== null && raw !== '') {
                const parsed = Number(raw);
                if (Number.isFinite(parsed) && parsed <= nowMs && (nowMs - parsed) <= SESSION_MAX_AGE_MS) {
                    return parsed;
                }
            }
            storage.setItem(SESSION_START_KEY, String(nowMs));
        } catch {
            // Fall back to in-memory start time if sessionStorage is unavailable
        }
    }
    return nowMs;
}

function initFooterTimer() {
    const el = document.getElementById('footer-time');
    if (!el) return;

    let storage = null;
    try {
        storage = window.sessionStorage;
    } catch {
        storage = null;
    }

    const startMs = getOrInitSessionStart(Date.now(), storage);
    let timerId = null;

    function updateTimer() {
        const currentEl = document.getElementById('footer-time');
        if (!currentEl) {
            if (timerId !== null) clearInterval(timerId);
            return;
        }
        currentEl.textContent = formatElapsed(Date.now() - startMs);
    }

    updateTimer();
    timerId = setInterval(updateTimer, 1000);

    document.addEventListener('visibilitychange', () => {
        if (document.visibilityState === 'visible') {
            updateTimer();
        }
    });
    window.addEventListener('pageshow', updateTimer);
}

if (typeof document !== 'undefined') {
    initFooterTimer();
}

// --- Smooth Scroll for Nav Links ---
function initSmoothScroll() {
    const scrollContainer = document.querySelector('.scroll-fade-wrapper');
    document.querySelectorAll('a[href^="#"]').forEach(anchor => {
        anchor.addEventListener('click', function (e) {
            const targetId = this.getAttribute('href');
            if (targetId === '#') return;

            const target = document.querySelector(targetId);
            if (target && scrollContainer) {
                e.preventDefault();
                const offsetTop = target.offsetTop - scrollContainer.offsetTop;
                const reduced = typeof window !== 'undefined' &&
                    typeof window.matchMedia === 'function' &&
                    window.matchMedia('(prefers-reduced-motion: reduce)').matches;
                scrollContainer.scrollTo({
                    top: offsetTop,
                    behavior: reduced ? 'auto' : 'smooth'
                });
            }
        });
    });
}

// --- Active Nav Highlight on Scroll ---
function initNavHighlight() {
    const sections = document.querySelectorAll('section[id]');
    const navLinks = document.querySelectorAll('.nav-link');
    const scrollContainer = document.querySelector('.scroll-fade-wrapper');

    if (sections.length === 0 || navLinks.length === 0 || !scrollContainer) return;

    const observer = new IntersectionObserver((entries) => {
        entries.forEach(entry => {
            if (entry.isIntersecting) {
                navLinks.forEach(link => {
                    link.style.color = '';
                    if (link.getAttribute('href') === `#${entry.target.id}`) {
                        link.style.color = 'var(--accent)';
                    }
                });
            }
        });
    }, {
        root: scrollContainer,
        rootMargin: '-20% 0px -80% 0px'
    });

    sections.forEach(section => observer.observe(section));
}

// --- Work Experience & Education Tabs ---
function initCareerTabs() {
    const tabs = document.querySelectorAll('.career-tab');
    const panels = document.querySelectorAll('.career-panel');
    if (tabs.length === 0 || panels.length === 0) return;

    tabs.forEach((tab) => {
        tab.addEventListener('click', () => {
            const targetId = tab.getAttribute('aria-controls');

            tabs.forEach((t) => {
                const isSelected = t === tab;
                t.classList.toggle('active', isSelected);
                t.setAttribute('aria-selected', String(isSelected));
            });

            panels.forEach((panel) => {
                const show = panel.id === targetId;
                panel.classList.toggle('active', show);
                panel.hidden = !show;
            });
        });
    });
}

// --- GitHub Activity & Latest Commit ---
const GITHUB_USER = ((document.body && document.body.dataset.githubUser) || '').trim();
const GITHUB_REPO = ((document.body && document.body.dataset.githubRepo) || '').trim();

const LANG_COLORS = {
    JavaScript: '#f9e2af',
    TypeScript: '#89b4fa',
    Python: '#a6e3a1',
    Go: '#89dceb',
    HTML: '#fab387',
    CSS: '#cba6f7',
    Rust: '#f38ba8',
    Shell: '#94e2d5',
};

async function initGitHubData() {
    if (!GITHUB_USER) return;

    const tasks = [];
    if (document.getElementById('github-activity')) {
        tasks.push(fetchRecentActivity());
    }
    if (GITHUB_REPO) {
        if (document.getElementById('footer-commit-sha') && document.getElementById('footer-commit')) {
            tasks.push(fetchLatestRepoCommit());
        }
        if (document.getElementById('contrib-graph')) {
            tasks.push(fetchLanguageBreakdown());
        }
    }
    if (tasks.length === 0) return;
    await Promise.all(tasks);
}

async function fetchLatestRepoCommit() {
    if (!GITHUB_USER || !GITHUB_REPO) return;
    const shaEl = document.getElementById('footer-commit-sha');
    const linkEl = document.getElementById('footer-commit');
    if (!shaEl || !linkEl) return;

    try {
        const res = await fetch(
            `https://api.github.com/repos/${GITHUB_USER}/${GITHUB_REPO}/commits?per_page=1`
        );
        if (!res.ok) return;
        const commits = await res.json();
        if (Array.isArray(commits) && commits.length > 0) {
            const latest = commits[0];
            const shortSha = latest.sha.slice(0, 7);
            shaEl.textContent = shortSha;
            linkEl.href = latest.html_url;
            linkEl.title = `${shortSha}: ${latest.commit.message.split('\n')[0]}`;
        }
    } catch {
        // Keep fallback commit hash
    }
}

async function fetchRecentActivity() {
    if (!GITHUB_USER) return;
    const container = document.getElementById('github-activity');
    if (!container) return;

    try {
        const eventsRes = await fetch(
            `https://api.github.com/users/${GITHUB_USER}/events/public?per_page=30`
        );
        const items = [];

        if (eventsRes.ok) {
            const events = await eventsRes.json();
            for (const ev of events) {
                const repoShort = ev.repo.name.split('/').pop();
                if (ev.type === 'PushEvent' && ev.payload) {
                    if (Array.isArray(ev.payload.commits) && ev.payload.commits.length > 0) {
                        for (const c of [...ev.payload.commits].reverse()) {
                            items.push({
                                repo: repoShort,
                                msg: c.message.split('\n')[0],
                                sha: c.sha.slice(0, 7),
                                url: `https://github.com/${ev.repo.name}/commit/${c.sha}`,
                            });
                            if (items.length >= 3) break;
                        }
                    } else if (ev.payload.head) {
                        const sha = ev.payload.head.slice(0, 7);
                        items.push({
                            repo: repoShort,
                            msg: `pushed to ${ev.payload.ref ? ev.payload.ref.replace('refs/heads/', '') : 'branch'}`,
                            sha,
                            url: `https://github.com/${ev.repo.name}/commit/${ev.payload.head}`,
                        });
                    }
                } else if (ev.type === 'CreateEvent' && ev.payload) {
                    items.push({
                        repo: repoShort,
                        msg: `created ${ev.payload.ref_type}${ev.payload.ref ? ` ${ev.payload.ref}` : ''}`,
                        sha: 'new',
                        url: `https://github.com/${ev.repo.name}`,
                    });
                } else if (ev.type === 'PullRequestEvent' && ev.payload && ev.payload.pull_request) {
                    const pr = ev.payload.pull_request;
                    items.push({
                        repo: repoShort,
                        msg: `${ev.payload.action} PR #${pr.number}: ${pr.title}`,
                        sha: `#${pr.number}`,
                        url: pr.html_url,
                    });
                } else if (ev.type === 'IssuesEvent' && ev.payload && ev.payload.issue) {
                    const issue = ev.payload.issue;
                    items.push({
                        repo: repoShort,
                        msg: `${ev.payload.action} issue #${issue.number}: ${issue.title}`,
                        sha: `#${issue.number}`,
                        url: issue.html_url,
                    });
                } else if (ev.type === 'ReleaseEvent' && ev.payload && ev.payload.release) {
                    items.push({
                        repo: repoShort,
                        msg: `released ${ev.payload.release.tag_name}`,
                        sha: ev.payload.release.tag_name,
                        url: ev.payload.release.html_url,
                    });
                } else if (ev.type === 'ForkEvent') {
                    items.push({
                        repo: repoShort,
                        msg: 'forked repository',
                        sha: 'fork',
                        url: `https://github.com/${ev.repo.name}`,
                    });
                } else if (ev.type === 'WatchEvent') {
                    items.push({
                        repo: repoShort,
                        msg: 'starred repository',
                        sha: 'star',
                        url: `https://github.com/${ev.repo.name}`,
                    });
                }
                if (items.length >= 3) break;
            }
        }

        // Fill remaining slots from recently pushed repos across the user's account
        if (items.length < 3) {
            const reposRes = await fetch(
                `https://api.github.com/users/${GITHUB_USER}/repos?sort=pushed&per_page=5`
            );
            if (reposRes.ok) {
                const repos = await reposRes.json();
                const seenUrls = new Set(items.map((i) => i.url));
                const commitResults = await Promise.all(
                    repos.map(async (repo) => {
                        try {
                            const cRes = await fetch(
                                `https://api.github.com/repos/${repo.full_name}/commits?per_page=1`
                            );
                            if (!cRes.ok) return null;
                            const commits = await cRes.json();
                            if (!Array.isArray(commits) || commits.length === 0) return null;
                            const c = commits[0];
                            return {
                                repo: repo.name,
                                msg: c.commit.message.split('\n')[0],
                                sha: c.sha.slice(0, 7),
                                url: c.html_url,
                            };
                        } catch {
                            return null;
                        }
                    })
                );
                for (const entry of commitResults) {
                    if (entry && !seenUrls.has(entry.url)) {
                        items.push(entry);
                        seenUrls.add(entry.url);
                        if (items.length >= 3) break;
                    }
                }
            }
        }

        if (items.length > 0) {
            renderCommitItems(container, items);
        }
    } catch {
        // Keep fallback content
    }
}

function renderCommitItems(container, items) {
    container.innerHTML = '';
    for (const item of items) {
        const row = document.createElement('div');
        row.className = 'commit-item';

        const link = document.createElement('a');
        link.className = 'commit-item-link';
        link.href = item.url;
        link.target = '_blank';
        link.rel = 'noopener';

        const repoSpan = document.createElement('span');
        repoSpan.className = 'commit-repo';
        repoSpan.textContent = `${item.repo}:`;

        const msgSpan = document.createElement('span');
        msgSpan.className = 'commit-msg';
        msgSpan.textContent = item.msg;

        const statsSpan = document.createElement('span');
        statsSpan.className = 'commit-stats';

        const shaSpan = document.createElement('span');
        shaSpan.className = 'commit-sha';
        shaSpan.textContent = item.sha;

        statsSpan.appendChild(shaSpan);
        link.append(repoSpan, msgSpan, statsSpan);
        row.appendChild(link);
        container.appendChild(row);
    }
}

async function fetchLanguageBreakdown() {
    if (!GITHUB_USER || !GITHUB_REPO) return;
    const graphEl = document.getElementById('contrib-graph');
    if (!graphEl) return;

    try {
        const res = await fetch(
            `https://api.github.com/repos/${GITHUB_USER}/${GITHUB_REPO}/languages`
        );
        if (!res.ok) return;
        const langs = await res.json();
        const total = Object.values(langs).reduce((sum, n) => sum + n, 0);
        if (!total) return;

        graphEl.innerHTML = '';
        for (const [lang, bytes] of Object.entries(langs)) {
            const pct = ((bytes / total) * 100).toFixed(1);
            const bar = document.createElement('div');
            bar.className = 'contrib-bar';
            bar.style.width = `${pct}%`;
            bar.style.background = LANG_COLORS[lang] || '#8a8a8a';
            bar.title = `${lang}: ${pct}%`;
            graphEl.appendChild(bar);
        }
    } catch {
        // Keep fallback bars
    }
}

// --- Code Block Copy Button ---
async function copyTextToClipboard(text) {
    if (
        typeof navigator !== 'undefined' &&
        navigator.clipboard &&
        typeof navigator.clipboard.writeText === 'function' &&
        (typeof window === 'undefined' || window.isSecureContext !== false)
    ) {
        await navigator.clipboard.writeText(text);
        return true;
    }

    if (typeof document === 'undefined' || typeof document.createElement !== 'function' || !document.body) {
        return false;
    }
    const ta = document.createElement('textarea');
    ta.value = text;
    ta.setAttribute('readonly', '');
    ta.style.position = 'fixed';
    ta.style.opacity = '0';
    ta.style.pointerEvents = 'none';
    document.body.appendChild(ta);
    ta.select();
    let ok = false;
    try {
        ok = typeof document.execCommand === 'function' && document.execCommand('copy');
    } finally {
        document.body.removeChild(ta);
    }
    return ok;
}

async function handleCodeCopyClick(btn) {
    const block = btn && btn.closest ? btn.closest('.code-block') : null;
    if (!block) return false;
    const pre = block.querySelector('pre');
    if (!pre) return false;

    const copyLabel = (btn.dataset && btn.dataset.copyLabel) || '';
    const copiedLabel = (btn.dataset && btn.dataset.copiedLabel) || '';

    let ok = false;
    try {
        ok = await copyTextToClipboard(pre.textContent || '');
    } catch {
        return false;
    }
    if (!ok) return false;

    if (btn.dataset) {
        btn.dataset.copied = 'true';
    }
    if (copiedLabel) {
        btn.setAttribute('aria-label', copiedLabel);
    }
    const statusEl = document.getElementById('copy-status');
    if (statusEl && copiedLabel) {
        statusEl.textContent = copiedLabel;
    }
    if (btn._copyTimer) {
        clearTimeout(btn._copyTimer);
    }
    btn._copyTimer = setTimeout(() => {
        if (btn.dataset) {
            delete btn.dataset.copied;
        }
        if (copyLabel) {
            btn.setAttribute('aria-label', copyLabel);
        }
        if (statusEl) {
            statusEl.textContent = '';
        }
    }, 1500);
    return true;
}

function initCodeCopy() {
    if (!document.querySelector('.code-block')) return;

    document.addEventListener('click', (e) => {
        const btn = e.target && e.target.closest ? e.target.closest('.code-copy') : null;
        if (!btn) return;
        void handleCodeCopyClick(btn);
    });
}

// --- Post TOC Active Section Highlight ---
function selectActiveHeadingId(headings, offsetTop) {
    if (!Array.isArray(headings) || headings.length === 0) return null;
    let activeId = headings[0].id;
    for (const h of headings) {
        if (h.top <= offsetTop) {
            activeId = h.id;
        } else {
            break;
        }
    }
    return activeId;
}

function initTocHighlight() {
    const toc = document.querySelector('.toc');
    if (!toc || typeof IntersectionObserver !== 'function') return;

    const links = Array.from(toc.querySelectorAll('.toc-link[href^="#"]'));
    if (links.length === 0) return;

    const entries = [];
    for (const link of links) {
        const id = link.getAttribute('href').slice(1);
        const target = id ? document.getElementById(id) : null;
        if (target) {
            entries.push({ id, link, target });
        }
    }
    if (entries.length === 0) return;

    const scrollContainer = document.querySelector('.scroll-fade-wrapper');

    function setActive(activeId) {
        for (const entry of entries) {
            const isActive = entry.id === activeId;
            entry.link.classList.toggle('active', isActive);
            if (isActive) {
                entry.link.setAttribute('aria-current', 'location');
            } else {
                entry.link.removeAttribute('aria-current');
            }
        }
    }

    function updateFromPositions() {
        const containerRect = scrollContainer
            ? scrollContainer.getBoundingClientRect()
            : { top: 0, height: window.innerHeight || 800 };
        const offsetTop = containerRect.top + Math.min(180, containerRect.height * 0.25);
        const positions = entries.map((e) => ({
            id: e.id,
            top: e.target.getBoundingClientRect().top,
        }));
        const activeId = selectActiveHeadingId(positions, offsetTop);
        if (activeId) {
            setActive(activeId);
        }
    }

    const observer = new IntersectionObserver(
        () => {
            updateFromPositions();
        },
        {
            root: scrollContainer || null,
            rootMargin: '-20% 0px -70% 0px',
            threshold: [0, 1],
        }
    );

    for (const entry of entries) {
        observer.observe(entry.target);
    }
    updateFromPositions();
}

// --- Mobile Navigation Panel ---
function initNavPanel() {
    const toggle = document.querySelector('.nav-toggle');
    const panel = document.getElementById('nav-panel');
    const backdrop = document.getElementById('nav-panel-backdrop');
    if (!toggle || !panel || !backdrop) return;

    const closeBtn = panel.querySelector('.nav-panel-close');
    const panelLinks = panel.querySelectorAll('a');
    const backgroundTargets = [
        document.querySelector('.sticky-nav'),
        document.querySelector('.scroll-fade-wrapper'),
    ].filter(Boolean);

    let isOpen = false;
    let hideTimer = null;

    function getFocusableElements() {
        return Array.from(
            panel.querySelectorAll('button:not([disabled]), a[href], [tabindex]:not([tabindex="-1"])')
        ).filter((el) => !el.hidden);
    }

    function setBackgroundInert(inertValue) {
        for (const el of backgroundTargets) {
            el.inert = inertValue;
            if (inertValue) {
                el.setAttribute('inert', '');
            } else {
                el.removeAttribute('inert');
            }
        }
    }

    function openPanel() {
        if (isOpen) return;
        isOpen = true;
        if (hideTimer) {
            clearTimeout(hideTimer);
            hideTimer = null;
        }
        panel.hidden = false;
        backdrop.hidden = false;
        // Force reflow so the slide-in transition runs after removing [hidden].
        void panel.offsetWidth;
        panel.classList.add('is-open');
        backdrop.classList.add('is-open');
        toggle.setAttribute('aria-expanded', 'true');
        if (document.body) {
            document.body.style.overflow = 'hidden';
        }
        setBackgroundInert(true);
        const focusTarget = closeBtn || getFocusableElements()[0];
        if (focusTarget && typeof focusTarget.focus === 'function') {
            focusTarget.focus();
        }
    }

    function closePanel(restoreFocus = true) {
        if (!isOpen) return;
        isOpen = false;
        panel.classList.remove('is-open');
        backdrop.classList.remove('is-open');
        toggle.setAttribute('aria-expanded', 'false');
        if (document.body) {
            document.body.style.overflow = '';
        }
        setBackgroundInert(false);

        const reducedMotion =
            typeof window.matchMedia === 'function' &&
            window.matchMedia('(prefers-reduced-motion: reduce)').matches;

        const finishHide = () => {
            if (!isOpen) {
                panel.hidden = true;
                backdrop.hidden = true;
            }
        };

        if (reducedMotion) {
            finishHide();
        } else {
            hideTimer = setTimeout(finishHide, 280);
        }

        if (restoreFocus && typeof toggle.focus === 'function') {
            toggle.focus();
        }
    }

    toggle.addEventListener('click', () => {
        if (isOpen) {
            closePanel(true);
        } else {
            openPanel();
        }
    });

    if (closeBtn) {
        closeBtn.addEventListener('click', () => closePanel(true));
    }

    backdrop.addEventListener('click', () => closePanel(true));

    for (const link of panelLinks) {
        link.addEventListener('click', () => closePanel(false));
    }

    document.addEventListener('keydown', (e) => {
        if (!isOpen) return;
        if (e.key === 'Escape') {
            e.preventDefault();
            closePanel(true);
            return;
        }
        if (e.key === 'Tab') {
            const focusables = getFocusableElements();
            if (focusables.length === 0) return;
            const first = focusables[0];
            const last = focusables[focusables.length - 1];
            if (e.shiftKey && document.activeElement === first) {
                e.preventDefault();
                last.focus();
            } else if (!e.shiftKey && document.activeElement === last) {
                e.preventDefault();
                first.focus();
            }
        }
    });

    if (typeof window.matchMedia === 'function') {
        const mobileQuery = window.matchMedia('(max-width: 900px)');
        const onMediaChange = (e) => {
            if (!e.matches && isOpen) {
                closePanel(false);
            }
        };
        if (typeof mobileQuery.addEventListener === 'function') {
            mobileQuery.addEventListener('change', onMediaChange);
        }
    }

    window.addEventListener('pageshow', () => {
        if (isOpen) {
            closePanel(false);
        }
    });
}
