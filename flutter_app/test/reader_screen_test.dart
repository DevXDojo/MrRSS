import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mrrss_flutter/src/app/mrrss_app.dart';
import 'package:mrrss_flutter/src/models/app_settings.dart';
import 'package:mrrss_flutter/src/models/article.dart';
import 'package:mrrss_flutter/src/models/article_content.dart';
import 'package:mrrss_flutter/src/models/feed.dart';
import 'package:mrrss_flutter/src/models/translation_result.dart';
import 'package:mrrss_flutter/src/models/version_info.dart';
import 'package:mrrss_flutter/src/platform/opml_file_service.dart';
import 'package:mrrss_flutter/src/reader/reader_repository.dart';

void main() {
  testWidgets('shows loading state while reader data is pending',
      (tester) async {
    final completer = Completer<ReaderSnapshot>();

    await tester.pumpWidget(
      MrRssApp(readerRepository: _FakeReaderRepository(completer.future)),
    );

    expect(find.byType(CircularProgressIndicator), findsOneWidget);
  });

  testWidgets('shows empty state when there are no articles', (tester) async {
    await tester.pumpWidget(
      MrRssApp(
        readerRepository: _FakeReaderRepository(
          Future.value(_snapshot(articles: const [])),
        ),
      ),
    );
    await tester.pumpAndSettle();

    expect(find.text('No articles'), findsOneWidget);
  });

  testWidgets('shows error state and can retry', (tester) async {
    var calls = 0;
    final repository = _CallbackReaderRepository(() {
      calls += 1;
      if (calls == 1) {
        return Future<ReaderSnapshot>.error(Exception('network failed'));
      }
      return Future.value(_snapshot());
    });

    await tester.pumpWidget(MrRssApp(readerRepository: repository));
    await tester.pumpAndSettle();

    expect(find.text('Failed to load reader data'), findsOneWidget);

    await tester.tap(find.text('Retry'));
    await tester.pumpAndSettle();

    expect(find.text('First Article'), findsWidgets);
  });

  testWidgets('renders feeds, article list, and selected detail',
      (tester) async {
    await tester.binding.setSurfaceSize(const Size(1200, 800));
    addTearDown(() => tester.binding.setSurfaceSize(null));

    await tester.pumpWidget(
      MrRssApp(
        readerRepository: _FakeReaderRepository(Future.value(_snapshot())),
      ),
    );
    await tester.pumpAndSettle();

    expect(find.text('All articles'), findsOneWidget);
    expect(find.text('Tech Feed'), findsWidgets);
    expect(find.text('First Article'), findsWidgets);
    expect(find.text('First summary'), findsOneWidget);
    expect(find.text('First Article body'), findsOneWidget);
    expect(find.text('Mark read'), findsOneWidget);
    expect(find.text('Unfavorite'), findsOneWidget);
    expect(find.byTooltip('Refresh feed'), findsOneWidget);
  });

  testWidgets('selects another article from the list', (tester) async {
    await tester.binding.setSurfaceSize(const Size(1200, 800));
    addTearDown(() => tester.binding.setSurfaceSize(null));

    await tester.pumpWidget(
      MrRssApp(
        readerRepository: _FakeReaderRepository(Future.value(_snapshot())),
      ),
    );
    await tester.pumpAndSettle();

    await tester.tap(find.text('Second Article').first);
    await tester.pumpAndSettle();

    expect(find.text('Second summary'), findsOneWidget);
  });

  testWidgets('opens article detail and returns on compact width',
      (tester) async {
    await tester.binding.setSurfaceSize(const Size(420, 800));
    addTearDown(() => tester.binding.setSurfaceSize(null));

    await tester.pumpWidget(
      MrRssApp(
        readerRepository: _FakeReaderRepository(Future.value(_snapshot())),
      ),
    );
    await tester.pumpAndSettle();

    expect(find.text('Second Article'), findsOneWidget);
    expect(find.text('Back to articles'), findsNothing);

    await tester.tap(find.text('Second Article'));
    await tester.pumpAndSettle();

    expect(find.text('Back to articles'), findsOneWidget);
    expect(find.text('Second Article body'), findsOneWidget);

    await tester.tap(find.text('Back to articles'));
    await tester.pumpAndSettle();

    expect(find.text('Back to articles'), findsNothing);
    expect(find.text('Second Article'), findsOneWidget);
  });

  testWidgets('invokes article status actions from detail panel',
      (tester) async {
    await tester.binding.setSurfaceSize(const Size(1200, 800));
    addTearDown(() => tester.binding.setSurfaceSize(null));

    final repository = _RecordingReaderRepository(Future.value(_snapshot()));

    await tester.pumpWidget(MrRssApp(readerRepository: repository));
    await tester.pumpAndSettle();

    await tester.tap(find.text('Mark read'));
    await tester.tap(find.text('Unfavorite'));
    await tester.pumpAndSettle();

    expect(repository.markReadCalls, ['1:true']);
    expect(repository.favoriteCalls, [1]);
  });

  testWidgets('translates article summary from detail panel', (tester) async {
    await tester.binding.setSurfaceSize(const Size(1200, 800));
    addTearDown(() => tester.binding.setSurfaceSize(null));

    final repository = _RecordingReaderRepository(Future.value(_snapshot()));

    await tester.pumpWidget(MrRssApp(readerRepository: repository));
    await tester.pumpAndSettle();

    await tester.tap(find.text('Translate summary'));
    await tester.pumpAndSettle();

    expect(repository.translationRequests.single, 'First summary:zh:true');
    expect(find.text('[ZH] First summary'), findsOneWidget);
  });

  testWidgets('translates article content from detail panel', (tester) async {
    await tester.binding.setSurfaceSize(const Size(1200, 800));
    addTearDown(() => tester.binding.setSurfaceSize(null));

    final repository = _RecordingReaderRepository(Future.value(_snapshot()));

    await tester.pumpWidget(MrRssApp(readerRepository: repository));
    await tester.pumpAndSettle();

    await tester.tap(find.text('Translate content'));
    await tester.pumpAndSettle();

    expect(repository.translationRequests.single, 'First Article body:zh:true');
    expect(find.text('[ZH] First Article body'), findsOneWidget);

    await tester.tap(find.text('Second Article').first);
    await tester.pumpAndSettle();

    expect(find.text('[ZH] First Article body'), findsNothing);
    expect(find.text('Second Article body'), findsOneWidget);
  });

  testWidgets('invokes feed refresh from desktop sidebar', (tester) async {
    await tester.binding.setSurfaceSize(const Size(1200, 800));
    addTearDown(() => tester.binding.setSurfaceSize(null));

    final repository = _RecordingReaderRepository(Future.value(_snapshot()));

    await tester.pumpWidget(MrRssApp(readerRepository: repository));
    await tester.pumpAndSettle();

    await tester.tap(find.byTooltip('Refresh feed'));
    await tester.pumpAndSettle();

    expect(repository.refreshFeedCalls, [1]);
  });

  testWidgets('opens settings and saves bootstrap settings', (tester) async {
    final repository = _RecordingReaderRepository(Future.value(_snapshot()));

    await tester.pumpWidget(MrRssApp(readerRepository: repository));
    await tester.pumpAndSettle();

    await tester.tap(find.byTooltip('Settings'));
    await tester.pumpAndSettle();

    expect(find.text('Settings'), findsOneWidget);
    expect(find.text('Save settings'), findsOneWidget);

    await tester.enterText(find.widgetWithText(TextField, 'Language'), 'zh-CN');
    await tester.enterText(
      find.widgetWithText(TextField, 'Refresh interval'),
      '45',
    );
    await tester.ensureVisible(find.text('Enable translation'));
    await tester.tap(find.text('Enable translation'));
    await tester.pumpAndSettle();

    await tester.tap(find.text('Translation only mode'));
    await tester.pumpAndSettle();

    await tester
        .ensureVisible(find.widgetWithText(TextField, 'Target language'));
    await tester.enterText(
      find.widgetWithText(TextField, 'Target language'),
      'ja',
    );

    await tester.ensureVisible(find.text('Google'));
    await tester.tap(find.text('Google').last);
    await tester.pumpAndSettle();
    await tester.tap(find.text('DeepL').last);
    await tester.pumpAndSettle();

    await tester.enterText(
      find.widgetWithText(TextField, 'DeepL API key'),
      'deepl-secret',
    );
    await tester.enterText(
      find.widgetWithText(TextField, 'DeepL endpoint'),
      'https://api-free.deepl.com/v2/translate',
    );

    await tester.ensureVisible(find.text('Save settings'));
    await tester.tap(find.text('Save settings'));
    await tester.pumpAndSettle();

    expect(repository.settingsUpdates.single['language'], 'zh-CN');
    expect(repository.settingsUpdates.single['update_interval'], '45');
    expect(repository.settingsUpdates.single.containsKey('theme'), isTrue);
    expect(repository.settingsUpdates.single['translation_enabled'], 'true');
    expect(repository.settingsUpdates.single['translation_only_mode'], 'true');
    expect(repository.settingsUpdates.single['target_language'], 'ja');
    expect(repository.settingsUpdates.single['translation_provider'], 'deepl');
    expect(repository.settingsUpdates.single['deepl_api_key'], 'deepl-secret');
    expect(
      repository.settingsUpdates.single['deepl_endpoint'],
      'https://api-free.deepl.com/v2/translate',
    );
  });

  testWidgets('saves Baidu translation provider settings', (tester) async {
    final repository = _RecordingReaderRepository(Future.value(_snapshot()));

    await tester.pumpWidget(MrRssApp(readerRepository: repository));
    await tester.pumpAndSettle();

    await tester.tap(find.byTooltip('Settings'));
    await tester.pumpAndSettle();

    await tester.ensureVisible(find.text('Google'));
    await tester.tap(find.text('Google').last);
    await tester.pumpAndSettle();
    await tester.tap(find.text('Baidu').last);
    await tester.pumpAndSettle();

    await tester.enterText(
      find.widgetWithText(TextField, 'Baidu app ID'),
      'baidu-app',
    );
    await tester.enterText(
      find.widgetWithText(TextField, 'Baidu secret key'),
      'baidu-secret',
    );

    await tester.ensureVisible(find.text('Save settings'));
    await tester.tap(find.text('Save settings'));
    await tester.pumpAndSettle();

    expect(repository.settingsUpdates.single['translation_provider'], 'baidu');
    expect(repository.settingsUpdates.single['baidu_app_id'], 'baidu-app');
    expect(
      repository.settingsUpdates.single['baidu_secret_key'],
      'baidu-secret',
    );
  });

  testWidgets('saves Microsoft translation provider settings', (tester) async {
    final repository = _RecordingReaderRepository(Future.value(_snapshot()));

    await tester.pumpWidget(MrRssApp(readerRepository: repository));
    await tester.pumpAndSettle();

    await tester.tap(find.byTooltip('Settings'));
    await tester.pumpAndSettle();

    await tester.ensureVisible(find.text('Google'));
    await tester.tap(find.text('Google').last);
    await tester.pumpAndSettle();
    await tester.tap(find.text('Microsoft').last);
    await tester.pumpAndSettle();

    await tester.enterText(
      find.widgetWithText(TextField, 'Microsoft API key'),
      'ms-key',
    );
    await tester.enterText(
      find.widgetWithText(TextField, 'Microsoft endpoint'),
      'https://api.cognitive.microsofttranslator.com',
    );
    await tester.enterText(
      find.widgetWithText(TextField, 'Microsoft region'),
      'eastasia',
    );

    await tester.ensureVisible(find.text('Save settings'));
    await tester.tap(find.text('Save settings'));
    await tester.pumpAndSettle();

    expect(
      repository.settingsUpdates.single['translation_provider'],
      'microsoft',
    );
    expect(repository.settingsUpdates.single['microsoft_api_key'], 'ms-key');
    expect(
      repository.settingsUpdates.single['microsoft_endpoint'],
      'https://api.cognitive.microsofttranslator.com',
    );
    expect(repository.settingsUpdates.single['microsoft_region'], 'eastasia');
  });

  testWidgets('saves Tencent translation provider settings', (tester) async {
    final repository = _RecordingReaderRepository(Future.value(_snapshot()));

    await tester.pumpWidget(MrRssApp(readerRepository: repository));
    await tester.pumpAndSettle();

    await tester.tap(find.byTooltip('Settings'));
    await tester.pumpAndSettle();

    await tester.ensureVisible(find.text('Google'));
    await tester.tap(find.text('Google').last);
    await tester.pumpAndSettle();
    await tester.tap(find.text('Tencent').last);
    await tester.pumpAndSettle();

    await tester.enterText(
      find.widgetWithText(TextField, 'Tencent secret ID'),
      'tencent-id',
    );
    await tester.enterText(
      find.widgetWithText(TextField, 'Tencent secret key'),
      'tencent-key',
    );
    await tester.enterText(
      find.widgetWithText(TextField, 'Tencent region'),
      'ap-shanghai',
    );

    await tester.ensureVisible(find.text('Save settings'));
    await tester.tap(find.text('Save settings'));
    await tester.pumpAndSettle();

    expect(
      repository.settingsUpdates.single['translation_provider'],
      'tencent',
    );
    expect(
      repository.settingsUpdates.single['tencent_secret_id'],
      'tencent-id',
    );
    expect(
      repository.settingsUpdates.single['tencent_secret_key'],
      'tencent-key',
    );
    expect(repository.settingsUpdates.single['tencent_region'], 'ap-shanghai');
  });

  testWidgets('saves AI translation provider settings', (tester) async {
    final repository = _RecordingReaderRepository(Future.value(_snapshot()));

    await tester.pumpWidget(MrRssApp(readerRepository: repository));
    await tester.pumpAndSettle();

    await tester.tap(find.byTooltip('Settings'));
    await tester.pumpAndSettle();

    await tester.ensureVisible(find.text('Google'));
    await tester.tap(find.text('Google').last);
    await tester.pumpAndSettle();
    await tester.tap(find.text('AI').last);
    await tester.pumpAndSettle();

    await tester.enterText(
      find.widgetWithText(TextField, 'AI translation profile ID'),
      'profile-1',
    );
    await tester.enterText(
      find.widgetWithText(TextField, 'AI translation prompt'),
      'Translate precisely.',
    );

    await tester.ensureVisible(find.text('Save settings'));
    await tester.tap(find.text('Save settings'));
    await tester.pumpAndSettle();

    expect(repository.settingsUpdates.single['translation_provider'], 'ai');
    expect(
      repository.settingsUpdates.single['ai_translation_profile_id'],
      'profile-1',
    );
    expect(
      repository.settingsUpdates.single['ai_translation_prompt'],
      'Translate precisely.',
    );
  });

  testWidgets('saves custom translation provider core settings',
      (tester) async {
    final repository = _RecordingReaderRepository(Future.value(_snapshot()));

    await tester.pumpWidget(MrRssApp(readerRepository: repository));
    await tester.pumpAndSettle();

    await tester.tap(find.byTooltip('Settings'));
    await tester.pumpAndSettle();

    await tester.ensureVisible(find.text('Google'));
    await tester.tap(find.text('Google').last);
    await tester.pumpAndSettle();
    await tester.tap(find.text('Custom').last);
    await tester.pumpAndSettle();

    await tester.tap(find.text('Enable custom translation'));
    await tester.pumpAndSettle();
    await tester.enterText(
      find.widgetWithText(TextField, 'Custom name'),
      'Acme Translate',
    );
    await tester.enterText(
      find.widgetWithText(TextField, 'Custom endpoint'),
      'https://translate.example.com',
    );
    await tester.enterText(
      find.widgetWithText(TextField, 'Custom headers JSON'),
      '{"X-API-Key":"secret"}',
    );
    await tester.enterText(
      find.widgetWithText(TextField, 'Custom body template'),
      '{"text":"{{text}}"}',
    );
    await tester.enterText(
      find.widgetWithText(TextField, 'Custom response path'),
      'data.text',
    );
    await tester.enterText(
      find.widgetWithText(TextField, 'Custom language mapping'),
      '{"zh":"zh-CN","en":"en-US"}',
    );
    await tester.enterText(
      find.widgetWithText(TextField, 'Custom timeout'),
      '15',
    );

    await tester.drag(
      find.byType(SingleChildScrollView),
      const Offset(0, -320),
    );
    await tester.pumpAndSettle();
    await tester.tap(find.widgetWithText(FilledButton, 'Save settings'));
    await tester.pumpAndSettle();

    expect(repository.settingsUpdates.single['translation_provider'], 'custom');
    expect(
      repository.settingsUpdates.single['custom_translation_enabled'],
      'true',
    );
    expect(
      repository.settingsUpdates.single['custom_translation_name'],
      'Acme Translate',
    );
    expect(
      repository.settingsUpdates.single['custom_translation_endpoint'],
      'https://translate.example.com',
    );
    expect(
      repository.settingsUpdates.single['custom_translation_method'],
      'POST',
    );
    expect(
      repository.settingsUpdates.single['custom_translation_headers'],
      '{"X-API-Key":"secret"}',
    );
    expect(
      repository.settingsUpdates.single['custom_translation_body_template'],
      '{"text":"{{text}}"}',
    );
    expect(
      repository.settingsUpdates.single['custom_translation_response_path'],
      'data.text',
    );
    expect(
      repository.settingsUpdates.single['custom_translation_lang_mapping'],
      '{"zh":"zh-CN","en":"en-US"}',
    );
    expect(
      repository.settingsUpdates.single['custom_translation_timeout'],
      '15',
    );
  });

  testWidgets('clears translation cache from settings', (tester) async {
    final repository = _RecordingReaderRepository(Future.value(_snapshot()));

    await tester.pumpWidget(MrRssApp(readerRepository: repository));
    await tester.pumpAndSettle();

    await tester.tap(find.byTooltip('Settings'));
    await tester.pumpAndSettle();

    await tester.ensureVisible(find.text('Clear translation cache'));
    await tester.tap(find.text('Clear translation cache'));
    await tester.pumpAndSettle();

    expect(repository.clearTranslationCalls, 1);
    expect(find.text('Translation cache cleared.'), findsOneWidget);
  });

  testWidgets('exports and imports OPML from settings', (tester) async {
    final repository = _RecordingReaderRepository(Future.value(_snapshot()));

    await tester.pumpWidget(MrRssApp(readerRepository: repository));
    await tester.pumpAndSettle();

    await tester.tap(find.byTooltip('Settings'));
    await tester.pumpAndSettle();

    await tester.ensureVisible(find.text('Export OPML'));
    await tester.tap(find.text('Export OPML'));
    await tester.pumpAndSettle();

    expect(repository.opmlExportCalls, 1);
    expect(find.textContaining('<opml'), findsOneWidget);

    const opml = '<?xml version="1.0"?><opml version="2.0"></opml>';
    await tester.enterText(
      find.widgetWithText(TextField, 'OPML import text'),
      opml,
    );
    await tester.ensureVisible(find.text('Import OPML'));
    await tester.tap(find.text('Import OPML'));
    await tester.pumpAndSettle();

    expect(repository.opmlImports, [opml]);
  });

  testWidgets('exports and imports OPML files from settings', (tester) async {
    final repository = _RecordingReaderRepository(Future.value(_snapshot()));
    final fileService = _RecordingOpmlFileService(
      pickedText: '<?xml version="1.0"?><opml version="2.0"></opml>',
    );

    await tester.pumpWidget(
      MrRssApp(
        readerRepository: repository,
        opmlFileService: fileService,
      ),
    );
    await tester.pumpAndSettle();

    await tester.tap(find.byTooltip('Settings'));
    await tester.pumpAndSettle();

    await tester.ensureVisible(find.text('Export OPML file'));
    await tester.tap(find.text('Export OPML file'));
    await tester.pumpAndSettle();

    expect(repository.opmlExportCalls, 1);
    expect(fileService.exportedTexts, [_opml()]);
    expect(find.textContaining('OPML exported to'), findsOneWidget);

    await tester.ensureVisible(find.text('Import OPML file'));
    await tester.tap(find.text('Import OPML file'));
    await tester.pumpAndSettle();

    expect(repository.opmlImports, [fileService.pickedText]);
    expect(find.text('OPML file imported.'), findsOneWidget);
  });
}

