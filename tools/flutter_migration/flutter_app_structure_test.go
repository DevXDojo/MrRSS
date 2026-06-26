package flutter_migration

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFlutterAppScaffoldHasRequiredFiles(t *testing.T) {
	root := repositoryRoot(t)

	required := []string{
		"flutter_app/README.md",
		"flutter_app/analysis_options.yaml",
		"flutter_app/pubspec.yaml",
		"flutter_app/lib/main.dart",
		"flutter_app/lib/src/app/mrrss_app.dart",
		"flutter_app/lib/src/api/api_config.dart",
		"flutter_app/lib/src/api/api_exception.dart",
		"flutter_app/lib/src/api/article_content_api.dart",
		"flutter_app/lib/src/api/article_status_api.dart",
		"flutter_app/lib/src/api/articles_api.dart",
		"flutter_app/lib/src/api/feed_actions_api.dart",
		"flutter_app/lib/src/api/feeds_api.dart",
		"flutter_app/lib/src/api/opml_api.dart",
		"flutter_app/lib/src/api/settings_api.dart",
		"flutter_app/lib/src/api/translation_api.dart",
		"flutter_app/lib/src/api/version_api.dart",
		"flutter_app/lib/src/models/article.dart",
		"flutter_app/lib/src/models/article_content.dart",
		"flutter_app/lib/src/models/app_settings.dart",
		"flutter_app/lib/src/models/feed.dart",
		"flutter_app/lib/src/models/translation_result.dart",
		"flutter_app/lib/src/models/version_info.dart",
		"flutter_app/lib/src/platform/opml_file_service.dart",
		"flutter_app/lib/src/reader/http_reader_repository.dart",
		"flutter_app/lib/src/reader/reader_repository.dart",
		"flutter_app/lib/src/reader/reader_screen.dart",
		"flutter_app/lib/src/reader/widgets/article_detail_panel.dart",
		"flutter_app/lib/src/reader/widgets/article_list_panel.dart",
		"flutter_app/lib/src/reader/widgets/feed_sidebar.dart",
		"flutter_app/lib/src/reader/widgets/settings_sheet.dart",
		"flutter_app/lib/src/utils/html_text.dart",
		"flutter_app/test/api_config_test.dart",
		"flutter_app/test/article_content_test.dart",
		"flutter_app/test/article_content_api_test.dart",
		"flutter_app/test/article_status_api_test.dart",
		"flutter_app/test/article_test.dart",
		"flutter_app/test/articles_api_test.dart",
		"flutter_app/test/app_settings_test.dart",
		"flutter_app/test/feed_test.dart",
		"flutter_app/test/feed_actions_api_test.dart",
		"flutter_app/test/feeds_api_test.dart",
		"flutter_app/test/html_text_test.dart",
		"flutter_app/test/opml_api_test.dart",
		"flutter_app/test/reader_screen_test.dart",
		"flutter_app/test/settings_api_test.dart",
		"flutter_app/test/translation_api_test.dart",
		"flutter_app/test/translation_result_test.dart",
		"flutter_app/test/version_api_test.dart",
		"flutter_app/android/app/src/main/kotlin/com/example/mrrss_flutter/MainActivity.kt",
		"flutter_app/android/app/src/main/AndroidManifest.xml",
		"flutter_app/ios/Runner/AppDelegate.swift",
		"flutter_app/ios/Runner/Info.plist",
		"flutter_app/linux/runner/main.cc",
		"flutter_app/macos/Runner/AppDelegate.swift",
		"flutter_app/macos/Runner/Info.plist",
		"flutter_app/windows/runner/main.cpp",
		"flutter_app/windows/runner/runner.exe.manifest",
	}

	for _, name := range required {
		if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(name))); err != nil {
			t.Fatalf("required Flutter migration file %s is missing: %v", name, err)
		}
	}
}

func TestFlutterPubspecDeclaresExpectedAppAndTestDependencies(t *testing.T) {
	root := repositoryRoot(t)
	body := readText(t, filepath.Join(root, "flutter_app", "pubspec.yaml"))

	required := []string{
		"name: mrrss_flutter",
		"flutter:",
		"flutter_test:",
		"http:",
		"file_picker:",
		"flutter_lints:",
		"uses-material-design: true",
	}

	for _, want := range required {
		if !strings.Contains(body, want) {
			t.Fatalf("pubspec.yaml missing %q", want)
		}
	}
}

