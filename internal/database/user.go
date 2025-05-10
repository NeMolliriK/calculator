package database

type User struct {
	ID       uint   `gorm:"primaryKey"`
	Login    string `gorm:"not null;unique"`
	Password string `gorm:"not null"        json:"-"`
}

func (u *User) SetPassword(hash string) {
	u.Password = hash
}

func (u *User) GetPassword() string {
	return u.Password
}

func CreateUser(expr *User) error {
	return DB.Create(expr).Error
}

func GetUserByLogin(login string) (*User, error) {
	var expr User
	err := DB.First(&expr, "login = ?", login).Error
	if err != nil {
		return nil, err
	}
	return &expr, nil
}

func GetUserByID(id uint) (*User, error) {
	var expr User
	err := DB.First(&expr, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &expr, nil
}
