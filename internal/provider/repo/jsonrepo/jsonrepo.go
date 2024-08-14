package jsonrepo

import (
	"encoding/json"
	"io"
	"os"
	"strings"

	"github.com/danielmesquitta/openfinance/internal/domain/entity"
	"github.com/danielmesquitta/openfinance/pkg/crypto"
)

var users []entity.User

func getDefaultUsers(cr crypto.Encrypter) []entity.User {
	if len(users) != 0 {
		return users
	}

	file, err := os.Open("users.json")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		panic(err)
	}

	if err = json.Unmarshal(data, &users); err != nil {
		panic(err)
	}

	for i, u := range users {
		if u.Setting == nil {
			panic("Setting is nil for user " + u.ID)
		}

		meuPluggyAccountIDs := u.Setting.MeuPluggyAccountIDs
		for i, v := range u.Setting.MeuPluggyAccountIDs {
			val, err := cr.Encrypt(strings.TrimSpace(v))
			if err != nil {
				panic(err)
			}
			meuPluggyAccountIDs[i] = val
		}

		notionToken, err := cr.Encrypt(u.Setting.NotionToken)
		if err != nil {
			panic(err)
		}

		notionPageID, err := cr.Encrypt(u.Setting.NotionPageID)
		if err != nil {
			panic(err)
		}

		meuPluggyClientID, err := cr.Encrypt(u.Setting.MeuPluggyClientID)
		if err != nil {
			panic(err)
		}

		meuPluggyClientSecret, err := cr.Encrypt(
			u.Setting.MeuPluggyClientSecret,
		)
		if err != nil {
			panic(err)
		}

		u.Setting = &entity.Setting{
			ID:                    u.Setting.ID,
			UserID:                u.ID,
			MeuPluggyAccountIDs:   meuPluggyAccountIDs,
			NotionToken:           notionToken,
			NotionPageID:          notionPageID,
			MeuPluggyClientID:     meuPluggyClientID,
			MeuPluggyClientSecret: meuPluggyClientSecret,
		}

		users[i] = u
	}

	return users
}
