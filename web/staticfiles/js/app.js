/**
 * Sistema Financeiro — Formatura Villa Lobos
 * JavaScript Interativo
 */

document.addEventListener('DOMContentLoaded', function () {

  // ── 1. SIDEBAR TOGGLE (mobile) ─────────────────────────────
  const sidebar   = document.getElementById('sidebar');
  const overlay   = document.getElementById('sidebarOverlay');
  const toggleBtn = document.getElementById('sidebarToggle');

  function openSidebar() {
    sidebar && sidebar.classList.add('open');
    overlay && overlay.classList.add('visible');
    document.body.style.overflow = 'hidden';
  }

  function closeSidebar() {
    sidebar && sidebar.classList.remove('open');
    overlay && overlay.classList.remove('visible');
    document.body.style.overflow = '';
  }

  toggleBtn && toggleBtn.addEventListener('click', openSidebar);
  overlay   && overlay.addEventListener('click', closeSidebar);

  // Close on ESC
  document.addEventListener('keydown', function (e) {
    if (e.key === 'Escape') closeSidebar();
  });

  // ── 2. ACTIVE SIDEBAR LINK ─────────────────────────────────
  const currentPath = window.location.pathname;
  document.querySelectorAll('.sidebar-link').forEach(function (link) {
    const href = link.getAttribute('href');
    if (href && href !== '#' && currentPath.startsWith(href) && href !== '/') {
      link.classList.add('active');
    } else if (href === '/' && currentPath === '/') {
      link.classList.add('active');
    }
  });

  // ── 3. ALERT AUTO-CLOSE & CLOSE BUTTON ────────────────────
  document.querySelectorAll('.alert').forEach(function (alert) {
    // Close button
    const btn = alert.querySelector('.alert-close');
    btn && btn.addEventListener('click', function () {
      dismissAlert(alert);
    });

    // Auto close after 5s
    setTimeout(function () { dismissAlert(alert); }, 5000);
  });

  function dismissAlert(el) {
    el.style.transition = 'opacity 0.3s ease, transform 0.3s ease';
    el.style.opacity    = '0';
    el.style.transform  = 'translateY(-6px)';
    setTimeout(function () { el.remove(); }, 300);
  }

  // ── 4. ANIMATE STAT CARD VALUES (countUp) ──────────────────
  document.querySelectorAll('[data-count-up]').forEach(function (el) {
    const raw   = el.dataset.countUp.replace(/\./g, '').replace(',', '.');
    const end   = parseFloat(raw) || 0;
    const isInt = Number.isInteger(end);
    const dur   = 900;
    const start = performance.now();

    function step(now) {
      const progress = Math.min((now - start) / dur, 1);
      const ease     = 1 - Math.pow(1 - progress, 3);
      const value    = end * ease;

      if (isInt) {
        el.textContent = Math.round(value).toLocaleString('pt-BR');
      } else {
        el.textContent = value.toLocaleString('pt-BR', { minimumFractionDigits: 2, maximumFractionDigits: 2 });
      }

      if (progress < 1) requestAnimationFrame(step);
    }

    // Trigger when element is visible
    const observer = new IntersectionObserver(function (entries) {
      if (entries[0].isIntersecting) {
        requestAnimationFrame(step);
        observer.disconnect();
      }
    }, { threshold: 0.3 });
    observer.observe(el);
  });

  // ── 5. ANIMATE PROGRESS BARS ───────────────────────────────
  document.querySelectorAll('.progress-fill[data-width]').forEach(function (bar) {
    const target = parseFloat(bar.dataset.width) || 0;
    bar.style.width = '0%';
    setTimeout(function () {
      bar.style.width = Math.min(target, 100) + '%';
    }, 200);
  });

  // ── 6. EXPENSE GROUP ACCORDION ────────────────────────────
  document.querySelectorAll('.expense-group-header').forEach(function (header) {
    header.addEventListener('click', function () {
      const group = header.closest('.expense-group');
      group && group.classList.toggle('open');
    });
  });

  // Open first group by default
  const firstGroup = document.querySelector('.expense-group');
  firstGroup && firstGroup.classList.add('open');

  // ── 6b. SIDEBAR ACTIVE LINK ICON ANIMATION ─────────────────
  document.querySelectorAll('.sidebar-link.active i[class*="ph"]').forEach(function(icon) {
    icon.style.opacity = '1';
  });

  // ── 7. FORM VALIDATION ──────────────────────────────────────
  document.querySelectorAll('form').forEach(function (form) {
    form.addEventListener('submit', function (e) {
      let valid = true;
      form.querySelectorAll('input[required], select[required], textarea[required]').forEach(function (field) {
        if (!field.value.trim()) {
          valid = false;
          field.classList.add('error');
          field.addEventListener('input', function () {
            if (this.value.trim()) this.classList.remove('error');
          }, { once: true });
        }
      });
      if (!valid) {
        e.preventDefault();
        showToast('Preencha todos os campos obrigatórios.', 'error');
        return;
      }

      // Loading state on submit button
      const submitBtn = form.querySelector('[type="submit"]');
      if (submitBtn) {
        submitBtn.classList.add('btn-loading');
        submitBtn.disabled = true;
      }
    });
  });

  // ── 8. INSTALLMENT PREVIEW ──────────────────────────────────
  const amountInput  = document.getElementById('id_total_amount');
  const countInput   = document.getElementById('id_installment_count');
  const previewBox   = document.getElementById('installmentPreview');

  function updatePreview() {
    if (!amountInput || !countInput || !previewBox) return;
    const amount = parseFloat(amountInput.value) || 0;
    const count  = parseInt(countInput.value) || 1;
    if (amount > 0 && count > 0) {
      const per = (amount / count).toLocaleString('pt-BR', { minimumFractionDigits: 2 });
      previewBox.textContent = '📋 ' + count + ' parcela(s) de R$ ' + per;
      previewBox.classList.add('visible');
    } else {
      previewBox.classList.remove('visible');
    }
  }

  amountInput && amountInput.addEventListener('input', updatePreview);
  countInput  && countInput.addEventListener('input', updatePreview);

  // ── 9. TOAST NOTIFICATIONS ─────────────────────────────────
  window.showToast = function (message, type) {
    type = type || 'info';
    let container = document.querySelector('.toast-container');
    if (!container) {
      container = document.createElement('div');
      container.className = 'toast-container';
      document.body.appendChild(container);
    }

    const icons = {
      success: '<svg fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"/></svg>',
      error:   '<svg fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/></svg>',
      info:    '<svg fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/></svg>',
    };

    const toast = document.createElement('div');
    toast.className = 'toast ' + type;
    toast.innerHTML = '<span class="toast-icon">' + (icons[type] || icons.info) + '</span>' + message;
    container.appendChild(toast);

    setTimeout(function () {
      toast.classList.add('out');
      setTimeout(function () { toast.remove(); }, 260);
    }, 3500);
  };

  // ── 10. CONFIRM DANGER ACTIONS ─────────────────────────────
  document.querySelectorAll('[data-confirm]').forEach(function (el) {
    el.addEventListener('click', function (e) {
      if (!confirm(el.dataset.confirm)) e.preventDefault();
    });
  });

  // ── 11. AUTO-RESIZE TEXTAREA ───────────────────────────────
  document.querySelectorAll('textarea.form-input').forEach(function (ta) {
    ta.addEventListener('input', function () {
      this.style.height = 'auto';
      this.style.height = this.scrollHeight + 'px';
    });
  });

});
