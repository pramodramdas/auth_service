package models

type TokenInterface interface {
	CreateTokenTable() (bool, error)
	InsertUsedToken(tokenString string) (bool, error)
	IsTokenUsed(tokenString string) (bool, error)
}

type TokenMembers struct {
	TokenInterface
}
