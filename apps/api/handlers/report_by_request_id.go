package handlers

import (
	"errors"
	"log"
	"net/http"

	api_types "github.com/DomNidy/saint_sim/pkg/go-shared/api_types"
	"github.com/DomNidy/saint_sim/pkg/go-shared/db"
	utils "github.com/DomNidy/saint_sim/pkg/go-shared/utils"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func GetReportByRequestID(c *gin.Context, dbClient *db.Queries) {
	requestID := c.Param("requestId")

	var requestUUID pgtype.UUID
	if err := requestUUID.Scan(requestID); err != nil {
		log.Printf("invalid request id %q: %v", requestID, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing or invalid request id"})
		return
	}

	_, err := dbClient.GetSimulationRequestOptions(c, requestUUID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Could not find simulation request with this id"})
			return
		}

		log.Printf("error checking simulation request %q: %v", requestID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	simData, err := dbClient.GetSimulationDataByRequestID(c, requestUUID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.Status(http.StatusAccepted)
			return
		}

		log.Printf("error getting simulation data for request %q: %v", requestID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, api_types.SimulationData{
		Id:        utils.IntPtr(int(simData.ID)),
		RequestId: utils.StrPtr(requestID),
		SimResult: utils.StrPtr(simData.SimResult),
	})
}
