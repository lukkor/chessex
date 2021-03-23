CREATE TABLE IF NOT EXISTS games (played_on DATE PRIMARY KEY, movetext VARCHAR)
WITH COMPACTION = {'class': 'TimeWindowCompactionStrategy', 'max_sstable_age_days': 1};
