# Live mode with custom config
docker run your-indexer live --config /path/to/config.yml --skip-db-check

# Historic mode
docker run your-indexer historic --from-height 1000 --to-height 2000

# Show all available commands
docker run your-indexer --help

# Show help for specific command
docker run your-indexer live --help