# darts-counter

Counts remaining values for a game of darts and shows possible finishes.

## Reference Configuration

```yaml
# database: stores game settings and statistics to disk
database:
  # the path to the database file
  path: darts-counter.db

# logging: writes an application log file to the specified path, overwritten on app restart
logging:
  # enables logging
  enabled: true
  # the log level
  level: info
  # the path to the log file
  path: darts-counter.log
```
