package jamfpro

import (
	"context"
	"net/http"
	"strconv"
)

const computersBasePath = "JSSResource/computers"

type ComputersService interface {
	List(context.Context) ([]Computer, *Response, error)
	GetByID(context.Context, int) (*Computer, *Response, error)
	GetByName(context.Context, string) (*Computer, *Response, error)
	GetBySerialNumber(context.Context, string) (*Computer, *Response, error)
}

// ComputersServiceOp handles communication with the computer-related
// methods of the Jamf Pro API.
type ComputersServiceOp struct {
	client *Client
}

var _ ComputersService = &ComputersServiceOp{}

// Computer represents a Jamf Pro Computer
type Computer struct {
	Id      int             `json:"id"`
	Name    string          `json:"name"`
	General ComputerGeneral `json:"general,omitempty"`
}

type ComputerGeneral struct {
	Id           int    `json:"id"`
	Name         string `json:"name"`
	AssetTag     string `json:"asset_tag"`
	Platform     string `json:"platform"`
	SerialNumber string `json:"serial_number"`
	Udid         string `json:"udid"`
}

type ComputerGetResponse struct {
	Computer Computer `json:"computer"`
}

// ComputerListResponse represents the raw API response to getting all computers
type ComputerListResponse struct {
	Computers *[]Computer `json:"computers"`
}

func (c *ComputersServiceOp) List(ctx context.Context) ([]Computer, *Response, error) {
	return c.list(ctx)
}

func (c *ComputersServiceOp) GetByID(ctx context.Context, Id int) (*Computer, *Response, error) {
	path := computersBasePath + "/id/" + strconv.Itoa(Id)

	req, err := c.client.NewRequest(ctx, http.MethodGet, path, nil, "application/json")
	if err != nil {
		return nil, nil, err
	}

	var computerResponse ComputerGetResponse
	resp, err := c.client.Do(ctx, req, &computerResponse)
	if err != nil {
		return nil, resp, err
	}

	computerResponse.Computer.Id = computerResponse.Computer.General.Id

	return &computerResponse.Computer, resp, err
}

func (c *ComputersServiceOp) GetByName(ctx context.Context, computerName string) (*Computer, *Response, error) {
	computers, _, err := c.list(ctx)
	var id int
	if err != nil {
		return nil, nil, err
	}

	for i := range computers {
		if computers[i].Name == computerName {
			id = computers[i].Id
			break
		}
	}

	if err != nil {
		return nil, nil, err
	}

	computer, resp, err := c.GetByID(ctx, id)
	if err != nil {
		return nil, resp, err
	}

	computer.Id = computer.General.Id

	return computer, resp, err
}

func (c *ComputersServiceOp) GetBySerialNumber(ctx context.Context, serialNumber string) (*Computer, *Response, error) {
	path := computersBasePath + "/serialnumber/" + serialNumber
	req, err := c.client.NewRequest(ctx, http.MethodGet, path, nil, "application/json")
	if err != nil {
		return nil, nil, err
	}

	var computerResponse ComputerGetResponse
	resp, err := c.client.Do(ctx, req, &computerResponse)
	if err != nil {
		return nil, resp, err
	}

	computerResponse.Computer.Id = computerResponse.Computer.General.Id

	return &computerResponse.Computer, resp, err
}

func (c *ComputersServiceOp) list(ctx context.Context) ([]Computer, *Response, error) {
	path := computersBasePath
	req, err := c.client.NewRequest(ctx, http.MethodGet, path, nil, "application/json")
	if err != nil {
		return nil, nil, err
	}

	var computerResponse ComputerListResponse
	resp, err := c.client.Do(ctx, req, &computerResponse)
	if err != nil {
		return nil, resp, err
	}

	return *computerResponse.Computers, resp, err

}
