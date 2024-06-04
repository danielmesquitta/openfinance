package pgrepo

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/danielmesquitta/openfinance/internal/domain/entity"
	"github.com/danielmesquitta/openfinance/internal/provider/db/pgdb"
	"github.com/danielmesquitta/openfinance/internal/provider/repo"
	"github.com/google/uuid"
	"github.com/jinzhu/copier"
)

type UserPgRepo struct {
	db *pgdb.Queries
}

func NewUserPgRepo(db *pgdb.Queries) *UserPgRepo {
	return &UserPgRepo{
		db: db,
	}
}

func (b UserPgRepo) GetFullUserByID(id string) (entity.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	dbUser, err := b.db.GetFullUserByID(ctx, id)
	if err == sql.ErrNoRows {
		return entity.User{}, nil
	}

	if err != nil {
		return entity.User{}, fmt.Errorf(
			"error getting user with setting by id: %w",
			err,
		)
	}

	user := entity.User{
		ID:        dbUser.ID,
		Email:     dbUser.Email,
		UpdatedAt: dbUser.UpdatedAt,
		Setting: &entity.Setting{
			ID:                    dbUser.ID_2.String,
			NotionToken:           dbUser.NotionToken.String,
			NotionPageID:          dbUser.NotionPageID.String,
			MeuPluggyClientID:     dbUser.MeuPluggyClientID.String,
			MeuPluggyClientSecret: dbUser.MeuPluggyClientSecret.String,
			MeuPluggyAccountIDs:   dbUser.MeuPluggyAccountIds,
			UserID:                dbUser.UserID.String,
			UpdatedAt:             dbUser.UpdatedAt_2.Time,
		},
	}

	return user, nil
}

func (b UserPgRepo) GetUserByEmail(email string) (entity.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	dbUser, err := b.db.GetUserByEmail(ctx, email)
	if err == sql.ErrNoRows {
		return entity.User{}, nil
	}

	if err != nil {
		return entity.User{}, fmt.Errorf("error getting user by email: %w", err)
	}

	var user entity.User
	err = copier.Copy(&user, dbUser)
	if err != nil {
		return entity.User{}, fmt.Errorf("error copying user: %w", err)
	}

	return user, nil
}

func (b UserPgRepo) CreateUser(dto repo.CreateUserDTO) (entity.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	arg := pgdb.CreateUserParams{}
	if err := copier.Copy(&arg, dto); err != nil {
		return entity.User{}, fmt.Errorf("error copying dto to db arg: %w", err)
	}

	id, err := uuid.NewV7()
	if err != nil {
		return entity.User{}, fmt.Errorf("error generating uuid: %w", err)
	}

	arg.ID = id.String()
	arg.UpdatedAt = time.Now()

	result, err := b.db.CreateUser(ctx, arg)
	if err != nil {
		return entity.User{}, fmt.Errorf("error creating user: %w", err)
	}

	var user entity.User
	err = copier.Copy(&user, result)
	if err != nil {
		return entity.User{}, fmt.Errorf("error copying user: %w", err)
	}

	return entity.User{}, nil
}
