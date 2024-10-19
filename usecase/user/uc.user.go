package user

import (
	"context"
	"net/http"
	"regexp"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/streadway/amqp"
	"gitlab.com/trisaptono/producer/config" // Import config package
	"gitlab.com/trisaptono/producer/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var ctx = context.Background()

// ValidateEmail checks if the provided email is in a valid format
func ValidateEmail(email string) bool {
	re := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return re.MatchString(email)
}

// MaxNameLength adalah panjang maksimum untuk nama
const MaxNameLength = 10 // Maksimum panjang nama
const MinNameLength = 5  // Minimum panjang nama

// ValidateNameLength checks if the name is within the allowed length
func ValidateNameLength(name string) bool {
	return len(name) >= MinNameLength && len(name) <= MaxNameLength
}

// CreateUser handles the creation of a new user
func CreateUser(c *gin.Context) {
	var user model.User

	// Bind JSON input to user struct
	if err := c.ShouldBindJSON(&user); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	// Validate email format
	if !ValidateEmail(user.Email) {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid email format"})
		return
	}

	// Validate name length
	if !ValidateNameLength(user.Nama) {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Name must be between 5 and 10 characters"})
		return
	}

	// Store user in PostgreSQL
	if err := model.DB.Create(&user).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user", "message": err.Error()})
		return
	}

	// Store user in MongoDB
	mongoClient := config.GetMongoClient()
	collection := mongoClient.Database("user").Collection("users") // Ganti dengan nama database MongoDB Anda
	_, err := collection.InsertOne(ctx, user)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to store user in MongoDB", "message": err.Error()})
		return
	}

	// Store session in Redis
	redisClient := config.GetRedisClient()
	err = redisClient.Set(ctx, user.Email, user.ID, 0).Err() // Simpan email dan ID pengguna
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to store session in Redis", "message": err.Error()})
		return
	}

	// Send message to RabbitMQ
	err = SendMessage("user_created", user.Email)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to send message to RabbitMQ", "message": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"user": user})
}

// SendMessage sends a message to RabbitMQ
func SendMessage(queueName string, message string) error {
	conn := config.GetRabbitConnection()
	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	_, err = ch.QueueDeclare(
		queueName, // name
		false,     // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		return err
	}

	err = ch.Publish(
		"",        // exchange
		queueName, // routing key
		false,     // mandatory
		false,     // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(message),
		},
	)
	return err
}

// GetUsers retrieves all users
func GetUsers(c *gin.Context) {
	var users []model.User

	// Retrieve all users from PostgreSQL
	if err := model.DB.Find(&users).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve users", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, users)
}

// GetUser retrieves a user by ID
func GetUser(c *gin.Context) {
	id := c.Param("id")
	var user model.User

	// Find user by ID in PostgreSQL
	if err := model.DB.First(&user, id).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// UpdateUser updates an existing user
func UpdateUser(c *gin.Context) {
	id := c.Param("id")
	var user model.User

	// Check if the user exists in PostgreSQL
	if err := model.DB.First(&user, id).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Bind JSON input to user struct
	if err := c.ShouldBindJSON(&user); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	// Validate email format
	if !ValidateEmail(user.Email) {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid email format"})
		return
	}

	// Validate name length
	if !ValidateNameLength(user.Nama) {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Name must be between 5 and 10 characters"})
		return
	}

	// Update user in PostgreSQL
	if err := model.DB.Save(&user).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user", "message": err.Error()})
		return
	}

	// Update user in MongoDB
	mongoClient := config.GetMongoClient()
	collection := mongoClient.Database("user").Collection("users") // Ganti dengan nama database MongoDB Anda

	update := bson.M{
		"$set": bson.M{
			"nama":  user.Nama, // Sesuaikan dengan field yang ada di model
			"email": user.Email,
		},
	}

	filter := bson.M{"id": user.ID} // Pastikan Anda menggunakan ID yang sesuai
	_, err := collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user in MongoDB", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}

// DeleteUser deletes a user by ID
// DeleteUser deletes a user by ID
func DeleteUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64) // Konversi ID string ke int64
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Delete user from PostgreSQL
	if err := model.DB.Delete(&model.User{}, id).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Convert ID to ObjectID for MongoDB (jika Anda ingin menyimpan ObjectID)
	mongoClient := config.GetMongoClient()
	collection := mongoClient.Database("user").Collection("users")             // Ganti dengan nama database MongoDB Anda
	_, err = collection.DeleteOne(context.Background(), primitive.M{"id": id}) // Ganti sesuai field yang ada di MongoDB
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user from MongoDB", "message": err.Error()})
		return
	}

	// Delete session from Redis
	redisClient := config.GetRedisClient()
	err = redisClient.Del(context.Background(), idStr).Err()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete session from Redis", "message": err.Error()})
		return
	}

	c.Status(http.StatusNoContent) // Return 204 No Content
}
