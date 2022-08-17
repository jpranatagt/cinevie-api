package data

import (
  "context"
  "crypto/rand"
  "crypto/sha256"
  "encoding/base32"
  "database/sql"
  "time"

  "api.cinevie.jpranata.tech/internal/validator"
)

const (
  ScopeActivation = "activation"
)

type Token struct{
  Plaintext string
  Hash      []byte
  UserID    int64
  Expiry    time.Time
  Scope     string
}

func generateToken(userID int64, ttl time.Duration, scope string) (*Token, error) {
  // ttl or time to live to get the expiry time
  token := &Token {
    UserID: userID,
    Expiry: time.Now().Add(ttl),
    Scope: scope,
  }

  randomBytes := make([]byte, 16)

  // fill the byte slice with random bytes using Read() and OS CSPRNG
  _, err := rand.Read(randomBytes)
  if err != nil {
    return nil, err
  }

  // encode the byte slice to a base-32-encoded string
  token.Plaintext = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomBytes)

  // generate a SHA-256 hash of the plain text token string
  hash := sha256.Sum256([]byte(token.Plaintext))
  token.Hash = hash[:] // convert to a slice using the [:] operator before storing it

  return token, nil
}

// is the plain text token that has been provided and is exactly 52 bytes long
func ValidateTokenPlaintext(v *validator.Validator, tokenPlaintext string) {
  v.Check(tokenPlaintext != "", "token", "must be provided.")
  v.Check(len(tokenPlaintext) == 26, "token", "must be 26 bytes long.")
}

// TokenModel type
type TokenModel struct {
  DB *sql.DB
}

// create a new Token struct and then inserts the data in the tokens table
func (m TokenModel) New(userID int64, ttl time.Duration, scope string) (*Token, error) {
  token, err := generateToken(userID, ttl, scope)
  if err != nil {
    return nil, err
  }

  err = m.Insert(token)
  return token, err
}

// add the data to tokens table
func (m TokenModel) Insert(token *Token) error {
  query := `
    INSERT INTO tokens (hash, user_id, expiry, scope)
    VALUES ($1, $2, $3, $4)
  `

  args := []interface{} {
    token.Hash,
    token.UserID,
    token.Expiry,
    token.Scope,
  }

  ctx, cancel := context.WithTimeout(context.Background(), 3 * time.Second)
  defer cancel()

  _, err := m.DB.ExecContext(ctx, query, args...)

  return err
}

// delete all tokens for a specific user and scope
func (m TokenModel) DeleteAllForUser(scope string, userID int64) error {
  query := `
    DELETE FROM tokens
    WHERE scope = $1 and user_id = $2
  `

  ctx, cancel := context.WithTimeout(context.Background(), 3 * time.Second)
  defer cancel()

  _, err := m.DB.ExecContext(ctx, query, scope, userID)

  return err
}
