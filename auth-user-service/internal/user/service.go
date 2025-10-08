package user

type Service interface {
	GetProfile(userID int) (*Profile, error)
	UpdateProfile(userID int, profile *Profile) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) GetProfile(userID int) (*Profile, error) {
	return s.repo.GetProfile(userID)
}

func (s *service) UpdateProfile(userID int, profile *Profile) error {
	return s.repo.UpdateProfile(userID, profile)
}
