package jamfpro

import (
	"context"
	"net/http"
	"strconv"
)

const departmentsBasePath = "uapi/v1/departments"

type DepartmentsService interface {
	List(context.Context) ([]Department, *Response, error)
	GetByID(context.Context, int) (*Department, *Response, error)
	GetByName(context.Context, string) (*Department, *Response, error)
	Create(context.Context, *DepartmentCreateRequest) (*Department, *Response, error)
	Update(context.Context, int, *DepartmentUpdateRequest) (*Department, *Response, error)
	Delete(context.Context, int) (*Response, error)
}

// DepartmentsServiceOp handles communication with the categories-related
// methods of the Jamf Pro API.
type DepartmentsServiceOp struct {
	client *Client
}

var _ DepartmentsService = &DepartmentsServiceOp{}

// Department represents a Jamf Pro Department
type Department struct {
	Id   string `json:"id,omitempty"`
	Name string `json:"name"`
	Href string `json:"href,omitempty"`
}

// DepartmentListResponse represents the raw API response to getting all departments
type DepartmentListResponse struct {
	DepartmentCount *int64        `json:"totalCount"`
	Departments     *[]Department `json:"results"`
}

// DepartmentCreateRequest represents a request to create a department.
type DepartmentCreateRequest struct {
	Name string `json:"name"`
}

// DepartmentCreateResponse represents an API response to creating a department
type DepartmentCreateResponse struct {
	Id   string `json:"id"`
	Href string `json:"href"`
}

// DepartmentUpdateRequest represents an API response to creating a department.
type DepartmentUpdateRequest struct {
	Id   int    `json:"id,omitempty"`
	Name string `json:"name"`
}

// DepartmentUpdateResponse represents an API response to updating a department.
type DepartmentUpdateResponse struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

func (d *DepartmentsServiceOp) List(ctx context.Context) ([]Department, *Response, error) {
	return d.list(ctx)
}

func (d *DepartmentsServiceOp) GetByID(ctx context.Context, i int) (*Department, *Response, error) {
	path := departmentsBasePath + "/" + strconv.Itoa(i)

	req, err := d.client.NewRequest(ctx, http.MethodGet, path, nil, "application/json")
	if err != nil {
		return nil, nil, err
	}

	var department Department
	resp, err := d.client.Do(ctx, req, &department)

	if err != nil {
		return nil, resp, err
	}

	return &department, resp, err
}

func (d *DepartmentsServiceOp) GetByName(ctx context.Context, name string) (*Department, *Response, error) {
	departments, _, err := d.list(ctx)
	var id string
	if err != nil {
		return nil, nil, err
	}

	for i := range departments {
		if departments[i].Name == name {
			id = departments[i].Id
			break
		}
	}
	intId, err := strconv.ParseInt(id, 10, 64)

	if err != nil {
		return nil, nil, err
	}

	department, resp, err := d.GetByID(ctx, int(intId))
	if err != nil {
		return nil, resp, err
	}

	return department, resp, err
}

func (d *DepartmentsServiceOp) Create(ctx context.Context, request *DepartmentCreateRequest) (*Department, *Response, error) {
	if request == nil {
		return nil, nil, NewArgError("createRequest", "cannot be nil")
	}

	req, err := d.client.NewRequest(ctx, http.MethodPost, departmentsBasePath, request, "application/json")
	if err != nil {
		return nil, nil, err
	}

	departmentCreation := new(DepartmentCreateResponse)
	resp, err := d.client.Do(ctx, req, departmentCreation)
	if err != nil {
		return nil, resp, err
	}

	if departmentCreation.Id == "" {
		return nil, resp, err
	}

	department := d.createDepartmentFromCreationResponse(*departmentCreation, *request)
	return &department, resp, err
}

func (d *DepartmentsServiceOp) Update(ctx context.Context, i int, request *DepartmentUpdateRequest) (*Department, *Response, error) {
	path := departmentsBasePath + "/" + strconv.Itoa(i)

	if request == nil {
		return nil, nil, NewArgError("createRequest", "cannot be nil")
	}

	req, err := d.client.NewRequest(ctx, http.MethodPut, path, request, "application/json")
	if err != nil {
		return nil, nil, err
	}

	departmentUpdate := new(DepartmentUpdateResponse)
	resp, err := d.client.Do(ctx, req, departmentUpdate)
	if err != nil {
		return nil, resp, err
	}

	building := d.createDepartmentFromUpdateResponse(*departmentUpdate, *request)
	return &building, resp, err
}

func (d *DepartmentsServiceOp) Delete(ctx context.Context, i int) (*Response, error) {
	path := departmentsBasePath + "/" + strconv.Itoa(i)

	req, err := d.client.NewRequest(ctx, http.MethodDelete, path, nil, "application/json")
	if err != nil {
		return nil, err
	}

	resp, err := d.client.Do(ctx, req, nil)
	if err != nil && err.Error() != "EOF" {
		return resp, err
	}

	return resp, err
}

func (d *DepartmentsServiceOp) list(ctx context.Context) ([]Department, *Response, error) {
	path := departmentsBasePath
	req, err := d.client.NewRequest(ctx, http.MethodGet, path, nil, "application/json")
	if err != nil {
		return nil, nil, err
	}

	var departmentResponse DepartmentListResponse
	resp, err := d.client.Do(ctx, req, &departmentResponse)
	if err != nil {
		return nil, resp, err
	}

	return *departmentResponse.Departments, resp, err

}

func (d *DepartmentsServiceOp) createDepartmentFromCreationResponse(response DepartmentCreateResponse, request DepartmentCreateRequest) Department {
	department := new(Department)
	department.Id = response.Id
	department.Href = response.Href
	department.Name = request.Name
	return *department
}

func (d *DepartmentsServiceOp) createDepartmentFromUpdateResponse(response DepartmentUpdateResponse, request DepartmentUpdateRequest) Department {
	department := new(Department)
	department.Id = response.Id
	department.Name = request.Name
	return *department
}
