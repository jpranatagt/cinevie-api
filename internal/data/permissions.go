package data

import (
	"context"
	"database/sql"
	"time"

	"github.com/lib/pq"
)

// Permissions slice which would hold the permission codes
// like movie:read and movie:write
type Permissions []string

// check whether the Permissions slice contains a specific
// permission code
func (p Permissions) Include(code string) bool {
	for i := range p {
		if code == p[i] {
			return true
		}
	}

	return false
}

// PermissionModel type
type PermissionModel struct {
	DB *sql.DB
}

// returns the entire permission codes for a specific user in
// Permissions slice
func (m PermissionModel) GetAllForUser(userID int64) (Permissions, error) {
	query := `
    SELECT permissions.code
    FROM permissions
    INNER JOIN users_permissions on users_permissions.permission_id = permissions.id
    INNER JOIN users ON users_permissions.user_id = users.id
    WHERE users.id = $1
  `

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var permissions Permissions

	for rows.Next() {
		var permission string

		err := rows.Scan(&permission)
		if err != nil {
			return nil, err
		}

		permissions = append(permissions, permission)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return permissions, nil
}

// use variadic parameter to assign multiple permission codes for specific user
// in single call
func (m PermissionModel) AddForUser(userID int64, codes ...string) error {
	query := `
    INSERT INTO users_permissions
    SELECT $1, permissions.id FROM permissions WHERE permissions.code = ANY($2)
  `

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, userID, pq.Array(codes))

	return err
}
