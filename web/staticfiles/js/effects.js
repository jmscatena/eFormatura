/**
 * Sistema de Controle de Formatura - Colégio Villa Lobos
 * Scripts de Interatividade e Efeitos Visuais
 */

// Função para criar partículas de fundo
function createParticles() {
    const container = document.createElement('div');
    container.className = 'particles-bg';
    document.body.appendChild(container);
    
    const particleCount = 50;
    
    for (let i = 0; i < particleCount; i++) {
        const particle = document.createElement('div');
        particle.className = 'particle';
        particle.style.left = Math.random() * 100 + '%';
        particle.style.top = Math.random() * 100 + '%';
        particle.style.animationDelay = Math.random() * 15 + 's';
        particle.style.animationDuration = (Math.random() * 10 + 10) + 's';
        container.appendChild(particle);
    }
}

// Função para adicionar efeito de ripple em botões
function addRippleEffect() {
    const buttons = document.querySelectorAll('.modern-btn');
    
    buttons.forEach(button => {
        button.addEventListener('click', function(e) {
            const rect = button.getBoundingClientRect();
            const x = e.clientX - rect.left;
            const y = e.clientY - rect.top;
            
            const ripple = document.createElement('span');
            ripple.style.cssText = `
                position: absolute;
                background: rgba(255, 255, 255, 0.3);
                border-radius: 50%;
                transform: scale(0);
                animation: ripple 0.6s linear;
                left: ${x}px;
                top: ${y}px;
                width: 100px;
                height: 100px;
                margin-left: -50px;
                margin-top: -50px;
            `;
            
            button.appendChild(ripple);
            
            setTimeout(() => ripple.remove(), 600);
        });
    });
}

// Função para animar elementos ao scroll
function animateOnScroll() {
    const elements = document.querySelectorAll('.animate-fade-up, .animate-fade-down, .animate-slide-in');
    
    const observer = new IntersectionObserver((entries) => {
        entries.forEach(entry => {
            if (entry.isIntersecting) {
                entry.target.style.opacity = '1';
                entry.target.style.transform = 'translateY(0)';
            }
        });
    }, {
        threshold: 0.1
    });
    
    elements.forEach(element => {
        element.style.opacity = '0';
        element.style.transform = 'translateY(20px)';
        element.style.transition = 'opacity 0.6s ease, transform 0.6s ease';
        observer.observe(element);
    });
}

// Função para adicionar efeito de hover em cards
function addCardHoverEffect() {
    const cards = document.querySelectorAll('.modern-card, .feature-card');
    
    cards.forEach(card => {
        card.addEventListener('mouseenter', function() {
            this.style.transform = 'translateY(-8px) scale(1.02)';
        });
        
        card.addEventListener('mouseleave', function() {
            this.style.transform = 'translateY(0) scale(1)';
        });
    });
}

// Função para animar barras de progresso
function animateProgressBars() {
    const progressBars = document.querySelectorAll('.modern-progress-fill');
    
    progressBars.forEach(bar => {
        const width = bar.style.width;
        bar.style.width = '0';
        
        setTimeout(() => {
            bar.style.transition = 'width 1s ease';
            bar.style.width = width;
        }, 300);
    });
}

// Função para adicionar efeito de typing em títulos
function typeWriter(element, text, speed = 50) {
    let i = 0;
    element.textContent = '';
    
    function type() {
        if (i < text.length) {
            element.textContent += text.charAt(i);
            i++;
            setTimeout(type, speed);
        }
    }
    
    type();
}

// Função para criar tooltip personalizado
function createTooltip(element, text) {
    element.setAttribute('data-tooltip', text);
    element.classList.add('tooltip');
}

// Função para adicionar efeito de parallax
function addParallaxEffect() {
    window.addEventListener('scroll', () => {
        const scrolled = window.pageYOffset;
        const parallaxElements = document.querySelectorAll('.parallax');
        
        parallaxElements.forEach(element => {
            const speed = element.dataset.speed || 0.5;
            element.style.transform = `translateY(${scrolled * speed}px)`;
        });
    });
}

// Função para validar formulários
function validateForm(form) {
    const inputs = form.querySelectorAll('input[required], select[required], textarea[required]');
    let isValid = true;
    
    inputs.forEach(input => {
        if (!input.value.trim()) {
            isValid = false;
            input.style.borderColor = 'var(--error-500)';
            
            input.addEventListener('input', function() {
                if (this.value.trim()) {
                    this.style.borderColor = 'var(--primary-500)';
                }
            });
        }
    });
    
    return isValid;
}

// Função para mostrar notificação toast
function showToast(message, type = 'success') {
    const toast = document.createElement('div');
    toast.className = `modern-alert modern-alert-${type}`;
    toast.style.cssText = `
        position: fixed;
        top: 20px;
        right: 20px;
        z-index: 10000;
        min-width: 300px;
        animation: slideInRight 0.3s ease;
    `;
    
    const icon = type === 'success' ? '✓' : type === 'error' ? '✕' : '!';
    
    toast.innerHTML = `
        <span style="font-weight: bold; margin-right: 8px;">${icon}</span>
        ${message}
    `;
    
    document.body.appendChild(toast);
    
    setTimeout(() => {
        toast.style.animation = 'slideOutRight 0.3s ease';
        setTimeout(() => toast.remove(), 300);
    }, 3000);
}

// Função para formatar moeda
function formatCurrency(value) {
    return new Intl.NumberFormat('pt-BR', {
        style: 'currency',
        currency: 'BRL'
    }).format(value);
}

// Função para formatar data
function formatDate(date) {
    return new Intl.DateTimeFormat('pt-BR').format(new Date(date));
}

