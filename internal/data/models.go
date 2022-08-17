package data

import (
  "database/sql"
  "errors"
)

// define ErrRecordNotFound error return this from Get()
// while looking up movie that doesn't exist
var (
  ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict = errors.New("edit conflict")
)

// create models which wrap MovieModel
type Models struct {
	Permissions PermissionModel
  Movies MovieModel
	Users UserModel
	Tokens TokenModel
}

// return the initialized MovieModel
func NewModels(db *sql.DB) Models {
  return Models{
		Permissions: PermissionModel{DB: db},
    Movies: MovieModel{DB: db},
		Users: UserModel{DB: db},
		Tokens: TokenModel{DB: db},
  }
}
