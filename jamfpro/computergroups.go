package jamfpro

import (
	"context"
	"encoding/xml"
	"net/http"
	"strconv"
	"time"
)

const computerGroupsBasePath = "JSSResource/computergroups"

type ComputerGroupsService interface {
	List(context.Context) ([]ComputerGroup, *Response, error)
	GetByID(context.Context, int) (*ComputerGroup, *Response, error)
	GetByName(context.Context, string) (*ComputerGroup, *Response, error)
	Create(context.Context, *ComputerGroupRequest) (*ComputerGroup, *Response, error)
	Update(context.Context, int, *ComputerGroupRequest) (*ComputerGroup, *Response, error)
	Delete(context.Context, int) (*Response, error)
}

// ComputerGroupsServiceOp handles communication with the computer group-related
// methods of the Jamf Pro API.
type ComputerGroupsServiceOp struct {
	client *Client
}

var _ ComputerGroupsService = &ComputerGroupsServiceOp{}

// ComputerGroup represents a Jamf Pro ComputerGroup
type ComputerGroup struct {
	Id      int    `xml:"id"`
	Name    string `xml:"name"`
	IsSmart bool   `xml:"is_smart"`
	//TODO: Sites
	//Site         Site   `json:"site"`
	Criteria  []ComputerGroupCriteria `xml:"criteria>criterion,omitempty"`
	Computers []Computer              `xml:"computers>computer,omitempty"`
}

type ComputerGroupCriteria struct {
	Name         string `xml:"name"`
	Priority     int    `xml:"priority"`
	AndOr        string `xml:"and_or"`
	SearchType   string `xml:"search_type"`
	Value        string `xml:"value"`
	OpeningParen bool   `xml:"opening_paren"`
	ClosingParen bool   `xml:"closing_paren"`
}

type ComputerGroupRequest struct {
	XMLName xml.Name `xml:"computer_group"`
	Name    string   `xml:"name"`
	IsSmart bool     `xml:"is_smart"`
	//TODO: Sites
	//Site         Site   `json:"site"`
	Criteria  []ComputerGroupCriteria `xml:"criteria>criterion,omitempty"`
	Computers []Computer              `xml:"computers>computer,omitempty"`
}

type ComputerGroupResponse struct {
	Id int `xml:"id"`
}

// ComputerGroupListResponse represents the raw API response to getting all computerGroups
type ComputerGroupListResponse struct {
	ComputerGroups *[]ComputerGroup `json:"computer_groups"`
}

func (c *ComputerGroupsServiceOp) Create(ctx context.Context, request *ComputerGroupRequest) (*ComputerGroup, *Response, error) {
	path := computerGroupsBasePath + "/id/0"
	if request == nil {
		return nil, nil, NewArgError("createRequest", "cannot be nil")
	}

	if request.IsSmart && len(request.Criteria) < 0 {
		return nil, nil, NewArgError("Criteria", "Criteria must be supplied for a Smart Group")
	}

	req, err := c.client.NewRequest(ctx, http.MethodPost, path, request, "application/xml")
	if err != nil {
		return nil, nil, err
	}

	computerGroupCreation := new(ComputerGroupResponse)
	resp, err := c.client.Do(ctx, req, computerGroupCreation)
	if err != nil {
		return nil, resp, err
	}

	if computerGroupCreation.Id == 0 {
		return nil, resp, err
	}

	// Below, we are attempting to work around Jamf Pro replication lag. It may take a while for the API changes to
	// actually take place on the server, so we wait until the API shows us it has happened.
	intendedComputerGroup := c.createComputerGroupFromRequest(*request)
	updatedComputerGroup, resp, err := c.client.ComputerGroups.GetByID(ctx, computerGroupCreation.Id)
	interval := 1
	for resp.StatusCode != http.StatusOK && !AreGroupsEquivalent(&intendedComputerGroup, updatedComputerGroup) {
		time.Sleep(time.Duration(interval) * time.Second)
		updatedComputerGroup, resp, err = c.client.ComputerGroups.GetByID(ctx, computerGroupCreation.Id)
		interval = interval * 2
	}
	computerGroup := c.createComputerGroupFromResponse(*computerGroupCreation, *request)
	return &computerGroup, resp, err
}

