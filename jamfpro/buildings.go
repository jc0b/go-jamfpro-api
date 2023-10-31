package jamfpro

import (
	"context"
	"net/http"
	"strconv"
)

const buildingsBasePath = "uapi/v1/buildings"

type BuildingsService interface {
	List(context.Context) ([]Building, *Response, error)
	GetByID(context.Context, int) (*Building, *Response, error)
	GetByName(context.Context, string) (*Building, *Response, error)
	Create(context.Context, *BuildingCreateRequest) (*Building, *Response, error)
	Update(context.Context, int, *BuildingUpdateRequest) (*Building, *Response, error)
	Delete(context.Context, int) (*Response, error)
}

// BuildingsServiceOp handles communication with the buildings related
// methods of the Jamf Pro API.
type BuildingsServiceOp struct {
	client *Client
}

var _ BuildingsService = &BuildingsServiceOp{}

// Building represents a Jamf Pro Building
type Building struct {
	Id             *string `json:"id,omitempty"` // The response type to be returned is a string
	Name           *string `json:"name,omitempty"`
	StreetAddress1 *string `json:"streetAddress1,omitempty"`
	StreetAddress2 *string `json:"streetAddress2,omitempty"`
	City           *string `json:"city,omitempty"`
	StateProvince  *string `json:"stateProvince,omitempty"`
	ZipPostalCode  *string `json:"zipPostalCode,omitempty"`
	Country        *string `json:"country,omitempty"`
	Href           *string `json:"href,omitempty"`
}

// BuildingGetResponse represents the raw API response to getting all buildings
type BuildingGetResponse struct {
	TotalCount *int64      `json:"totalCount"`
	Buildings  *[]Building `json:"results"`
}

// BuildingCreateRequest represents a request to create a building.
type BuildingCreateRequest struct {
	Name           string `json:"name"`
	StreetAddress1 string `json:"streetAddress1,omitempty"`
	StreetAddress2 string `json:"streetAddress2,omitempty"`
	City           string `json:"city,omitempty"`
	StateProvince  string `json:"stateProvince,omitempty"`
	ZipPostalCode  string `json:"zipPostalCode,omitempty"`
	Country        string `json:"country,omitempty"`
}

// BuildingCreateResponse represents an API response to creating a building
type BuildingCreateResponse struct {
	Id   *string `json:"id"`
	Href *string `json:"href"`
}

// BuildingUpdateRequest represents a request to update a building.
type BuildingUpdateRequest struct {
	Name           string `json:"name"`
	StreetAddress1 string `json:"streetAddress1,omitempty"`
	StreetAddress2 string `json:"streetAddress2,omitempty"`
	City           string `json:"city,omitempty"`
	StateProvince  string `json:"stateProvince,omitempty"`
	ZipPostalCode  string `json:"zipPostalCode,omitempty"`
	Country        string `json:"country,omitempty"`
}

// BuildingUpdateResponse represents an API response to updating a building
type BuildingUpdateResponse struct {
	Id             string `json:"id"` // The response type to be returned is a string
	Name           string `json:"name"`
	StreetAddress1 string `json:"streetAddress1,omitempty"`
	StreetAddress2 string `json:"streetAddress2,omitempty"`
	City           string `json:"city,omitempty"`
	StateProvince  string `json:"stateProvince,omitempty"`
	ZipPostalCode  string `json:"zipPostalCode,omitempty"`
	Country        string `json:"country,omitempty"`
}

func (b *BuildingsServiceOp) List(ctx context.Context) ([]Building, *Response, error) {
	return b.list(ctx)
}

func (b *BuildingsServiceOp) GetByID(ctx context.Context, i int) (*Building, *Response, error) {
	path := buildingsBasePath + "/" + strconv.Itoa(i)

	req, err := b.client.NewRequest(ctx, http.MethodGet, path, nil, "application/json")
	if err != nil {
		return nil, nil, err
	}

	var building Building
	resp, err := b.client.Do(ctx, req, &building)
	if err != nil {
		return nil, resp, err
	}

	return &building, resp, err
}

