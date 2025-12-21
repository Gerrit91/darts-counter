package stats

type noopImpl struct{}

func (*noopImpl) ListGameStats(filterOpts ...filter) ([]*GameStats, error) {
	return nil, nil
}

func (*noopImpl) CreateGameStats(_ *GameStats) error {
	return nil
}

func (*noopImpl) DeleteGameStats(id string) error {
	return nil
}

func (*noopImpl) Close() {
}

func (*noopImpl) Enabled() bool {
	return false
}
