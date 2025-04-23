package controllers

import (
	. "backend/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"net/http"
	"reflect"
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

func (qc *QueryController) FetchQuery(c *gin.Context) {

	var request QueryRequest

	if err := c.ShouldBindJSON(&request); err != nil {

		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})

		return
	}

	queryMap := QueryMap{

		RequestID: uint64(uuid.New().ID()),

		QueryRequest: request,

		Response: make(chan Response, 1),
	}

	qc.queryChannel <- queryMap

	response := <-queryMap.Response

	//parsed := parseMap(reflect.ValueOf(response.Data))

	c.JSON(http.StatusOK, response)
}

func parseMap(dataValue reflect.Value) any {

	if dataValue.Kind() == reflect.Interface {

		dataValue = dataValue.Elem()
	}

	switch dataValue.Kind() {

	case reflect.Map:

		result := make(map[string]any)

		iter := dataValue.MapRange()

		for iter.Next() {

			key := iter.Key().Interface().(string)

			value := iter.Value()

			result[key] = parseMap(value)
		}

		return result

	case reflect.Slice:

		var result []any

		for i := 0; i < dataValue.Len(); i++ {

			result = append(result, parseMap(dataValue.Index(i)))
		}

		return result

	default:

		return dataValue.Interface()
	}
}
