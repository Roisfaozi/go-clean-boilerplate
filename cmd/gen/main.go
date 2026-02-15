package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"unicode"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// Config holds the generator configuration
type Config struct {
	ModuleName      string // e.g. "product"
	ModuleNameTitle string // e.g. "Product" (PascalCase)
	ModuleNameCamel string // e.g. "product" (camelCase)
	BasePath        string
	ProjectModule   string // Go module name from go.mod
}

func main() {
	moduleNamePtr := flag.String("name", "", "Name of the module to generate (e.g. product)")
	flag.Parse()

	if *moduleNamePtr == "" {
		fmt.Println("Please provide a module name using -name flag")
		os.Exit(1)
	}

	moduleName := strings.ToLower(*moduleNamePtr)
	moduleNameTitle := toPascalCase(moduleName)
	moduleNameCamel := toCamelCase(moduleName)

	// Hardcoded for this project, ideally read from go.mod
	projectModule := "github.com/Roisfaozi/go-clean-boilerplate"

	config := Config{
		ModuleName:      moduleName,
		ModuleNameTitle: moduleNameTitle,
		ModuleNameCamel: moduleNameCamel,
		BasePath:        filepath.Join("internal", "modules", moduleName),
		ProjectModule:   projectModule,
	}

	fmt.Printf("Generating module: %s (Title: %s, Camel: %s)...", moduleName, moduleNameTitle, moduleNameCamel)

	createDirectories(config)
	generateFiles(config)

	fmt.Println("✅ Module generated successfully!")
	fmt.Println("Next steps:")
	fmt.Printf("1. Run 'go mod tidy'\n")
	fmt.Printf("2. Register the module in 'internal/config/app.go' (New%sModule)\n", moduleNameTitle)
	fmt.Printf("3. Add migration file if needed.")
}

func createDirectories(cfg Config) {
	dirs := []string{
		filepath.Join(cfg.BasePath, "entity"),
		filepath.Join(cfg.BasePath, "model"),
		filepath.Join(cfg.BasePath, "repository"),
		filepath.Join(cfg.BasePath, "usecase"),
		filepath.Join(cfg.BasePath, "delivery", "http"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Printf("Error creating directory %s: %v\n", dir, err)
			os.Exit(1)
		}
	}
}

func generateFiles(cfg Config) {
	files := []struct {
		Path     string
		Template string
	}{
		{
			Path:     filepath.Join(cfg.BasePath, "entity", fmt.Sprintf("%s_entity.go", cfg.ModuleName)),
			Template: entityTemplate,
		},
		{
			Path:     filepath.Join(cfg.BasePath, "model", fmt.Sprintf("%s_model.go", cfg.ModuleName)),
			Template: modelTemplate,
		},
		{
			Path:     filepath.Join(cfg.BasePath, "repository", "interface.go"),
			Template: repoInterfaceTemplate,
		},
		{
			Path:     filepath.Join(cfg.BasePath, "repository", fmt.Sprintf("%s_repository.go", cfg.ModuleName)),
			Template: repoImplTemplate,
		},
		{
			Path:     filepath.Join(cfg.BasePath, "usecase", "interface.go"),
			Template: usecaseInterfaceTemplate,
		},
		{
			Path:     filepath.Join(cfg.BasePath, "usecase", fmt.Sprintf("%s_usecase.go", cfg.ModuleName)),
			Template: usecaseImplTemplate,
		},
		{
			Path:     filepath.Join(cfg.BasePath, "delivery", "http", fmt.Sprintf("%s_controller.go", cfg.ModuleName)),
			Template: controllerTemplate,
		},
		{
			Path:     filepath.Join(cfg.BasePath, "delivery", "http", fmt.Sprintf("%s_routes.go", cfg.ModuleName)),
			Template: routesTemplate,
		},
		{
			Path:     filepath.Join(cfg.BasePath, "module.go"),
			Template: moduleTemplate,
		},
	}

	for _, file := range files {
		f, err := os.Create(file.Path)
		if err != nil {
			fmt.Printf("Error creating file %s: %v\n", file.Path, err)
			os.Exit(1)
		}
		defer func(f *os.File) {
			err := f.Close()
			if err != nil {
				fmt.Printf("Warning: failed to close file: %v\n", err)
			}
		}(f)

		tmpl, err := template.New("module").Parse(file.Template)
		if err != nil {
			fmt.Printf("Error parsing template for %s: %v\n", file.Path, err)
			os.Exit(1)
		}

		if err := tmpl.Execute(f, cfg); err != nil {
			fmt.Printf("Error executing template for %s: %v\n", file.Path, err)
			os.Exit(1)
		}
		fmt.Printf("Created: %s\n", file.Path)
	}
}

