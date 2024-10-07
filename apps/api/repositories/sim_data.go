package repositories

import (
	"context"

	"github.com/DomNidy/saint_sim/pkg/interfaces"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SimDataRepository interface {
	GetSimData(id int) (*interfaces.SimDataGet, error)
}

type SimDataRepositoryImpl struct {
	db *pgxpool.Pool
}

func NewSimDataRepository(db *pgxpool.Pool) SimDataRepository {
	return &SimDataRepositoryImpl{db: db}
}

func (repo *SimDataRepositoryImpl) GetSimData(id int) (*interfaces.SimDataGet, error) {
	simData := &interfaces.SimDataGet{}

	err := repo.db.QueryRow(context.Background(),
		"SELECT id, from_request, sim_result FROM simulation_data WHERE id=$1", id).Scan(
		&simData.ID,
		&simData.FromRequest,
		&simData.SimResult,
	)

	if err != nil {
		return nil, err
	}

	return simData, nil
}
