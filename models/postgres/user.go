package postgres

import (
	"database/sql/driver"
	"time"
)

// UserRole represents the role of a user
type UserRole string

const (
	UserRoleUsuario UserRole = "usuario"
	UserRoleAdmin   UserRole = "admin"
)

// UserStatus represents the status of a user
type UserStatus string

const (
	UserStatusActivo   UserStatus = "activo"
	UserStatusInactivo UserStatus = "inactivo"
	UserStatusBaneado  UserStatus = "baneado"
)

// Value implements the driver.Valuer interface for UserRole
func (ur UserRole) Value() (driver.Value, error) {
	return string(ur), nil
}

// Value implements the driver.Valuer interface for UserStatus
func (us UserStatus) Value() (driver.Value, error) {
	return string(us), nil
}

/*
 * 'User' contains the blueprint definition of a User. It contains a reference to GameProfile
 */
type User struct {
	Email                  string     `gorm:"primaryKey;size:255;not null;index:idx_users_email"`
	ProfileUsername        string     `gorm:"column:username;size:50;not null;uniqueIndex:idx_users_profile_username"`
	PasswordHash           string     `gorm:"size:255;not null"`
	Role                   UserRole   `gorm:"type:varchar(20);default:'usuario'"`
	Status                 UserStatus `gorm:"type:varchar(20);default:'activo'"`
	AvatarURL              *string    `gorm:"column:avatar_url;type:text"`
	Bio                    *string    `gorm:"type:text"`
	BirthDate              *time.Time `gorm:"column:birth_date;type:date"`
	Country                *string    `gorm:"size:100"`
	EmailVerified          bool       `gorm:"default:false"`
	EmailVerificationToken *string    `gorm:"size:255"`
	PasswordResetToken     *string    `gorm:"size:255"`
	PasswordResetExpires   *time.Time `gorm:"column:password_reset_expires"`
	LastLogin              *time.Time `gorm:"column:last_login"`
	CreatedAt              time.Time  `gorm:"column:created_at;default:CURRENT_TIMESTAMP"`
	UpdatedAt              time.Time  `gorm:"column:updated_at;default:CURRENT_TIMESTAMP"`
}
