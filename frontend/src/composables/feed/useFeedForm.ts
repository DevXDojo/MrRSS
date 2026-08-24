import { ref, computed, onMounted } from 'vue';
import { useI18n } from 'vue-i18n';
import type { Feed } from '@/types/models';
import { useAppStore } from '@/stores/app';

type FeedType = 'url' | 'script' | 'xpath' | 'email';
type ProxyMode = 'global' | 'custom' | 'none';
type RefreshMode = 'global' | 'fixed' | 'intelligent' | 'custom' | 'never';

export function useFeedForm(feed?: Feed) {
  const { t } = useI18n();
  const store = useAppStore();

  // Check if image gallery feature is enabled
  const imageGalleryEnabled = ref(false);

  const feedType = ref<FeedType>('url');
  const title = ref('');
  const url = ref('');
  const category = ref('');
  const categorySelection = ref('');
  const showCustomCategory = ref(false);
  const scriptPath = ref('');
  const hideFromTimeline = ref(false);
  const isImageMode = ref(false);

  // XPath fields
  const xpathType = ref<'HTML+XPath' | 'XML+XPath'>('HTML+XPath');
  const xpathItem = ref('');
  const xpathItemTitle = ref('');
  const xpathItemContent = ref('');
  const xpathItemUri = ref('');
  const xpathItemAuthor = ref('');
  const xpathItemTimestamp = ref('');
  const xpathItemTimeFormat = ref('');
  const xpathItemThumbnail = ref('');
  const xpathItemCategories = ref('');
  const xpathItemUid = ref('');

  // Email/Newsletter fields
  const emailAddress = ref('');
  const imapServer = ref('');
  const imapPort = ref(993);
  const emailUsername = ref('');
  const emailPassword = ref('');
  const emailFolder = ref('INBOX');

  // Tags
  const selectedTags = ref<number[]>([]);

  // Article view mode
  const articleViewMode = ref<'global' | 'webpage' | 'rendered' | 'external'>('global');

  // Auto expand content mode
  const autoExpandContent = ref<'global' | 'enabled' | 'disabled'>('global');

  // Proxy settings
  const proxyMode = ref<ProxyMode>('global');
  const proxyType = ref('http');
  const proxyHost = ref('');
  const proxyPort = ref('');
  const proxyUsername = ref('');
  const proxyPassword = ref('');

  // Refresh settings
  const refreshMode = ref<RefreshMode>('global');
  const refreshInterval = ref(30);

  const isSubmitting = ref(false);
  const showAdvancedSettings = ref(false);

  // Available scripts from the scripts directory
  const availableScripts = ref<Array<{ name: string; path: string; type: string }>>([]);
  const scriptsDir = ref('');

  // Get unique categories from existing feeds, excluding FreshRSS-only categories
  const existingCategories = computed(() => {
    const categoryFeedsMap = new Map<string, boolean>();

    // Build a map of category -> whether it has non-FreshRSS feeds
    store.feeds.forEach((feed) => {
      if (feed.category && feed.category.trim() !== '') {
        if (!categoryFeedsMap.has(feed.category)) {
          categoryFeedsMap.set(feed.category, !feed.is_freshrss_source);
        } else {
          // Update if we find a non-FreshRSS feed in this category
          if (!feed.is_freshrss_source) {
            categoryFeedsMap.set(feed.category, true);
          }
        }
      }
    });

    // Filter out categories where all feeds are from FreshRSS
    // or category name ends with " (FreshRSS)" or matches pattern " (FreshRSS \d+)$"
    const categories = Array.from(categoryFeedsMap.entries())
      .filter(([_, hasNonFreshRSS]) => hasNonFreshRSS)
      .filter(([categoryName]) => {
        return !categoryName.endsWith(' (FreshRSS)') && !categoryName.match(/ \(FreshRSS \d+\)$/);
      })
      .map(([categoryName]) => categoryName)
      .sort();

    return categories;
  });

  // Watch for category selection changes
  function handleCategoryChange(value?: string) {
    if (value !== undefined) {
      categorySelection.value = value;
    }
    if (categorySelection.value === '__custom__') {
      showCustomCategory.value = true;
      category.value = '';
    } else {
      showCustomCategory.value = false;
      category.value = categorySelection.value;
    }
  }

  async function loadImageGallerySetting() {
    try {
      const res = await fetch('/api/settings');
      if (res.ok) {
        const data = await res.json();
        imageGalleryEnabled.value = data.image_gallery_enabled === 'true';
      }
    } catch (e) {
      console.error('Failed to load settings:', e);
    }
  }

  async function loadScripts() {
    try {
      const res = await fetch('/api/scripts/list');
      if (res.ok) {
        const data = await res.json();
        availableScripts.value = data.scripts || [];
        scriptsDir.value = data.scripts_dir || '';
      }
    } catch (e) {
      console.error('Failed to load scripts:', e);
    }
  }

  const isFormValid = computed(() => {
    if (feedType.value === 'url') {
      return url.value.trim() !== '';
    } else if (feedType.value === 'script') {
      return scriptPath.value.trim() !== '';
    } else if (feedType.value === 'xpath') {
      return url.value.trim() !== '' && xpathItem.value.trim() !== '';
    } else if (feedType.value === 'email') {
      return (
        emailAddress.value.trim() !== '' &&
        imapServer.value.trim() !== '' &&
        emailUsername.value.trim() !== '' &&
        emailPassword.value.trim() !== ''
      );
    }
    return false;
  });

  // Validation for URL field
  const isUrlInvalid = computed(() => {
    return (feedType.value === 'url' || feedType.value === 'xpath') && !url.value.trim();
  });

  // Validation for script field
  const isScriptInvalid = computed(() => {
    return feedType.value === 'script' && !scriptPath.value.trim();
  });

  // Validation for XPath item field
  const isXpathItemInvalid = computed(() => {
    return feedType.value === 'xpath' && !xpathItem.value.trim();
  });

  function buildProxyUrl(): string {
    if (proxyMode.value !== 'custom' || !proxyHost.value || !proxyPort.value) {
      return '';
    }

    let auth = '';
    if (proxyUsername.value) {
      auth = proxyPassword.value
        ? `${proxyUsername.value}:${proxyPassword.value}@`
        : `${proxyUsername.value}@`;
    }

    return `${proxyType.value}://${auth}${proxyHost.value}:${proxyPort.value}`;
  }

  function getRefreshInterval(): number {
    // Return 0 for global, -1 for intelligent, -2 for never, or the custom interval
    switch (refreshMode.value) {
      case 'global':
        return 0;
      case 'intelligent':
        return -1;
      case 'never':
        return -2;
      case 'custom':
        return refreshInterval.value;
      default:
        return 0;
    }
  }

  function initializeFromFeed(feed: Feed) {
    title.value = feed.title;
    url.value = feed.url;
    category.value = feed.category;
    scriptPath.value = feed.script_path || '';
    hideFromTimeline.value = feed.hide_from_timeline || false;
    isImageMode.value = feed.is_image_mode || false;

    // Initialize XPath fields
    xpathType.value =
      feed.type === 'HTML+XPath' || feed.type === 'XML+XPath'
        ? (feed.type as 'HTML+XPath' | 'XML+XPath')
        : 'HTML+XPath';
    xpathItem.value = feed.xpath_item || '';
    xpathItemTitle.value = feed.xpath_item_title || '';
    xpathItemContent.value = feed.xpath_item_content || '';
    xpathItemUri.value = feed.xpath_item_uri || '';
    xpathItemAuthor.value = feed.xpath_item_author || '';
    xpathItemTimestamp.value = feed.xpath_item_timestamp || '';
    xpathItemTimeFormat.value = feed.xpath_item_time_format || '';
    xpathItemThumbnail.value = feed.xpath_item_thumbnail || '';
    xpathItemCategories.value = feed.xpath_item_categories || '';
    xpathItemUid.value = feed.xpath_item_uid || '';

    // Initialize article view mode
    articleViewMode.value =
      (feed.article_view_mode as 'global' | 'webpage' | 'rendered' | 'external') || 'global';

    // Initialize auto expand content mode
    autoExpandContent.value =
      (feed.auto_expand_content as 'global' | 'enabled' | 'disabled') || 'global';

    // Determine feed type based on feed properties
    if (feed.script_path) {
      feedType.value = 'script';
    } else if (feed.xpath_item) {
      feedType.value = 'xpath';
    } else if (feed.type === 'email') {
      feedType.value = 'email';
      // Initialize email fields
      emailAddress.value = feed.email_address || '';
      imapServer.value = feed.email_imap_server || '';
      imapPort.value = feed.email_imap_port || 993;
      emailUsername.value = feed.email_username || '';
      emailPassword.value = feed.email_password || '';
      emailFolder.value = feed.email_folder || 'INBOX';
    } else {
      feedType.value = 'url';
    }

    // Initialize tags
    if (feed.tags) {
      selectedTags.value = feed.tags.map((tag) => tag.id);
    }

    // Initialize proxy settings
    if (feed.proxy_url) {
      proxyMode.value = 'custom';
      // Parse proxy URL: protocol://[username:password@]host:port
      try {
        const proxyUrlObj = new URL(feed.proxy_url);
        proxyType.value = proxyUrlObj.protocol.replace(':', '');
        proxyHost.value = proxyUrlObj.hostname;
        proxyPort.value = proxyUrlObj.port;
        proxyUsername.value = proxyUrlObj.username;
        proxyPassword.value = proxyUrlObj.password;
      } catch (e) {
        // Fallback for invalid URL format
        console.error('Failed to parse proxy URL:', e);
        window.showToast(t('setting.network.invalidProxyUrl'), 'error');
      }
    } else if (feed.proxy_enabled) {
      proxyMode.value = 'global';
    } else {
      proxyMode.value = 'none';
    }

    // Initialize refresh settings
    const interval = feed.refresh_interval || 0;
    if (interval === 0) {
      refreshMode.value = 'global';
    } else if (interval === -1) {
      refreshMode.value = 'intelligent';
    } else if (interval === -2) {
      refreshMode.value = 'never';
    } else {
      refreshMode.value = 'custom';
      refreshInterval.value = interval;
    }

    // Initialize category selection
    if (category.value && existingCategories.value.includes(category.value)) {
      categorySelection.value = category.value;
    } else if (category.value) {
      // If category doesn't exist in list, show custom input
      showCustomCategory.value = true;
    }
  }

  function resetForm() {
    title.value = '';
    url.value = '';
    category.value = '';
    scriptPath.value = '';
    hideFromTimeline.value = false;
    isImageMode.value = false;
    xpathType.value = 'HTML+XPath';
    xpathItem.value = '';
    xpathItemTitle.value = '';
    xpathItemContent.value = '';
    xpathItemUri.value = '';
    xpathItemAuthor.value = '';
    xpathItemTimestamp.value = '';
    xpathItemTimeFormat.value = '';
    xpathItemThumbnail.value = '';
    xpathItemCategories.value = '';
    xpathItemUid.value = '';
    // Reset email fields
    emailAddress.value = '';
    imapServer.value = '';
    imapPort.value = 993;
    emailUsername.value = '';
    emailPassword.value = '';
    emailFolder.value = 'INBOX';
    articleViewMode.value = 'global';
    autoExpandContent.value = 'global';
    proxyMode.value = 'global';
    proxyType.value = 'http';
    proxyHost.value = '';
    proxyPort.value = '';
    proxyUsername.value = '';
    proxyPassword.value = '';
    refreshMode.value = 'global';
    refreshInterval.value = 30;
  }

  async function openScriptsFolder() {
    try {
      await fetch('/api/scripts/open', { method: 'POST' });
      window.showToast(t('setting.customization.scriptsFolderOpened'), 'success');
    } catch (e) {
      console.error('Failed to open scripts folder:', e);
    }
  }

  onMounted(async () => {
    await loadScripts();
    await loadImageGallerySetting();

    // Listen for settings changes
    window.addEventListener('image-gallery-setting-changed', (e: Event) => {
      const customEvent = e as CustomEvent;
      imageGalleryEnabled.value = customEvent.detail.enabled;
    });

    if (feed) {
      initializeFromFeed(feed);
    }
  });

  return {
    // State
    imageGalleryEnabled,
    feedType,
    title,
    url,
    category,
    categorySelection,
    showCustomCategory,
    scriptPath,
    hideFromTimeline,
    isImageMode,
    xpathType,
    xpathItem,
    xpathItemTitle,
    xpathItemContent,
    xpathItemUri,
    xpathItemAuthor,
    xpathItemTimestamp,
    xpathItemTimeFormat,
    xpathItemThumbnail,
    xpathItemCategories,
    xpathItemUid,
    // Email fields
    emailAddress,
    imapServer,
    imapPort,
    emailUsername,
    emailPassword,
    emailFolder,
    // Tags
    selectedTags,
    articleViewMode,
    autoExpandContent,
    proxyMode,
    proxyType,
    proxyHost,
    proxyPort,
    proxyUsername,
    proxyPassword,
    refreshMode,
    refreshInterval,
    isSubmitting,
    showAdvancedSettings,
    availableScripts,
    scriptsDir,
    existingCategories,

    // Computed
    isFormValid,
    isUrlInvalid,
    isScriptInvalid,
    isXpathItemInvalid,

    // Methods
    handleCategoryChange,
    buildProxyUrl,
    getRefreshInterval,
    resetForm,
    openScriptsFolder,
  };
}
