package controllers

import (
	. "backend/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"net/http"
)

type QueryController struct {
	DB *sqlx.DB

	queryChannel chan QueryMap
}

func NewQueryController(db *sqlx.DB, queryChannel chan QueryMap) *QueryController {

	return &QueryController{
		DB: db,

		queryChannel: queryChannel,
	}
}

func (controller *QueryController) FetchQuery(context *gin.Context) {

	var request QueryRequest

	if err := context.ShouldBindJSON(&request); err != nil {

		context.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})

		return
	}

	queryMap := QueryMap{

		RequestID: uint64(uuid.New().ID()),

		QueryRequest: request,

		Response: make(chan Response, 1),
	}

	controller.queryChannel <- queryMap

	response := <-queryMap.Response

	context.JSON(http.StatusOK, response)
}
