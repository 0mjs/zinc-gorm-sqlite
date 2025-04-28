package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/0mjs/zinc"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type User struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Name      string         `json:"name"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

type CreateUserRequest struct {
	Name string `json:"name"`
}

type UpdateUserRequest struct {
	Name string `json:"name"`
}

type DeleteUserRequest struct {
	DeletedAt gorm.DeletedAt `json:"deleted_at"`
}

func main() {
	app := zinc.New()
	var orm *gorm.DB

	err := app.ConnectDB("sqlite", "./sqlite.db")
	if err != nil {
		log.Fatalf("Connection failed: %v", err)
	}
	db := app.GetDB()
	if db == nil {
		log.Fatalf("Failed to get *sql.DB connection from Zinc")
	}
	log.Println("Connected to SQLite")

	orm, err = gorm.Open(sqlite.New(sqlite.Config{Conn: db}), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to initialize GORM: %v", err)
	}
	log.Println("Initialized GORM")
	err = orm.AutoMigrate(&User{})
	if err != nil {
		log.Fatalf("GORM AutoMigrate failed: %v", err)
		return
	}
	log.Println("Migrated User model")

	users := app.Group("/users")
	log.Println("Created users group")

	users.Get("/", func(c *zinc.Context) error {
		var users []User
		includeDeleted := c.Query("include_deleted")

		var result *gorm.DB
		if includeDeleted == "true" {
			result = orm.Unscoped().Find(&users)
		} else {
			result = orm.Find(&users)
		}
		if result.Error != nil {
			return fmt.Errorf("failed to query users: %w", result.Error)
		}
		return c.JSON(users)
	})

	users.Post("/", func(c *zinc.Context) error {
		var req CreateUserRequest
		if err := c.Bind(&req); err != nil {
			return c.Status(http.StatusBadRequest).Send("Invalid request body")
		}

		user := User{Name: req.Name}
		result := orm.Create(&user)
		if result.Error != nil {
			return fmt.Errorf("failed to insert user: %w", result.Error)
		}

		return c.Status(http.StatusCreated).JSON(user)
	})

	users.Get("/:id", func(c *zinc.Context) error {
		id := c.Param("id")
		var user User
		includeDeleted := c.Query("include_deleted")

		result := orm.First(&user, id)
		if result.Error != nil {
			if includeDeleted == "true" {
				result = orm.Unscoped().First(&user, id)
			} else {
				return fmt.Errorf("failed to find user: %w", result.Error)
			}
		}

		return c.JSON(user)
	})

	users.Put("/:id", func(c *zinc.Context) error {
		id := c.Param("id")
		var user User

		result := orm.First(&user, id)
		if result.Error != nil {
			return fmt.Errorf("failed to find user: %w", result.Error)
		}
		if err := c.Bind(&user); err != nil {
			return c.Status(http.StatusBadRequest).Send("Invalid request body")
		}

		result = orm.Save(&user)
		if result.Error != nil {
			return fmt.Errorf("failed to update user: %w", result.Error)
		}

		return c.JSON(user)
	})

	users.Delete("/:id", func(c *zinc.Context) error {
		id := c.Param("id")

		result := orm.Model(&User{}).Where("id = ?", id).Update("deleted_at", gorm.DeletedAt{Time: time.Now(), Valid: true})
		if result.Error != nil {
			return fmt.Errorf("failed to delete user: %w", result.Error)
		}

		return c.Status(http.StatusNoContent).Send("User deleted")
	})

	app.Serve(":8080")
}
