package envrepo

import (
	"strings"

	"github.com/danielmesquitta/openfinance/internal/domain/entity"
	"github.com/danielmesquitta/openfinance/pkg/crypto"
	"github.com/google/uuid"
)

var defaultSettings entity.Setting

func getDefaultSetting(env *Env, cr crypto.Encrypter) entity.Setting {
	if defaultSettings.ID != "" {
		return defaultSettings
	}

	meuPluggyAccountIDs := strings.Split(env.MeuPluggyAccountIDs, ",")
	for i, v := range meuPluggyAccountIDs {
		val, err := cr.Encrypt(strings.TrimSpace(v))
		if err != nil {
			panic(err)
		}
		meuPluggyAccountIDs[i] = val
	}

	notionToken, err := cr.Encrypt(env.NotionToken)
	if err != nil {
		panic(err)
	}

	notionPageID, err := cr.Encrypt(env.NotionPageID)
	if err != nil {
		panic(err)
	}

	meuPluggyClientID, err := cr.Encrypt(env.MeuPluggyClientID)
	if err != nil {
		panic(err)
	}

	meuPluggyClientSecret, err := cr.Encrypt(env.MeuPluggyClientSecret)
	if err != nil {
		panic(err)
	}

	defaultSettings = entity.Setting{
		UserID:                uuid.NewString(),
		ID:                    uuid.NewString(),
		NotionToken:           notionToken,
		NotionPageID:          notionPageID,
		MeuPluggyClientID:     meuPluggyClientID,
		MeuPluggyClientSecret: meuPluggyClientSecret,
		MeuPluggyAccountIDs:   meuPluggyAccountIDs,
	}

	return defaultSettings
}