func TestFirstFlutterApiClientTargetsInventoryRoute(t *testing.T) {
	root := repositoryRoot(t)

	inventoryBody := readText(t, filepath.Join(root, "docs", "flutter", "API_ROUTE_INVENTORY.json"))
	var routes []string
	if err := json.Unmarshal([]byte(inventoryBody), &routes); err != nil {
		t.Fatalf("parse route inventory: %v", err)
	}

	hasVersionRoute := false
	for _, route := range routes {
		if route == "/api/version" {
			hasVersionRoute = true
			break
		}
	}
	if !hasVersionRoute {
		t.Fatalf("route inventory must include /api/version for the first Flutter API client")
	}

	versionAPI := readText(t, filepath.Join(root, "flutter_app", "lib", "src", "api", "version_api.dart"))
	if !strings.Contains(versionAPI, "/api/version") {
		t.Fatalf("version_api.dart must call /api/version")
	}

	versionTest := readText(t, filepath.Join(root, "flutter_app", "test", "version_api_test.dart"))
	if !strings.Contains(versionTest, "loads version from backend") {
		t.Fatalf("version_api_test.dart must cover the backend version call")
	}
}

func TestFlutterFeedsApiClientTargetsInventoryRoute(t *testing.T) {
	root := repositoryRoot(t)

	inventoryBody := readText(t, filepath.Join(root, "docs", "flutter", "API_ROUTE_INVENTORY.json"))
	var routes []string
	if err := json.Unmarshal([]byte(inventoryBody), &routes); err != nil {
		t.Fatalf("parse route inventory: %v", err)
	}

	hasFeedsRoute := false
	for _, route := range routes {
		if route == "/api/feeds" {
			hasFeedsRoute = true
			break
		}
	}
	if !hasFeedsRoute {
		t.Fatalf("route inventory must include /api/feeds for the Flutter feeds API client")
	}

	feedsAPI := readText(t, filepath.Join(root, "flutter_app", "lib", "src", "api", "feeds_api.dart"))
	if !strings.Contains(feedsAPI, "/api/feeds") {
		t.Fatalf("feeds_api.dart must call /api/feeds")
	}

	feedsTest := readText(t, filepath.Join(root, "flutter_app", "test", "feeds_api_test.dart"))
	if !strings.Contains(feedsTest, "loads feeds from backend") {
		t.Fatalf("feeds_api_test.dart must cover the backend feeds call")
	}
}

func TestFlutterFeedActionsApiClientTargetsInventoryRoutes(t *testing.T) {
	root := repositoryRoot(t)

	inventoryBody := readText(t, filepath.Join(root, "docs", "flutter", "API_ROUTE_INVENTORY.json"))
	var routes []string
	if err := json.Unmarshal([]byte(inventoryBody), &routes); err != nil {
		t.Fatalf("parse route inventory: %v", err)
	}

	hasRefreshRoute := false
	for _, route := range routes {
		if route == "/api/feeds/refresh" {
			hasRefreshRoute = true
			break
		}
	}
	if !hasRefreshRoute {
		t.Fatalf("route inventory must include /api/feeds/refresh for Flutter feed refresh actions")
	}

	feedActionsAPI := readText(t, filepath.Join(root, "flutter_app", "lib", "src", "api", "feed_actions_api.dart"))
	for _, want := range []string{"/api/feeds/refresh", "refreshFeed"} {
		if !strings.Contains(feedActionsAPI, want) {
			t.Fatalf("feed_actions_api.dart missing %q", want)
		}
	}

	feedActionsTest := readText(t, filepath.Join(root, "flutter_app", "test", "feed_actions_api_test.dart"))
	if !strings.Contains(feedActionsTest, "refreshes feed through backend") {
		t.Fatalf("feed_actions_api_test.dart must cover the backend feed refresh call")
	}
	if !strings.Contains(feedActionsTest, "refreshing") {
		t.Fatalf("feed_actions_api_test.dart must assert the backend refresh status")
	}
}

