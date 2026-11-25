// ============================================
// Abhijeet — Backend Engineer Portfolio
// Interactive JavaScript
// ============================================

document.addEventListener('DOMContentLoaded', () => {
    initClickCounter();
    initClock();
    initUptime();
    initSmoothScroll();
    initNavHighlight();
});

// --- Click Counter ---
function initClickCounter() {
    const btn = document.getElementById('clickBtn');
    const countEl = document.getElementById('clickCount');
    let count = parseInt(localStorage.getItem('clickCount') || '0', 10);

    countEl.textContent = count;
    updateClickText(count);

    btn.addEventListener('click', () => {
        count++;
        countEl.textContent = count;
        localStorage.setItem('clickCount', count);
        updateClickText(count);

        // Ripple effect
        btn.style.transform = 'scale(0.94)';
        setTimeout(() => {
            btn.style.transform = 'scale(1)';
        }, 100);
    });
}

function updateClickText(count) {
    const countEl = document.getElementById('clickCount');
    const parent = countEl.parentElement;
    parent.textContent = '';
    parent.appendChild(document.createTextNode("you've clicked "));
    const span = document.createElement('span');
    span.id = 'clickCount';
    span.textContent = count;
    span.style.color = 'var(--accent)';
    span.style.fontWeight = '600';
    parent.appendChild(span);
    parent.appendChild(document.createTextNode(count === 1 ? ' time' : ' times'));
}

// --- Live Clock (IST) ---
function initClock() {
    function updateTime() {
        const now = new Date();
        const ist = new Date(now.toLocaleString('en-US', { timeZone: 'Asia/Kolkata' }));
        const hours = String(ist.getHours()).padStart(2, '0');
        const minutes = String(ist.getMinutes()).padStart(2, '0');
        const seconds = String(ist.getSeconds()).padStart(2, '0');
        const timeStr = `☀ ${hours}:${minutes}:${seconds}`;

        const el = document.getElementById('current-time');
        if (el) el.textContent = timeStr;
    }

    updateTime();
    setInterval(updateTime, 1000);
}

// --- Uptime Counter ---
function initUptime() {
    const startTime = Date.now();

    function updateUptime() {
        const elapsed = Date.now() - startTime;
        const hours = String(Math.floor(elapsed / 3600000)).padStart(1, '0');
        const minutes = String(Math.floor((elapsed % 3600000) / 60000)).padStart(2, '0');
        const seconds = String(Math.floor((elapsed % 60000) / 1000)).padStart(2, '0');

        const el = document.getElementById('uptime');
        if (el) el.textContent = `${hours}:${minutes}:${seconds}`;
    }

    updateUptime();
    setInterval(updateUptime, 1000);
}

// --- Smooth Scroll for Nav Links ---
function initSmoothScroll() {
    document.querySelectorAll('a[href^="#"]').forEach(anchor => {
        anchor.addEventListener('click', function (e) {
            const targetId = this.getAttribute('href');
            if (targetId === '#') return;

            const target = document.querySelector(targetId);
            if (target) {
                e.preventDefault();
                target.scrollIntoView({
                    behavior: 'smooth',
                    block: 'start'
                });
            }
        });
    });
}

// --- Active Nav Highlight on Scroll ---
function initNavHighlight() {
    const sections = document.querySelectorAll('section[id]');
    const navLinks = document.querySelectorAll('.nav-link');

    if (sections.length === 0 || navLinks.length === 0) return;

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
        rootMargin: '-20% 0px -80% 0px'
    });

    sections.forEach(section => observer.observe(section));
}
