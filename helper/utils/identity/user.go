package identity

import (
	"pulse/helper/utils/toolkit/mapx"
)

type UserClaims struct {
	*MapClaims
	TokenAlias string `json:"alias"`
	Email      string `json:"email"`
	Workplace  string `json:"workplace"`
	Position   string `json:"position"`
}

func (c *UserClaims) GetEmail() string      { return c.Email }
func (c *UserClaims) GetPosition() string   { return c.Position }
func (c *UserClaims) GetWorkplace() string  { return c.Workplace }
func (c *UserClaims) GetTokenAlias() string { return c.TokenAlias }

func UserClaimsFrom(c *MapClaims) *UserClaims {
	return &UserClaims{
		MapClaims:  c,
		TokenAlias: mapx.GetStringFromMap(c.GetAttributes(), "alias"),
		Email:      mapx.GetStringFromMap(c.GetAttributes(), "email"),
		Workplace:  mapx.GetStringFromMap(c.GetAttributes(), "workplace"),
		Position:   mapx.GetStringFromMap(c.GetAttributes(), "position"),
	}
}
