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
  const startDateInput = document.getElementById('id_start_date');
  const firstPaymentInput = document.getElementById('id_first_payment_date');

  function formatDateBR(dateStr) {
    if (!dateStr) return '';
    const parts = dateStr.split('-');
    return parts[2] + '/' + parts[1] + '/' + parts[0];
  }

  function formatCurrency(value) {
    return value.toLocaleString('pt-BR', { minimumFractionDigits: 2, maximumFractionDigits: 2 });
  }

  function updatePreview() {
    if (!amountInput || !countInput || !previewBox) return;
    const amount = parseFloat(amountInput.value) || 0;
    const count  = parseInt(countInput.value) || 1;
    if (amount > 0 && count > 0) {
      const per = (amount / count);
      let html = '<strong>📋 Resumo:</strong><br>';
      html += count + ' parcela(s) de R$ ' + formatCurrency(per) + '<br><br>';
      html += '<strong>📅 Cronograma:</strong><br>';
      
      let startDate;
      let firstPaymentDate;
      let monthsInterval = 1;
      
      if (startDateInput && startDateInput.value) {
        startDate = new Date(startDateInput.value);
      } else {
        startDate = new Date();
      }
      
      if (firstPaymentInput && firstPaymentInput.value) {
        firstPaymentDate = new Date(firstPaymentInput.value);
      } else {
        firstPaymentDate = new Date(startDate);
      }
      
      // Calculate interval same as backend (Go)
      const diffMs = firstPaymentDate - startDate;
      const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24));
      const diffMonths = Math.floor(diffDays / 30);
      if (diffMonths > 0) {
        monthsInterval = diffMonths;
      }
      
      for (let i = 0; i < Math.min(count, 12); i++) {
        const paymentDate = new Date(firstPaymentDate);
        paymentDate.setMonth(paymentDate.getMonth() + (i * monthsInterval));
        html += 'Parcela ' + (i + 1) + ': ' + formatDateBR(paymentDate.toISOString().split('T')[0]) + '<br>';
      }
      
      if (count > 12) {
        html += '... e mais ' + (count - 12) + ' parcelas';
      }
      
      previewBox.innerHTML = html;
      previewBox.classList.add('visible');
    } else {
      previewBox.classList.remove('visible');
    }
  }

  amountInput && amountInput.addEventListener('input', updatePreview);
  countInput  && countInput.addEventListener('input', updatePreview);
  startDateInput && startDateInput.addEventListener('change', updatePreview);
  firstPaymentInput && firstPaymentInput.addEventListener('change', updatePreview);

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
