<template>
  <footer class="footer">
    <div class="footer-background">
      <div class="footer-gradient"></div>
    </div>

    <div class="container">
      <div class="footer-main">
        <div class="footer-brand">
          <img src="../../assets/logo.svg" alt="MrRSS Logo" width="45" height="45" />
          <h3 class="brand-name">MrRSS</h3>
          <p class="brand-tagline">Read Smarter with AI</p>
        </div>

        <div class="footer-links">
          <div class="link-group">
            <h4 class="link-title">Product</h4>
            <ul class="link-list">
              <li><a href="#features">Features</a></li>
              <li><a href="#download">Download</a></li>
              <li>
                <a href="https://github.com/DevXDojo/MrRSS/blob/main/CHANGELOG.md">Changelog</a>
              </li>
              <li><a href="https://github.com/users/DevXDojo/projects/2">Roadmap</a></li>
            </ul>
          </div>

          <div class="link-group">
            <h4 class="link-title">Resources</h4>
            <ul class="link-list">
              <li>
                <a href="https://github.com/DevXDojo/MrRSS/blob/main/README.md">Documentation</a>
              </li>
              <li>
                <a href="https://github.com/DevXDojo/MrRSS/blob/main/CODE_OF_CONDUCT.md"
                  >User Guide</a
                >
              </li>
              <li>
                <a href="https://github.com/DevXDojo/MrRSS/blob/main/docs/SERVER_MODE.md"
                  >API Reference</a
                >
              </li>
              <li><a href="https://github.com/DevXDojo/MrRSS/discussions">Community</a></li>
            </ul>
          </div>

          <div class="link-group">
            <h4 class="link-title">Legal</h4>
            <ul class="link-list">
              <li>
                <a href="https://github.com/DevXDojo/MrRSS/blob/main/SECURITY.md">Security Policy</a>
              </li>
              <li><a href="https://github.com/DevXDojo/MrRSS/blob/main/LICENSE">License</a></li>
              <li><a href="https://github.com/DevXDojo/MrRSS/issues">Contact</a></li>
            </ul>
          </div>

          <div class="link-group">
            <h4 class="link-title">Connect</h4>
            <ul class="link-list">
              <li>
                <a href="https://github.com/DevXDojo/MrRSS" class="social-link">
                  <PhGithubLogo :size="20" />
                  GitHub
                </a>
              </li>
              <!-- <li>
                <a href="#" class="social-link">
                  <svg width="20" height="20" viewBox="0 0 24 24" fill="currentColor">
                    <path d="M18.244 2.25h3.308l-7.227 8.26 8.502 11.24H16.17l-5.214-6.817L4.99 21.75H1.68l7.73-8.835L1.254 2.25H8.08l4.713 6.231zm-1.161 17.52h1.833L7.084 4.126H5.117z"/>
                  </svg>
                  Twitter
                </a>
              </li>
              <li>
                <a href="#" class="social-link">
                  <svg width="20" height="20" viewBox="0 0 24 24" fill="currentColor">
                    <path d="M20.317 4.37a19.791 19.791 0 0 0-4.885-1.515.074.074 0 0 0-.079.037c-.21.375-.444.864-.608 1.25a18.27 18.27 0 0 0-5.487 0 12.64 12.64 0 0 0-.617-1.25.077.077 0 0 0-.079-.037A19.736 19.736 0 0 0 3.677 4.37a.07.07 0 0 0-.032.027C.533 9.046-.32 13.58.099 18.057a.082.082 0 0 0 .031.057 19.9 19.9 0 0 0 5.993 3.03.078.078 0 0 0 .084-.028 14.09 14.09 0 0 0 1.226-1.994.076.076 0 0 0-.041-.106 13.107 13.107 0 0 1-1.872-.892.077.077 0 0 1-.008-.128 10.2 10.2 0 0 0 .372-.292.074.074 0 0 1 .077-.01c3.928 1.793 8.18 1.793 12.062 0a.074.074 0 0 1 .078.01c.12.098.246.198.373.292a.077.077 0 0 1-.006.127 12.299 12.299 0 0 1-1.873.892.077.077 0 0 0-.041.107c.36.698.772 1.362 1.225 1.993a.076.076 0 0 0 .084.028 19.839 19.839 0 0 0 6.002-3.03.077.077 0 0 0 .032-.054c.5-5.177-.838-9.674-3.549-13.66a.061.061 0 0 0-.031-.03zM8.02 15.33c-1.183 0-2.157-1.085-2.157-2.419 0-1.333.956-2.419 2.157-2.419 1.21 0 2.176 1.096 2.157 2.42 0 1.333-.956 2.418-2.157 2.418zm7.975 0c-1.183 0-2.157-1.085-2.157-2.419 0-1.333.955-2.419 2.157-2.419 1.21 0 2.176 1.096 2.157 2.42 0 1.333-.946 2.418-2.157 2.418z"/>
                  </svg>
                  Discord
                </a>
              </li> -->
            </ul>
          </div>
        </div>
      </div>

      <div class="footer-bottom">
        <div class="footer-copyright">
          <p>&copy; 2026 MrRSS. All rights reserved.</p>
          <p class="copyright-note">Open source and available under GPL-3.0 License.</p>
        </div>

        <div class="footer-badge">
          <span class="badge-icon">♥</span>
          <span>Made for the community</span>
        </div>
      </div>
    </div>
  </footer>
