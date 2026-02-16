package services

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// APIDocumentationService provides API documentation generation and management
type APIDocumentationService struct {
	BaseURL string
	Title   string
	Version string
}

// NewAPIDocumentationService creates a new API documentation service
func NewAPIDocumentationService(baseURL, title, version string) *APIDocumentationService {
	return &APIDocumentationService{
		BaseURL: baseURL,
		Title:   title,
		Version: version,
	}
}

// APIEndpoint represents a single API endpoint
type APIEndpoint struct {
	Method      string                 `json:"method"`
	Path        string                 `json:"path"`
	Summary     string                 `json:"summary"`
	Description string                 `json:"description"`
	Tags        []string               `json:"tags"`
	Parameters  []APIParameter         `json:"parameters"`
	RequestBody *APIRequestBody        `json:"request_body,omitempty"`
	Responses   map[string]APIResponse `json:"responses"`
	Security    []APISecurity          `json:"security,omitempty"`
}

// APIParameter represents an API parameter
type APIParameter struct {
	Name        string `json:"name"`
	In          string `json:"in"` // query, header, path, cookie
	Description string `json:"description"`
	Required    bool   `json:"required"`
	Type        string `json:"type"`
	Example     string `json:"example,omitempty"`
}

// APIRequestBody represents the request body for an endpoint
type APIRequestBody struct {
	Description string                  `json:"description"`
	Content     map[string]APIMediaType `json:"content"`
	Required    bool                    `json:"required"`
}

// APIMediaType represents a media type in request/response
type APIMediaType struct {
	Schema APISchema `json:"schema"`
}

// APISchema represents a schema definition
type APISchema struct {
	Type       string                 `json:"type"`
	Properties map[string]APIProperty `json:"properties,omitempty"`
	Items      *APIProperty           `json:"items,omitempty"`
	Example    interface{}            `json:"example,omitempty"`
}

// APIProperty represents a property in a schema
type APIProperty struct {
	Type        string                 `json:"type"`
	Description string                 `json:"description,omitempty"`
	Format      string                 `json:"format,omitempty"`
	Enum        []string               `json:"enum,omitempty"`
	Example     interface{}            `json:"example,omitempty"`
	Properties  map[string]APIProperty `json:"properties,omitempty"`
	Required    []string               `json:"required,omitempty"`
}

// APIResponse represents an API response
type APIResponse struct {
	Description string                  `json:"description"`
	Content     map[string]APIMediaType `json:"content,omitempty"`
}

// APISecurity represents security requirements
type APISecurity struct {
	Type   string `json:"type"`
	Scheme string `json:"scheme"`
}

// APIDocumentation represents the complete API documentation
type APIDocumentation struct {
	OpenAPI    string                   `json:"openapi"`
	Info       APIInfo                  `json:"info"`
	Servers    []APIServer              `json:"servers"`
	Paths      map[string]APIPathItem   `json:"paths"`
	Components APIComponents            `json:"components,omitempty"`
	Security   []APISecurityRequirement `json:"security,omitempty"`
	Tags       []APITag                 `json:"tags,omitempty"`
}

// APIInfo contains API metadata
type APIInfo struct {
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Version     string     `json:"version"`
	Contact     APIContact `json:"contact,omitempty"`
	License     APILicense `json:"license,omitempty"`
}

// APIContact contains contact information
type APIContact struct {
	Name  string `json:"name,omitempty"`
	URL   string `json:"url,omitempty"`
	Email string `json:"email,omitempty"`
}

// APILicense contains license information
type APILicense struct {
	Name string `json:"name"`
	URL  string `json:"url,omitempty"`
}

// APIServer represents a server
type APIServer struct {
	URL         string `json:"url"`
	Description string `json:"description"`
}

