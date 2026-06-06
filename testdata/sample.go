package sample

import (
    "fmt"
    "strings"
)

type User struct {
    Name  string
    Age   int
    Email string
}

type Repository interface {
    FindByID(id int) (*User, error)
    Save(u *User) error
}

type UserService struct {
    repo Repository
}

func NewUserService(repo Repository) *UserService {
    return &UserService{repo: repo}
}

func (s *UserService) GetUser(id int) (*User, error) {
    return s.repo.FindByID(id)
}

func FormatUser(u *User) string {
    return fmt.Sprintf("%s <%s>", strings.ToUpper(u.Name), u.Email)
}

func ProcessUsers(users []*User, fn func(*User)) {
    for _, u := range users {
        fn(u)
    }
}