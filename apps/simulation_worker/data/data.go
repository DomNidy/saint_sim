package data

import (
	"context"

	"github.com/DomNidy/saint_sim/pkg/interfaces"
	"github.com/jackc/pgx/v5/pgxpool"
)

func InsertSimulationData(db *pgxpool.Pool, data *interfaces.SimDataInsert) error {
	_, err := db.Exec(context.Background(), "INSERT INTO simulation_data (request_id, sim_result) VALUES ($1, $2)", data.RequestID, data.SimResult)
	if err != nil {
		return err
	}
	return nil
}