// APIPathItem represents operations for a path
type APIPathItem struct {
	Get     *APIOperation `json:"get,omitempty"`
	Post    *APIOperation `json:"post,omitempty"`
	Put     *APIOperation `json:"put,omitempty"`
	Delete  *APIOperation `json:"delete,omitempty"`
	Patch   *APIOperation `json:"patch,omitempty"`
	Head    *APIOperation `json:"head,omitempty"`
	Options *APIOperation `json:"options,omitempty"`
	Trace   *APIOperation `json:"trace,omitempty"`
}

// APIOperation represents an API operation
type APIOperation struct {
	Summary     string                 `json:"summary"`
	Description string                 `json:"description"`
	Tags        []string               `json:"tags,omitempty"`
	Parameters  []APIParameter         `json:"parameters,omitempty"`
	RequestBody *APIRequestBody        `json:"requestBody,omitempty"`
	Responses   map[string]APIResponse `json:"responses"`
	Security    []APISecurity          `json:"security,omitempty"`
}

// APIComponents contains reusable components
type APIComponents struct {
	Schemas         map[string]APISchema         `json:"schemas,omitempty"`
	SecuritySchemes map[string]APISecurityScheme `json:"securitySchemes,omitempty"`
	Responses       map[string]APIResponse       `json:"responses,omitempty"`
	Parameters      map[string]APIParameter      `json:"parameters,omitempty"`
	Examples        map[string]interface{}       `json:"examples,omitempty"`
}

// APISecurityScheme represents a security scheme
type APISecurityScheme struct {
	Type         string `json:"type"`
	Scheme       string `json:"scheme,omitempty"`
	BearerFormat string `json:"bearerFormat,omitempty"`
	Description  string `json:"description,omitempty"`
}

// APISecurityRequirement represents a security requirement
type APISecurityRequirement map[string][]string

// APITag represents a tag for grouping operations
type APITag struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// GenerateDocumentation generates API documentation for the application
func (ads *APIDocumentationService) GenerateDocumentation(endpoints []APIEndpoint) *APIDocumentation {
	doc := &APIDocumentation{
		OpenAPI: "3.1.0",
		Info: APIInfo{
			Title:       ads.Title,
			Description: "InsightEngine API Documentation",
			Version:     ads.Version,
			Contact: APIContact{
				Name:  "InsightEngine Team",
				Email: "api@insightengine.ai",
			},
			License: APILicense{
				Name: "MIT",
				URL:  "https://opensource.org/licenses/MIT",
			},
		},
		Servers: []APIServer{
			{
				URL:         ads.BaseURL,
				Description: "Production server",
			},
			{
				URL:         strings.Replace(ads.BaseURL, "api.", "dev-api.", 1),
				Description: "Development server",
			},
		},
		Paths: make(map[string]APIPathItem),
		Components: APIComponents{
			Schemas: make(map[string]APISchema),
			SecuritySchemes: map[string]APISecurityScheme{
				"bearerAuth": {
					Type:         "http",
					Scheme:       "bearer",
					BearerFormat: "JWT",
					Description:  "JWT token for authentication",
				},
			},
		},
		Security: []APISecurityRequirement{
			{
				"bearerAuth": []string{},
			},
		},
		Tags: []APITag{
			{Name: "Authentication", Description: "User authentication and registration"},
			{Name: "Queries", Description: "Query management and execution"},
			{Name: "Connections", Description: "Database connection management"},
			{Name: "Dashboards", Description: "Dashboard creation and management"},
			{Name: "Analytics", Description: "Advanced analytics features"},
		},
	}

	// Group endpoints by path
	for _, endpoint := range endpoints {
		pathItem, exists := doc.Paths[endpoint.Path]
		if !exists {
			pathItem = APIPathItem{}
		}

		operation := &APIOperation{
			Summary:     endpoint.Summary,
			Description: endpoint.Description,
			Tags:        endpoint.Tags,
			Parameters:  endpoint.Parameters,
			RequestBody: endpoint.RequestBody,
			Responses:   endpoint.Responses,
			Security:    endpoint.Security,
		}

		switch strings.ToUpper(endpoint.Method) {
		case "GET":
			pathItem.Get = operation
		case "POST":
			pathItem.Post = operation
		case "PUT":
			pathItem.Put = operation
		case "DELETE":
			pathItem.Delete = operation
		case "PATCH":
			pathItem.Patch = operation
		case "HEAD":
			pathItem.Head = operation
		case "OPTIONS":
			pathItem.Options = operation
		}

		doc.Paths[endpoint.Path] = pathItem
	}

	return doc
}

