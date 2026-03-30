package auth

type AuthRepository struct{}

func (r *AuthRepository) FindByUsername(username string) *User {
    // dummy data
    if username == "admin" {
        return &User{
            ID:       1,
            Username: "admin",
            Password: "123456", // nanti harus hash
        }
    }
    return nil
}