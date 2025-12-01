#!/usr/bin/env python3

import os
import sys
import json
from loguru import logger

out_file = '../assets/allcards.json'
oracle_path = os.path.expanduser('~/Downloads/oracle-cards-20251130220622.json')
year = '1993'

# configure logger
logger.remove()
logger.add(
    sys.stdout,
    format="{time:YYYY-MM-DD HH:mm:ss} | <lvl>{level}</lvl> | {message}",
    level="DEBUG",
    filter=lambda record: True,  # Include all records
)

# check if we have permission to write to the file
def verify_write_permission(path):
  try:
    open(out_file, 'w')
  except:
    logger.error(f"Permission denied or directory not existing to {out_file}")
    sys.exit(1)

def convert_data(oracle_path):
  all_cards = []
  with open(oracle_path, 'r') as fp:
    bulk_data = json.load(fp)

  for card in bulk_data:
    if not card['set_type'] in ['core', 'expansion', 'masters']:
      continue

    year = card['released_at'].split("-")[0]
    if not any(d['year'] == year for d in all_cards):
      logger.warning(f"Added year {year} entry added to all_cards")
      all_cards.append({"year": year, "cards": [] })

    year_index = next((index for (index, d) in enumerate(all_cards) if d["year"] == year), None)
    if year_index is not None:
      logger.info(f"Adding {card['name']} ({card['set']}/{card['collector_number']}) type {card['set_type']}")
      all_cards[year_index]["cards"].append(card['set']+"/"+card['collector_number'])

    return all_cards


# save all_cards to disk as json
def persist_data(path, data):
  if len(data) <= 0:
    logger.error(f"No data to be persisted")
    sys.exit(1)
         
  with open(out_file, 'w') as fp:
      json.dump(data, fp)

if __name__ == "__main__":
  verify_write_permission(out_file)
  cards = convert_data(oracle_path)
  persist_data(out_file, cards)

