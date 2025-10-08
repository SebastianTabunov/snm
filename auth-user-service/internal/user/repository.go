package user

import (
	"database/sql"
	"time"
)

type Repository interface {
	GetProfile(userID int) (*Profile, error)
	UpdateProfile(userID int, profile *Profile) error
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

type Profile struct {
	ID        int       `json:"id"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Phone     string    `json:"phone"`
	Address   string    `json:"address"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (r *repository) GetProfile(userID int) (*Profile, error) {
	var profile Profile
	err := r.db.QueryRow(
		`SELECT id, email, first_name, last_name, phone, address, created_at, updated_at 
		 FROM user_profiles 
		 WHERE id = $1`,
		userID,
	).Scan(
		&profile.ID, &profile.Email, &profile.FirstName, &profile.LastName,
		&profile.Phone, &profile.Address, &profile.CreatedAt, &profile.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &profile, nil
}

func (r *repository) UpdateProfile(userID int, profile *Profile) error {
	_, err := r.db.Exec(
		`UPDATE user_profiles 
		 SET first_name = $1, last_name = $2, phone = $3, address = $4, updated_at = NOW()
		 WHERE id = $5`,
		profile.FirstName, profile.LastName, profile.Phone, profile.Address, userID,
	)
	return err
}
