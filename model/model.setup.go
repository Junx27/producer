package model

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	DB      *gorm.DB
	MongoDB *mongo.Database
)

// ConnectDatabase menghubungkan ke database PostgreSQL dan MongoDB
func ConnectDatabase(DbUser, DbPwd, DbHost, DbPort, DbName, mongoUri, mongoDbName string) (*gorm.DB, *mongo.Database) {
	// Menyusun Data Source Name (DSN) untuk PostgreSQL
	dsn := "host=" + DbHost + " user=" + DbUser + " password=" + DbPwd + " dbname=" + DbName + " port=" + DbPort + " sslmode=disable TimeZone=Asia/Shanghai"
	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Gagal terhubung ke database PostgreSQL: %v", err)
		return nil, nil
	}

	DB = database

	// Melakukan auto-migrate untuk model User
	if err := DB.AutoMigrate(&User{}); err != nil {
		log.Fatalf("Auto-migration gagal: %v", err)
	}

	// Menghubungkan ke MongoDB
	clientOptions := options.Client().ApplyURI(mongoUri)
	client, err := mongo.NewClient(clientOptions)
	if err != nil {
		log.Fatalf("Gagal membuat client MongoDB: %v", err)
		return DB, nil
	}

	// Menghubungkan ke MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	if err != nil {
		log.Fatalf("Gagal terhubung ke MongoDB: %v", err)
		return DB, nil
	}

	// Menentukan database MongoDB
	MongoDB = client.Database(mongoDbName)

	return DB, MongoDB
}
