package pgrepo

import (
	"context"
	"database/sql"
	"time"

	"github.com/danielmesquitta/openfinance/internal/domain/entity"
	"github.com/danielmesquitta/openfinance/internal/provider/db/pgdb"
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

func (b UserPgRepo) GetUserByID(id string) (entity.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	dbUser, err := b.db.GetUserByID(ctx, id)
	if err == sql.ErrNoRows {
		return entity.User{}, nil
	}

	if err != nil {
		return entity.User{}, err
	}

	var user entity.User
	err = copier.Copy(&user, dbUser)
	if err != nil {
		return entity.User{}, err
	}

	return user, nil
}

func (b UserPgRepo) GetUserWithSettingByID(id string) (entity.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	dbUser, err := b.db.GetUserWithSettingByID(ctx, id)
	if err == sql.ErrNoRows {
		return entity.User{}, nil
	}

	if err != nil {
		return entity.User{}, err
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
		return entity.User{}, err
	}

	var user entity.User
	err = copier.Copy(&user, dbUser)
	if err != nil {
		return entity.User{}, err
	}

	return user, nil
}

func (b UserPgRepo) CreateUser(user *entity.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	params := pgdb.CreateUserParams{}

	err := copier.Copy(&params, user)
	if err != nil {
		return err
	}

	id, err := uuid.NewV7()
	if err != nil {
		return err
	}

	params.ID = id.String()
	params.UpdatedAt = time.Now()

	err = b.db.CreateUser(ctx, params)
	if err != nil {
		return err
	}

	user.ID = params.ID

	return nil
}
