package example

type CreateExample struct {
	Message string `json:"message" validate:"required,min=3,max=255"`
}

type UpdateExample struct {
	Message *string `json:"message" validate:"omitempty,min=3,max=255"`
}