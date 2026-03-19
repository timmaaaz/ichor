# Manitowoc Customer Seed

Seed data pipeline for Manitowoc, a relay manufacturer. Populates a fresh Ichor database with platform baseline data, real customer configuration, and placeholder historical depth.

## Prerequisites

- Python 3.9+
- `psql` CLI (PostgreSQL client)
- A fresh database with the Ichor schema and all migrations applied

## Setup

```bash
pip install -r requirements.txt
```

## Running

```bash
./seed.sh postgresql://user:pass@host:5432/manitowoc_db
```

## Resuming After Failure

Each SQL file is numbered. If the run fails partway through, pass the file number to resume from:

```bash
./seed.sh postgresql://user:pass@host:5432/manitowoc_db 22
```

Files with a number less than `start_from` are skipped.

## File Structure

```
manitowoc/
├── data/           # YAML config files consumed by generate.py
├── seed/           # Generated SQL files (output of generate.py)
├── generate.py     # Reads data/ YAML configs, writes SQL files to seed/
├── seed.sh         # Entry point: runs generate.py then applies SQL in order
├── requirements.txt
└── README.md
```

## Three-Tier Seed Model

| Range | Tier | Description |
|-------|------|-------------|
| 00–13 | Platform baseline | Users, roles, permissions, locations, asset types — same across all customers |
| 20–21 | Customer real data | Manitowoc-specific products, suppliers, warehouses, and configuration |
| 22–27 | Placeholder historical depth | Synthetic inventory transactions, receipts, and movements for a realistic starting state |
