package entity

type Team struct {
	Name    string    `json:"team_name" validate:"required,min=1,max=128"`
	Members []*Member `json:"members" validate:"required,dive"`
}

type Member struct {
	ID       string `json:"user_id" validate:"required,min=1,max=64"`
	Username string `json:"username" validate:"required,min=1,max=128"`
	IsActive bool   `json:"is_active" validate:"required"`
}
