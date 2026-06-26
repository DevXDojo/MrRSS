import 'package:flutter/material.dart';

import '../../models/article.dart';
import '../../models/article_content.dart';
import '../../models/translation_result.dart';
import '../../utils/html_text.dart';

class ArticleDetailPanel extends StatefulWidget {
  const ArticleDetailPanel({
    required this.article,
    required this.loadContent,
    required this.onMarkRead,
    required this.onToggleFavorite,
    required this.onTranslateText,
    required this.targetLanguage,
    super.key,
  });

  final Article article;
  final Future<ArticleContent> Function(int articleId) loadContent;
  final Future<void> Function({required int articleId, required bool read})
      onMarkRead;
  final Future<void> Function(int articleId) onToggleFavorite;
  final Future<TranslationResult> Function({
    required String text,
    required String targetLanguage,
    bool force,
  }) onTranslateText;
  final String targetLanguage;

  @override
  State<ArticleDetailPanel> createState() => _ArticleDetailPanelState();
}

class _ArticleDetailPanelState extends State<ArticleDetailPanel> {
  late Future<ArticleContent> _contentFuture;
  String? _translatedSummary;
  String? _translatedContent;
  bool _translatingSummary = false;
  bool _translatingContent = false;
  String? _translationError;

  @override
  void initState() {
    super.initState();
    _contentFuture = widget.loadContent(widget.article.id);
  }

  @override
  void didUpdateWidget(covariant ArticleDetailPanel oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (oldWidget.article.id != widget.article.id) {
      _contentFuture = widget.loadContent(widget.article.id);
      _translatedSummary = null;
      _translatedContent = null;
      _translationError = null;
      _translatingSummary = false;
      _translatingContent = false;
    }
  }

  @override
  Widget build(BuildContext context) {
    final article = widget.article;
    final title = article.translatedTitle.isEmpty
        ? article.title
        : article.translatedTitle;
    return SingleChildScrollView(
      padding: const EdgeInsets.all(24),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            title,
            style: Theme.of(context).textTheme.headlineSmall,
          ),
          const SizedBox(height: 8),
          Text(
            article.feedTitle,
            style: Theme.of(context).textTheme.labelLarge,
          ),
          if (article.author.isNotEmpty) ...[
            const SizedBox(height: 4),
            Text(article.author),
          ],
          const SizedBox(height: 16),
          Wrap(
            spacing: 8,
            runSpacing: 8,
            children: [
              OutlinedButton.icon(
                onPressed: () => widget.onMarkRead(
                  articleId: article.id,
                  read: !article.isRead,
                ),
                icon:
                    Icon(article.isRead ? Icons.mark_email_unread : Icons.done),
                label: Text(article.isRead ? 'Mark unread' : 'Mark read'),
              ),
              OutlinedButton.icon(
                onPressed: () => widget.onToggleFavorite(article.id),
                icon: Icon(article.isFavorite ? Icons.star : Icons.star_border),
                label: Text(article.isFavorite ? 'Unfavorite' : 'Favorite'),
              ),
              OutlinedButton.icon(
                onPressed: _translatingSummary ? null : _translateSummary,
                icon: const Icon(Icons.translate),
                label: Text(
                  _translatingSummary ? 'Translating...' : 'Translate summary',
                ),
              ),
              OutlinedButton.icon(
                onPressed: _translatingContent ? null : _translateContent,
                icon: const Icon(Icons.article),
                label: Text(
                  _translatingContent
                      ? 'Translating content...'
                      : 'Translate content',
                ),
              ),
            ],
          ),
          const SizedBox(height: 24),
          if (article.summary.isNotEmpty) ...[
            Text(
              article.summary,
              style: Theme.of(context).textTheme.bodyLarge,
            ),
            if (_translatedSummary != null) ...[
              const SizedBox(height: 12),
              Text(
                _translatedSummary!,
                style: Theme.of(context).textTheme.bodyLarge,
              ),
            ],
            if (_translationError != null) ...[
              const SizedBox(height: 12),
              Text(
                _translationError!,
                style: TextStyle(color: Theme.of(context).colorScheme.error),
              ),
            ],
            const SizedBox(height: 24),
          ],
          FutureBuilder<ArticleContent>(
            key: ValueKey(article.id),
            future: _contentFuture,
            builder: (context, snapshot) {
              if (snapshot.connectionState != ConnectionState.done) {
                return const Center(child: CircularProgressIndicator());
              }
              if (snapshot.hasError) {
                return Text('Failed to load article content');
              }

              final content = htmlToReadableText(snapshot.requireData.content);
              if (content.isEmpty) {
                return const Text('No article content available.');
              }

              return Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  SelectableText(content),
                  if (_translatedContent != null) ...[
                    const SizedBox(height: 16),
                    SelectableText(_translatedContent!),
                  ],
                ],
              );
            },
          ),
        ],
      ),
    );
  }

  Future<void> _translateSummary() async {
    final text = widget.article.summary.isEmpty
        ? widget.article.title
        : widget.article.summary;
    setState(() {
      _translatingSummary = true;
      _translationError = null;
    });

    try {
      final result = await widget.onTranslateText(
        text: text,
        targetLanguage: widget.targetLanguage,
        force: true,
      );
      if (mounted) {
        setState(() {
          _translatedSummary = result.translatedText;
          _translatingSummary = false;
        });
      }
    } catch (error) {
      if (mounted) {
        setState(() {
          _translationError = error.toString();
          _translatingSummary = false;
        });
      }
    }
  }

  Future<void> _translateContent() async {
    setState(() {
      _translatingContent = true;
      _translationError = null;
    });

    try {
      final articleContent = await _contentFuture;
      final text = htmlToReadableText(articleContent.content);
      if (text.isEmpty) {
        throw StateError('No article content available to translate.');
      }

      final result = await widget.onTranslateText(
        text: text,
        targetLanguage: widget.targetLanguage,
        force: true,
      );
      if (mounted) {
        setState(() {
          _translatedContent = result.translatedText;
          _translatingContent = false;
        });
      }
    } catch (error) {
      if (mounted) {
        setState(() {
          _translationError = error.toString();
          _translatingContent = false;
        });
      }
    }
  }
}
