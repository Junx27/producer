package model

type User struct {
	ID    int64  `gorm:"primaryKey" json:"id"`
	Nama  string `gorm:"type:varchar(300)" json:"nama"`
	Email string `gorm:"type:varchar(300)" json:"email"`
}
