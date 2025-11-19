# Database Schema

## Sections Table
- `id`: TEXT PK
- `title`: TEXT
- `part`: TEXT
- `section`: TEXT
- `agency_id`: TEXT
- `path`: TEXT
- `text`: TEXT
- `rev_date`: DATETIME
- `checksum_sha256`: TEXT
- `word_count`: INTEGER
- `def_count`: INTEGER
- `xref_count`: INTEGER
- `modal_count`: INTEGER
- `rscs_raw`: INTEGER
- `rscs_per_1k`: REAL
- `snapshot_date`: TEXT

## Summaries
- `kind`: TEXT
- `key`: TEXT
- `text`: TEXT
- `model`: TEXT
- `created_at`: DATETIME

## LSA Activity
- `key_kind`: TEXT
- `key`: TEXT
- `since_rev_date`: DATETIME
- `proposals_count`: INTEGER
- `amendments_count`: INTEGER
- `finals_count`: INTEGER
- `captured_at`: DATETIME
- `source_hint`: TEXT

Relationships: Sections FK agency_id to Agencies, etc.
Constraints: PKs, not nulls as per domain.
