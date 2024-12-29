package test

import (
	"context"
	"os"
	"strconv"
	"testing"
	"time"

	"fullcycle-auction_go/internal/entity/auction_entity"
	"fullcycle-auction_go/internal/infra/database/auction"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Ajuste se precisar de outra porta ou URL para Mongo
var mongoURI = "mongodb://localhost:27017"
var testDBName = "auction_test_db"

func TestCreateAuctionAndAutoClose(t *testing.T) {
	// Passo 1: Configura variáveis de ambiente específicas para o teste
	// Definindo uma duração curta (ex.: 0 minutos) e checagem rápida (1s)
	os.Setenv("AUCTION_DURATION_MINUTES", "0")         // cria leilão com EndTime = now
	os.Setenv("AUCTION_CHECK_INTERVAL_SECONDS", "1")   // rotina roda a cada 1s

	ctx := context.Background()

	// Passo 2: Conecta a um Mongo de teste
	clientOpts := options.Client().ApplyURI(mongoURI)
	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		t.Fatalf("Erro ao conectar no Mongo: %v", err)
	}
	defer client.Disconnect(ctx)

	// Usa uma base de dados separada para teste (evita colidir com dados de produção)
	testDB := client.Database(testDBName)

	// Passo 3: Cria o repositório de Auction
	auctionRepo := auction.NewAuctionRepository(testDB)

	// Importante: Limpamos a coleção "auctions" antes do teste
	if err := cleanCollection(ctx, testDB, "auctions"); err != nil {
		t.Fatalf("Erro ao limpar collection de teste: %v", err)
	}

	// Passo 4: Inicia a goroutine de fechamento automático
	auctionRepo.StartAuctionExpirationChecker(ctx)

	// Passo 5: Cria um Auction com status "open" e EndTime = now (instantâneo)
	// Assim, teoricamente deve fechar rapidinho.
	now := time.Now()
	a := auction_entity.Auction{
		Id:            "auction_test_001",
		ProductName:   "Produto Teste",
		Category:      "TestCategory",
		Description:   "Desc teste",
		Condition:     auction_entity.New,
		Status:        auction_entity.Active,
		Timestamp:     now,
		EndTime:       now, // EndTime = now => significa que já está 'vencido'
	}

	// Chamamos a função CreateAuction, que deve setar EndTime (se AUCTION_DURATION_MINUTES != 0)
	if err := auctionRepo.CreateAuction(ctx, &a); err != nil {
		t.Fatalf("Erro ao criar auction: %v", err)
	}

	// Verifica se a CreateAuction recalculou EndTime adequadamente
	// Se "AUCTION_DURATION_MINUTES" = 0, deve ficar igual a 'now'
	// Se fosse 5, seria 'now+5min', etc.
	if durationStr := os.Getenv("AUCTION_DURATION_MINUTES"); durationStr != "" {
		dur, _ := strconv.Atoi(durationStr)
		// Se dur = 0, então Auction.EndTime deve ficar = time.Now() (aprox)
		// Fazemos uma checagem de tolerância de no máximo 2s.
		if dur == 0 {
			// Tolerância de 2s de diferença
			elapsed := time.Since(a.EndTime)
			if elapsed < 0 || elapsed > 2*time.Second {
				t.Errorf("EndTime esperado ~ now, mas diff = %v", elapsed)
			}
		}
	}

	// Passo 6: Aguarda alguns segundos para que a goroutine rode e feche o leilão
	time.Sleep(3 * time.Second)

	// Passo 7: Lê novamente o auction do Mongo e checa status
	var result struct {
		Status string `bson:"status"`
	}
	if err := testDB.Collection("auctions").FindOne(ctx, bson.M{"_id": a.Id}).Decode(&result); err != nil {
		t.Fatalf("Erro ao buscar auction no Mongo: %v", err)
	}

	if result.Status != string(auction_entity.Completed) {
		t.Errorf("Esperado que o status do Auction fosse 'closed', mas está '%s'", result.Status)
	}
}

// Função auxiliar para limpar dados da coleção antes do teste
func cleanCollection(ctx context.Context, db *mongo.Database, collection string) error {
	return db.Collection(collection).Drop(ctx)
}
