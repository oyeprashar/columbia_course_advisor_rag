#!/bin/bash

# Exit immediately if a command exits with a non-zero status
set -e

echo "Starting parsing sequence..."

python3 scrapper/fetch.py

python3 parser/parse_breadth_requirements.py
python3 parser/parse_bulletin.py
python3 parser/parse_core_ms_cs_requirements.py
python3 parser/parse_faq.py
python3 parser/parse_pathways.py

echo "All parsers completed successfully!"