func (b *BuildingsServiceOp) GetByName(ctx context.Context, name string) (*Building, *Response, error) {
	buildings, _, err := b.list(ctx)
	var id string
	if err != nil {
		return nil, nil, err
	}

	for i := range buildings {
		if *buildings[i].Name == name {
			id = *buildings[i].Id
			break
		}
	}
	intId, err := strconv.ParseInt(id, 10, 64)

	if err != nil {
		return nil, nil, err
	}

	building, resp, err := b.GetByID(ctx, int(intId))
	if err != nil {
		return nil, resp, err
	}

	return building, resp, err
}

func (b *BuildingsServiceOp) Create(ctx context.Context, request *BuildingCreateRequest) (*Building, *Response, error) {
	if request == nil {
		return nil, nil, NewArgError("createRequest", "cannot be nil")
	}

	req, err := b.client.NewRequest(ctx, http.MethodPost, buildingsBasePath, request, "application/json")
	if err != nil {
		return nil, nil, err
	}

	buildingCreation := new(BuildingCreateResponse)
	resp, err := b.client.Do(ctx, req, buildingCreation)
	if err != nil {
		return nil, resp, err
	}

	if buildingCreation.Id == nil {
		return nil, resp, err
	}

	building := b.createBuildingFromCreationResponse(*buildingCreation, *request)
	return &building, resp, err
}

func (b *BuildingsServiceOp) Update(ctx context.Context, i int, request *BuildingUpdateRequest) (*Building, *Response, error) {
	path := buildingsBasePath + "/" + strconv.Itoa(i)

	if request == nil {
		return nil, nil, NewArgError("createRequest", "cannot be nil")
	}

	req, err := b.client.NewRequest(ctx, http.MethodPut, path, request, "application/json")
	if err != nil {
		return nil, nil, err
	}

	buildingUpdate := new(BuildingUpdateResponse)
	resp, err := b.client.Do(ctx, req, buildingUpdate)
	if err != nil {
		return nil, resp, err
	}

	building := b.createBuildingFromUpdateResponse(*buildingUpdate, *request)
	return &building, resp, err
}

func (b *BuildingsServiceOp) Delete(ctx context.Context, i int) (*Response, error) {
	path := buildingsBasePath + "/" + strconv.Itoa(i)

	req, err := b.client.NewRequest(ctx, http.MethodDelete, path, nil, "application/json")
	if err != nil {
		return nil, err
	}

	resp, err := b.client.Do(ctx, req, nil)
	if err != nil && err.Error() != "EOF" {
		return resp, err
	}

	return resp, err
}

func (b *BuildingsServiceOp) list(ctx context.Context) ([]Building, *Response, error) {
	path := buildingsBasePath

	req, err := b.client.NewRequest(ctx, http.MethodGet, path, nil, "application/json")
	if err != nil {
		return nil, nil, err
	}

	var buildingResponse BuildingGetResponse
	resp, err := b.client.Do(ctx, req, &buildingResponse)
	if err != nil {
		return nil, resp, err
	}

	return *buildingResponse.Buildings, resp, err
}

func (b *BuildingsServiceOp) createBuildingFromCreationResponse(response BuildingCreateResponse, request BuildingCreateRequest) Building {
	building := new(Building)
	building.Id = response.Id
	building.Href = response.Href
	building.Name = &request.Name
	building.StreetAddress1 = &request.StreetAddress1
	building.StreetAddress2 = &request.StreetAddress2
	building.City = &request.City
	building.StateProvince = &request.StateProvince
	building.ZipPostalCode = &request.ZipPostalCode
	building.Country = &request.Country
	return *building
}

func (b *BuildingsServiceOp) createBuildingFromUpdateResponse(response BuildingUpdateResponse, request BuildingUpdateRequest) Building {
	building := new(Building)
	building.Id = &response.Id
	building.Name = &request.Name
	building.StreetAddress1 = &request.StreetAddress1
	building.StreetAddress2 = &request.StreetAddress2
	building.City = &request.City
	building.StateProvince = &request.StateProvince
	building.ZipPostalCode = &request.ZipPostalCode
	building.Country = &request.Country
	return *building
}
