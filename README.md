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

# checkout specifies how a player must finish the leg
#
# - straight-out: player can finish with any shot
# - double-out: player must finish with a double
checkout: double-out

# checkin specifies how a player must start the leg
#
# - straight-in: player can start with any shot
# - double-in: player must start with a double
checkin: straight-in

# players: specifies the players that are playing the game
players:
  - name: Andreas
  - name: Conni
  - name: Gerrit
```