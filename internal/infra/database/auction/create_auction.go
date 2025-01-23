package auction

import (
	"context"
	"os"
	"sync"
	"time"

	"github.com/jhonasalves/go-expert-fc-labs-auction/configuration/logger"
	"github.com/jhonasalves/go-expert-fc-labs-auction/internal/entity/auction_entity"
	"github.com/jhonasalves/go-expert-fc-labs-auction/internal/internal_error"

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
	Collection            *mongo.Collection
	auctionDuration       time.Duration
	auctionStatusMap      map[string]auction_entity.AuctionStatus
	auctionEndTimeMap     map[string]time.Time
	auctionStatusMapMutex *sync.Mutex
	auctionEndTimeMutex   *sync.Mutex
}

func NewAuctionRepository(database *mongo.Database) *AuctionRepository {
	ar := &AuctionRepository{
		Collection:            database.Collection("auctions"),
		auctionDuration:       getAuctionDuration(),
		auctionStatusMap:      make(map[string]auction_entity.AuctionStatus),
		auctionEndTimeMap:     make(map[string]time.Time),
		auctionStatusMapMutex: &sync.Mutex{},
		auctionEndTimeMutex:   &sync.Mutex{},
	}

	ar.StartAuctionMonitor(context.Background(), ar.auctionDuration)

	return ar
}

func (ar *AuctionRepository) StartAuctionMonitor(ctx context.Context, checkInterval time.Duration) {
	ticker := time.NewTicker(checkInterval)

	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				ar.closeExpiredAuctions(ctx)
			}
		}
	}()
}

func (ar *AuctionRepository) closeExpiredAuctions(ctx context.Context) {
	now := time.Now()
	ar.auctionEndTimeMutex.Lock()
	defer ar.auctionEndTimeMutex.Unlock()

	for id, endTime := range ar.auctionEndTimeMap {
		if now.After(endTime) {

			ar.auctionStatusMapMutex.Lock()
			ar.auctionStatusMap[id] = auction_entity.Completed
			ar.auctionStatusMapMutex.Unlock()

			filter := bson.M{"_id": id}
			update := bson.M{"$set": bson.M{"status": auction_entity.Completed}}

			_, err := ar.Collection.UpdateOne(
				ctx,
				filter,
				update,
			)
			if err != nil {
				logger.Error("Error trying to update auction status", err)
			}

			delete(ar.auctionEndTimeMap, id)
			delete(ar.auctionStatusMap, id)
		}
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

	ar.auctionEndTimeMutex.Lock()
	ar.auctionEndTimeMap[auctionEntity.Id] = auctionEntity.Timestamp.Add(ar.auctionDuration)
	ar.auctionEndTimeMutex.Unlock()

	ar.auctionStatusMapMutex.Lock()
	ar.auctionStatusMap[auctionEntity.Id] = auctionEntity.Status
	ar.auctionStatusMapMutex.Unlock()

	return nil
}

func getAuctionDuration() time.Duration {
	auctionDuration := os.Getenv("AUCTION_DURATION")

	duration, err := time.ParseDuration(auctionDuration)
	if err != nil {
		return time.Minute * 10
	}

	return duration
}
