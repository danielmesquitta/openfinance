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

4. Create your .env file

```bash
cp .env.example .env
```

5. Create config files

```bash
cp config/categories.json.example config/categories.json
cp config/mappings.json.example config/mappings.json
cp config/users.json.example config/users.json
```

5. Configure your .env file and the config files with your credentials

6. Execute the script

```bash
make
```
