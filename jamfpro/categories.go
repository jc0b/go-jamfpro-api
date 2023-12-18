package jamfpro

import (
	"context"
	"net/http"
	"strconv"
)

const categoriesBasePath = "uapi/v1/categories"

type CategoriesService interface {
	List(context.Context) ([]Category, *Response, error)
	GetByID(context.Context, int) (*Category, *Response, error)
	GetByName(context.Context, string) (*Category, *Response, error)
	Create(context.Context, *CategoryCreateRequest) (*Category, *Response, error)
	Update(context.Context, int, *CategoryUpdateRequest) (*Category, *Response, error)
	Delete(context.Context, int) (*Response, error)
}

// CategoriesServiceOp handles communication with the categories-related
// methods of the Jamf Pro API.
type CategoriesServiceOp struct {
	client *Client
}

var _ CategoriesService = &CategoriesServiceOp{}

// Category represents a Jamf Pro Category
type Category struct {
	Id       string `json:"id,omitempty"`
	Name     string `json:"name"`
	Priority int    `json:"priority"`
	Href     string `json:"href,omitempty"`
}

// CategoryListResponse represents the raw API response to getting all categories
type CategoryListResponse struct {
	CategoryCount *int64      `json:"totalCount"`
	Categories    *[]Category `json:"results"`
}

// CategoryCreateRequest represents a request to create a category.
type CategoryCreateRequest struct {
	Name     string `json:"name"`
	Priority int    `json:"priority"`
}

// CategoryCreateResponse represents an API response to creating a category
type CategoryCreateResponse struct {
	Id   string `json:"id"`
	Href string `json:"href"`
}

// CategoryUpdateRequest represents an API response to creating a category.
type CategoryUpdateRequest struct {
	Id       int    `json:"id,omitempty"`
	Name     string `json:"name"`
	Priority int    `json:"priority"`
}

type CategoryUpdateResponse struct {
	Id       string `json:"id"`
	Name     string `json:"name"`
	Priority int    `json:"priority"`
}

func (c *CategoriesServiceOp) List(ctx context.Context) ([]Category, *Response, error) {
	return c.list(ctx)
}

func (c *CategoriesServiceOp) GetByID(ctx context.Context, i int) (*Category, *Response, error) {
	path := categoriesBasePath + "/" + strconv.Itoa(i)

	req, err := c.client.NewRequest(ctx, http.MethodGet, path, nil, "application/json")
	if err != nil {
		return nil, nil, err
	}

	var category Category
	resp, err := c.client.Do(ctx, req, &category)
	if err != nil {
		return nil, resp, err
	}

	return &category, resp, err
}

func (c *CategoriesServiceOp) GetByName(ctx context.Context, name string) (*Category, *Response, error) {
	categories, _, err := c.list(ctx)
	var id string
	if err != nil {
		return nil, nil, err
	}

	for i := range categories {
		if categories[i].Name == name {
			id = categories[i].Id
			break
		}
	}
	intId, err := strconv.ParseInt(id, 10, 64)

	if err != nil {
		return nil, nil, err
	}

	category, resp, err := c.GetByID(ctx, int(intId))
	if err != nil {
		return nil, resp, err
	}

	return category, resp, err
}

func (c CategoriesServiceOp) Create(ctx context.Context, request *CategoryCreateRequest) (*Category, *Response, error) {
	if request == nil {
		return nil, nil, NewArgError("createRequest", "cannot be nil")
	}

	req, err := c.client.NewRequest(ctx, http.MethodPost, categoriesBasePath, request, "application/json")
	if err != nil {
		return nil, nil, err
	}

	categoryCreation := new(CategoryCreateResponse)
	resp, err := c.client.Do(ctx, req, categoryCreation)
	if err != nil {
		return nil, resp, err
	}

	if categoryCreation.Id == "" {
		return nil, resp, err
	}

	category := c.createCategoryFromCreationResponse(*categoryCreation, *request)
	return &category, resp, err
}

func (c *CategoriesServiceOp) Update(ctx context.Context, i int, request *CategoryUpdateRequest) (*Category, *Response, error) {
	path := categoriesBasePath + "/" + strconv.Itoa(i)
	if request == nil {
		return nil, nil, NewArgError("updateRequest", "cannot be nil")
	} else if i == 0 {
		return nil, nil, NewArgError("category ID", "cannot be 0")
	}

	req, err := c.client.NewRequest(ctx, http.MethodPut, path, request, "application/json")
	if err != nil {
		return nil, nil, err
	}

	categoryUpdate := new(CategoryUpdateResponse)
	resp, err := c.client.Do(ctx, req, categoryUpdate)
	if err != nil {
		return nil, resp, err
	}

	building := c.createCategoryFromUpdateResponse(*categoryUpdate, *request)
	return &building, resp, err
}

func (c *CategoriesServiceOp) Delete(ctx context.Context, i int) (*Response, error) {
	path := categoriesBasePath + "/" + strconv.Itoa(i)

	req, err := c.client.NewRequest(ctx, http.MethodDelete, path, nil, "application/json")
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(ctx, req, nil)
	if err != nil && err.Error() != "EOF" {
		return resp, err
	}

	return resp, err
}

func (c *CategoriesServiceOp) list(ctx context.Context) ([]Category, *Response, error) {
	path := categoriesBasePath
	req, err := c.client.NewRequest(ctx, http.MethodGet, path, nil, "application/json")
	if err != nil {
		return nil, nil, err
	}

	var categoryResponse CategoryListResponse
	resp, err := c.client.Do(ctx, req, &categoryResponse)
	if err != nil {
		return nil, resp, err
	}

	return *categoryResponse.Categories, resp, err

}

func (c *CategoriesServiceOp) createCategoryFromCreationResponse(response CategoryCreateResponse, request CategoryCreateRequest) Category {
	category := new(Category)
	category.Id = response.Id
	category.Href = response.Href
	category.Name = request.Name
	category.Priority = request.Priority
	return *category
}

func (c *CategoriesServiceOp) createCategoryFromUpdateResponse(response CategoryUpdateResponse, request CategoryUpdateRequest) Category {
	category := new(Category)
	category.Id = response.Id
	category.Name = request.Name
	category.Priority = request.Priority
	return *category
}
