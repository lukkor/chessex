CREATE TABLE IF NOT EXISTS games_by_opening (id blob, opening text, outcome text, tags map<text, text>, raw text, PRIMARY KEY ((opening, outcome), id))
WITH COMPACTION = {'class': 'SizeTieredCompactionStrategy'};