// Helper to convert "my_module" or "my module" to "MyModule"
func toPascalCase(s string) string {
	s = strings.ReplaceAll(s, "_", " ")
	c := cases.Title(language.English, cases.NoLower)
	return strings.ReplaceAll(c.String(s), " ", "")
}

// Helper to convert "my_module" to "myModule"
func toCamelCase(s string) string {
	s = toPascalCase(s)
	if len(s) == 0 {
		return s
	}
	r := []rune(s)
	r[0] = unicode.ToLower(r[0])
	return string(r)
}

// --- TEMPLATES ---

const entityTemplate = `package entity

import "gorm.io/gorm"

type {{.ModuleNameTitle}} struct {
	ID        string         ` + "`" + `gorm:"primaryKey;type:varchar(36)"` + "`" + `
	Name      string         ` + "`" + `gorm:"type:varchar(100);not null"` + "`" + `
	CreatedAt int64          ` + "`" + `gorm:"autoCreateTime:milli"` + "`" + `
	UpdatedAt int64          ` + "`" + `gorm:"autoUpdateTime:milli"` + "`" + `
	DeletedAt gorm.DeletedAt ` + "`" + `gorm:"index"` + "`" + `
}

func ({{.ModuleNameTitle}}) TableName() string {
	return "{{.ModuleName}}s"
}
`

const modelTemplate = `package model

type Create{{.ModuleNameTitle}}Request struct {
	Name string ` + "`" + `json:"name" validate:"required,min=3,max=100"` + "`" + `
}

type Update{{.ModuleNameTitle}}Request struct {
	ID   string ` + "`" + `json:"-" validate:"required"` + "`" + `
	Name string ` + "`" + `json:"name" validate:"required,min=3,max=100"` + "`" + `
}

type {{.ModuleNameTitle}}Response struct {
	ID        string ` + "`" + `json:"id"` + "`" + `
	Name      string ` + "`" + `json:"name"` + "`" + `
	CreatedAt int64  ` + "`" + `json:"created_at"` + "`" + `
}
`

const repoInterfaceTemplate = `package repository

import (
	"context"

	"{{.ProjectModule}}/internal/modules/{{.ModuleName}}/entity"
	"{{.ProjectModule}}/pkg/querybuilder"
)

type {{.ModuleNameTitle}}Repository interface {
	Create(ctx context.Context, data *entity.{{.ModuleNameTitle}}) error
	FindByID(ctx context.Context, id string) (*entity.{{.ModuleNameTitle}}, error)
	Update(ctx context.Context, data *entity.{{.ModuleNameTitle}}) error
	Delete(ctx context.Context, id string) error
	FindAllDynamic(ctx context.Context, filter *querybuilder.DynamicFilter) ([]*entity.{{.ModuleNameTitle}}, error)
}
`

