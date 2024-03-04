package jamfpro

import (
	"context"
	"net/http"
	"strconv"
)

const apiRolesBasePath = "uapi/v1/api-roles"

type ApiRolesService interface {
	List(context.Context) ([]ApiRole, *Response, error)
	GetByID(context.Context, int) (*ApiRole, *Response, error)
	GetByName(context.Context, string) (*ApiRole, *Response, error)
	Create(context.Context, *ApiRoleCreateRequest) (*ApiRole, *Response, error)
	Update(context.Context, int, *ApiRoleUpdateRequest) (*ApiRole, *Response, error)
	Delete(context.Context, int) (*Response, error)
}

// ApiRolesServiceOp handles communication with the API roles related
// methods of the Jamf Pro API.
type ApiRolesServiceOp struct {
	client *Client
}

var _ ApiRolesService = &ApiRolesServiceOp{}

// ApiRole represents a Jamf Pro API Role
type ApiRole struct {
	Id          *string   `json:"id,omitempty"` // The response type to be returned is a string
	DisplayName *string   `json:"displayName,omitempty"`
	Privileges  *[]string `json:"privileges,omitempty"`
}

// ApiRoleGetResponse represents the raw API response to getting all API roles
type ApiRoleGetResponse struct {
	TotalCount *int64     `json:"totalCount"`
	ApiRoles   *[]ApiRole `json:"results"`
}

// ApiRoleCreateRequest represents a request to create an API role.
type ApiRoleCreateRequest struct {
	DisplayName string   `json:"displayName,omitempty"`
	Privileges  []string `json:"privileges,omitempty"`
}

// ApiRoleCreateResponse represents an API response to creating an API role
type ApiRoleCreateResponse struct {
	Id          *string   `json:"id,omitempty"` // The response type to be returned is a string
	DisplayName *string   `json:"displayName,omitempty"`
	Privileges  *[]string `json:"privileges,omitempty"`
}

// ApiRoleUpdateRequest represents a request to update an API role.
type ApiRoleUpdateRequest struct {
	DisplayName string   `json:"displayName,omitempty"`
	Privileges  []string `json:"privileges,omitempty"`
}

// ApiRoleUpdateResponse represents an API response to updating an API role
type ApiRoleUpdateResponse struct {
	Id          *string   `json:"id,omitempty"` // The response type to be returned is a string
	DisplayName *string   `json:"displayName,omitempty"`
	Privileges  *[]string `json:"privileges,omitempty"`
}

func (a *ApiRolesServiceOp) List(ctx context.Context) ([]ApiRole, *Response, error) {
	return a.list(ctx)
}

func (a *ApiRolesServiceOp) GetByID(ctx context.Context, id int) (*ApiRole, *Response, error) {
	path := apiRolesBasePath + "/" + strconv.Itoa(id)

	req, err := a.client.NewRequest(ctx, http.MethodGet, path, nil, "application/json")
	if err != nil {
		return nil, nil, err
	}

	var apiRole ApiRole
	resp, err := a.client.Do(ctx, req, &apiRole)
	if err != nil {
		return nil, resp, err
	}

	return &apiRole, resp, err
}

func (a *ApiRolesServiceOp) GetByName(ctx context.Context, name string) (*ApiRole, *Response, error) {
	apiRoles, _, err := a.list(ctx)
	var id string
	if err != nil {
		return nil, nil, err
	}

	for i := range apiRoles {
		if *apiRoles[i].DisplayName == name {
			id = *apiRoles[i].Id
			break
		}
	}
	intId, err := strconv.ParseInt(id, 10, 64)

	if err != nil {
		return nil, nil, err
	}

	apiRole, resp, err := a.GetByID(ctx, int(intId))
	if err != nil {
		return nil, resp, err
	}

	return apiRole, resp, err
}

func (a *ApiRolesServiceOp) Create(ctx context.Context, request *ApiRoleCreateRequest) (*ApiRole, *Response, error) {
	if request == nil {
		return nil, nil, NewArgError("createRequest", "cannot be nil")
	}

	req, err := a.client.NewRequest(ctx, http.MethodPost, apiRolesBasePath, request, "application/json")
	if err != nil {
		return nil, nil, err
	}

	apiRoleCreation := new(ApiRole)
	resp, err := a.client.Do(ctx, req, &apiRoleCreation)
	if err != nil {
		return nil, resp, err
	}

	if apiRoleCreation.Id == nil {
		return nil, resp, err
	}

	return apiRoleCreation, resp, err
}

func (a *ApiRolesServiceOp) Update(ctx context.Context, id int, request *ApiRoleUpdateRequest) (*ApiRole, *Response, error) {
	path := apiRolesBasePath + "/" + strconv.Itoa(id)

	if request == nil {
		return nil, nil, NewArgError("createRequest", "cannot be nil")
	}

	req, err := a.client.NewRequest(ctx, http.MethodPut, path, request, "application/json")
	if err != nil {
		return nil, nil, err
	}

	apiRoleUpdate := new(ApiRole)
	resp, err := a.client.Do(ctx, req, apiRoleUpdate)
	if err != nil {
		return nil, resp, err
	}

	return apiRoleUpdate, resp, err
}

func (a *ApiRolesServiceOp) Delete(ctx context.Context, id int) (*Response, error) {
	path := apiRolesBasePath + "/" + strconv.Itoa(id)

	req, err := a.client.NewRequest(ctx, http.MethodDelete, path, nil, "application/json")
	if err != nil {
		return nil, err
	}

	resp, err := a.client.Do(ctx, req, nil)
	if err != nil && err.Error() != "EOF" {
		return resp, err
	}

	return resp, err
}

func (a *ApiRolesServiceOp) list(ctx context.Context) ([]ApiRole, *Response, error) {
	path := apiRolesBasePath

	req, err := a.client.NewRequest(ctx, http.MethodGet, path, nil, "application/json")
	if err != nil {
		return nil, nil, err
	}

	var apiRoleResponse ApiRoleGetResponse
	resp, err := a.client.Do(ctx, req, &apiRoleResponse)
	if err != nil {
		return nil, resp, err
	}

	return *apiRoleResponse.ApiRoles, resp, err
}
