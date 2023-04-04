package controllers

import (
	"context"
	"crypto/tls"
	"gin-mongo-api/configs"
	"gin-mongo-api/models"
	"gin-mongo-api/responses"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"gopkg.in/gomail.v2"
	"gopkg.in/mgo.v2/bson"
)

var userCollection *mongo.Collection = configs.GetCollection(configs.DB, "users")
var validate = validator.New()

func CreateUser() gin.HandlerFunc {
    return func(c *gin.Context) {
        ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
        var user models.User
        defer cancel()

        if err := c.BindJSON(&user); err != nil {
            c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"data": err.Error()}})
            return
        }

        if validationErr := validate.Struct(&user); validationErr != nil {
            c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"data": validationErr.Error()}})
            return
        }

        // Check if user with the same name already exists
        existingUser, err := userCollection.FindOne(ctx, bson.M{"name": user.Name}).DecodeBytes()
        if err == nil && existingUser != nil {
            c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"data": "User with the same name already exists"}})
            return
        }

        newUser := models.User{
            Id:       primitive.NewObjectID(),
            Name:     user.Name,
            Location: user.Location,
            Title:    user.Title,
        }

        result, err := userCollection.InsertOne(ctx, newUser)
        if err != nil {
            c.JSON(http.StatusInternalServerError, responses.UserResponse{Status: http.StatusInternalServerError, Message: "error", Data: map[string]interface{}{"data": err.Error()}})
            return
        }
        if result.InsertedID != nil {
            go func() {
                err := sendEmailToAdmin(newUser.Name)
                if err != nil {
                    log.Println("Failed to send email to admin:", err)
                }
            }()
        }

        c.JSON(http.StatusCreated, responses.UserResponse{Status: http.StatusCreated, Message: "success", Data: map[string]interface{}{"data": result}})
    }
}

func sendEmailToAdmin(name string) error {
    // Set up SMTP connection parameters
    smtpHost := "smtp.gmail.com"
    smtpPort := 465
    smtpUsername := os.Getenv("SMTP_EMAIL")
    smtpPassword := os.Getenv("SMTP_PASSWORD")

    // Set up email message
    subject := "New User Registration"
    body := "Hi Admin, <br><br> A new user has been registered with the name " + name + ". <br><br> Thanks, <br> Team"
    from := os.Getenv("SMTP_EMAIL")
    to := []string{"avijitsen.me@gmail.com"}

    // Create email message
    message := gomail.NewMessage()
    message.SetHeader("From", from)
    message.SetHeader("To", to...)
    message.SetHeader("Subject", subject)
    message.SetBody("text/html", body)

    // Set up SMTP authentication and send email
    dialer := gomail.NewDialer(smtpHost, smtpPort, smtpUsername, smtpPassword)
    dialer.TLSConfig = &tls.Config{InsecureSkipVerify: true}
    dialer.Auth = smtp.PlainAuth("", smtpUsername, smtpPassword, smtpHost)
    err := dialer.DialAndSend(message)
    if err != nil {
        print(err.Error())
    }

    return err
}

func sendEmailToAdminDelete(name string) error {
    // Set up SMTP connection parameters
    smtpHost := "smtp.gmail.com"
    smtpPort := 465
    smtpUsername := os.Getenv("SMTP_EMAIL")
    smtpPassword := os.Getenv("SMTP_PASSWORD")

    // Set up email message
    subject := "New User Registration"
    body := "Hi Admin, <br><br> A new user has been Deleted with the name " + name + ". <br><br> Thanks, <br> Team"
    from :=os.Getenv("SMTP_EMAIL")
    to := []string{"avijitsen.me@gmail.com"}

    // Create email message
    message := gomail.NewMessage()
    message.SetHeader("From", from)
    message.SetHeader("To", to...)
    message.SetHeader("Subject", subject)
    message.SetBody("text/html", body)

    // Set up SMTP authentication and send email
    dialer := gomail.NewDialer(smtpHost, smtpPort, smtpUsername, smtpPassword)
    dialer.TLSConfig = &tls.Config{InsecureSkipVerify: true}
    dialer.Auth = smtp.PlainAuth("", smtpUsername, smtpPassword, smtpHost)
    err := dialer.DialAndSend(message)
    if err != nil {
        print(err.Error())
    }

    return err
}

func GetAUser() gin.HandlerFunc {
    return func(c *gin.Context) {
        ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
        userId := c.Param("userId")
        var user models.User
        defer cancel()

        objId, _ := primitive.ObjectIDFromHex(userId)

        err := userCollection.FindOne(ctx, bson.M{"id": objId}).Decode(&user)
        if err != nil {
            c.JSON(http.StatusInternalServerError, responses.UserResponse{Status: http.StatusInternalServerError, Message: "error", Data: map[string]interface{}{"data": err.Error()}})
            return
        }

        c.JSON(http.StatusOK, responses.UserResponse{Status: http.StatusOK, Message: "success", Data: map[string]interface{}{"data": user}})
    }
}

