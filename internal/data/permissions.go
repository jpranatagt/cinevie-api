package data

import (
  "context"
  "database/sql"
  "time"
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
    INNDER JOIN users ON users_permissions.user_id = users.id
    WHERE users.id = $1
  `

  ctx, cancel := context.WithTimeout(context.Background(), 3 * time.Second)
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

    permission = append(permissions, permission)
  }

  if err = rows.Err(); err != nil {
    return nil, err
  }

  return permissions, nil
}