class _FakeReaderRepository implements ReaderRepository {
  const _FakeReaderRepository(this.future);

  final Future<ReaderSnapshot> future;

  @override
  Future<ReaderSnapshot> loadInitial() => future;

  @override
  Future<ArticleContent> loadArticleContent(int articleId) {
    return Future.value(_content(articleId));
  }

  @override
  Future<void> markArticleRead({
    required int articleId,
    required bool read,
  }) async {}

  @override
  Future<void> toggleArticleFavorite(int articleId) async {}

  @override
  Future<void> refreshFeed(int feedId) async {}

  @override
  Future<AppSettings> updateSettings(Map<String, String> settings) async {
    return _snapshot().settings;
  }

  @override
  Future<String> exportOpml() async {
    return _opml();
  }

  @override
  Future<void> importOpmlText(String opmlText) async {}

  @override
  Future<TranslationResult> translateText({
    required String text,
    required String targetLanguage,
    bool force = false,
  }) async {
    return _translation(text, targetLanguage);
  }

  @override
  Future<void> clearTranslations() async {}
}

class _CallbackReaderRepository implements ReaderRepository {
  const _CallbackReaderRepository(this.callback);

  final Future<ReaderSnapshot> Function() callback;

  @override
  Future<ReaderSnapshot> loadInitial() => callback();

