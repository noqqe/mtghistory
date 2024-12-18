# mtghistory

Is a small web application written in Golang where you can visualize your Magic: The Gathering collection.

# Data formats

* Archidekt
* Deckbox
* ManaBox
* Moxfield

# Manual

If you have your cards in a different (or your own) format, you can still upload your collection.

Just Pick "ManaBox" and create a csv file where **setcode** is column 4 and **collectornumber** is column 6.
Everything else can be empty.

Example:

```
,,,usg,,321
,,,blb,,21
```
