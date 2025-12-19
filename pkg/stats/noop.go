package stats

type noopImpl struct{}

func (b *noopImpl) GetPlayerStats(filterOpts ...filter) ([]*PlayerStats, error) {
	return nil, nil
}

func (*noopImpl) GetGameStats(filterOpts ...filter) ([]*GameStats, error) {
	return nil, nil
}

func (*noopImpl) CreateGameStats(_ *GameStats) error {
	return nil
}

func (*noopImpl) DeleteGameStats(id string) error {
	return nil
}
