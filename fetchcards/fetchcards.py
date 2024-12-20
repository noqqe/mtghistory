#!/usr/bin/env python3

import sys
import requests
import json
from loguru import logger

url = 'https://api.scryfall.com/cards/search?order=released&dir=desc&q=(st:masters%20or%20st:core%20or%20st:expansion)%20-set:plst%20lang:en%20unique:prints%20game:paper'
filepath = '../assets/allcards.json'
all_cards = []

# configure logger
logger.remove()
logger.add(
    sys.stdout,
    format="{time:YYYY-MM-DD HH:mm:ss} | <lvl>{level}</lvl> | {message}",
    level="DEBUG",
    filter=lambda record: True,  # Include all records
)

# check if we have permission to write to the file
try:
  open(filepath, 'w')
except:
  logger.error(f"Permission denied or directory not existing to {filepath}")
  sys.exit(1)
  
# Fetch all cards from scryfall
response = requests.get(url).json()
cards = response["data"]

c = 1
for card in cards:
  year = card['released_at'].split("-")[0]
  logger.info(f"Processing {card['set']}/{card['collector_number']} from {year}")

  if not any(d['year'] == year for d in all_cards):
    logger.warning(f"Added year {year} entry added to all_cards")
    all_cards.append({"year": year, "cards": [] })

  year_index = next((index for (index, d) in enumerate(all_cards) if d["year"] == year), None)
  if year_index is not None:
    logger.info(f"Adding {card['set']}/{card['collector_number']} to all_cards")
    all_cards[year_index]["cards"].append(card['set']+"/"+card['collector_number'])

# pagination
while response["has_more"]:
  c = c + 1
  logger.info(f"Fetching page {c}")
  response = requests.get(response["next_page"]).json()
  cards = response["data"]

  for card in  cards:
    year = card['released_at'].split("-")[0]
    logger.info(f"Processing {card['set']}/{card['collector_number']} from {year}")

    if not any(d['year'] == year for d in all_cards):
      logger.warning(f"Added year {year} entry added to all_cards")
      all_cards.append({"year": year, "cards": [] })

    year_index = next((index for (index, d) in enumerate(all_cards) if d["year"] == year), None)
    if year_index is not None:
      logger.info(f"Adding {card['set']}/{card['collector_number']} to all_cards")
      all_cards[year_index]["cards"].append(card['set']+"/"+card['collector_number'])

# save all_cards to disk as json
with open('../assets/allcards.json', 'w') as fp:
    json.dump(all_cards, fp)
