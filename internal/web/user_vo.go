package web

type LoginJwtReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type SignUpReq struct {
	Email           string `json:"email"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirmPassword"`
}

type EditReq struct {
	NickName string `json:"nickname"`
	BirthDay string `json:"birthday"`
	Brief    string `json:"aboutMe"`
}

type SendSMSReq struct {
	Phone string `json:"phone"`
}

type LoginSMSReq struct {
	Phone string `json:"phone"`
	Code  string `json:"code"`
}