  @override
  Future<ArticleContent> loadArticleContent(int articleId) {
    return Future.value(_content(articleId));
  }

  @override
  Future<void> markArticleRead({
    required int articleId,
    required bool read,
  }) async {}

  @override
  Future<void> toggleArticleFavorite(int articleId) async {}

  @override
  Future<void> refreshFeed(int feedId) async {}

  @override
  Future<AppSettings> updateSettings(Map<String, String> settings) async {
    return _snapshot().settings;
  }

  @override
  Future<String> exportOpml() async {
    return _opml();
  }

  @override
  Future<void> importOpmlText(String opmlText) async {}

  @override
  Future<TranslationResult> translateText({
    required String text,
    required String targetLanguage,
    bool force = false,
  }) async {
    return _translation(text, targetLanguage);
  }

  @override
  Future<void> clearTranslations() async {}
}

class _RecordingReaderRepository extends _FakeReaderRepository {
  _RecordingReaderRepository(super.future);

  final List<String> markReadCalls = [];
  final List<int> favoriteCalls = [];
  final List<int> refreshFeedCalls = [];
  final List<Map<String, String>> settingsUpdates = [];
  final List<String> translationRequests = [];
  final List<String> opmlImports = [];
  int opmlExportCalls = 0;
  int clearTranslationCalls = 0;