// Função para debounce
function debounce(func, wait) {
    let timeout;
    return function executedFunction(...args) {
        const later = () => {
            clearTimeout(timeout);
            func(...args);
        };
        clearTimeout(timeout);
        timeout = setTimeout(later, wait);
    };
}

// Função para throttle
function throttle(func, limit) {
    let inThrottle;
    return function(...args) {
        if (!inThrottle) {
            func.apply(this, args);
            inThrottle = true;
            setTimeout(() => inThrottle = false, limit);
        }
    };
}

// Função para copiar para clipboard
function copyToClipboard(text) {
    navigator.clipboard.writeText(text).then(() => {
        showToast('Copiado para a área de transferência!', 'success');
    }).catch(() => {
        showToast('Erro ao copiar!', 'error');
    });
}

// Função para gerar cores aleatórias
function getRandomColor() {
    const colors = [
        '#4CAF50', '#2196F3', '#FF9800', '#E91E63', 
        '#9C27B0', '#00BCD4', '#FF5722', '#795548'
    ];
    return colors[Math.floor(Math.random() * colors.length)];
}

// Função para criar gráfico simples
function createSimpleChart(canvas, data, colors) {
    const ctx = canvas.getContext('2d');
    const barWidth = canvas.width / data.length - 10;
    const maxValue = Math.max(...data);
    
    data.forEach((value, index) => {
        const barHeight = (value / maxValue) * (canvas.height - 20);
        const x = index * (barWidth + 10) + 5;
        const y = canvas.height - barHeight - 10;
        
        ctx.fillStyle = colors[index % colors.length];
        ctx.fillRect(x, y, barWidth, barHeight);
    });
}

// Função para adicionar efeito de loading
function showLoading(element) {
    const loading = document.createElement('div');
    loading.className = 'spinner';
    loading.style.cssText = `
        position: absolute;
        top: 50%;
        left: 50%;
        transform: translate(-50%, -50%);
    `;
    element.style.position = 'relative';
    element.appendChild(loading);
    element.classList.add('loading');
}

function hideLoading(element) {
    const loading = element.querySelector('.spinner');
    if (loading) {
        loading.remove();
    }
    element.classList.remove('loading');
}

// Função para criar modal
function createModal(title, content, actions = []) {
    const modal = document.createElement('div');
    modal.className = 'modal-backdrop';
    modal.style.cssText = `
        position: fixed;
        top: 0;
        left: 0;
        width: 100%;
        height: 100%;
        display: flex;
        align-items: center;
        justify-content: center;
        z-index: 10000;
        animation: fadeIn 0.3s ease;
    `;
    
    const modalContent = document.createElement('div');
    modalContent.className = 'modern-card';
    modalContent.style.cssText = `
        max-width: 500px;
        width: 90%;
        max-height: 80vh;
        overflow-y: auto;
        animation: slideUp 0.3s ease;
    `;
    
    let actionsHTML = '';
    if (actions.length > 0) {
        actionsHTML = '<div style="display: flex; gap: 8px; justify-content: flex-end; margin-top: 16px;">';
        actions.forEach(action => {
            actionsHTML += `<button class="modern-btn ${action.class || ''}" onclick="${action.onclick}">${action.text}</button>`;
        });
        actionsHTML += '</div>';
    }
    
    modalContent.innerHTML = `
        <h2 style="margin-bottom: 16px;">${title}</h2>
        <div style="margin-bottom: 16px;">${content}</div>
        ${actionsHTML}
    `;
    
    modal.appendChild(modalContent);
    document.body.appendChild(modal);
    
    modal.addEventListener('click', (e) => {
        if (e.target === modal) {
            modal.remove();
        }
    });
    
    return modal;
}

// Função para confirmar ação
function confirmAction(message, callback) {
    const modal = createModal('Confirmação', message, [
        {
            text: 'Cancelar',
            class: 'modern-btn-secondary',
            onclick: 'this.closest(".modal-backdrop").remove()'
        },
        {
            text: 'Confirmar',
            class: 'modern-btn-primary',
            onclick: `this.closest(".modal-backdrop").remove(); ${callback}()`
        }
    ]);
}

// Função para inicializar todos os efeitos
function initializeEffects() {
    createParticles();
    addRippleEffect();
    animateOnScroll();
    addCardHoverEffect();
    animateProgressBars();
    addParallaxEffect();
}

// Inicializar quando o DOM estiver pronto
document.addEventListener('DOMContentLoaded', () => {
    initializeEffects();
    
    // Adicionar validação aos formulários
    const forms = document.querySelectorAll('form');
    forms.forEach(form => {
        form.addEventListener('submit', (e) => {
            if (!validateForm(form)) {
                e.preventDefault();
                showToast('Por favor, preencha todos os campos obrigatórios.', 'error');
            }
        });
    });
    
    // Adicionar efeito de hover nos links da navegação
    const navLinks = document.querySelectorAll('.modern-link');
    navLinks.forEach(link => {
        link.addEventListener('mouseenter', function() {
            this.style.transform = 'translateY(-2px)';
        });
        
        link.addEventListener('mouseleave', function() {
            this.style.transform = 'translateY(0)';
        });
    });
});

// Exportar funções para uso global
window.FormaturaEffects = {
    createParticles,
    addRippleEffect,
    animateOnScroll,
    addCardHoverEffect,
    animateProgressBars,
    typeWriter,
    createTooltip,
    addParallaxEffect,
    validateForm,
    showToast,
    formatCurrency,
    formatDate,
    debounce,
    throttle,
    copyToClipboard,
    getRandomColor,
    createSimpleChart,
    showLoading,
    hideLoading,
    createModal,
    confirmAction,
    initializeEffects
};