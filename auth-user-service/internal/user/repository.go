package user

import (
	"database/sql"
	"time"
)

type Repository interface {
	GetProfile(userID int) (*Profile, error)
	UpdateProfile(userID int, profile *Profile) error
	CreateProfile(userID int, email string) error
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
	FirstName string    `json:"first_name,omitempty"`
	LastName  string    `json:"last_name,omitempty"`
	Phone     string    `json:"phone,omitempty"`
	Address   string    `json:"address,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (r *repository) GetProfile(userID int) (*Profile, error) {
	var profile Profile
	err := r.db.QueryRow(
		`SELECT u.id, u.email, p.first_name, p.last_name, p.phone, p.address, u.created_at, p.updated_at 
		 FROM users u 
		 LEFT JOIN user_profiles p ON u.id = p.id 
		 WHERE u.id = $1`,
		userID,
	).Scan(
		&profile.ID, &profile.Email, &profile.FirstName, &profile.LastName,
		&profile.Phone, &profile.Address, &profile.CreatedAt, &profile.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	return &profile, err
}

func (r *repository) UpdateProfile(userID int, profile *Profile) error {
	// Сначала проверяем, существует ли профиль
	var exists bool
	err := r.db.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM user_profiles WHERE id = $1)",
		userID,
	).Scan(&exists)

	if err != nil {
		return err
	}

	if exists {
		// Обновляем существующий профиль
		_, err = r.db.Exec(
			`UPDATE user_profiles 
			 SET first_name = $1, last_name = $2, phone = $3, address = $4, updated_at = NOW()
			 WHERE id = $5`,
			profile.FirstName, profile.LastName, profile.Phone, profile.Address, userID,
		)
	} else {
		// Создаем новый профиль
		_, err = r.db.Exec(
			`INSERT INTO user_profiles (id, first_name, last_name, phone, address) 
			 VALUES ($1, $2, $3, $4, $5)`,
			userID, profile.FirstName, profile.LastName, profile.Phone, profile.Address,
		)
	}

	return err
}

func (r *repository) CreateProfile(userID int, email string) error {
	_, err := r.db.Exec(
		`INSERT INTO user_profiles (id) VALUES ($1)`,
		userID,
	)
	return err
}
