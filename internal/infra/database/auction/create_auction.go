package auction

import (
	"context"
	"fmt"
	"fullcycle-auction_go/configuration/logger"
	"fullcycle-auction_go/internal/entity/auction_entity"
	"fullcycle-auction_go/internal/internal_error"
	"os"
	"strconv"
	"time"

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
	EndTime     int64                           `bson:"end_time"`
}
type AuctionRepository struct {
	Collection *mongo.Collection
}

func NewAuctionRepository(database *mongo.Database) *AuctionRepository {
	return &AuctionRepository{
		Collection: database.Collection("auctions"),
	}
}

func (ar *AuctionRepository) CreateAuction(
	ctx context.Context,
	auctionEntity *auction_entity.Auction,
) *internal_error.InternalError {

	// Lê a variável de ambiente para definir a duração em minutos
	durationStr := os.Getenv("AUCTION_DURATION_MINUTES")
	if durationStr == "" {
		durationStr = "5" // valor default caso não esteja setada
	}
	durationMinutes, err := strconv.Atoi(durationStr)
	if err != nil {
		logger.Error("Error parsing AUCTION_DURATION_MINUTES", err)
		// Ou retorne um erro apropriado...
		return internal_error.NewInternalServerError("Invalid duration for auction")
	}

	// Calcula o horário de encerramento do leilão
	endTime := time.Now().Add(time.Duration(durationMinutes) * time.Minute)

	// Ajusta os campos na entidade antes de salvar
	auctionEntity.Status = auction_entity.Active // status inicial "active"
	auctionEntity.EndTime = endTime       // define a data/hora de encerramento

	// Cria a estrutura AuctionEntityMongo para salvar no Mongo
	auctionEntityMongo := &AuctionEntityMongo{
		Id:           auctionEntity.Id,
		ProductName:  auctionEntity.ProductName,
		Category:     auctionEntity.Category,
		Description:  auctionEntity.Description,
		Condition:    auctionEntity.Condition,
		Status:       auctionEntity.Status,
		Timestamp:    auctionEntity.Timestamp.Unix(),
		EndTime:      auctionEntity.EndTime.Unix(),
	}

	// 5. Insere no MongoDB
	_, err = ar.Collection.InsertOne(ctx, auctionEntityMongo)
	if err != nil {
		logger.Error("Error trying to insert auction", err)
		return internal_error.NewInternalServerError("Error trying to insert auction")
	}

	logger.Info(fmt.Sprintf("Auction inserted with ID %s, EndTimestamp = %v", auctionEntity.Id, auctionEntity.EndTime))
	return nil
}

// ----------------------------------------------------------------------------
// Goroutine para fechamento automático dos leilões
// ----------------------------------------------------------------------------

// StartAuctionExpirationChecker inicia uma goroutine que, a cada X segundos,
// verifica se há leilões abertos cujo EndTime<= now e os fecha.
func (ar *AuctionRepository) StartAuctionExpirationChecker(ctx context.Context) {
	go func() {
		// Lê do .env o intervalo de checagem (em segundos)
		checkIntervalStr := os.Getenv("AUCTION_CHECK_INTERVAL_SECONDS")
		if checkIntervalStr == "" {
			checkIntervalStr = "10" // valor padrão
		}
		interval, err := strconv.Atoi(checkIntervalStr)
		if err != nil {
			interval = 10
		}

		for {
			// Aguarda o intervalo configurado
			time.Sleep(time.Duration(interval) * time.Second)

			now := time.Now().Unix()

			// Filtro: status = "active" e end_time <= now
			filter := bson.M{
				"status":        auction_entity.Active,
				"end_timestamp": bson.M{"$lte": now},
			}
			update := bson.M{
				"$set": bson.M{
					"status": auction_entity.Completed,
				},
			}

			// Atualiza todos os documentos que satisfazem o filtro
			res, err := ar.Collection.UpdateMany(ctx, filter, update)
			if err != nil {
				logger.Error("Error trying to close expired auctions", err)
				continue
			}

			if res.ModifiedCount > 0 {
				logger.Info(fmt.Sprintf("Closed %d auctions automatically", res.ModifiedCount))
			}
		}
	}()
}