func TestFlutterSettingsApiClientTargetsInventoryRoute(t *testing.T) {
	root := repositoryRoot(t)

	inventoryBody := readText(t, filepath.Join(root, "docs", "flutter", "API_ROUTE_INVENTORY.json"))
	var routes []string
	if err := json.Unmarshal([]byte(inventoryBody), &routes); err != nil {
		t.Fatalf("parse route inventory: %v", err)
	}

	hasSettingsRoute := false
	for _, route := range routes {
		if route == "/api/settings" {
			hasSettingsRoute = true
			break
		}
	}
	if !hasSettingsRoute {
		t.Fatalf("route inventory must include /api/settings for the Flutter settings API client")
	}

	settingsAPI := readText(t, filepath.Join(root, "flutter_app", "lib", "src", "api", "settings_api.dart"))
	for _, want := range []string{"/api/settings", "getSettings", "updateSettings", "jsonEncode(settings)"} {
		if !strings.Contains(settingsAPI, want) {
			t.Fatalf("settings_api.dart missing %q", want)
		}
	}

	settingsTest := readText(t, filepath.Join(root, "flutter_app", "test", "settings_api_test.dart"))
	for _, want := range []string{"loads settings from backend", "updates settings through backend"} {
		if !strings.Contains(settingsTest, want) {
			t.Fatalf("settings_api_test.dart missing %q", want)
		}
	}
}

func TestFlutterArticlesApiClientTargetsInventoryRoute(t *testing.T) {
	root := repositoryRoot(t)

	inventoryBody := readText(t, filepath.Join(root, "docs", "flutter", "API_ROUTE_INVENTORY.json"))
	var routes []string
	if err := json.Unmarshal([]byte(inventoryBody), &routes); err != nil {
		t.Fatalf("parse route inventory: %v", err)
	}

	hasArticlesRoute := false
	for _, route := range routes {
		if route == "/api/articles" {
			hasArticlesRoute = true
			break
		}
	}
	if !hasArticlesRoute {
		t.Fatalf("route inventory must include /api/articles for the Flutter articles API client")
	}

	articlesAPI := readText(t, filepath.Join(root, "flutter_app", "lib", "src", "api", "articles_api.dart"))
	if !strings.Contains(articlesAPI, "/api/articles") {
		t.Fatalf("articles_api.dart must call /api/articles")
	}

	articlesTest := readText(t, filepath.Join(root, "flutter_app", "test", "articles_api_test.dart"))
	if !strings.Contains(articlesTest, "loads articles from backend") {
		t.Fatalf("articles_api_test.dart must cover the backend articles call")
	}
}

func TestFlutterArticleContentApiClientTargetsInventoryRoute(t *testing.T) {
	root := repositoryRoot(t)

	inventoryBody := readText(t, filepath.Join(root, "docs", "flutter", "API_ROUTE_INVENTORY.json"))
	var routes []string
	if err := json.Unmarshal([]byte(inventoryBody), &routes); err != nil {
		t.Fatalf("parse route inventory: %v", err)
	}

	hasContentRoute := false
	for _, route := range routes {
		if route == "/api/articles/content" {
			hasContentRoute = true
			break
		}
	}
	if !hasContentRoute {
		t.Fatalf("route inventory must include /api/articles/content for the Flutter article content API client")
	}

	contentAPI := readText(t, filepath.Join(root, "flutter_app", "lib", "src", "api", "article_content_api.dart"))
	if !strings.Contains(contentAPI, "/api/articles/content") {
		t.Fatalf("article_content_api.dart must call /api/articles/content")
	}

	contentTest := readText(t, filepath.Join(root, "flutter_app", "test", "article_content_api_test.dart"))
	if !strings.Contains(contentTest, "loads article content from backend") {
		t.Fatalf("article_content_api_test.dart must cover the backend article content call")
	}
}

