package auction_controller_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/rafaelsouzaribeiro/labs-auction/configuration/database/mongodb"
	"github.com/rafaelsouzaribeiro/labs-auction/internal/entity/auction_entity"
	"github.com/rafaelsouzaribeiro/labs-auction/internal/infra/api/web/controller/auction_controller"
	"github.com/rafaelsouzaribeiro/labs-auction/internal/infra/database/auction"
	"github.com/rafaelsouzaribeiro/labs-auction/internal/infra/database/bid"
	"github.com/rafaelsouzaribeiro/labs-auction/internal/usecase/auction_usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func setupTestEnvironment(t *testing.T) (*gin.Engine, *mongo.Database) {
	gin.SetMode(gin.TestMode)

	err := godotenv.Load("../../../../../../cmd/auction/.env")
	require.NoError(t, err, "erro ao carregar arquivo .env")

	os.Setenv("MONGODB_DB", "auctions_integration_test")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	database, err := mongodb.NewMongoDBConnection(ctx)
	require.NoError(t, err, "erro ao conectar no MongoDB")

	pingCtx, pingCancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer pingCancel()
	require.NoError(t, database.Client().Ping(pingCtx, nil), "mongo não respondeu ao ping")

	auctionRepo := auction.NewAuctionRepository(database)
	bidRepo := bid.NewBidRepository(database, auctionRepo)
	useCase := auction_usecase.NewAuctionUseCase(auctionRepo, bidRepo)
	controller := auction_controller.NewAuctionController(useCase)

	router := gin.Default()
	router.POST("/auction", controller.CreateAuction)
	router.GET("/auction/:auctionId", controller.FindAuctionById)

	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cleanupCancel()
		_ = database.Collection("auctions").Drop(cleanupCtx)
		_ = database.Collection("bids").Drop(cleanupCtx)
	})

	return router, database
}

func TestCreateAuction_Success(t *testing.T) {
	os.Setenv("AUCTION_DURATION", "5s")
	router, database := setupTestEnvironment(t)

	body := map[string]interface{}{
		"product_name": "Notebook Dell",
		"category":     "Eletrônicos",
		"description":  "Notebook novo em folha para testes",
		"condition":    1,
	}
	jsonBody, err := json.Marshal(body)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/auction", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusCreated, recorder.Code)

	var savedAuction auction.AuctionEntityMongo
	err = database.Collection("auctions").FindOne(context.Background(), bson.M{"product_name": "Notebook Dell"}).Decode(&savedAuction)
	require.NoError(t, err, "o leilão deveria estar salvo no banco de dados")

	assert.Equal(t, "Notebook Dell", savedAuction.ProductName)
	assert.Equal(t, auction_entity.Active, savedAuction.Status)
}

func TestCreateAuction_AutoClose_Integration(t *testing.T) {
	os.Setenv("AUCTION_DURATION", "1s")
	router, database := setupTestEnvironment(t)

	body := map[string]interface{}{
		"product_name": "iPhone 15",
		"category":     "Smartphones",
		"description":  "iPhone excelente estado",
		"condition":    2,
	}
	jsonBody, err := json.Marshal(body)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/auction", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)
	assert.Equal(t, http.StatusCreated, recorder.Code)

	var savedAuction auction.AuctionEntityMongo
	err = database.Collection("auctions").FindOne(context.Background(), bson.M{"product_name": "iPhone 15"}).Decode(&savedAuction)
	require.NoError(t, err)
	assert.Equal(t, auction_entity.Active, savedAuction.Status)

	time.Sleep(1500 * time.Millisecond)

	err = database.Collection("auctions").FindOne(context.Background(), bson.M{"_id": savedAuction.Id}).Decode(&savedAuction)
	require.NoError(t, err)
	assert.Equal(t, auction_entity.Closed, savedAuction.Status, "o status deveria mudar para Closed após o tempo configurado")
}

func TestCreateAuction_BadRequest(t *testing.T) {
	router, _ := setupTestEnvironment(t)

	body := map[string]interface{}{
		"product_name": "",
		"category":     "El",
		"description":  "Curta",
		"condition":    1,
	}
	jsonBody, err := json.Marshal(body)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/auction", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}
