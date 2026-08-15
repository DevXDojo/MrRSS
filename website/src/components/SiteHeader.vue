<script setup>
import { ref, onMounted, onUnmounted } from 'vue';
import { PhArrowUpRight } from '@phosphor-icons/vue';

const props = defineProps({
  isDark: {
    type: Boolean,
    default: false,
  },
});

const isScrolled = ref(false);
const isMobileMenuOpen = ref(false);

const handleScroll = () => {
  isScrolled.value = window.scrollY > 50;
};

const toggleMobileMenu = () => {
  isMobileMenuOpen.value = !isMobileMenuOpen.value;
  // Prevent body scroll when menu is open
  document.body.style.overflow = isMobileMenuOpen.value ? 'hidden' : '';
};

const closeMobileMenu = () => {
  isMobileMenuOpen.value = false;
  document.body.style.overflow = '';
};

onMounted(() => {
  window.addEventListener('scroll', handleScroll);
});

onUnmounted(() => {
  window.removeEventListener('scroll', handleScroll);
  document.body.style.overflow = '';
});
</script>

<template>
  <header class="site-header" :class="{ scrolled: isScrolled, 'dark-mode': isDark }">
    <div class="header-container">
      <!-- Logo -->
      <a href="#" class="header-logo">
        <img src="../../assets/logo.svg" alt="MrRSS Logo" width="36" height="36" />
        <span class="logo-text">MrRSS</span>
      </a>

      <!-- Navigation -->
      <nav class="header-nav">
        <a href="#features" class="nav-link">Features</a>
        <a href="#download" class="nav-link">Download</a>
        <a href="https://github.com/DevXDojo/MrRSS/blob/main/README.md" class="nav-link"
          >Documentation</a
        >
        <a href="https://github.com/DevXDojo/MrRSS" target="_blank" class="nav-link">GitHub</a>
      </nav>

      <!-- CTA Button -->
      <div class="header-actions">
        <a href="#download" class="cta-button">
          Get Started
          <PhArrowUpRight :size="16" weight="bold" />
        </a>

        <!-- Mobile Menu Toggle -->
        <button class="mobile-toggle" @click="toggleMobileMenu" :aria-expanded="isMobileMenuOpen">
          <span class="hamburger-line"></span>
          <span class="hamburger-line"></span>
          <span class="hamburger-line"></span>
        </button>
      </div>
    </div>

    <!-- Mobile Menu -->
    <div class="mobile-menu" :class="{ active: isMobileMenuOpen }">
      <nav class="mobile-nav">
        <a href="#features" class="mobile-nav-link" @click="closeMobileMenu">Features</a>
        <a href="#download" class="mobile-nav-link" @click="closeMobileMenu">Download</a>
        <a href="#" class="mobile-nav-link" @click="closeMobileMenu">Documentation</a>
        <a href="https://github.com/mrrss/app" target="_blank" class="mobile-nav-link">GitHub</a>
        <a href="#download" class="mobile-cta" @click="closeMobileMenu">Get Started</a>
      </nav>
    </div>
  </header>
</template>

<style scoped>
.site-header {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  z-index: 1000;
  transition: all 0.3s ease;
  background: transparent;
  padding: 1.25rem 0;
}

.site-header.scrolled {
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(20px);
  box-shadow: 0 2px 20px rgba(0, 0, 0, 0.06);
  padding: 0.875rem 0;
}

.site-header.dark-mode {
  background: rgba(13, 17, 23, 0.95);
}

.site-header.dark-mode.scrolled {
  background: rgba(13, 17, 23, 0.98);
  box-shadow: 0 2px 20px rgba(0, 0, 0, 0.3);
}

.header-container {
  max-width: 1400px;
  margin: 0 auto;
  padding: 0 2rem;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 2rem;
}

.header-logo {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  text-decoration: none;
  transition: transform 0.3s ease;
}

.header-logo:hover {
  transform: scale(1.02);
}

.header-logo svg {
  color: #007bff;
  flex-shrink: 0;
}

.logo-text {
  font-family: 'Space Grotesk', sans-serif;
  font-size: 1.5rem;
  font-weight: 700;
  color: #007bff;
  transition: color 0.3s ease;
}

.site-header.dark-mode .logo-text {
  color: #58a6ff;
}

.header-nav {
  display: flex;
  align-items: center;
  gap: 2.5rem;
}

