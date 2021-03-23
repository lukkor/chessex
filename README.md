# chessex

## Why?

A personal project to learn more about part of the future stack I'll work with,
namely: [Go](https://golang.org/), [ScyllaDB](https://www.scylladb.com/) and
[Gin web framework](https://github.com/gin-gonic/gin).

## What is it?

Not 100% clear on the full outcome yet, but I'd like to make available
aggregated and explorable data about the 1 million chess games I found online.

## How?

 — [x] Find an open chess database of, at least, 1 million games.
 — [x] Setup Go project and the ScyllaDB locally.
 — [ ] Load the million games into ScyllaDB.
 — [ ] Create the API endpoint for statistics on openings.

## The lexical analysis rabbit hole

I went down the lexical analysis rabbit hole at the third step after trying to
extract as much information as possible from the Portable Game Notation
([PGN](https://en.wikipedia.org/wiki/Portable_Game_Notation)). It is possible
with a regurlar expression but after looking into it, I wanted to work on
parsing the notation with lexical analysis (refreshing old memories).
