# API Documentation

- `GET /agencies`: List agencies with totals (word_count, rscs_*, lsa, last_updated)
  Params: `sort=rscs_per_1k&dir=desc&limit=10`

- `GET /agencies/{id}`: Overview, top titles by RSCS

- `GET /titles`: List all titles
- `GET /titles/{t}`: Title details, metrics, LSA counts

- `GET /sections/{id}`: Section details, text excerpt, summary

- `GET /snapshots/diff`: Compare snapshots
