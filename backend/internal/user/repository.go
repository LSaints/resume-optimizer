package user

import (
	"database/sql"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

func (r *UserRepository) GetUsers() ([]User, error) {
	query := `
		SELECT id, name, email, password
		FROM users
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User

	for rows.Next() {
		var user User

		err := rows.Scan(
			&user.ID,
			&user.Name,
			&user.Email,
			&user.Password,
		)

		if err != nil {
			return nil, err
		}

		users = append(users, user)
	}

	return users, nil
}

func (r *UserRepository) GetUserById(
	id string,
) (User, error) {

	query := `
		SELECT id, name, email, password
		FROM users
		WHERE id = ?
	`

	var user User

	err := r.db.QueryRow(query, id).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Password,
	)

	if err != nil {
		return User{}, err
	}

	return user, nil
}

func (r *UserRepository) GetByEmail(
	email string,
) (User, error) {

	query := `
		SELECT id, name, email, password
		FROM users
		WHERE email = ?
	`

	var user User

	err := r.db.QueryRow(query, email).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Password,
	)

	if err != nil {
		return User{}, err
	}

	return user, nil
}

func (r *UserRepository) CreateUser(user User) error {
	query := `
		INSERT INTO users (
			id,
			name,
			email,
			password
		)
		VALUES (?, ?, ?, ?)
	`

	_, err := r.db.Exec(
		query,
		user.ID,
		user.Name,
		user.Email,
		user.Password,
	)

	return err
}
