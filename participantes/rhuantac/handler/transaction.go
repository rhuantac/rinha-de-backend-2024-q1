package handler

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/rhuantac/rinha-concurrency/internal/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type TransactionType string

const (
	Credit TransactionType = "c"
	Debit  TransactionType = "d"
)

type TransactionRequest struct {
	Value           int             `json:"valor"`
	TransactionType TransactionType `json:"tipo"`
	Description     string          `json:"description"`
}

type TransactionResponse struct {
	Limit   int `json:"limite"`
	Balance int `json:"saldo"`
}

func TransactionHandler(db *mongo.Database) gin.HandlerFunc {

	return func(c *gin.Context) {
		userId, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		var req TransactionRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		if req.TransactionType == Debit {
			req.Value *= -1
		}

		coll := db.Collection("users")
		filter := bson.D{{Key: "_id", Value: userId}}
		update := bson.D{{Key: "$inc", Value: bson.D{{Key: "current_balance", Value: req.Value}}}}

		var user model.User
		result := coll.FindOne(c, filter)
		result.Decode(&user)

		if result.Err() != nil {
			c.Status(http.StatusNotFound)
			return
		}

		newBalance := user.CurrentBalance + req.Value
		//Balance cannot be lower than limit value
		if newBalance < user.Limit * -1 {
			c.Status(http.StatusUnprocessableEntity)
			return
		}

		_, err = coll.UpdateOne(c, filter, update)

		if err != nil {
			c.Status(http.StatusBadRequest)
			log.Printf("Error updating user %d", userId)
		}

		c.JSON(http.StatusOK, TransactionResponse{Limit: user.Limit, Balance: newBalance})
	}
}