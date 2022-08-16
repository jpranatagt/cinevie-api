package data

import (
  "context"
  "database/sql"
  "errors"
  "time"

  "api.cinevie.jpranata.tech/internal/validator"

  "golang.org/x/crypto/bcrypt"
)

var (
  ErrDuplicateEmail = errors.New("duplicate email.")
)

type UserModel struct {
  DB *sql.DB
}

// json "-" to prevent password  and version fields appearing in any output
// alsp password use custom type
type User struct {
  ID            int64       `json:"id"`
  CreatedAt     time.Time   `json:"created_at"`
  Name          string      `json:"name"`
  Email         string      `json:"email"`
  Password      password    `json:"-"`
  Activated     bool        `json:"activated"`
  Version       int         `json:"-"`
}

// custom struct of password containing a plain text and a hashed version
// pointers to distinguish between plain text being not present in struct
// at all versus plain text containing empty string ""
type password struct {
  plaintext *string
  hash      []byte
}

// calculates the bcrypt hash of plain text password
// also stores both the hash and the plain text version in the struct
func (p *password) Set(plaintextPassword string) error {
  hash, err := bcrypt.GenerateFromPassword([]byte(plaintextPassword), 12)
  if err != nil {
    return err
  }

  p.plaintext = &plaintextPassword
  p.hash = hash

  return nil
}

// checks whether the provided plain text password matches the hashed one
func (p *password) Matches(plaintextPassword string) (bool, error) {
  err := bcrypt.CompareHashAndPassword(p.hash, []byte(plaintextPassword))

  // password didn't match
  if err != nil {
    switch {
    case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
      return false, nil
    default:
      return false, err
    }
  }

  // password match
  return true, nil
}

func ValidateEmail(v *validator.Validator, email string) {
  v.Check(email != "", "email", "must be provided.")
  v.Check(validator.Matches(email, validator.EmailRX), "email", "must be a valid email address.")
}

func ValidatePasswordPlaintext(v *validator.Validator, password string) {
  v.Check(password != "", "password", "must be provided.")
  v.Check(len(password) >= 8, "password", "must be at least 8 characters long.")
  v.Check(len(password) <= 72, "password", "must be less than 72 characters long.")
}

func ValidateUser(v *validator.Validator, user *User) {
  v.Check(user.Name != "", "name", "must be provided.")
  v.Check(len(user.Name) <= 500, "name", "must be less than 500 characters long.")

  // call standalone ValidateEmail() helper
  ValidateEmail(v, user.Email)

  // check first if plain text password is not nill
  // then call the standalone ValidatePasswordPlaintext
  if user.Password.plaintext != nil {
    ValidatePasswordPlaintext(v, *user.Password.plaintext)
  }

  // if hashed password is nil then something wrong with codebase
  if user.Password.hash == nil {
    panic("missing password hash for user.")
  }
}

func (m UserModel) Insert(user *User) error {
  query := `
    INSERT INTO users (name, email, password_hash, activated)
    VALUES ($1, $2, $3, $4)
    RETURNING id, created_at, version
  `

  args := []interface{}{
    user.Name,
    user.Email,
    user.Password.hash,
    user.Activated,
  }

  ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
  defer cancel()

  // return error for duplicated email condition
  err := m.DB.QueryRowContext(ctx, query, args...).Scan(&user.ID, &user.CreatedAt, &user.Version)
  if err != nil {
    switch {
      case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
        return ErrDuplicateEmail
      default:
        return err
    }
  }

  return nil
}

// retrieve the user details based on user's email
func (m UserModel) GetByEmail(email string) (*User, error) {
  query := `
    SELECT id, created_at, name, email, password_hash, activated, version
    FROM users
    WHERE email = $1
  `

  var user User

  ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
  defer cancel()

  err := m.DB.QueryRowContext(ctx, query, email).Scan(
    &user.ID,
    &user.CreatedAt,
    &user.Name,
    &user.Email,
    &user.Password.hash,
    &user.Activated,
    &user.Version,
  )

  if err != nil {
    switch {
      case errors.Is(err, sql.ErrNoRows):
        return nil, ErrRecordNotFound
      default:
        return nil, err
    }
  }

  return &user, nil
}

// update details for specific user
// check against version and email violation
func (m UserModel) Update(user *User) error {
  query := `
    UPDATE users
    SET name = $1, email = $2, password_hash = $3, activated = $4, version = version + 1
    WHERE id = $5 and version = $6
    RETURNING version
  `

  args := []interface{}{
    user.Name,
    user.Email,
    user.Password.hash,
    user.Activated,
    user.ID,
    user.Version,
  }

  ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
  defer cancel()

  err := m.DB.QueryRowContext(ctx, query, args...).Scan(&user.Version)
  if err != nil {

    switch {
      case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
        return ErrDuplicateEmail
      case errors.Is(err, sql.ErrNoRows):
        return ErrEditConflict
      default:
        return err
    }
  }

  return nil
}
