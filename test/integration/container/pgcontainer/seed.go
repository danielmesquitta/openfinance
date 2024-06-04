package pgcontainer

import (
	"time"

	"github.com/danielmesquitta/openfinance/internal/domain/entity"
)

type Seed string

const (
	SeedTestUserWithoutSetting Seed = "new_test_user_without_setting"
	SeedTestUserWithSetting    Seed = "new_test_user_with_setting"
)

var TestUserWithoutSetting entity.User = entity.User{
	ID:        "018fe49d-3b5d-77a1-a3e6-82eaecdff193",
	Email:     "testuser@danielmesquitta.com",
	UpdatedAt: time.Now(),
}

var TestUserWithSetting entity.User = entity.User{
	ID:        "018fe4b3-a1ea-77ca-9ecb-cac5b64756b5",
	Email:     "testuser@danielmesquitta.com",
	UpdatedAt: time.Now(),
}