const repoImplTemplate = `package repository

import (
	"context"

	"{{.ProjectModule}}/internal/modules/{{.ModuleName}}/entity"
	"{{.ProjectModule}}/pkg/querybuilder"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type {{.ModuleNameCamel}}Repository struct {
	db  *gorm.DB
	log *logrus.Logger
}

func New{{.ModuleNameTitle}}Repository(db *gorm.DB, log *logrus.Logger) {{.ModuleNameTitle}}Repository {
	return &{{.ModuleNameCamel}}Repository{
		db:  db,
		log: log,
	}
}

func (r *{{.ModuleNameCamel}}Repository) Create(ctx context.Context, data *entity.{{.ModuleNameTitle}}) error {
	if data.ID == "" {
		id, err := uuid.NewV7()
			if err != nil {
				return err
			}
			data.ID = id.String()
		}
	return r.db.WithContext(ctx).Create(data).Error
}

func (r *{{.ModuleNameCamel}}Repository) FindByID(ctx context.Context, id string) (*entity.{{.ModuleNameTitle}}, error) {
	var data entity.{{.ModuleNameTitle}}
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&data).Error; err != nil {
		return nil, err
	}
	return &data, nil
}

func (r *{{.ModuleNameCamel}}Repository) Update(ctx context.Context, data *entity.{{.ModuleNameTitle}}) error {
	return r.db.WithContext(ctx).Save(data).Error
}

func (r *{{.ModuleNameCamel}}Repository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&entity.{{.ModuleNameTitle}}{}, "id = ?", id).Error
}

func (r *{{.ModuleNameCamel}}Repository) FindAllDynamic(ctx context.Context, filter *querybuilder.DynamicFilter) ([]*entity.{{.ModuleNameTitle}}, error) {
	var results []*entity.{{.ModuleNameTitle}}
	query := r.db.WithContext(ctx)

	query, err := querybuilder.GenerateDynamicQuery(query, &entity.{{.ModuleNameTitle}}{}, filter)
	if err != nil {
		return nil, err
	}

	query, err = querybuilder.GenerateDynamicSort(query, &entity.{{.ModuleNameTitle}}{}, filter)
	if err != nil {
		return nil, err
	}

	if err := query.Find(&results).Error; err != nil {
		return nil, err
	}
	return results, nil
}
`

const usecaseInterfaceTemplate = `package usecase

import (
	"context"

	"{{.ProjectModule}}/internal/modules/{{.ModuleName}}/model"
	"{{.ProjectModule}}/pkg/querybuilder"
)

type {{.ModuleNameTitle}}UseCase interface {
	Create(ctx context.Context, req model.Create{{.ModuleNameTitle}}Request) (*model.{{.ModuleNameTitle}}Response, error)
	GetByID(ctx context.Context, id string) (*model.{{.ModuleNameTitle}}Response, error)
	Update(ctx context.Context, req model.Update{{.ModuleNameTitle}}Request) (*model.{{.ModuleNameTitle}}Response, error)
	Delete(ctx context.Context, id string) error
	GetAllDynamic(ctx context.Context, filter *querybuilder.DynamicFilter) ([]model.{{.ModuleNameTitle}}Response, error)
}
`

