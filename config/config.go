package config

import (
	"context"
	"log"
	"os"

	"gitlab.com/trisaptono/producer/model" // Pastikan path ini sesuai dengan struktur project Anda

	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	mongoClient *mongo.Client
	redisClient *redis.Client
	rabbitConn  *amqp.Connection
	ctx         = context.Background()
)

// LoadEnv loads environment variables from .env file
func LoadEnv() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
}

// InitializeDatabase initializes PostgreSQL, MongoDB, Redis, and RabbitMQ connections
func InitializeDatabase() {
	// PostgreSQL
	DbHost := os.Getenv("DB_HOST")
	DbPort := os.Getenv("DB_PORT")
	DbUser := os.Getenv("DB_USER")
	DbPwd := os.Getenv("DB_PASSWORD")
	DbName := os.Getenv("DB_NAME")

	log.Printf("Connecting to PostgreSQL database at %s:%s...", DbHost, DbPort)
	model.ConnectDatabase(DbUser, DbPwd, DbHost, DbPort, DbName, os.Getenv("MONGODB_URI"), "user")
	log.Printf("PostgreSQL database successfully connected!")

	// MongoDB
	MongoURI := os.Getenv("MONGODB_URI")
	log.Printf("Connecting to MongoDB at %s...", MongoURI)

	var err error
	mongoClient, err = mongo.Connect(ctx, options.Client().ApplyURI(MongoURI))
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	if err = mongoClient.Ping(ctx, nil); err != nil {
		log.Fatalf("MongoDB connection error: %v", err)
	}
	log.Printf("MongoDB successfully connected!")

	// Redis
	log.Printf("Connecting to Redis...")
	redisClient = redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_ADDR"),
	})
	_, err = redisClient.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	log.Printf("Redis successfully connected!")

	// RabbitMQ
	RabbitMQURL := os.Getenv("RABBITMQ_URL")
	log.Printf("Connecting to RabbitMQ at %s...", RabbitMQURL)

	rabbitConn, err = amqp.Dial(RabbitMQURL)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	log.Printf("RabbitMQ successfully connected!")
}

// GetMongoClient returns the MongoDB client
func GetMongoClient() *mongo.Client {
	return mongoClient
}

// GetRedisClient returns the Redis client
func GetRedisClient() *redis.Client {
	return redisClient
}

// GetRabbitConnection returns the RabbitMQ connection
func GetRabbitConnection() *amqp.Connection {
	return rabbitConn
}
