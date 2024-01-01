# darts-counter

Counts remaining values for a game of darts and shows possible finishes.

## Reference Configuration

```yaml
# game specifies the game type being played
#
# - 301: Game301
# - 501: Game501
# - 701: Game701
# - 1001: Game1001
game: "301"

# checkout specifies the checkout variant being played
#
# - single-out: player can finish with any score
# - double-out: player must finish with a double score
checkout: double-out

# players: specifies the players that are playing the game
players:
  - name: Andreas
  - name: Conni
  - name: Gerrit
```