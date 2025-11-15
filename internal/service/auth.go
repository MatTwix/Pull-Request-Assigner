package service

type AuthService struct {
	UserAPIKey  string
	AdminAPIKey string
}

func NewAuthService(userAPIKey, adminAPIKey string) *AuthService {
	return &AuthService{
		UserAPIKey:  userAPIKey,
		AdminAPIKey: adminAPIKey,
	}
}

func (s *AuthService) IsUserKey(key string) bool {
	return key == s.UserAPIKey
}

func (s *AuthService) IsAdminKey(key string) bool {
	return key == s.AdminAPIKey
}