.nav-link {
  font-size: 0.9375rem;
  font-weight: 500;
  color: rgba(0, 0, 0, 0.7);
  text-decoration: none;
  transition: all 0.3s ease;
  position: relative;
}

.site-header.dark-mode .nav-link {
  color: rgba(240, 246, 252, 0.7);
}

.nav-link::after {
  content: '';
  position: absolute;
  bottom: -4px;
  left: 0;
  width: 0;
  height: 2px;
  background: #007bff;
  transition: width 0.3s ease;
}

.site-header.dark-mode .nav-link::after {
  background: #58a6ff;
}

.nav-link:hover {
  color: #007bff;
}

.site-header.dark-mode .nav-link:hover {
  color: #58a6ff;
}

.nav-link:hover::after {
  width: 100%;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.cta-button {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.75rem 1.5rem;
  background: linear-gradient(135deg, #007bff, #0056b3);
  color: white;
  border-radius: 10px;
  font-size: 0.9375rem;
  font-weight: 600;
  text-decoration: none;
  transition: all 0.3s ease;
  box-shadow: 0 4px 14px rgba(0, 123, 255, 0.2);
}

.cta-button:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 20px rgba(0, 123, 255, 0.3);
}

.cta-button svg {
  transition: transform 0.3s ease;
}

.cta-button:hover svg {
  transform: translate(3px, -3px);
}

.mobile-toggle {
  display: none;
  flex-direction: column;
  gap: 5px;
  padding: 0.5rem;
  background: transparent;
  border: none;
  cursor: pointer;
}

.hamburger-line {
  width: 24px;
  height: 2px;
  background: #1a1a1a;
  border-radius: 2px;
  transition: all 0.3s ease;
}

.site-header.dark-mode .hamburger-line {
  background: #f0f6fc;
}

.mobile-menu {
  display: none;
}

.site-header.dark-mode .mobile-menu {
  background: rgba(13, 17, 23, 0.98);
  border-top: 1px solid rgba(240, 246, 252, 0.1);
}

@media (max-width: 768px) {
  .site-header {
    padding: 1rem 0;
  }

  .header-container {
    padding: 0 1.5rem;
    gap: 1rem;
  }

  .header-nav {
    display: none;
  }

  .mobile-toggle {
    display: flex;
  }

  .mobile-toggle[aria-expanded='true'] .hamburger-line:nth-child(1) {
    transform: rotate(45deg) translate(5px, 5px);
  }

  .mobile-toggle[aria-expanded='true'] .hamburger-line:nth-child(2) {
    opacity: 0;
  }

  .mobile-toggle[aria-expanded='true'] .hamburger-line:nth-child(3) {
    transform: rotate(-45deg) translate(7px, -6px);
  }

  .mobile-menu {
    display: block;
    position: absolute;
    top: 100%;
    left: 0;
    width: 100%;
    background: rgba(255, 255, 255, 0.98);
    backdrop-filter: blur(20px);
    border-top: 1px solid rgba(0, 0, 0, 0.06);
    padding: 0;
    max-height: 0;
    overflow: hidden;
    transition: all 0.3s ease;
    box-shadow: 0 10px 30px rgba(0, 0, 0, 0.1);
  }

  .mobile-menu.active {
    max-height: 400px;
    padding: 2rem 1.5rem;
  }

  .mobile-nav {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  .mobile-nav-link {
    display: block;
    padding: 1rem;
    font-size: 1rem;
    font-weight: 500;
    color: rgba(0, 0, 0, 0.7);
    text-decoration: none;
    border-radius: 10px;
    transition: all 0.3s ease;
  }

  .site-header.dark-mode .mobile-nav-link {
    color: rgba(240, 246, 252, 0.7);
  }

  .mobile-nav-link:hover {
    background: rgba(0, 123, 255, 0.08);
    color: #007bff;
  }

  .site-header.dark-mode .mobile-nav-link:hover {
    background: rgba(88, 166, 255, 0.15);
    color: #58a6ff;
  }

  .mobile-cta {
    display: block;
    padding: 1rem;
    background: linear-gradient(135deg, #007bff, #0056b3);
    color: white;
    text-align: center;
    border-radius: 10px;
    font-weight: 600;
    text-decoration: none;
    margin-top: 1rem;
  }

  .site-header.dark-mode .mobile-cta {
    background: linear-gradient(135deg, #58a6ff, #1f6feb);
  }
}

@media (max-width: 640px) {
  .logo-text {
    font-size: 1.25rem;
  }

  .cta-button {
    display: none;
  }
}
</style>
