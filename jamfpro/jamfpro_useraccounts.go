package jamfpro

import (
	"context"
	"net/http"
)

const accountsBasePath = "JSSResource/accounts"

type UserAccountsService interface {
	List(context.Context) ([]UserAccount, *Response, error)
	GetByID(context.Context, int) (*UserAccount, *Response, error)
	GetByName(context.Context, string) (*UserAccount, *Response, error)
	Create(context.Context, *UserAccountRequest) (*UserAccount, *Response, error)
	Update(context.Context, int, *UserAccountRequest) (*UserAccount, *Response, error)
	Delete(context.Context, int) (*Response, error)
}

type UserAccountsServiceOp struct {
	client *Client
}

var _ UserAccountsService = &UserAccountsServiceOp{}

type UserAccount struct {
	Id                  int              `xml:"id"`
	Name                string           `xml:"name"`
	IsDirectoryUser     bool             `xml:"directory_user"`
	FullName            string           `xml:"full_name"`
	Email               string           `xml:"email"`
	EmailAddress        string           `xml:"email_address"`
	PasswordSha256      string           `xml:"password_sha256"`
	Enabled             string           `xml:"enabled"`
	ForcePasswordChange bool             `xml:"force_password_change"`
	AccessLevel         string           `xml:"access_level"`
	PrivilegeSet        string           `xml:"privilege_set"`
	Privileges          PrivilegesObject `xml:"privileges"`
}

type PrivilegesObject struct {
	JssObjects  []Privilege `xml:"jss_objects"`
	JssSettings []Privilege `xml:"jss_settings"`
	JssActions  []Privilege `xml:"jss_actions"`
	CasperAdmin []Privilege `xml:"casper_admin"`
}

type Privilege struct {
	Privilege string `xml:"privilege"`
}

type UserAccountRequest struct {
	Name                string           `xml:"name"`
	IsDirectoryUser     bool             `xml:"directory_user"`
	FullName            string           `xml:"full_name"`
	Email               string           `xml:"email"`
	EmailAddress        string           `xml:"email_address"`
	Password            string           `xml:"password"`
	Enabled             string           `xml:"enabled"`
	ForcePasswordChange bool             `xml:"force_password_change"`
	AccessLevel         string           `xml:"access_level"`
	PrivilegeSet        string           `xml:"privilege_set"`
	Privileges          PrivilegesObject `xml:"privileges"`
}

type UserAccountListResponse struct {
	Accounts AccountsObject `xml:"accounts"`
}

type AccountsObject struct {
	Users []UserAccount `xml:"users"`
}

func (u *UserAccountsServiceOp) List(ctx context.Context) ([]UserAccount, *Response, error) {
	//TODO implement me
	panic("implement me")
}

func (u *UserAccountsServiceOp) GetByID(ctx context.Context, i int) (*UserAccount, *Response, error) {
	//TODO implement me
	panic("implement me")
}

func (u *UserAccountsServiceOp) GetByName(ctx context.Context, s string) (*UserAccount, *Response, error) {
	//TODO implement me
	panic("implement me")
}

func (u *UserAccountsServiceOp) Create(ctx context.Context, request *UserAccountRequest) (*UserAccount, *Response, error) {
	//TODO implement me
	panic("implement me")
}

func (u *UserAccountsServiceOp) Update(ctx context.Context, i int, request *UserAccountRequest) (*UserAccount, *Response, error) {
	//TODO implement me
	panic("implement me")
}

func (u *UserAccountsServiceOp) Delete(ctx context.Context, i int) (*Response, error) {
	//TODO implement me
	panic("implement me")
}

func (u *UserAccountsServiceOp) list(ctx context.Context) ([]UserAccount, *Response, error) {
	path := accountsBasePath

	req, err := u.client.NewRequest(ctx, http.MethodGet, path, nil, "application/xml")
	if err != nil {
		return nil, nil, err
	}

	var userResponse UserAccountListResponse
	resp, err := u.client.Do(ctx, req, &userResponse)
	if err != nil {
		return nil, resp, err
	}

	return userResponse.Accounts.Users, resp, err
}
