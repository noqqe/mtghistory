#!/usr/bin/env python3

import sys
import requests
import json
from loguru import logger

logger.remove()
logger.add(
    sys.stdout,
    format="{time:YYYY-MM-DD HH:mm:ss} | <lvl>{level}</lvl> | {message}",
    level="DEBUG",
    filter=lambda record: True,  # Include all records
)

all_cards = {} 

url = 'https://api.scryfall.com/cards/search?order=released&dir=desc&q=(st:masters%20or%20st:core%20or%20st:expansion)%20-set:plst%20lang:en%20unique:prints%20game:paper'
response = requests.get(url).json()
cards = response["data"]

c = 1
for card in  cards:
  year = card['released_at'].split("-")[0]
  logger.info(f"Processing {card['set']}/{card['collector_number']} from {year}")
  if all_cards.get(year) == None:
    logger.warning(f"Added year {year} to dict")
    all_cards[year] = []
  all_cards[year].append(card['set']+"/"+card['collector_number'])

# fucking pagination
while response["has_more"]:
  c = c + 1
  logger.info(f"Fetching page {c}")
  response = requests.get(response["next_page"]).json()
  cards = response["data"]

  for card in  cards:
    year = card['released_at'].split("-")[0]
    logger.info(f"Processing {card['set']}/{card['collector_number']} from {year}")
    if all_cards.get(year) == None:
      logger.warning(f"Added year {year} to dict")
      all_cards[year] = []
    all_cards[year].append(card['set']+"/"+card['collector_number'])


# save all_cards to disk as json
with open('assets/allcards.json', 'w') as fp:
    json.dump(all_cards, fp)
