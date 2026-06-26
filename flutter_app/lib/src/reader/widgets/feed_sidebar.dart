import 'package:flutter/material.dart';

import '../../models/feed.dart';

class FeedSidebar extends StatelessWidget {
  const FeedSidebar({
    required this.feeds,
    required this.onRefreshFeed,
    super.key,
  });

  final List<Feed> feeds;
  final Future<void> Function(int feedId) onRefreshFeed;

  @override
  Widget build(BuildContext context) {
    return ListView(
      padding: const EdgeInsets.symmetric(vertical: 8),
      children: [
        const ListTile(
          leading: Icon(Icons.all_inbox_outlined),
          title: Text('All articles'),
        ),
        const Divider(),
        for (final feed in feeds)
          ListTile(
            dense: true,
            leading: const Icon(Icons.rss_feed),
            title: Text(
              feed.title,
              maxLines: 1,
              overflow: TextOverflow.ellipsis,
            ),
            subtitle: feed.category.isEmpty
                ? null
                : Text(
                    feed.category,
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                  ),
            trailing: IconButton(
              tooltip: 'Refresh feed',
              icon: const Icon(Icons.refresh),
              onPressed: () => onRefreshFeed(feed.id),
            ),
          ),
      ],
    );
  }
}