const usecaseImplTemplate = `package usecase

import (
	"context"
	"errors"

	"{{.ProjectModule}}/internal/modules/{{.ModuleName}}/entity"
	"{{.ProjectModule}}/internal/modules/{{.ModuleName}}/model"
	"{{.ProjectModule}}/internal/modules/{{.ModuleName}}/repository"
	"{{.ProjectModule}}/pkg/exception"
	"{{.ProjectModule}}/pkg/querybuilder"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type {{.ModuleNameCamel}}UseCase struct {
	repo repository.{{.ModuleNameTitle}}Repository
	log  *logrus.Logger
}

func New{{.ModuleNameTitle}}UseCase(repo repository.{{.ModuleNameTitle}}Repository, log *logrus.Logger) {{.ModuleNameTitle}}UseCase {
	return &{{.ModuleNameCamel}}UseCase{
		repo: repo,
		log:  log,
	}
}

func (uc *{{.ModuleNameCamel}}UseCase) Create(ctx context.Context, req model.Create{{.ModuleNameTitle}}Request) (*model.{{.ModuleNameTitle}}Response, error) {
	entity := &entity.{{.ModuleNameTitle}}{
		Name: req.Name,
	}

	if err := uc.repo.Create(ctx, entity); err != nil {
		uc.log.WithError(err).Error("Failed to create {{.ModuleName}}")
		return nil, exception.ErrInternalServer
	}

	return &model.{{.ModuleNameTitle}}Response{
		ID:        entity.ID,
		Name:      entity.Name,
		CreatedAt: entity.CreatedAt,
	}, nil
}

func (uc *{{.ModuleNameCamel}}UseCase) GetByID(ctx context.Context, id string) (*model.{{.ModuleNameTitle}}Response, error) {
	entity, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, exception.ErrNotFound
		}
		uc.log.WithError(err).Error("Failed to get {{.ModuleName}} by ID")
		return nil, exception.ErrInternalServer
	}

	return &model.{{.ModuleNameTitle}}Response{
		ID:        entity.ID,
		Name:      entity.Name,
		CreatedAt: entity.CreatedAt,
	}, nil
}

func (uc *{{.ModuleNameCamel}}UseCase) Update(ctx context.Context, req model.Update{{.ModuleNameTitle}}Request) (*model.{{.ModuleNameTitle}}Response, error) {
	entity, err := uc.repo.FindByID(ctx, req.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, exception.ErrNotFound
		}
		return nil, exception.ErrInternalServer
	}

	entity.Name = req.Name

	if err := uc.repo.Update(ctx, entity); err != nil {
		uc.log.WithError(err).Error("Failed to update {{.ModuleName}}")
		return nil, exception.ErrInternalServer
	}

	return &model.{{.ModuleNameTitle}}Response{
		ID:        entity.ID,
		Name:      entity.Name,
		CreatedAt: entity.CreatedAt,
	}, nil
}

func (uc *{{.ModuleNameCamel}}UseCase) Delete(ctx context.Context, id string) error {
	if err := uc.repo.Delete(ctx, id); err != nil {
		uc.log.WithError(err).Error("Failed to delete {{.ModuleName}}")
		return exception.ErrInternalServer
	}
	return nil
}

func (uc *{{.ModuleNameCamel}}UseCase) GetAllDynamic(ctx context.Context, filter *querybuilder.DynamicFilter) ([]model.{{.ModuleNameTitle}}Response, error) {
	entities, err := uc.repo.FindAllDynamic(ctx, filter)
	if err != nil {
		uc.log.WithError(err).Error("Failed to get all {{.ModuleName}}s")
		return nil, exception.ErrInternalServer
	}

	var responses []model.{{.ModuleNameTitle}}Response
	for _, entity := range entities {
		responses = append(responses, model.{{.ModuleNameTitle}}Response{
			ID:        entity.ID,
			Name:      entity.Name,
			CreatedAt: entity.CreatedAt,
		})
	}
	return responses, nil
}
`

