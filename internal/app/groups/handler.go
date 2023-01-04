package groups

import (
	"github.com/gofiber/fiber/v2"
	"github.com/miniyus/keyword-search-backend/internal/core/api_error"
	"github.com/miniyus/keyword-search-backend/internal/utils"
	"strconv"
)

type Handler interface {
	Create(ctx *fiber.Ctx) error
	Update(ctx *fiber.Ctx) error
	Patch(ctx *fiber.Ctx) error
	All(ctx *fiber.Ctx) error
	Find(ctx *fiber.Ctx) error
	FindByName(ctx *fiber.Ctx) error
	Delete(ctx *fiber.Ctx) error
}

type HandlerStruct struct {
	service Service
}

func NewHandler(service Service) Handler {
	return &HandlerStruct{
		service,
	}
}

// Create
// @Summary create group
// @Description create group
// @Tags Groups
// @Param request body CreateGroup ture "creat group"
// @Success 201 {object} ResponseGroup
// @Failure 400 {object} api_error.ErrorResponse
// @Failure 403 {object} api_error.ErrorResponse
// @Accept json
// @Produce json
// @Router /api/groups [post]
// @Security BearerAuth
func (h *HandlerStruct) Create(ctx *fiber.Ctx) error {
	dto := &CreateGroup{}
	err := ctx.BodyParser(dto)
	if err != nil {
		errRes := api_error.NewBadRequestError(ctx)
		return errRes.Response()
	}

	validate := utils.Validate(dto)
	if validate != nil {
		errRes := api_error.NewBadRequestError(ctx)
		return errRes.Response()
	}

	errRes := utils.HandleValidate(ctx, dto)
	if errRes != nil {
		return errRes.Response()
	}

	result, err := h.service.Create(dto)
	if err != nil {
		return err
	}
	return ctx.Status(fiber.StatusCreated).JSON(result)
}

// Update
// @Summary update group
// @Description update group
// @Tags Groups
// @param id path int true "group pk"
// @Param request body UpdateGroup ture "creat group"
// @Success 200 {object} ResponseGroup
// @Failure 400 {object} api_error.ErrorResponse
// @Failure 403 {object} api_error.ErrorResponse
// @Accept json
// @Produce json
// @Router /api/groups [put]
// @Security BearerAuth
func (h *HandlerStruct) Update(ctx *fiber.Ctx) error {
	param := ctx.Params("id")
	pk, err := strconv.Atoi(param)
	if err != nil {
		errRes := api_error.NewBadRequestError(ctx)
		return errRes.Response()
	}

	dto := &UpdateGroup{}
	err = ctx.BodyParser(dto)
	if err != nil {
		errRes := api_error.NewBadRequestError(ctx)
		return errRes.Response()
	}

	errRes := utils.HandleValidate(ctx, dto)
	if errRes != nil {
		return errRes.Response()
	}

	result, err := h.service.Update(uint(pk), dto)

	return ctx.JSON(result)
}

func (h *HandlerStruct) Patch(ctx *fiber.Ctx) error {
	param := ctx.Params("id")
	pk, err := strconv.Atoi(param)
	if err != nil {
		errRes := api_error.NewBadRequestError(ctx)
		return errRes.Response()
	}

	dto := &UpdateGroup{}
	err = ctx.BodyParser(dto)
	if err != nil {
		return err
	}

	errRes := utils.HandleValidate(ctx, dto)
	if errRes != nil {
		return errRes.Response()
	}

	result, err := h.service.Update(uint(pk), dto)

	return ctx.JSON(result)
}

// All
// @Summary get all groups
// @Description get all group
// @Tags Groups
// @param page query int true "page number"
// @param page_size query int true "page size"
// @Success 200 {object} ListResponse
// @Failure 403 {object} api_error.ErrorResponse
// @Accept json
// @Produce json
// @Router /api/groups [get]
// @Security BearerAuth
func (h *HandlerStruct) All(ctx *fiber.Ctx) error {
	page, err := utils.GetPageFromCtx(ctx)

	result, err := h.service.All(page)
	if err != nil {
		return err
	}

	res := ListResponse{Paginator: result, Data: result.Data.([]ResponseGroup)}
	return ctx.JSON(res)
}

// Find
// @Summary get group by pk
// @Description get group by pk
// @Tags Groups
// @param id path int true "pk"
// @Success 200 {object} ResponseGroup
// @Failure 403 {object} api_error.ErrorResponse
// @Accept json
// @Produce json
// @Router /api/groups/:id [get]
// @Security BearerAuth
func (h *HandlerStruct) Find(ctx *fiber.Ctx) error {
	param := ctx.Params("id")
	pk, err := strconv.Atoi(param)
	if err != nil {
		return err
	}

	result, err := h.service.Find(uint(pk))
	if err != nil {
		return err
	}

	return ctx.JSON(result)
}

// FindByName
// @Summary get group by name
// @Description get group by name
// @Tags Groups
// @param name path int true "name"
// @Success 200 {object} ResponseGroup
// @Failure 403 {object} api_error.ErrorResponse
// @Accept json
// @Produce json
// @Router /api/groups/name/:name [get]
// @Security BearerAuth
func (h *HandlerStruct) FindByName(ctx *fiber.Ctx) error {
	name := ctx.Params("name")

	result, err := h.service.FindByName(name)
	if err != nil {
		return err
	}

	return ctx.JSON(result)
}

// Delete
// @Summary delete group
// @Description delete group
// @Tags Groups
// @param name path int true "name"
// @Success 204 {bool} utils.StatusResponse
// @Failure 403 {object} api_error.ErrorResponse
// @Accept json
// @Produce json
// @Router /api/groups/:id [delete]
// @Security BearerAuth
func (h *HandlerStruct) Delete(ctx *fiber.Ctx) error {
	param := ctx.Params("id")
	pk, err := strconv.Atoi(param)
	if err != nil {
		return err
	}

	result, err := h.service.Delete(uint(pk))
	if err != nil {
		return err
	}

	return ctx.Status(fiber.StatusNoContent).JSON(utils.StatusResponse{Status: result})
}