  @override
  Future<void> markArticleRead({
    required int articleId,
    required bool read,
  }) async {
    markReadCalls.add('$articleId:$read');
  }

  @override
  Future<void> toggleArticleFavorite(int articleId) async {
    favoriteCalls.add(articleId);
  }

  @override
  Future<void> refreshFeed(int feedId) async {
    refreshFeedCalls.add(feedId);
  }

  @override
  Future<AppSettings> updateSettings(Map<String, String> settings) async {
    settingsUpdates.add(settings);
    return _snapshot().settings;
  }

  @override
  Future<String> exportOpml() async {
    opmlExportCalls += 1;
    return _opml();
  }

  @override
  Future<void> importOpmlText(String opmlText) async {
    opmlImports.add(opmlText);
  }

  @override
  Future<TranslationResult> translateText({
    required String text,
    required String targetLanguage,
    bool force = false,
  }) async {
    translationRequests.add('$text:$targetLanguage:$force');
    return _translation(text, targetLanguage);
  }

  @override
  Future<void> clearTranslations() async {
    clearTranslationCalls += 1;
  }
}

ReaderSnapshot _snapshot({
  List<Article>? articles,
}) {
  return ReaderSnapshot(
    version: const VersionInfo(version: '1.3.23'),
    settings: AppSettings.fromJson({
      'language': 'en-US',
      'theme': 'auto',
      'update_interval': '30',
      'refresh_mode': 'fixed',
      'default_view_mode': 'rendered',
      'translation_enabled': 'false',
      'translation_only_mode': 'false',
      'target_language': 'zh',
      'translation_provider': 'google',
      'deepl_api_key': '',
      'deepl_endpoint': '',
      'baidu_app_id': '',
      'baidu_secret_key': '',
      'microsoft_api_key': '',
      'microsoft_endpoint': '',
      'microsoft_region': '',
      'tencent_secret_id': '',
      'tencent_secret_key': '',
      'tencent_region': 'ap-guangzhou',
      'ai_translation_profile_id': '',
      'ai_translation_prompt': 'Translate precisely.',
      'custom_translation_enabled': 'false',
      'custom_translation_name': '',
      'custom_translation_endpoint': '',
      'custom_translation_method': 'POST',
      'custom_translation_headers': '',
      'custom_translation_body_template': '',
      'custom_translation_response_path': '',
      'custom_translation_lang_mapping': '',
      'custom_translation_timeout': '10',
    }),
    feeds: [
      Feed.fromJson({
        'id': 1,
        'title': 'Tech Feed',
        'url': 'https://example.com/feed.xml',
        'link': 'https://example.com',
        'description': '',
        'category': 'Tech',
        'image_url': '',
        'position': 1,
        'last_updated': '2026-06-25T09:30:00Z',
        'hide_from_timeline': false,
        'proxy_enabled': false,
        'refresh_interval': 0,
        'is_image_mode': false,
        'email_address': '',
        'tags': [],
      }),
    ],
    articles: articles ??
        [_article(1, 'First Article'), _article(2, 'Second Article')],
  );
}

