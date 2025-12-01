#!/usr/bin/env python3

import sys
import json
import requests
from loguru import logger
import tempfile

out_file = '../assets/allcards.json'
bulk_url = "https://api.scryfall.com/bulk-data"
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

def find_oracle_file(bulk_url):
  r = requests.get(bulk_url)
  for e in r.json()['data']:
    if e['type'] == 'oracle_cards':
      logger.info(f"Found oracle file at {e['download_uri']}")
      return e['download_uri']

def get_oracle_file(url):
  """Download a large file and save it to a temporary file."""
  with tempfile.NamedTemporaryFile(delete=False) as temp_file:
    response = requests.get(url, stream=True)

    if response.status_code == 200:
      for chunk in response.iter_content(chunk_size=8192):
        temp_file.write(chunk)
      logger.info(f"File downloaded and saved to: {temp_file.name}")
    else:
      logger.error(f"Failed to download file. Status code: {response.status_code}")

  return temp_file.name


def convert_data(oracle_path):
  all_cards = []
  with open(oracle_path, 'r') as fp:
    bulk_data = json.load(fp)

  for card in bulk_data:
    if not card['set_type'] in ['core', 'expansion', 'masters']:
      continue

    if any(card['games']) == 'paper':
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
  file_url = find_oracle_file(bulk_url)
  oracle_path = get_oracle_file(file_url)
  cards = convert_data(oracle_path)
  persist_data(out_file, cards)