func TestFlutterArticleStatusApiClientTargetsInventoryRoutes(t *testing.T) {
	root := repositoryRoot(t)

	inventoryBody := readText(t, filepath.Join(root, "docs", "flutter", "API_ROUTE_INVENTORY.json"))
	var routes []string
	if err := json.Unmarshal([]byte(inventoryBody), &routes); err != nil {
		t.Fatalf("parse route inventory: %v", err)
	}

	requiredRoutes := map[string]bool{
		"/api/articles/read":     false,
		"/api/articles/favorite": false,
	}
	for _, route := range routes {
		if _, ok := requiredRoutes[route]; ok {
			requiredRoutes[route] = true
		}
	}
	for route, found := range requiredRoutes {
		if !found {
			t.Fatalf("route inventory must include %s for Flutter article status actions", route)
		}
	}

	statusAPI := readText(t, filepath.Join(root, "flutter_app", "lib", "src", "api", "article_status_api.dart"))
	for _, want := range []string{"/api/articles/read", "/api/articles/favorite", "markRead", "toggleFavorite"} {
		if !strings.Contains(statusAPI, want) {
			t.Fatalf("article_status_api.dart missing %q", want)
		}
	}

	statusTest := readText(t, filepath.Join(root, "flutter_app", "test", "article_status_api_test.dart"))
	for _, want := range []string{"marks article read through backend", "toggles favorite through backend"} {
		if !strings.Contains(statusTest, want) {
			t.Fatalf("article_status_api_test.dart missing %q", want)
		}
	}
}

func TestFlutterTranslationApiClientTargetsInventoryRoute(t *testing.T) {
	root := repositoryRoot(t)

	inventoryBody := readText(t, filepath.Join(root, "docs", "flutter", "API_ROUTE_INVENTORY.json"))
	var routes []string
	if err := json.Unmarshal([]byte(inventoryBody), &routes); err != nil {
		t.Fatalf("parse route inventory: %v", err)
	}

	hasTranslateRoute := false
	for _, route := range routes {
		if route == "/api/articles/translate-text" {
			hasTranslateRoute = true
			break
		}
	}
	if !hasTranslateRoute {
		t.Fatalf("route inventory must include /api/articles/translate-text for the Flutter translation API client")
	}

	translationAPI := readText(t, filepath.Join(root, "flutter_app", "lib", "src", "api", "translation_api.dart"))
	for _, want := range []string{"/api/articles/translate-text", "/api/articles/clear-translations", "translateText", "clearTranslations", "target_language"} {
		if !strings.Contains(translationAPI, want) {
			t.Fatalf("translation_api.dart missing %q", want)
		}
	}

	translationTest := readText(t, filepath.Join(root, "flutter_app", "test", "translation_api_test.dart"))
	for _, want := range []string{"translates text through backend", "clears translation cache through backend"} {
		if !strings.Contains(translationTest, want) {
			t.Fatalf("translation_api_test.dart missing %q", want)
		}
	}
}

func TestFlutterOpmlApiClientTargetsInventoryRoutes(t *testing.T) {
	root := repositoryRoot(t)

	inventoryBody := readText(t, filepath.Join(root, "docs", "flutter", "API_ROUTE_INVENTORY.json"))
	var routes []string
	if err := json.Unmarshal([]byte(inventoryBody), &routes); err != nil {
		t.Fatalf("parse route inventory: %v", err)
	}

	requiredRoutes := map[string]bool{
		"/api/opml/export": false,
		"/api/opml/import": false,
	}
	for _, route := range routes {
		if _, ok := requiredRoutes[route]; ok {
			requiredRoutes[route] = true
		}
	}
	for route, found := range requiredRoutes {
		if !found {
			t.Fatalf("route inventory must include %s for the Flutter OPML API client", route)
		}
	}

	opmlAPI := readText(t, filepath.Join(root, "flutter_app", "lib", "src", "api", "opml_api.dart"))
	for _, want := range []string{"/api/opml/export", "/api/opml/import", "exportOpml", "importOpmlText"} {
		if !strings.Contains(opmlAPI, want) {
			t.Fatalf("opml_api.dart missing %q", want)
		}
	}

	opmlTest := readText(t, filepath.Join(root, "flutter_app", "test", "opml_api_test.dart"))
	for _, want := range []string{"exports OPML through backend", "imports OPML text through backend"} {
		if !strings.Contains(opmlTest, want) {
			t.Fatalf("opml_api_test.dart missing %q", want)
		}
	}
}

