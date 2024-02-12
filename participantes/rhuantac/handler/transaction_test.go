package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/rhuantac/rinha-concurrency/config"
	"github.com/rhuantac/rinha-concurrency/internal"
)

type Account struct {
}

func setupServer() *gin.Engine{
	mongoClient := config.SetupMongo()
	internal.ClearDb(mongoClient)
	internal.SeedDb(mongoClient)
	
	router := gin.Default()
	api := router.Group("/clientes/:id")

	api.POST("/transacoes", TransactionHandler(mongoClient.Database(os.Getenv("MONGO_DATABASE"))))

	return router
}

func TestTransactions(t *testing.T) {
	godotenv.Load(filepath.Join("../", ".env"))
	t.Run("credits increase amount", func(t *testing.T) { 
		router := setupServer()
		accId := 1
		endpoint := fmt.Sprintf("/clientes/%d/transacoes", accId)
		testPayload, _ := json.Marshal(TransactionRequest{
			Value: 500,
			TransactionType: "c",
			Description: "Test transaction",
		})
		wantedResponse := TransactionResponse{
			Limit: 100000,
			Balance: 500,
		}

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, endpoint, strings.NewReader(string(testPayload)))
		router.ServeHTTP(w, req)

		if(w.Code != http.StatusOK) {
			t.Errorf("request didn't returned the right status code. Got %d expected %d", w.Code, http.StatusOK)
		}

		var got TransactionResponse
		json.NewDecoder(w.Body).Decode(&got)

		if(!reflect.DeepEqual(wantedResponse, got)) {
			t.Errorf("got %v want %v", got, wantedResponse) 
		}

	})

	t.Run("debits decrease amount", func(t *testing.T) { 
		router := setupServer()
		accId := 1
		endpoint := fmt.Sprintf("/clientes/%d/transacoes", accId)
		testPayload, _ := json.Marshal(TransactionRequest{
			Value: 500,
			TransactionType: "d",
			Description: "Test transaction",
		})
		wantedResponse := TransactionResponse{
			Limit: 100000,
			Balance: -500,
		}

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, endpoint, strings.NewReader(string(testPayload)))
		router.ServeHTTP(w, req)

		if(w.Code != http.StatusOK) {
			t.Errorf("request didn't returned the right status code. Got %d expected %d", w.Code, http.StatusOK)
		}

		var got TransactionResponse
		json.NewDecoder(w.Body).Decode(&got)

		if(!reflect.DeepEqual(wantedResponse, got)) {
			t.Errorf("got %v want %v", got, wantedResponse) 
		}

	})

	t.Run("debits won't exceed limit", func(t *testing.T) { 
		router := setupServer()
		accId := 1
		endpoint := fmt.Sprintf("/clientes/%d/transacoes", accId)
		testPayload, _ := json.Marshal(TransactionRequest{
			Value: 100001,
			TransactionType: "d",
			Description: "Test transaction",
		})

		wantedCode := http.StatusUnprocessableEntity

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, endpoint, strings.NewReader(string(testPayload)))
		router.ServeHTTP(w, req)

		if(w.Code != wantedCode) {
			t.Errorf("request didn't returned the right status code. Got %d expected %d", w.Code, wantedCode)
		}
	})
	
	t.Run("unknown ID returns 404", func(t *testing.T) { 
		router := setupServer()
		accId := 10
		endpoint := fmt.Sprintf("/clientes/%d/transacoes", accId)
		testPayload, _ := json.Marshal(TransactionRequest{
			Value: 100001,
			TransactionType: "d",
			Description: "Test transaction",
		})

		wantedCode := http.StatusNotFound

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, endpoint, strings.NewReader(string(testPayload)))
		router.ServeHTTP(w, req)

		if(w.Code != wantedCode) {
			t.Errorf("request didn't returned the right status code. Got %d expected %d", w.Code, wantedCode)
		}
	})
}