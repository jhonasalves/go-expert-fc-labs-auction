package auction

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/jhonasalves/go-expert-fc-labs-auction/internal/entity/auction_entity"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

func TestAuctionMonitorClosesExpiredAuctions(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("closeExpiredAuctions", func(mt *mtest.T) {
		mockCollection := mt.Coll

		auctionID := primitive.NewObjectID().Hex()
		expiredTime := time.Now().Add(-2 * time.Second)

		repo := &AuctionRepository{
			Collection:            mockCollection,
			auctionDuration:       time.Second * 1,
			auctionStatusMap:      map[string]auction_entity.AuctionStatus{auctionID: auction_entity.Active},
			auctionEndTimeMap:     map[string]time.Time{auctionID: expiredTime},
			auctionStatusMapMutex: &sync.Mutex{},
			auctionEndTimeMutex:   &sync.Mutex{},
		}

		mt.AddMockResponses(
			mtest.CreateSuccessResponse(),
		)

		go repo.StartAuctionMonitor(context.Background(), time.Second*1)

		time.Sleep(time.Second * 2)

		assert.NotContains(t, repo.auctionEndTimeMap, auctionID, "Auction expired should be removed from the end time map")
		assert.NotContains(t, repo.auctionStatusMap, auctionID, "Auction expired should be removed from the status map")
	})
}
