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

func (b BuildingsServiceOp) List(ctx context.Context) ([]Building, *Response, error) {
	return b.list(ctx)
}

func (b BuildingsServiceOp) GetByID(ctx context.Context, i int) (*Building, *Response, error) {
	path := buildingsBasePath + "/" + strconv.Itoa(i)

	req, err := b.client.NewRequest(ctx, http.MethodGet, path, nil)
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

func (b BuildingsServiceOp) GetByName(ctx context.Context, name string) (*Building, *Response, error) {
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

func (b BuildingsServiceOp) Create(ctx context.Context, request *BuildingCreateRequest) (*Building, *Response, error) {
	if request == nil {
		return nil, nil, NewArgError("createRequest", "cannot be nil")
	}

	req, err := b.client.NewRequest(ctx, http.MethodPost, buildingsBasePath, request)
	if err != nil {
		return nil, nil, err
	}

	building := new(Building)

	resp, err := b.client.Do(ctx, req, building)
	if err != nil {
		return nil, resp, err
	}

	return building, resp, err
}

func (b BuildingsServiceOp) Update(ctx context.Context, i int, request *BuildingUpdateRequest) (*Building, *Response, error) {
	//TODO implement me
	panic("implement me")
}

func (b BuildingsServiceOp) Delete(ctx context.Context, i int) (*Response, error) {
	//TODO implement me
	panic("implement me")
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

type BuildingGetResponse struct {
	TotalCount *int64      `json:"totalCount"`
	Buildings  *[]Building `json:"results"`
}

// BuildingCreateRequest represents a request to create a building.
type BuildingCreateRequest struct {
	Name           *string `json:"name"`
	StreetAddress1 *string `json:"streetAddress1,omitempty"`
	StreetAddress2 *string `json:"streetAddress2,omitempty"`
	City           *string `json:"city,omitempty"`
	StateProvince  *string `json:"stateProvince,omitempty"`
	ZipPostalCode  *string `json:"zipPostalCode,omitempty"`
	Country        *string `json:"country,omitempty"`
}

// BuildingUpdateRequest represents a request to update a building.
// TODO: Come back to this
type BuildingUpdateRequest struct {
	Name           *string `json:"name,omitempty"`
	StreetAddress1 *string `json:"streetAddress1,omitempty"`
	StreetAddress2 *string `json:"streetAddress2,omitempty"`
	City           *string `json:"city,omitempty"`
	StateProvince  *string `json:"stateProvince,omitempty"`
	ZipPostalCode  *string `json:"zipPostalCode,omitempty"`
	Country        *string `json:"country,omitempty"`
}

//func (building Building) String() string {
//	return Stringify(tag)
//}

type listBuildingOptions struct {
	Name string `url:"name,omitempty"`
}

func (b BuildingsServiceOp) list(ctx context.Context) ([]Building, *Response, error) {
	path := buildingsBasePath
	//path, err := addOptions(path, opt)
	//if err != nil {
	//	return nil, nil, err
	//}
	//path, err = addOptions(path, buildingOpt)
	//if err != nil {
	//	return nil, nil, err
	//}

	req, err := b.client.NewRequest(ctx, http.MethodGet, path, nil)
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
