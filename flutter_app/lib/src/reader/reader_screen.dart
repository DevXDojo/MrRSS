import 'package:flutter/material.dart';

import '../models/article.dart';
import '../platform/opml_file_service.dart';
import 'reader_repository.dart';
import 'widgets/article_detail_panel.dart';
import 'widgets/article_list_panel.dart';
import 'widgets/feed_sidebar.dart';
import 'widgets/settings_sheet.dart';

class ReaderScreen extends StatefulWidget {
  const ReaderScreen({
    required this.repository,
    required this.opmlFileService,
    super.key,
  });

  final ReaderRepository repository;
  final OpmlFileService opmlFileService;

  @override
  State<ReaderScreen> createState() => _ReaderScreenState();
}

class _ReaderScreenState extends State<ReaderScreen> {
  late Future<ReaderSnapshot> _snapshotFuture;
  int? _selectedArticleId;
  bool _showCompactDetail = false;

  @override
  void initState() {
    super.initState();
    _snapshotFuture = widget.repository.loadInitial();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('MrRSS'),
        actions: [
          IconButton(
            tooltip: 'Settings',
            icon: const Icon(Icons.settings),
            onPressed: () => _openSettings(context),
          ),
        ],
      ),
      body: FutureBuilder<ReaderSnapshot>(
        future: _snapshotFuture,
        builder: (context, snapshot) {
          if (snapshot.connectionState != ConnectionState.done) {
            return const Center(child: CircularProgressIndicator());
          }

          if (snapshot.hasError) {
            return _ReaderError(
              message: snapshot.error.toString(),
              onRetry: _reload,
            );
          }

          final data = snapshot.requireData;
          if (data.articles.isEmpty) {
            return const Center(child: Text('No articles'));
          }

          final selectedArticle = _selectedArticle(data.articles);
          return LayoutBuilder(
            builder: (context, constraints) {
              if (constraints.maxWidth >= 900) {
                return Row(
                  children: [
                    SizedBox(
                      width: 260,
                      child: FeedSidebar(
                        feeds: data.feeds,
                        onRefreshFeed: widget.repository.refreshFeed,
                      ),
                    ),
                    const VerticalDivider(width: 1),
                    SizedBox(
                      width: 360,
                      child: ArticleListPanel(
                        articles: data.articles,
                        selectedArticleId: selectedArticle.id,
                        onSelected: _selectArticle,
                      ),
                    ),
                    const VerticalDivider(width: 1),
                    Expanded(
                      child: ArticleDetailPanel(
                        article: selectedArticle,
                        loadContent: widget.repository.loadArticleContent,
                        onMarkRead: widget.repository.markArticleRead,
                        onToggleFavorite:
                            widget.repository.toggleArticleFavorite,
                        onTranslateText: widget.repository.translateText,
                        targetLanguage: data.settings.targetLanguage,
                      ),
                    ),
                  ],
                );
              }

              if (_showCompactDetail) {
                return Column(
                  children: [
                    Align(
                      alignment: Alignment.centerLeft,
                      child: TextButton.icon(
                        onPressed: _showCompactList,
                        icon: const Icon(Icons.arrow_back),
                        label: const Text('Back to articles'),
                      ),
                    ),
                    const Divider(height: 1),
                    Expanded(
                      child: ArticleDetailPanel(
                        article: selectedArticle,
                        loadContent: widget.repository.loadArticleContent,
                        onMarkRead: widget.repository.markArticleRead,
                        onToggleFavorite:
                            widget.repository.toggleArticleFavorite,
                        onTranslateText: widget.repository.translateText,
                        targetLanguage: data.settings.targetLanguage,
                      ),
                    ),
                  ],
                );
              }

              return ArticleListPanel(
                articles: data.articles,
                selectedArticleId: selectedArticle.id,
                onSelected: _selectArticleAndShowCompactDetail,
              );
            },
          );
        },
      ),
    );
  }

  Article _selectedArticle(List<Article> articles) {
    final selectedId = _selectedArticleId;
    if (selectedId != null) {
      for (final article in articles) {
        if (article.id == selectedId) {
          return article;
        }
      }
    }
    return articles.first;
  }

  void _selectArticle(Article article) {
    setState(() {
      _selectedArticleId = article.id;
    });
  }

  void _selectArticleAndShowCompactDetail(Article article) {
    setState(() {
      _selectedArticleId = article.id;
      _showCompactDetail = true;
    });
  }

  void _showCompactList() {
    setState(() {
      _showCompactDetail = false;
    });
  }

  void _reload() {
    setState(() {
      _selectedArticleId = null;
      _showCompactDetail = false;
      _snapshotFuture = widget.repository.loadInitial();
    });
  }

  Future<void> _openSettings(BuildContext context) async {
    final snapshot = await _snapshotFuture;
    if (!context.mounted) {
      return;
    }

    await showModalBottomSheet<void>(
      context: context,
      isScrollControlled: true,
      builder: (context) {
        return SettingsSheet(
          settings: snapshot.settings,
          opmlFileService: widget.opmlFileService,
          onExportOpml: widget.repository.exportOpml,
          onImportOpmlText: (opmlText) async {
            await widget.repository.importOpmlText(opmlText);
            _reload();
          },
          onClearTranslations: () async {
            await widget.repository.clearTranslations();
            _reload();
          },
          onSave: (settings) async {
            await widget.repository.updateSettings(settings);
            _reload();
          },
        );
      },
    );
  }
}

class _ReaderError extends StatelessWidget {
  const _ReaderError({
    required this.message,
    required this.onRetry,
  });

  final String message;
  final VoidCallback onRetry;

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(24),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Text(
              'Failed to load reader data',
              style: Theme.of(context).textTheme.titleMedium,
            ),
            const SizedBox(height: 8),
            Text(
              message,
              textAlign: TextAlign.center,
            ),
            const SizedBox(height: 16),
            FilledButton(
              onPressed: onRetry,
              child: const Text('Retry'),
            ),
          ],
        ),
      ),
    );
  }
}