func TestFlutterReaderWorkflowUsesReadOnlyRepository(t *testing.T) {
	root := repositoryRoot(t)

	app := readText(t, filepath.Join(root, "flutter_app", "lib", "src", "app", "mrrss_app.dart"))
	for _, want := range []string{"ReaderScreen(", "opmlFileService", "FilePickerOpmlFileService"} {
		if !strings.Contains(app, want) {
			t.Fatalf("MrRssApp missing %q", want)
		}
	}
	if !strings.Contains(app, "repository: repository") {
		t.Fatalf("MrRssApp must mount ReaderScreen with an injectable repository")
	}

	opmlFileService := readText(t, filepath.Join(root, "flutter_app", "lib", "src", "platform", "opml_file_service.dart"))
	for _, want := range []string{"abstract class OpmlFileService", "FilePicker.saveFile", "FilePicker.pickFiles", "allowedExtensions: const ['opml', 'xml']", "utf8.decode"} {
		if !strings.Contains(opmlFileService, want) {
			t.Fatalf("opml_file_service.dart missing %q", want)
		}
	}

	repository := readText(t, filepath.Join(root, "flutter_app", "lib", "src", "reader", "reader_repository.dart"))
	for _, want := range []string{"abstract class ReaderRepository", "Future<ReaderSnapshot> loadInitial()", "Future<ArticleContent> loadArticleContent", "markArticleRead", "toggleArticleFavorite", "refreshFeed", "updateSettings", "exportOpml", "importOpmlText", "translateText", "clearTranslations", "List<Feed>", "List<Article>"} {
		if !strings.Contains(repository, want) {
			t.Fatalf("reader_repository.dart missing %q", want)
		}
	}

	httpRepository := readText(t, filepath.Join(root, "flutter_app", "lib", "src", "reader", "http_reader_repository.dart"))
	for _, want := range []string{"VersionApi", "SettingsApi", "FeedsApi", "ArticlesApi", "ArticleContentApi", "ArticleStatusApi", "FeedActionsApi", "OpmlApi", "TranslationApi", "listArticles(filter: 'all'", "getContent(articleId)", "markRead", "toggleFavorite", "refreshFeed", "updateSettings", "exportOpml", "importOpmlText", "translateText", "clearTranslations"} {
		if !strings.Contains(httpRepository, want) {
			t.Fatalf("http_reader_repository.dart missing %q", want)
		}
	}

	screen := readText(t, filepath.Join(root, "flutter_app", "lib", "src", "reader", "reader_screen.dart"))
	for _, want := range []string{"FutureBuilder<ReaderSnapshot>", "FeedSidebar", "ArticleListPanel", "ArticleDetailPanel", "SettingsSheet", "opmlFileService", "loadArticleContent", "_showCompactDetail", "Back to articles", "No articles", "Retry"} {
		if !strings.Contains(screen, want) {
			t.Fatalf("reader_screen.dart missing %q", want)
		}
	}

	detail := readText(t, filepath.Join(root, "flutter_app", "lib", "src", "reader", "widgets", "article_detail_panel.dart"))
	for _, want := range []string{"FutureBuilder<ArticleContent>", "htmlToReadableText", "SelectableText", "Mark read", "Favorite", "Translate summary", "Translate content", "_contentFuture", "No article content available"} {
		if !strings.Contains(detail, want) {
			t.Fatalf("article_detail_panel.dart missing %q", want)
		}
	}

	htmlText := readText(t, filepath.Join(root, "flutter_app", "lib", "src", "utils", "html_text.dart"))
	for _, want := range []string{"htmlToReadableText", "_decodeHtmlEntities", "<script", "<style"} {
		if !strings.Contains(htmlText, want) {
			t.Fatalf("html_text.dart missing %q", want)
		}
	}

	sidebar := readText(t, filepath.Join(root, "flutter_app", "lib", "src", "reader", "widgets", "feed_sidebar.dart"))
	for _, want := range []string{"onRefreshFeed", "Refresh feed", "Icons.refresh"} {
		if !strings.Contains(sidebar, want) {
			t.Fatalf("feed_sidebar.dart missing %q", want)
		}
	}

	settingsSheet := readText(t, filepath.Join(root, "flutter_app", "lib", "src", "reader", "widgets", "settings_sheet.dart"))
	for _, want := range []string{"Save settings", "update_interval", "language", "theme", "translation_enabled", "translation_only_mode", "translation_provider", "target_language", "deepl_api_key", "deepl_endpoint", "baidu_app_id", "baidu_secret_key", "microsoft_api_key", "microsoft_endpoint", "microsoft_region", "tencent_secret_id", "tencent_secret_key", "tencent_region", "ai_translation_profile_id", "ai_translation_prompt", "custom_translation_enabled", "custom_translation_name", "custom_translation_endpoint", "custom_translation_method", "custom_translation_headers", "custom_translation_body_template", "custom_translation_response_path", "custom_translation_lang_mapping", "custom_translation_timeout", "Enable translation", "Translation provider", "Target language", "DeepL API key", "DeepL endpoint", "Baidu app ID", "Baidu secret key", "Microsoft API key", "Microsoft endpoint", "Microsoft region", "Tencent secret ID", "Tencent secret key", "Tencent region", "AI translation profile ID", "AI translation prompt", "Enable custom translation", "Custom endpoint", "Custom method", "Custom headers JSON", "Custom body template", "Custom response path", "Custom language mapping", "Custom timeout", "Clear translation cache", "Export OPML", "Import OPML", "Export OPML file", "Import OPML file", "OPML import text", "Refresh interval must be a positive number"} {
		if !strings.Contains(settingsSheet, want) {
			t.Fatalf("settings_sheet.dart missing %q", want)
		}
	}
}

