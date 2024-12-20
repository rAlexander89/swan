package project

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rAlexander89/swan/utils"
)

func WriteHandler(projectPath, domain string) error {
	projectName, err := utils.GetProjectName()
	if err != nil {
		return fmt.Errorf("failed to get project name: %w", err)
	}

	// prepare names
	domainLower := strings.ToLower(domain)
	domainTitle := utils.ToUpperFirst(domain)

	handlerContent := fmt.Sprintf(`package api

import (
    "encoding/json"
    "net/http"
    "fmt"

    "%s/internal/core/services/%s_service/service"
)

type Create%sRequest struct {
    // TODO: add fields based on domain struct
}

type Create%sResponse struct {
    ID string `+"`json:\"id\"`"+`
}

type %sHandler struct {
    service service.%sService
}

func New%sHandler(service service.%sService) *%sHandler {
    return &%sHandler{
        service: service,
    }
}

func (h *%sHandler) Create(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
        return
    }

    var req Create%sRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, fmt.Sprintf("invalid request: %%v", err), http.StatusBadRequest)
        return
    }
    defer r.Body.Close()

    // TODO: map request to domain model and call service

    resp := Create%sResponse{
        ID: "generated-id", // TODO: get from service response
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    if err := json.NewEncoder(w).Encode(resp); err != nil {
        http.Error(w, fmt.Sprintf("error encoding response: %%v", err), http.StatusInternalServerError)
        return
    }
}`, projectName, utils.PascalToSnake(domain), domainTitle, domainTitle, domainTitle, domainTitle, domainTitle, domainTitle, domainTitle, domainTitle, domainTitle, domainTitle, domainTitle)

	// create handler directory if it doesn't exist
	handlerDir := filepath.Join(projectPath, "internal", "app", "handlers", "api")
	if err := os.MkdirAll(handlerDir, 0755); err != nil {
		return fmt.Errorf("failed to create handler directory: %v", err)
	}

	// write handler file
	handlerPath := filepath.Join(handlerDir, fmt.Sprintf("%s_handler.go", domainLower))
	if err := os.WriteFile(handlerPath, []byte(handlerContent), 0644); err != nil {
		return fmt.Errorf("failed to write handler file: %v", err)
	}

	return nil
}
