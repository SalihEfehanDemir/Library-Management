package main

import (
	"context"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)


var userCollection *mongo.Collection
var bookCollection *mongo.Collection


type User struct {
	ID       primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	Username string               `bson:"username" json:"username"`
	Password string               `bson:"password,omitempty" json:"-"` 
	Books    []primitive.ObjectID `bson:"books" json:"books"`         
}


type Book struct {
	ID         primitive.ObjectID  `bson:"_id,omitempty" json:"id"`
	Title      string              `bson:"title" json:"title"`
	BorrowerID *primitive.ObjectID `bson:"borrower_id,omitempty" json:"borrower_id,omitempty"`
}


func connectDB() *mongo.Client {
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal("MongoDB Client oluşturulamadı:", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := client.Connect(ctx); err != nil {
		log.Fatal("MongoDB'ye bağlanılamadı:", err)
	}
	return client
}


func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}


func checkPasswordHash(password, hashed string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(password))
	return err == nil
}

func main() {
	
	client := connectDB()
	db := client.Database("library")
	userCollection = db.Collection("users")
	bookCollection = db.Collection("books")

	
	app := fiber.New()

	
	app.Use(logger.New())

	
	app.Post("/register", registerUser)
	app.Post("/login", loginUser)
	app.Get("/user/:id", getUser)
	app.Delete("/user/:id", deleteUser)

	app.Post("/book", addBook)
	app.Get("/books", listBooks)

	app.Post("/borrow", borrowBook)
	app.Post("/return", returnBook)

	
	log.Fatal(app.Listen(":3000"))
}


func registerUser(c *fiber.Ctx) error {
	type request struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	var body request

	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Geçersiz JSON"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()


	count, err := userCollection.CountDocuments(ctx, bson.M{"username": body.Username})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Veritabanı hatası"})
	}
	if count > 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Kullanıcı adı zaten mevcut"})
	}

	hashed, err := hashPassword(body.Password)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Şifre hashlenemedi"})
	}

	user := User{
		Username: body.Username,
		Password: hashed,
		Books:    []primitive.ObjectID{},
	}

	res, err := userCollection.InsertOne(ctx, user)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Kullanıcı eklenemedi"})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"inserted_id": res.InsertedID})
}

func loginUser(c *fiber.Ctx) error {
	type request struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	var body request

	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Geçersiz JSON"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	
	var user User
	if err := userCollection.FindOne(ctx, bson.M{"username": body.Username}).Decode(&user); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Kullanıcı bulunamadı"})
	}

	
	if !checkPasswordHash(body.Password, user.Password) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Hatalı şifre"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Giriş başarılı",
		"user_id": user.ID,
	})
}

func getUser(c *fiber.Ctx) error {
	id := c.Params("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Geçersiz kullanıcı ID"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var user User
	if err := userCollection.FindOne(ctx, bson.M{"_id": objID}).Decode(&user); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Kullanıcı bulunamadı"})
	}

	user.Password = ""
	return c.Status(fiber.StatusOK).JSON(user)
}

func deleteUser(c *fiber.Ctx) error {
	id := c.Params("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Geçersiz kullanıcı ID"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := userCollection.DeleteOne(ctx, bson.M{"_id": objID})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Kullanıcı silinemedi"})
	}
	if res.DeletedCount == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Kullanıcı bulunamadı"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Kullanıcı silindi"})
}


func addBook(c *fiber.Ctx) error {
	type request struct {
		Title string `json:"title"`
	}
	var body request

	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Geçersiz JSON"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	book := Book{
		Title:      body.Title,
		BorrowerID: nil,
	}

	res, err := bookCollection.InsertOne(ctx, book)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Kitap eklenemedi"})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"inserted_id": res.InsertedID})
}

func listBooks(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := bookCollection.Find(ctx, bson.M{})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Kitaplar alınamadı"})
	}
	defer cursor.Close(ctx)

	var books []Book
	if err := cursor.All(ctx, &books); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Kitaplar parse edilemedi"})
	}

	return c.Status(fiber.StatusOK).JSON(books)
}


func borrowBook(c *fiber.Ctx) error {
	type request struct {
		UserID string `json:"user_id"`
		BookID string `json:"book_id"`
	}
	var body request

	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Geçersiz JSON"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	userObjID, err := primitive.ObjectIDFromHex(body.UserID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Geçersiz user_id"})
	}
	bookObjID, err := primitive.ObjectIDFromHex(body.BookID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Geçersiz book_id"})
	}


	var user User
	if err := userCollection.FindOne(ctx, bson.M{"_id": userObjID}).Decode(&user); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Kullanıcı bulunamadı"})
	}
	
	if len(user.Books) >= 2 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Kullanıcının 2 kitap limiti doldu"})
	}


	var book Book
	if err := bookCollection.FindOne(ctx, bson.M{"_id": bookObjID}).Decode(&book); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Kitap bulunamadı"})
	}

	if book.BorrowerID != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Kitap zaten ödünç alınmış"})
	}

	
	_, err = bookCollection.UpdateOne(ctx,
		bson.M{"_id": bookObjID},
		bson.M{"$set": bson.M{"borrower_id": userObjID}},
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Kitap güncellenemedi"})
	}

	
	_, err = userCollection.UpdateOne(ctx,
		bson.M{"_id": userObjID},
		bson.M{"$push": bson.M{"books": bookObjID}},
	)
	if err != nil {
		bookCollection.UpdateOne(ctx, bson.M{"_id": bookObjID}, bson.M{"$set": bson.M{"borrower_id": nil}})
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Kullanıcı güncellenemedi"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Kitap başarıyla ödünç alındı"})
}

func returnBook(c *fiber.Ctx) error {
	type request struct {
		UserID string `json:"user_id"`
		BookID string `json:"book_id"`
	}
	var body request

	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Geçersiz JSON"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	userObjID, err := primitive.ObjectIDFromHex(body.UserID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Geçersiz user_id"})
	}
	bookObjID, err := primitive.ObjectIDFromHex(body.BookID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Geçersiz book_id"})
	}


	var book Book
	if err := bookCollection.FindOne(ctx, bson.M{"_id": bookObjID}).Decode(&book); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Kitap bulunamadı"})
	}
	
	if book.BorrowerID == nil || *book.BorrowerID != userObjID {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Bu kitap bu kullanıcıya ait değil"})
	}

	
	_, err = bookCollection.UpdateOne(ctx,
		bson.M{"_id": bookObjID},
		bson.M{"$set": bson.M{"borrower_id": nil}},
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Kitap güncellenemedi"})
	}

	
	_, err = userCollection.UpdateOne(ctx,
		bson.M{"_id": userObjID},
		bson.M{"$pull": bson.M{"books": bookObjID}},
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Kullanıcı güncellenemedi"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Kitap başarıyla iade edildi"})
}