func TestFlutterReaderWorkflowHasWidgetCoverage(t *testing.T) {
	root := repositoryRoot(t)

	testFile := readText(t, filepath.Join(root, "flutter_app", "test", "reader_screen_test.dart"))
	for _, want := range []string{
		"shows loading state",
		"shows empty state",
		"shows error state",
		"renders feeds, article list, and selected detail",
		"selects another article",
		"First Article body",
		"opens article detail and returns on compact width",
		"Back to articles",
		"invokes article status actions",
		"invokes feed refresh",
		"markReadCalls",
		"favoriteCalls",
		"refreshFeedCalls",
		"opens settings and saves bootstrap settings",
		"translation_provider",
		"deepl_api_key",
		"saves Baidu translation provider settings",
		"baidu_secret_key",
		"saves Microsoft translation provider settings",
		"microsoft_region",
		"saves Tencent translation provider settings",
		"tencent_secret_key",
		"saves AI translation provider settings",
		"ai_translation_prompt",
		"saves custom translation provider core settings",
		"custom_translation_headers",
		"custom_translation_response_path",
		"custom_translation_lang_mapping",
		"settingsUpdates",
		"clears translation cache",
		"clearTranslationCalls",
		"exports and imports OPML",
		"opmlImports",
		"exports and imports OPML files",
		"OpmlFileService",
		"translates article summary",
		"translates article content",
		"translationRequests",
	} {
		if !strings.Contains(testFile, want) {
			t.Fatalf("reader_screen_test.dart missing widget coverage %q", want)
		}
	}
}

func TestFlutterGeneratedArtifactsAreIgnored(t *testing.T) {
	root := repositoryRoot(t)
	gitignore := readText(t, filepath.Join(root, ".gitignore"))

	required := []string{
		".dart_tool/",
		".flutter-plugins",
		".flutter-plugins-dependencies",
		"flutter_app/build/",
		"flutter_app/.dart_tool/",
	}

	for _, want := range required {
		if !strings.Contains(gitignore, want) {
			t.Fatalf(".gitignore missing Flutter generated artifact rule %q", want)
		}
	}
}

func readText(t *testing.T, path string) string {
	t.Helper()

	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}

	return string(body)
}

func repositoryRoot(t *testing.T) string {
	t.Helper()

	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("could not find repository root from %s", dir)
		}
		dir = parent
	}
}
