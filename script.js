// ============================================
// Abhijeet — Backend Engineer Portfolio
// Interactive JavaScript
// ============================================

document.addEventListener('DOMContentLoaded', () => {
    initClock();
    initFooterClock();
    initSmoothScroll();
    initNavHighlight();
    initCareerTabs();
    initGitHubData();
});

// --- Live Clock (IST) ---
function initClock() {
    function updateTime() {
        const now = new Date();
        const ist = new Date(now.toLocaleString('en-US', { timeZone: 'Asia/Kolkata' }));
        const hours = String(ist.getHours()).padStart(2, '0');
        const minutes = String(ist.getMinutes()).padStart(2, '0');
        const seconds = String(ist.getSeconds()).padStart(2, '0');
        const timeStr = `${hours}:${minutes}:${seconds}`;

        const el = document.getElementById('current-time');
        if (el) el.textContent = timeStr;
    }

    updateTime();
    setInterval(updateTime, 1000);
}

// --- Footer Live Clock ---
function initFooterClock() {
    function updateFooterTime() {
        const now = new Date();
        const ist = new Date(now.toLocaleString('en-US', { timeZone: 'Asia/Kolkata' }));
        const hours = String(ist.getHours()).padStart(2, '0');
        const minutes = String(ist.getMinutes()).padStart(2, '0');
        const seconds = String(ist.getSeconds()).padStart(2, '0');

        const el = document.getElementById('footer-time');
        if (el) el.textContent = `${hours}:${minutes}:${seconds}`;
    }

    updateFooterTime();
    setInterval(updateFooterTime, 1000);
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
                scrollContainer.scrollTo({
                    top: offsetTop,
                    behavior: 'smooth'
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
const GITHUB_USER = 'abhijeetydv';
const GITHUB_REPO = 'ay_x';

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
    await Promise.all([
        fetchLatestRepoCommit(),
        fetchRecentActivity(),
        fetchLanguageBreakdown(),
    ]);
}

async function fetchLatestRepoCommit() {
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
