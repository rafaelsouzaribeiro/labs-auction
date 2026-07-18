package auction_controller_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/rafaelsouzaribeiro/labs-auction/internal/entity/auction_entity"
	"github.com/rafaelsouzaribeiro/labs-auction/internal/entity/bid_entity"
	"github.com/rafaelsouzaribeiro/labs-auction/internal/infra/api/web/controller/auction_controller"
	"github.com/rafaelsouzaribeiro/labs-auction/internal/internal_error"
	"github.com/rafaelsouzaribeiro/labs-auction/internal/usecase/auction_usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type AuctionRepositoryMock struct {
	mock.Mock
}

func (m *AuctionRepositoryMock) CreateAuction(ctx context.Context, auction *auction_entity.Auction) *internal_error.InternalError {
	args := m.Called(ctx, auction)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*internal_error.InternalError)
}

func (m *AuctionRepositoryMock) FindAuctions(ctx context.Context, status auction_entity.AuctionStatus, category, productName string) ([]auction_entity.Auction, *internal_error.InternalError) {
	args := m.Called(ctx, status, category, productName)
	return args.Get(0).([]auction_entity.Auction), nil
}

func (m *AuctionRepositoryMock) FindAuctionById(ctx context.Context, id string) (*auction_entity.Auction, *internal_error.InternalError) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*internal_error.InternalError)
	}
	return args.Get(0).(*auction_entity.Auction), nil
}

type BidRepositoryMock struct {
	mock.Mock
}

func (m *BidRepositoryMock) CreateBid(ctx context.Context, bids []bid_entity.Bid) *internal_error.InternalError {
	args := m.Called(ctx, bids)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*internal_error.InternalError)
}

func (m *BidRepositoryMock) FindBidByAuctionId(ctx context.Context, auctionId string) ([]bid_entity.Bid, *internal_error.InternalError) {
	args := m.Called(ctx, auctionId)
	return args.Get(0).([]bid_entity.Bid), nil
}

func (m *BidRepositoryMock) FindWinningBidByAuctionId(ctx context.Context, auctionId string) (*bid_entity.Bid, *internal_error.InternalError) {
	args := m.Called(ctx, auctionId)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*internal_error.InternalError)
	}
	return args.Get(0).(*bid_entity.Bid), nil
}

func setupMockEnvironment(t *testing.T, auctionRepo *AuctionRepositoryMock, bidRepo *BidRepositoryMock) *gin.Engine {
	gin.SetMode(gin.TestMode)

	useCase := auction_usecase.NewAuctionUseCase(auctionRepo, bidRepo)
	controller := auction_controller.NewAuctionController(useCase)

	router := gin.Default()
	router.POST("/auction", controller.CreateAuction)
	router.GET("/auction/:auctionId", controller.FindAuctionById)

	return router
}

func TestCreateAuction_Success_Mock(t *testing.T) {
	auctionRepo := new(AuctionRepositoryMock)
	bidRepo := new(BidRepositoryMock)

	auctionRepo.On("CreateAuction", mock.Anything, mock.MatchedBy(func(a *auction_entity.Auction) bool {
		return a.ProductName == "Notebook Dell" && a.Status == auction_entity.Active
	})).Return(nil)

	router := setupMockEnvironment(t, auctionRepo, bidRepo)

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
	auctionRepo.AssertExpectations(t)
}

func TestCreateAuction_BadRequest_Mock(t *testing.T) {
	auctionRepo := new(AuctionRepositoryMock)
	bidRepo := new(BidRepositoryMock)

	router := setupMockEnvironment(t, auctionRepo, bidRepo)

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

	auctionRepo.AssertNotCalled(t, "CreateAuction")
}
