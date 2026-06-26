import 'package:flutter/material.dart';

import '../../models/article.dart';

class ArticleListPanel extends StatelessWidget {
  const ArticleListPanel({
    required this.articles,
    required this.selectedArticleId,
    required this.onSelected,
    super.key,
  });

  final List<Article> articles;
  final int selectedArticleId;
  final ValueChanged<Article> onSelected;

  @override
  Widget build(BuildContext context) {
    return ListView.separated(
      itemCount: articles.length,
      separatorBuilder: (context, index) => const Divider(height: 1),
      itemBuilder: (context, index) {
        final article = articles[index];
        final selected = article.id == selectedArticleId;
        return ListTile(
          selected: selected,
          title: Text(
            article.translatedTitle.isEmpty
                ? article.title
                : article.translatedTitle,
            maxLines: 2,
            overflow: TextOverflow.ellipsis,
          ),
          subtitle: Text(
            article.feedTitle,
            maxLines: 1,
            overflow: TextOverflow.ellipsis,
          ),
          trailing: article.isFavorite ? const Icon(Icons.star) : null,
          onTap: () => onSelected(article),
        );
      },
    );
  }
}
