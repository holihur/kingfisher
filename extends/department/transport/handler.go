// Package transport implements HTTP transport for the department module.
package transport

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"

	"kingfisher/core/query"
	"kingfisher/core/response"
	"kingfisher/extends/department/app"
	"kingfisher/extends/department/domain"
)

type DepartmentHandler struct{ svc *app.DepartmentService }

func NewDepartmentHandler(svc *app.DepartmentService) *DepartmentHandler {
	return &DepartmentHandler{svc: svc}
}

// appErrCode 从 error 中提取 errcode，若非 AppError 则返回 -1
func appErrCode(err error) int {
	var e *app.Error
	if errors.As(err, &e) {
		return e.Code
	}
	return -1
}

// handleSvcErr 统一映射 service 错误：AppError → errcode，否则 500。
func handleSvcErr(c *gin.Context, err error) {
	if code := appErrCode(err); code > 0 {
		response.ErrorJSON(c, code)
		return
	}
	response.InternalError(c)
}

var departmentQueryDefs = query.Defs{
	"name":       {Name: "name", Type: query.TypeString, Searchable: true, Filterable: true},
	"parent_id":  {Name: "parent_id", Type: query.TypeUint, Filterable: true},
	"status":     {Name: "status", Type: query.TypeInt, Filterable: true},
	"subtree_id": {Name: "subtree_id", Type: query.TypeUint, Filterable: true},
	"created_at": {Name: "created_at", Type: query.TypeTime, Filterable: true},
	"updated_at": {Name: "updated_at", Type: query.TypeTime, Filterable: true},
}

// @Summary 部门树
// @Description 返回全部部门树（含每个部门挂载的角色）
// @Tags Department
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=[]domain.Department} "部门树"
// @Router /departments/tree [get]
func (h *DepartmentHandler) Tree(c *gin.Context) {
	tree, err := h.svc.Tree(c.Request.Context())
	if err != nil {
		handleSvcErr(c, err)
		return
	}
	response.OKJSON(c, tree)
}

// @Summary 部门列表（分页）
// @Description 分页部门列表；filter 支持 subtree_id（某部门及其子孙）、parent_id、status、q=名称搜索
// @Tags Department
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Param q query string false "关键词（搜索名称）"
// @Param filter query string false "筛选 JSON，如 {\"subtree_id\":1,\"status\":1}"
// @Param sort query string false "排序" default(-id)
// @Success 200 {object} response.Response{data=response.PageData} "部门列表"
// @Router /departments [get]
func (h *DepartmentHandler) List(c *gin.Context) {
	pq, err := query.Parse(c, departmentQueryDefs)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	depts, total, err := h.svc.List(c.Request.Context(), pq)
	if err != nil {
		handleSvcErr(c, err)
		return
	}
	response.PageJSON(c, depts, total, pq.Page, pq.PageSize)
}

// @Summary 部门详情
// @Description 返回单个部门（含挂载的角色）
// @Tags Department
// @Produce json
// @Security BearerAuth
// @Param id path int true "部门ID"
// @Success 200 {object} response.Response{data=domain.Department} "部门详情"
// @Failure 10901 {object} response.Response "部门不存在"
// @Router /departments/{id} [get]
func (h *DepartmentHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "id 参数错误")
		return
	}
	d, err := h.svc.GetByID(c.Request.Context(), uint(id))
	if err != nil {
		handleSvcErr(c, err)
		return
	}
	response.OKJSON(c, d)
}

// CreateDeptReq 创建部门请求体
type CreateDeptReq struct {
	ParentID uint   `json:"parent_id"`
	Name     string `json:"name" binding:"required" example:"技术部"`
	Sort     *int   `json:"sort"`
	Status   *int   `json:"status"`
	Remark   string `json:"remark"`
}

// @Summary 创建部门
// @Tags Department
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body CreateDeptReq true "创建请求"
// @Success 200 {object} response.Response{data=domain.Department} "创建成功"
// @Router /departments [post]
func (h *DepartmentHandler) Create(c *gin.Context) {
	var req CreateDeptReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	d := &domain.Department{
		ParentID: req.ParentID, Name: req.Name,
		Sort: 0, Status: 1, Remark: req.Remark,
	}
	if req.Sort != nil {
		d.Sort = *req.Sort
	}
	if req.Status != nil {
		d.Status = *req.Status
	}
	if err := h.svc.Create(c.Request.Context(), d); err != nil {
		handleSvcErr(c, err)
		return
	}
	response.OKJSON(c, d)
}

// UpdateDeptReq 更新部门请求体（指针字段，仅更新传入的字段）
type UpdateDeptReq struct {
	ParentID *uint   `json:"parent_id"`
	Name     *string `json:"name"`
	Sort     *int    `json:"sort"`
	Status   *int    `json:"status"`
	Remark   *string `json:"remark"`
}

// @Summary 更新部门
// @Tags Department
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "部门ID"
// @Param body body UpdateDeptReq true "更新请求"
// @Success 200 {object} response.Response "更新成功"
// @Router /departments/{id} [put]
func (h *DepartmentHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "id 参数错误")
		return
	}
	var req UpdateDeptReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	updates := map[string]any{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.ParentID != nil {
		updates["parent_id"] = *req.ParentID
	}
	if req.Sort != nil {
		updates["sort"] = *req.Sort
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.Remark != nil {
		updates["remark"] = *req.Remark
	}
	if len(updates) == 0 {
		response.BadRequest(c, "没有可更新的字段")
		return
	}
	if err := h.svc.Update(c.Request.Context(), uint(id), updates); err != nil {
		handleSvcErr(c, err)
		return
	}
	response.OKJSON(c, gin.H{})
}

// @Summary 删除部门
// @Description 有子部门时拒绝删除（返回 10902）
// @Tags Department
// @Produce json
// @Security BearerAuth
// @Param id path int true "部门ID"
// @Success 200 {object} response.Response "删除成功"
// @Failure 10902 {object} response.Response "存在子部门，不可删除"
// @Router /departments/{id} [delete]
func (h *DepartmentHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "id 参数错误")
		return
	}
	if err := h.svc.Delete(c.Request.Context(), uint(id)); err != nil {
		handleSvcErr(c, err)
		return
	}
	response.OKJSON(c, gin.H{})
}

// AssignRolesReq 部门角色分配请求体
type AssignRolesReq struct {
	RoleIDs []uint `json:"role_ids" binding:"required"`
}

// @Summary 分配部门角色
// @Description 全量替换部门的角色关联（部门角色会合并进成员用户的有效权限）
// @Tags Department
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "部门ID"
// @Param body body AssignRolesReq true "角色ID列表"
// @Success 200 {object} response.Response "分配成功"
// @Router /departments/{id}/roles [put]
func (h *DepartmentHandler) AssignRoles(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "id 参数错误")
		return
	}
	var req AssignRolesReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.svc.AssignRoles(c.Request.Context(), uint(id), req.RoleIDs); err != nil {
		handleSvcErr(c, err)
		return
	}
	response.OKJSON(c, gin.H{})
}
