// ============================================
// Abhijeet — Backend Engineer Portfolio
// Interactive JavaScript
// ============================================

document.addEventListener('DOMContentLoaded', () => {
    initClock();
    initFooterClock();
    initSmoothScroll();
    initNavHighlight();
});

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
