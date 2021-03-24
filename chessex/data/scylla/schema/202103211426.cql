CREATE TABLE IF NOT EXISTS games (opening text, outcome text, tags map<text, text>, raw text, PRIMARY KEY (opening, outcome))
WITH COMPACTION = {'class': 'SizeTieredCompactionStrategy'};