func (c *ComputerGroupsServiceOp) Update(ctx context.Context, i int, request *ComputerGroupRequest) (*ComputerGroup, *Response, error) {
	path := computerGroupsBasePath + "/id/" + strconv.Itoa(i)
	if request == nil {
		return nil, nil, NewArgError("createRequest", "cannot be nil")
	}

	req, err := c.client.NewRequest(ctx, http.MethodPut, path, request, "application/xml")
	if err != nil {
		return nil, nil, err
	}

	computerGroupUpdate := new(ComputerGroupResponse)
	resp, err := c.client.Do(ctx, req, computerGroupUpdate)
	if err != nil {
		return nil, resp, err
	}
	retryCount := 5
	if resp.StatusCode == 404 {
		for resp.StatusCode == 404 && retryCount > 0 {
			time.Sleep(time.Duration(2) * time.Second)
			resp, err = c.client.Do(ctx, req, computerGroupUpdate)
			retryCount = retryCount - 1
		}
	}

	if computerGroupUpdate.Id == 0 {
		return nil, resp, err
	}

	// Below, we are attempting to work around Jamf Pro replication lag. It may take a while for the API changes to
	// actually take place on the server, so we wait until the API shows us it has happened.
	intendedComputerGroup := c.createComputerGroupFromRequest(*request)
	updatedComputerGroup, resp, err := c.client.ComputerGroups.GetByID(ctx, computerGroupUpdate.Id)
	interval := 1
	for resp.StatusCode != http.StatusOK && !AreGroupsEquivalent(&intendedComputerGroup, updatedComputerGroup) {
		time.Sleep(time.Duration(interval) * time.Second)
		updatedComputerGroup, resp, err = c.client.ComputerGroups.GetByID(ctx, computerGroupUpdate.Id)
		interval = interval * 2
	}
	computerGroup := c.createComputerGroupFromResponse(*computerGroupUpdate, *request)
	return &computerGroup, resp, err
}

func (c *ComputerGroupsServiceOp) Delete(ctx context.Context, i int) (*Response, error) {
	path := computerGroupsBasePath + "/id/" + strconv.Itoa(i)

	req, err := c.client.NewRequest(ctx, http.MethodDelete, path, nil, "application/xml")

	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(ctx, req, nil)
	if err != nil && err.Error() != "EOF" {
		return resp, err
	}

	return resp, err

}

func (c *ComputerGroupsServiceOp) List(ctx context.Context) ([]ComputerGroup, *Response, error) {
	return c.list(ctx)
}

func (c *ComputerGroupsServiceOp) GetByID(ctx context.Context, Id int) (*ComputerGroup, *Response, error) {
	path := computerGroupsBasePath + "/id/" + strconv.Itoa(Id)

	req, err := c.client.NewRequest(ctx, http.MethodGet, path, nil, "application/xml")
	if err != nil {
		return nil, nil, err
	}

	var computerGroupResponse ComputerGroup
	resp, err := c.client.Do(ctx, req, &computerGroupResponse)
	if err != nil {
		return nil, resp, err
	}

	if computerGroupResponse.IsSmart {
		computerGroupResponse.Computers = nil
	} else {
		computerGroupResponse.Criteria = nil
	}

	return &computerGroupResponse, resp, err
}

func (c *ComputerGroupsServiceOp) GetByName(ctx context.Context, computerGroupName string) (*ComputerGroup, *Response, error) {
	computerGroups, _, err := c.list(ctx)
	var id int
	if err != nil {
		return nil, nil, err
	}

	for i := range computerGroups {
		if computerGroups[i].Name == computerGroupName {
			id = computerGroups[i].Id
			break
		}
	}

	if err != nil {
		return nil, nil, err
	}

	computerGroup, resp, err := c.GetByID(ctx, id)
	if err != nil {
		return nil, resp, err
	}

	return computerGroup, resp, err
}

func (c *ComputerGroupsServiceOp) list(ctx context.Context) ([]ComputerGroup, *Response, error) {
	path := computerGroupsBasePath
	req, err := c.client.NewRequest(ctx, http.MethodGet, path, nil, "application/json")
	if err != nil {
		return nil, nil, err
	}

	var computerGroupResponse ComputerGroupListResponse
	resp, err := c.client.Do(ctx, req, &computerGroupResponse)
	if err != nil {
		return nil, resp, err
	}

	return *computerGroupResponse.ComputerGroups, resp, err

}

func (c *ComputerGroupsServiceOp) createComputerGroupFromRequest(request ComputerGroupRequest) ComputerGroup {
	computerGroup := new(ComputerGroup)
	computerGroup.Name = request.Name
	computerGroup.IsSmart = request.IsSmart
	computerGroup.Criteria = request.Criteria
	computerGroup.Computers = request.Computers
	return *computerGroup
}

func (c *ComputerGroupsServiceOp) createComputerGroupFromResponse(response ComputerGroupResponse, request ComputerGroupRequest) ComputerGroup {
	computerGroup := new(ComputerGroup)
	computerGroup.Id = response.Id
	computerGroup.Name = request.Name
	computerGroup.IsSmart = request.IsSmart
	computerGroup.Criteria = request.Criteria
	computerGroup.Computers = request.Computers
	return *computerGroup
}
