import { createApp } from 'vue';
import { createPinia } from 'pinia';
import PhosphorIcons from '@phosphor-icons/vue';
import i18n, { locale } from './i18n';
import './style.css';
import App from './App.vue';
import { setSettingsFromRawData } from './composables/core/useSettings';

const app = createApp(App);
const pinia = createPinia();

// Add global error handler for Vue errors
app.config.errorHandler = (err, instance, info) => {
  console.error('[Vue Error Handler] Error:', err);
  console.error('[Vue Error Handler] Component:', instance?.$?.type?.name || 'Unknown');
  console.error('[Vue Error Handler] Info:', info);
  // Log the full stack trace
  if (err instanceof Error) {
    console.error('[Vue Error Handler] Stack:', err.stack);
  }
};

app.use(pinia);
app.use(i18n);
app.use(PhosphorIcons);

// Initialize language setting before mounting
async function initializeApp() {
  try {
    const res = await fetch('/api/settings');
    if (!res.ok) {
      throw new Error(`HTTP ${res.status}: ${res.statusText}`);
    }

    // Get response text first to debug JSON parsing issues
    const text = await res.text();
    let data;

    try {
      data = JSON.parse(text);
    } catch (jsonError) {
      console.error('JSON parse error:', jsonError);
      console.error('Response text (first 500 chars):', text.substring(0, 500));
      // Use default empty object if JSON is invalid
      data = {};
    }

    if (data.language) {
      locale.value = data.language;
    }

    setSettingsFromRawData(data);

    // Start FreshRSS status polling if enabled
    // Note: Don't use store here - it will be initialized after mount
    // Store initialization will handle FreshRSS polling in App.vue
  } catch (e) {
    console.error('Error loading language setting:', e);
  }

  app.mount('#app');
}

// Initialize and mount
initializeApp();