Article _article(int id, String title) {
  return Article.fromJson({
    'id': id,
    'feed_id': 1,
    'title': title,
    'url': 'https://example.com/$id',
    'image_url': '',
    'audio_url': '',
    'video_url': '',
    'published_at': '2026-06-25T09:45:00Z',
    'is_read': false,
    'is_favorite': id == 1,
    'is_hidden': false,
    'is_read_later': false,
    'feed_title': 'Tech Feed',
    'author': 'Writer',
    'translated_title': '',
    'summary': id == 1 ? 'First summary' : 'Second summary',
    'freshrss_item_id': '',
  });
}

ArticleContent _content(int articleId) {
  return ArticleContent.fromJson({
    'content': articleId == 1
        ? '<p>First Article body</p>'
        : '<p>Second Article body</p>',
    'feed_url': 'https://example.com/feed.xml',
    'cached': true,
  });
}

String _opml() {
  return '<?xml version="1.0"?><opml version="2.0"></opml>';
}

TranslationResult _translation(String text, String targetLanguage) {
  return TranslationResult.fromJson({
    'translated_text': '[${targetLanguage.toUpperCase()}] $text',
    'html': '<p>[${targetLanguage.toUpperCase()}] $text</p>',
    'skipped': 'false',
  });
}

class _RecordingOpmlFileService implements OpmlFileService {
  _RecordingOpmlFileService({required this.pickedText});

  final String pickedText;
  final List<String> exportedTexts = [];

  @override
  Future<String?> exportText(
    String opmlText, {
    String suggestedFileName = 'mrrss-subscriptions.opml',
  }) async {
    exportedTexts.add(opmlText);
    return 'C:\\Users\\reader\\Downloads\\$suggestedFileName';
  }

  @override
  Future<String?> pickText() async {
    return pickedText;
  }
}
