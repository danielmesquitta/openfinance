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

4. Create the ingest profiles config file

```bash
cp config/ingest_profiles.json.example config/ingest_profiles.json
```

5. Configure your `.env` and `config/ingest_profiles.json` files with your credentials and per-ingest-profile categorization settings.

Each ingest profile must define a non-empty `categories` object and a `mappings` object, which may be empty. The optional `fallback` defaults to `Others`. If the fallback is absent from `categories`, it is added automatically with Notion's default color.

Each profile may also set `language` to `en` or `pt-BR`. When omitted, it defaults to `en` for backward compatibility. The language controls monthly Notion table titles, transaction column names, and generated payment-method options. Category names, mappings, fallback values, transaction data, dates, and currency are not translated.

Changing a profile's language does not rename or migrate existing Notion tables. Existing monthly tables in either supported language are reused with their original column language, while missing tables are created in the profile's currently configured language.

6. Execute the script

```bash
make
```
