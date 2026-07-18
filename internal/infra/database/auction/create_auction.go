package auction

import (
	"context"
	"os"
	"time"

	"github.com/rafaelsouzaribeiro/labs-auction/configuration/logger"
	"github.com/rafaelsouzaribeiro/labs-auction/internal/entity/auction_entity"
	"github.com/rafaelsouzaribeiro/labs-auction/internal/internal_error"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type AuctionEntityMongo struct {
	Id          string                          `bson:"_id"`
	ProductName string                          `bson:"product_name"`
	Category    string                          `bson:"category"`
	Description string                          `bson:"description"`
	Condition   auction_entity.ProductCondition `bson:"condition"`
	Status      auction_entity.AuctionStatus    `bson:"status"`
	Timestamp   int64                           `bson:"timestamp"`
}
type AuctionRepository struct {
	Collection      *mongo.Collection
	auctionDuration time.Duration
}

func getAuctionDuration() time.Duration {
	duration, err := time.ParseDuration(os.Getenv("AUCTION_DURATION"))
	if err != nil || duration <= 0 {
		return 5 * time.Minute
	}
	return duration
}

func NewAuctionRepository(database *mongo.Database) *AuctionRepository {
	return &AuctionRepository{
		Collection:      database.Collection("auctions"),
		auctionDuration: getAuctionDuration(),
	}
}

func (ar *AuctionRepository) CreateAuction(
	ctx context.Context,
	auctionEntity *auction_entity.Auction) *internal_error.InternalError {
	auctionEntityMongo := &AuctionEntityMongo{
		Id:          auctionEntity.Id,
		ProductName: auctionEntity.ProductName,
		Category:    auctionEntity.Category,
		Description: auctionEntity.Description,
		Condition:   auctionEntity.Condition,
		Status:      auctionEntity.Status,
		Timestamp:   auctionEntity.Timestamp.Unix(),
	}
	_, err := ar.Collection.InsertOne(ctx, auctionEntityMongo)
	if err != nil {
		logger.Error("Error trying to insert auction", err)
		return internal_error.NewInternalServerError("Error trying to insert auction")
	}

	go ar.closeAuctionAfterDuration(auctionEntity.Id)

	return nil
}

func (ar *AuctionRepository) closeAuctionAfterDuration(auctionId string) {
	time.Sleep(ar.auctionDuration)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"_id": auctionId, "status": auction_entity.Active}
	update := bson.M{"$set": bson.M{"status": auction_entity.Closed}}

	if _, err := ar.Collection.UpdateOne(ctx, filter, update); err != nil {
		logger.Error("Error trying to close auction automatically", err)
	}
}