// ValidateRequest validates a request against the API specification
func (ads *APIDocumentationService) ValidateRequest(c *fiber.Ctx, endpoint *APIEndpoint) error {
	// Validate path parameters
	for _, param := range endpoint.Parameters {
		if param.In == "path" && param.Required {
			value := c.Params(param.Name)
			if value == "" {
				return fmt.Errorf("required path parameter '%s' is missing", param.Name)
			}
		}
	}

	// Validate query parameters
	for _, param := range endpoint.Parameters {
		if param.In == "query" && param.Required {
			value := c.Query(param.Name)
			if value == "" {
				return fmt.Errorf("required query parameter '%s' is missing", param.Name)
			}
		}
	}

	// Validate request body if required
	if endpoint.RequestBody != nil && endpoint.RequestBody.Required {
		if len(c.Body()) == 0 {
			return fmt.Errorf("request body is required but empty")
		}
	}

	return nil
}

// GenerateExample generates example request/response for an endpoint
func (ads *APIDocumentationService) GenerateExample(endpoint *APIEndpoint) map[string]interface{} {
	example := make(map[string]interface{})

	// Generate example request
	if endpoint.RequestBody != nil && endpoint.RequestBody.Content != nil {
		for contentType, mediaType := range endpoint.RequestBody.Content {
			if mediaType.Schema.Example != nil {
				example["request_"+contentType] = mediaType.Schema.Example
			}
		}
	}

	// Generate example response
	for statusCode, response := range endpoint.Responses {
		if len(response.Content) > 0 {
			for contentType, mediaType := range response.Content {
				if mediaType.Schema.Example != nil {
					example["response_"+statusCode+"_"+contentType] = mediaType.Schema.Example
				}
			}
		}
	}

	return example
}

// GetEndpointDocumentation returns documentation for a specific endpoint
func (ads *APIDocumentationService) GetEndpointDocumentation(method, path string) *APIEndpoint {
	// This would typically come from a registry of all endpoints
	// For now, return a basic example
	return &APIEndpoint{
		Method:      method,
		Path:        path,
		Summary:     fmt.Sprintf("%s %s endpoint", strings.ToUpper(method), path),
		Description: fmt.Sprintf("Documentation for the %s %s endpoint", strings.ToUpper(method), path),
		Tags:        []string{"General"},
		Responses: map[string]APIResponse{
			"200": {
				Description: "Successful response",
				Content: map[string]APIMediaType{
					"application/json": {
						Schema: APISchema{
							Type: "object",
							Properties: map[string]APIProperty{
								"status": {
									Type:    "string",
									Example: "success",
								},
							},
						},
					},
				},
			},
		},
	}
}

// RegisterEndpoint registers an endpoint for documentation
func (ads *APIDocumentationService) RegisterEndpoint(endpoint APIEndpoint) {
	// In a real implementation, this would store the endpoint in a registry
	// For now, we just validate the structure
	_ = endpoint
}

// GetDocumentationHandler returns a Fiber handler that serves API documentation
func (ads *APIDocumentationService) GetDocumentationHandler(endpoints []APIEndpoint) fiber.Handler {
	return func(c *fiber.Ctx) error {
		doc := ads.GenerateDocumentation(endpoints)
		return c.JSON(doc)
	}
}

// GetInteractiveDocsHandler returns a handler for interactive API documentation
func (ads *APIDocumentationService) GetInteractiveDocsHandler(endpoints []APIEndpoint) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// This would serve an interactive documentation UI
		// For now, redirect to standard JSON documentation
		doc := ads.GenerateDocumentation(endpoints)
		return c.JSON(doc)
	}
}
