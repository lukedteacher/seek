(function () {
  'use strict';

  const toastTimers = new Map();

  // setup toast when it appears
  function setupToast(toast) {
    if (!toast || toastTimers.has(toast)) return;

    const duration = parseInt(toast.dataset.tuiToastDuration || '3000');
    const progress = toast.querySelector('.toast-progress');

    // initialize timer state
    const state = {
      timer: null,
      startTime: Date.now(),
      remaining: duration,
      paused: false,
      progressWidth: null,
    };
    toastTimers.set(toast, state);

    // animate progress bar if present
    if (progress && duration > 0) {
      progress.style.width = '100%';
      void progress.offsetWidth;
      progress.style.transition = `width ${duration}ms linear`;
      progress.style.width = '0px';
    }

    // auto-dismiss after duration
    if (duration > 0) {
      state.timer = setTimeout(() => dismissToast(toast), duration);
    }

    // pause on hover
    toast.addEventListener('mouseenter', () => {
      const state = toastTimers.get(toast);
      if (!state || state.paused) return;

      // clear the dismiss timer
      clearTimeout(state.timer);

      // calculate remaining time
      state.remaining = state.remaining - (Date.now() - state.startTime);
      state.paused = true;

      // pause progress animation
      if (progress) {
        state.progressWidth = getComputedStyle(progress).width;
        progress.style.transition = 'none';
        progress.style.width = state.progressWidth;
      }
    });

    // resume on mouse leave
    toast.addEventListener('mouseleave', () => {
      const state = toastTimers.get(toast);
      if (!state || !state.paused || state.remaining <= 0) return;

      // resume timer with remaining time
      state.startTime = Date.now();
      state.paused = false;
      state.timer = setTimeout(() => dismissToast(toast), state.remaining);

      // resume progress animation
      if (progress) {
        progress.style.width = state.progressWidth;
        void progress.offsetWidth;
        progress.style.transition = `width ${state.remaining}ms linear`;
        progress.style.width = '0px';
      }
    });
  }

  // dismiss toast with fade out
  function dismissToast(toast) {
    // clean up timer state
    toastTimers.delete(toast);

    // add transition for smooth fade out
    toast.style.transition = 'opacity 300ms, transform 300ms';
    toast.style.opacity = '0';
    toast.style.transform = 'translateY(1rem)';

    // remove after animation
    setTimeout(() => toast.remove(), 300);
  }

  // handle dismiss button clicks
  document.addEventListener('click', (e) => {
    const dismissBtn = e.target.closest('[data-tui-toast-dismiss]');
    if (dismissBtn) {
      const toast = dismissBtn.closest('[data-tui-toast]');
      if (toast) dismissToast(toast);
    }
  });

  function initializeToasts(root) {
    if (!root) return;

    if (root.matches?.('[data-tui-toast]')) {
      setupToast(root);
    }

    root.querySelectorAll?.('[data-tui-toast]').forEach(setupToast);
  }

  // initialize pre-rendered toasts
  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', () =>
      initializeToasts(document),
    );
  } else {
    initializeToasts(document);
  }

  // watch for new toasts
  new MutationObserver((mutations) => {
    mutations.forEach((m) => {
      m.addedNodes.forEach((node) => {
        if (node.nodeType === 1) {
          initializeToasts(node);
        }
      });
    });
  }).observe(document.body, { childList: true, subtree: true });
})();