func EditAUser() gin.HandlerFunc {
    return func(c *gin.Context) {
        ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
        userId := c.Param("userId")
        var user models.User
        defer cancel()

        objId, _ := primitive.ObjectIDFromHex(userId)

        //validate the request body
        if err := c.BindJSON(&user); err != nil {
            c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"data": err.Error()}})
            return
        }

        //use the validator library to validate required fields
        if validationErr := validate.Struct(&user); validationErr != nil {
            c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"data": validationErr.Error()}})
            return
        }

        update := bson.M{"name": user.Name, "location": user.Location, "title": user.Title}
        result, err := userCollection.UpdateOne(ctx, bson.M{"id": objId}, bson.M{"$set": update})

        if err != nil {
            c.JSON(http.StatusInternalServerError, responses.UserResponse{Status: http.StatusInternalServerError, Message: "error", Data: map[string]interface{}{"data": err.Error()}})
            return
        }

        //get updated user details
        var updatedUser models.User
        if result.MatchedCount == 1 {
            err := userCollection.FindOne(ctx, bson.M{"id": objId}).Decode(&updatedUser)
            if err != nil {
                c.JSON(http.StatusInternalServerError, responses.UserResponse{Status: http.StatusInternalServerError, Message: "error", Data: map[string]interface{}{"data": err.Error()}})
                return
            }
        }

        c.JSON(http.StatusOK, responses.UserResponse{Status: http.StatusOK, Message: "success", Data: map[string]interface{}{"data": updatedUser}})
    }
}

func DeleteAUser() gin.HandlerFunc {
    return func(c *gin.Context) {
        ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
        userId := c.Param("userId")
        defer cancel()

        objId, _ := primitive.ObjectIDFromHex(userId)

        // Get the user before deleting
        var user models.User
        err := userCollection.FindOne(ctx, bson.M{"id": objId}).Decode(&user)

        // Delete the user from the database
        result, err := userCollection.DeleteOne(ctx, bson.M{"id": objId})

        if err != nil {
            c.JSON(http.StatusInternalServerError, responses.UserResponse{Status: http.StatusInternalServerError, Message: "error", Data: map[string]interface{}{"data": err.Error()}})
            return
        }

        if result.DeletedCount < 1 {
            c.JSON(http.StatusNotFound,
                responses.UserResponse{Status: http.StatusNotFound, Message: "error", Data: map[string]interface{}{"data": "User with specified ID not found!"}},
            )
            return
        }

        // Send email to admin
        err = sendEmailToAdminDelete(user.Name)

        if err != nil {
            c.JSON(http.StatusInternalServerError, responses.UserResponse{Status: http.StatusInternalServerError, Message: "error", Data: map[string]interface{}{"data": "Error sending email to admin!"}})
            return
        }

        c.JSON(http.StatusOK,
            responses.UserResponse{Status: http.StatusOK, Message: "success", Data: map[string]interface{}{"data": "User successfully deleted!"}},
        )
    }
}


func GetAllUsers() gin.HandlerFunc {
    return func(c *gin.Context) {
        ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
        var users []models.User
        defer cancel()

        results, err := userCollection.Find(ctx, bson.M{})

        if err != nil {
            c.JSON(http.StatusInternalServerError, responses.UserResponse{Status: http.StatusInternalServerError, Message: "error", Data: map[string]interface{}{"data": err.Error()}})
            return
        }

        
        defer results.Close(ctx)
        for results.Next(ctx) {
            var singleUser models.User
            if err = results.Decode(&singleUser); err != nil {
                c.JSON(http.StatusInternalServerError, responses.UserResponse{Status: http.StatusInternalServerError, Message: "error", Data: map[string]interface{}{"data": err.Error()}})
            }

            users = append(users, singleUser)
        }

        c.JSON(http.StatusOK,
            responses.UserResponse{Status: http.StatusOK, Message: "success", Data: map[string]interface{}{"data": users}},
        )
    }
}

// const Emailtemplate = `<!DOCTYPE html>
// <html lang="en">
// <head>
//     <meta charset="UTF-8">
//     <title>New User Registration</title>
//     <style type="text/css">
//         /* CSS styles */
//         body {
//             font-family: Arial, sans-serif;
//             font-size: 16px;
//             line-height: 1.4;
//             background-color: #f4f4f4;
//         }
//         .container {
//             max-width: 600px;
//             margin: 0 auto;
//             padding: 20px;
//             background-color: #fff;
//             border-radius: 5px;
//             box-shadow: 0 2px 5px rgba(0,0,0,0.1);
//         }
//         h1 {
//             font-size: 24px;
//             margin-top: 0;
//             color: #333;
//         }
//         p {
//             margin-bottom: 20px;
//             color: #666;
//         }
//     </style>
// </head>
// <body>
//     <div class="container">
//         <h1>New User Registration</h1>
//         <p>A new user with name {{.Name}} has registered on your website.</p>
//     </div>
// </body>
// </html>
// `

