import '../models/app_settings.dart';
import '../models/article.dart';
import '../models/article_content.dart';
import '../models/feed.dart';
import '../models/translation_result.dart';
import '../models/version_info.dart';

abstract class ReaderRepository {
  Future<ReaderSnapshot> loadInitial();

  Future<ArticleContent> loadArticleContent(int articleId);

  Future<void> markArticleRead({
    required int articleId,
    required bool read,
  });

  Future<void> toggleArticleFavorite(int articleId);

  Future<void> refreshFeed(int feedId);

  Future<AppSettings> updateSettings(Map<String, String> settings);

  Future<String> exportOpml();

  Future<void> importOpmlText(String opmlText);

  Future<TranslationResult> translateText({
    required String text,
    required String targetLanguage,
    bool force,
  });

  Future<void> clearTranslations();
}

class ReaderSnapshot {
  const ReaderSnapshot({
    required this.version,
    required this.settings,
    required this.feeds,
    required this.articles,
  });

  final VersionInfo version;
  final AppSettings settings;
  final List<Feed> feeds;
  final List<Article> articles;
}
