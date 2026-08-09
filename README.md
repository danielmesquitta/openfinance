# Open Finance Integration Project

## Overview

This project provides a solution to fetch card and account information from Brazil's Open Finance and automatically populate a Notion database using the Notion API. It's designed to help users effortlessly manage their financial data within a Notion workspace.

## Dependencies

- [Go](https://go.dev/)
- [Make](https://sp21.datastructur.es/materials/guides/make-install.html)

## Getting started

1. Clone the repo

```bash
git clone https://github.com/danielmesquitta/openfinance

```

2. Install the required packages:

```bash
make install
```

3. Create your .env file

```bash
cp .env.example .env
```

4. Create the sync profiles config file

```bash
cp config/sync_profiles.json.example config/sync_profiles.json
```

5. Configure your `.env` and `config/sync_profiles.json` files with your credentials and per-sync-profile categorization settings.

Each sync profile must define a non-empty `categories` object and a `mappings` object, which may be empty. The optional `fallback` defaults to `Others`. If the fallback is absent from `categories`, it is added automatically with Notion's default color.

6. Execute the script

```bash
make
```
