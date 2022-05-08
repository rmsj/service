package db

import "time"

// Product represents an individual product.
type Product struct {
	ProductID   string    `gorm:"product_id;primaryKey"` // Unique identifier.
	Name        string    `gorm:"name"`                  // Display name of the product.
	Cost        int       `gorm:"cost"`                  // Price for one item in cents.
	Quantity    int       `gorm:"quantity"`              // Original number of items available.
	Sold        int       `gorm:"-"`                     // Aggregate field showing number of items sold.
	Revenue     int       `gorm:"-"`                     // Aggregate field showing total cost of sold items.
	UserID      string    `gorm:"user_id"`               // ID of the user who created the product.
	DateCreated time.Time `gorm:"date_created"`          // When the product was added.
	DateUpdated time.Time `gorm:"date_updated"`          // When the product record was last modified.
}
