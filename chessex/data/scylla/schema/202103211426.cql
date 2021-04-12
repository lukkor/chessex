CREATE TABLE IF NOT EXISTS games_by_moves (moves text, outcome text, count counter, PRIMARY KEY (moves, outcome))
WITH COMPACTION = {'class': 'SizeTieredCompactionStrategy'};