</template>

<script setup>
import { onMounted, onUnmounted } from 'vue';
import { PhGithubLogo } from '@phosphor-icons/vue';

const emit = defineEmits(['footerVisibility']);

const checkVisibility = () => {
  const footer = document.querySelector('.footer');
  if (!footer) return;

  const rect = footer.getBoundingClientRect();
  const viewportHeight = window.innerHeight;
  // 通知父组件footer顶部是否进入视窗超过一半
  emit('footerVisibility', rect.top < viewportHeight * 0.5);
};

const throttledCheck = (() => {
  let ticking = false;
  return () => {
    if (!ticking) {
      window.requestAnimationFrame(() => {
        checkVisibility();
        ticking = false;
      });
      ticking = true;
    }
  };
})();

onMounted(() => {
  window.addEventListener('scroll', throttledCheck, { passive: true });
  // 初始检查
  setTimeout(() => checkVisibility(), 100);
});

onUnmounted(() => {
  window.removeEventListener('scroll', throttledCheck);
});
</script>

<style scoped>
.footer {
  position: relative;
  padding: 4rem 2rem 2rem;
  background: linear-gradient(180deg, #0d1117 0%, #010409 100%);
  border-top: 1px solid rgba(240, 246, 252, 0.1);
  color: #f0f6fc;
}

.footer-background {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  pointer-events: none;
  overflow: hidden;
}

.footer-gradient {
  position: absolute;
  top: -50%;
  left: 50%;
  transform: translateX(-50%);
  width: 80%;
  height: 100%;
  background: radial-gradient(ellipse, rgba(0, 123, 255, 0.04) 0%, transparent 70%);
  filter: blur(100px);
}

.container {
  position: relative;
  max-width: 1200px;
  margin: 0 auto;
  z-index: 1;
}

.footer-main {
  display: grid;
  grid-template-columns: 300px 1fr;
  gap: 4rem;
  margin-bottom: 3rem;
}

.footer-brand {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.brand-name {
  font-family: 'Space Grotesk', sans-serif;
  font-size: 1.75rem;
  font-weight: 700;
  color: #007bff;
}

.brand-tagline {
  font-size: 1rem;
  color: rgba(240, 246, 252, 0.6);
}

.footer-links {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 2rem;
}

.link-group {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.link-title {
  font-family: 'Space Grotesk', sans-serif;
  font-size: 1rem;
  font-weight: 600;
  color: #f0f6fc;
  margin-bottom: 0.5rem;
}

.link-list {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  list-style: none;
}

.link-list a {
  font-size: 0.9375rem;
  color: rgba(240, 246, 252, 0.6);
  text-decoration: none;
  transition: color 0.3s;
}

.link-list a:hover {
  color: #58a6ff;
}

.social-link {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.social-link svg {
  opacity: 0.5;
  transition: opacity 0.3s;
}

.social-link:hover svg {
  opacity: 1;
}

.footer-bottom {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-top: 2rem;
  border-top: 1px solid rgba(240, 246, 252, 0.1);
}

.footer-copyright {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.footer-copyright p {
  font-size: 0.875rem;
  color: rgba(240, 246, 252, 0.5);
}

.copyright-note {
  font-size: 0.8125rem !important;
}

.footer-badge {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 1rem;
  background: rgba(88, 166, 255, 0.1);
  border: 1px solid rgba(88, 166, 255, 0.2);
  border-radius: 50px;
  font-size: 0.875rem;
  color: rgba(240, 246, 252, 0.7);
}

.badge-icon {
  color: #ff0095;
}

@media (max-width: 1024px) {
  .footer-main {
    grid-template-columns: 1fr;
    gap: 3rem;
  }

  .footer-links {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 768px) {
  .footer {
    padding: 3rem 1.5rem 2rem;
  }

  .footer-main {
    gap: 2rem;
  }
}

@media (max-width: 640px) {
  .footer {
    padding: 3rem 1rem 1.5rem;
  }

  .footer-links {
    grid-template-columns: 1fr;
    gap: 2rem;
  }

  .footer-bottom {
    flex-direction: column;
    gap: 1.5rem;
    text-align: center;
  }

  .footer-copyright {
    align-items: center;
  }

  .footer-badge {
    font-size: 0.8125rem;
    padding: 0.4375rem 0.875rem;
  }
}

@media (max-width: 480px) {
  .footer-brand {
    align-items: center;
    text-align: center;
  }

  .brand-name {
    font-size: 1.5rem;
  }

  .brand-tagline {
    font-size: 0.9375rem;
  }

  .link-title {
    font-size: 0.9375rem;
  }

  .link-list a {
    font-size: 0.875rem;
  }

  .footer-badge {
    font-size: 0.75rem;
  }
}
</style>