const controllerTemplate = `package http

import (
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/{{.ModuleName}}/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/{{.ModuleName}}/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/exception"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/querybuilder"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/response"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/validation"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
)

type {{.ModuleNameTitle}}Controller struct {
	UseCase  usecase.{{.ModuleNameTitle}}UseCase
	Log      *logrus.Logger
	Validate *validator.Validate
}

func New{{.ModuleNameTitle}}Controller(useCase usecase.{{.ModuleNameTitle}}UseCase, log *logrus.Logger, validate *validator.Validate) *{{.ModuleNameTitle}}Controller {
	return &{{.ModuleNameTitle}}Controller{
		UseCase:  useCase,
		Log:      log,
		Validate: validate,
	}
}

func (c *{{.ModuleNameTitle}}Controller) Create(ctx *gin.Context) {
	var req model.Create{{.ModuleNameTitle}}Request
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, exception.ErrBadRequest, "Invalid request body")
		return
	}

	if err := c.Validate.Struct(req); err != nil {
		msg := validation.FormatValidationErrors(err)
		response.ValidationError(ctx, exception.ErrValidationError, msg)
		return
	}

	resp, err := c.UseCase.Create(ctx.Request.Context(), req)
	if err != nil {
		response.HandleError(ctx, err, "Failed to create {{.ModuleName}}")
		return
	}

	response.Created(ctx, resp)
}

func (c *{{.ModuleNameTitle}}Controller) GetByID(ctx *gin.Context) {
	id := ctx.Param("id")
	resp, err := c.UseCase.GetByID(ctx.Request.Context(), id)
	if err != nil {
		response.HandleError(ctx, err, "Failed to get {{.ModuleName}}")
		return
	}

	response.Success(ctx, resp)
}

func (c *{{.ModuleNameTitle}}Controller) Update(ctx *gin.Context) {
	id := ctx.Param("id")
	var req model.Update{{.ModuleNameTitle}}Request
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, exception.ErrBadRequest, "Invalid request body")
		return
	}
	req.ID = id

	if err := c.Validate.Struct(req); err != nil {
		msg := validation.FormatValidationErrors(err)
		response.ValidationError(ctx, exception.ErrValidationError, msg)
		return
	}

	resp, err := c.UseCase.Update(ctx.Request.Context(), req)
	if err != nil {
		response.HandleError(ctx, err, "Failed to update {{.ModuleName}}")
		return
	}

	response.Success(ctx, resp)
}

func (c *{{.ModuleNameTitle}}Controller) Delete(ctx *gin.Context) {
	id := ctx.Param("id")
	if err := c.UseCase.Delete(ctx.Request.Context(), id); err != nil {
		response.HandleError(ctx, err, "Failed to delete {{.ModuleName}}")
		return
	}

	response.Success(ctx, gin.H{"message": "{{.ModuleNameTitle}} deleted successfully"})
}

func (c *{{.ModuleNameTitle}}Controller) GetDynamic(ctx *gin.Context) {
	var filter querybuilder.DynamicFilter
	if err := ctx.ShouldBindJSON(&filter); err != nil {
		response.BadRequest(ctx, exception.ErrBadRequest, "Invalid filter format")
		return
	}

	resp, err := c.UseCase.GetAllDynamic(ctx.Request.Context(), &filter)
	if err != nil {
		response.HandleError(ctx, err, "Failed to get {{.ModuleName}} list")
		return
	}

	response.Success(ctx, resp)
}
`

const routesTemplate = `package http

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.RouterGroup, controller *{{.ModuleNameTitle}}Controller) {
	group := router.Group("/{{.ModuleName}}s")
	{
		group.POST("", controller.Create)
		group.GET("/:id", controller.GetByID)
		group.PUT("/:id", controller.Update)
		group.DELETE("/:id", controller.Delete)
		group.POST("/search", controller.GetDynamic)
	}
}
`

const moduleTemplate = `package {{.ModuleName}}

import (
	"{{.ProjectModule}}/internal/modules/{{.ModuleName}}/delivery/http"
	"{{.ProjectModule}}/internal/modules/{{.ModuleName}}/repository"
	"{{.ProjectModule}}/internal/modules/{{.ModuleName}}/usecase"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type {{.ModuleNameTitle}}Module struct {
	{{.ModuleNameTitle}}Controller *http.{{.ModuleNameTitle}}Controller
}

func New{{.ModuleNameTitle}}Module(db *gorm.DB, log *logrus.Logger, validate *validator.Validate) *{{.ModuleNameTitle}}Module {
	repo := repository.New{{.ModuleNameTitle}}Repository(db, log)
	uc := usecase.New{{.ModuleNameTitle}}UseCase(repo, log)
	controller := http.New{{.ModuleNameTitle}}Controller(uc, log, validate)

	return &{{.ModuleNameTitle}}Module{
		{{.ModuleNameTitle}}Controller: controller,
	}
}

func (m *{{.ModuleNameTitle}}Module) Controller() *http.{{.ModuleNameTitle}}Controller {
	return m.{{.ModuleNameTitle}}Controller
}
`
