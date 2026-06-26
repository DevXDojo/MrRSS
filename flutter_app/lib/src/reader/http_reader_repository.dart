import '../api/api_config.dart';
import '../api/article_content_api.dart';
import '../api/article_status_api.dart';
import '../api/articles_api.dart';
import '../api/feed_actions_api.dart';
import '../api/feeds_api.dart';
import '../api/opml_api.dart';
import '../api/settings_api.dart';
import '../api/translation_api.dart';
import '../api/version_api.dart';
import '../models/article_content.dart';
import '../models/app_settings.dart';
import '../models/translation_result.dart';
import 'reader_repository.dart';

class HttpReaderRepository implements ReaderRepository {
  HttpReaderRepository({
    required ApiConfig config,
    VersionApi? versionApi,
    SettingsApi? settingsApi,
    FeedsApi? feedsApi,
    ArticlesApi? articlesApi,
    ArticleContentApi? articleContentApi,
    ArticleStatusApi? articleStatusApi,
    FeedActionsApi? feedActionsApi,
    OpmlApi? opmlApi,
    TranslationApi? translationApi,
  })  : _versionApi = versionApi ?? VersionApi(config: config),
        _settingsApi = settingsApi ?? SettingsApi(config: config),
        _feedsApi = feedsApi ?? FeedsApi(config: config),
        _articlesApi = articlesApi ?? ArticlesApi(config: config),
        _articleContentApi =
            articleContentApi ?? ArticleContentApi(config: config),
        _articleStatusApi =
            articleStatusApi ?? ArticleStatusApi(config: config),
        _feedActionsApi = feedActionsApi ?? FeedActionsApi(config: config),
        _opmlApi = opmlApi ?? OpmlApi(config: config),
        _translationApi = translationApi ?? TranslationApi(config: config);

  final VersionApi _versionApi;
  final SettingsApi _settingsApi;
  final FeedsApi _feedsApi;
  final ArticlesApi _articlesApi;
  final ArticleContentApi _articleContentApi;
  final ArticleStatusApi _articleStatusApi;
  final FeedActionsApi _feedActionsApi;
  final OpmlApi _opmlApi;
  final TranslationApi _translationApi;

  @override
  Future<ReaderSnapshot> loadInitial() async {
    final version = await _versionApi.getVersion();
    final settings = await _settingsApi.getSettings();
    final feeds = await _feedsApi.listFeeds();
    final articles = await _articlesApi.listArticles(filter: 'all', limit: 50);

    return ReaderSnapshot(
      version: version,
      settings: settings,
      feeds: feeds,
      articles: articles,
    );
  }

  @override
  Future<ArticleContent> loadArticleContent(int articleId) {
    return _articleContentApi.getContent(articleId);
  }

  @override
  Future<void> markArticleRead({
    required int articleId,
    required bool read,
  }) {
    return _articleStatusApi.markRead(articleId: articleId, read: read);
  }

  @override
  Future<void> toggleArticleFavorite(int articleId) {
    return _articleStatusApi.toggleFavorite(articleId);
  }

  @override
  Future<void> refreshFeed(int feedId) async {
    await _feedActionsApi.refreshFeed(feedId);
  }

  @override
  Future<AppSettings> updateSettings(Map<String, String> settings) {
    return _settingsApi.updateSettings(settings);
  }

  @override
  Future<String> exportOpml() {
    return _opmlApi.exportOpml();
  }

  @override
  Future<void> importOpmlText(String opmlText) {
    return _opmlApi.importOpmlText(opmlText);
  }

  @override
  Future<TranslationResult> translateText({
    required String text,
    required String targetLanguage,
    bool force = false,
  }) {
    return _translationApi.translateText(
      text: text,
      targetLanguage: targetLanguage,
      force: force,
    );
  }

  @override
  Future<void> clearTranslations() {
    return _translationApi.clearTranslations();
  }
}
