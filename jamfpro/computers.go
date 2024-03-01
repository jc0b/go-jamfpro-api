package jamfpro

import (
	"context"
	"encoding/xml"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

const computersBasePath = "JSSResource/computers"

type ComputersService interface {
	List(context.Context) ([]Computer, *Response, error)
	GetByID(context.Context, int) (*Computer, *Response, error)
	GetByName(context.Context, string) (*Computer, *Response, error)
	GetBySerialNumber(context.Context, string) (*Computer, *Response, error)
	Create(context.Context, *ComputerCreateRequest) (*Computer, *Response, error)
	Update(context.Context, int, *ComputerUpdateRequest) (*Computer, *Response, error)
	Delete(context.Context, int) (*Response, error)
}

// ComputersServiceOp handles communication with the computer-related
// methods of the Jamf Pro API.
type ComputersServiceOp struct {
	client *Client
}

var _ ComputersService = &ComputersServiceOp{}

// Computer represents a Jamf Pro Computer
type Computer struct {
	Id           int             `json:"id" xml:"id"`
	Name         string          `json:"name" xml:"name,omitempty"`
	General      ComputerGeneral `json:"general,omitempty" xml:"-"`
	SerialNumber string          `json:"serial_number,omitempty" xml:"serial_number,omitempty"`
	Udid         string          `json:"udid,omitempty" xml:"udid,omitempty"`
}

type ComputerGeneral struct {
	Id           int    `json:"id"`
	Name         string `json:"name"`
	AssetTag     string `json:"asset_tag"`
	Platform     string `json:"platform"`
	SerialNumber string `json:"serial_number"`
	Udid         string `json:"udid"`
}

type ComputerCreateRequest struct {
	XMLName xml.Name              `xml:"computer"`
	General ComputerCreateGeneral `xml:"general"`
}

type ComputerUpdateRequest struct {
	XMLName xml.Name              `xml:"computer"`
	General ComputerCreateGeneral `xml:"general"`
}

type ComputerCreateGeneral struct {
	Name         string `xml:"name"`
	SerialNumber string `xml:"serial_number"`
	Udid         string `xml:"udid,omitempty"`
}

type ComputerGetResponse struct {
	Computer Computer `json:"computer"`
}

type ComputerCreateResponse struct {
	Id int `xml:"id"`
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
	computerResponse.Computer.Name = computerResponse.Computer.General.Name
	computerResponse.Computer.SerialNumber = computerResponse.Computer.General.SerialNumber
	computerResponse.Computer.Udid = computerResponse.Computer.General.Udid

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
	computer.Name = computer.General.Name
	computer.SerialNumber = computer.General.SerialNumber
	computer.Udid = computer.General.Udid

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
	computerResponse.Computer.Name = computerResponse.Computer.General.Name
	computerResponse.Computer.SerialNumber = computerResponse.Computer.General.SerialNumber
	computerResponse.Computer.Udid = computerResponse.Computer.General.Udid

	return &computerResponse.Computer, resp, err
}

// Create creates a Computer record in Jamf Pro. Note that possibilities here are intentionally limited - this function
// really only serves to create dummy computer records for testing the datasource facility.
func (c *ComputersServiceOp) Create(ctx context.Context, request *ComputerCreateRequest) (*Computer, *Response, error) {
	path := computersBasePath + "/id/0"
	if request == nil {
		return nil, nil, NewArgError("createRequest", "cannot be nil")
	}

	req, err := c.client.NewRequest(ctx, http.MethodPost, path, request, "application/xml")
	if err != nil {
		return nil, nil, err
	}

	computerCreation := new(ComputerCreateResponse)
	resp, err := c.client.Do(ctx, req, computerCreation)
	if err != nil {
		return nil, resp, err
	}

	intendedComputerRecord := c.createComputerFromCreationResponse(*computerCreation, *request)

	createdComputerRecord, resp, err := c.client.Computers.GetByID(ctx, intendedComputerRecord.Id)
	interval := 1
	for resp.StatusCode != http.StatusOK && !AreComputerRecordsEquivalent(&intendedComputerRecord, createdComputerRecord) {
		time.Sleep(time.Duration(interval) * time.Second)
		createdComputerRecord, resp, err = c.client.Computers.GetByID(ctx, intendedComputerRecord.Id)
		interval = interval * 2
	}
	return &intendedComputerRecord, resp, err
}

// Update updates a Computer record in Jamf Pro. Note that possibilities here are intentionally limited - this function
// really only serves to create dummy computer records for testing the datasource facility.
func (c *ComputersServiceOp) Update(ctx context.Context, i int, request *ComputerUpdateRequest) (*Computer, *Response, error) {
	path := computersBasePath + "/id/" + strconv.Itoa(i)
	if request == nil {
		return nil, nil, NewArgError("updateRequest", "cannot be nil")
	} else if i == 0 {
		return nil, nil, NewArgError("computer ID", "cannot be 0")
	}

	req, err := c.client.NewRequest(ctx, http.MethodPut, path, request, "application/xml")
	if err != nil {
		return nil, nil, err
	}

	computerUpdate := new(ComputerCreateResponse)
	resp, err := c.client.Do(ctx, req, computerUpdate)
	if err != nil {
		return nil, resp, err
	}

	intendedComputerRecord := c.createComputerFromUpdateResponse(*computerUpdate, *request)

	updatedComputerRecord, resp, err := c.client.Computers.GetByID(ctx, intendedComputerRecord.Id)
	interval := 1
	for resp.StatusCode != http.StatusOK && !AreComputerRecordsEquivalent(&intendedComputerRecord, updatedComputerRecord) {
		time.Sleep(time.Duration(interval) * time.Second)
		updatedComputerRecord, resp, err = c.client.Computers.GetByID(ctx, intendedComputerRecord.Id)
		interval = interval * 2
	}
	return &intendedComputerRecord, resp, err
}

func (c *ComputersServiceOp) Delete(ctx context.Context, i int) (*Response, error) {
	path := computersBasePath + "/id/" + strconv.Itoa(i)

	req, err := c.client.NewRequest(ctx, http.MethodDelete, path, nil, "application/xml")
	if err != nil {
		return nil, err
	}

	deletionResp, deletionErr := c.client.Do(ctx, req, nil)
	if deletionErr != nil && deletionErr.Error() != "EOF" {
		return deletionResp, deletionErr
	}

	_, resp, err := c.client.Computers.GetByID(ctx, i)
	interval := 1
	limit := 5
	for resp.StatusCode != http.StatusNotFound && limit > 0 {
		time.Sleep(time.Duration(interval) * time.Second)
		_, resp, err = c.client.Computers.GetByID(ctx, i)
		interval = interval * 2
		limit = limit - 1
	}
	if limit == 0 {
		return nil, fmt.Errorf("failed to delete computer with id %d after 5 attempts", i)
	}

	return deletionResp, deletionErr
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

func (c *ComputersServiceOp) createComputerFromRequest(request ComputerCreateRequest) Computer {
	computer := new(Computer)
	return *computer
}

func (c *ComputersServiceOp) createComputerFromCreationResponse(response ComputerCreateResponse, request ComputerCreateRequest) Computer {
	return Computer{
		Id:           response.Id,
		Name:         request.General.Name,
		SerialNumber: request.General.SerialNumber,
	}

}

func (c *ComputersServiceOp) createComputerFromUpdateResponse(response ComputerCreateResponse, request ComputerUpdateRequest) Computer {
	return Computer{
		Id:           response.Id,
		Name:         request.General.Name,
		SerialNumber: request.General.SerialNumber,
	}
}
