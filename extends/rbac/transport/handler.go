package transport
import ("github.com/gin-gonic/gin"; "kingfisher/core/response"; "kingfisher/extends/rbac/app"; "kingfisher/extends/rbac/domain"; "strconv")
type RoleHandler struct{ svc *app.RoleService }
func NewRoleHandler(svc *app.RoleService) *RoleHandler { return &RoleHandler{svc: svc} }
func (h *RoleHandler) List(c *gin.Context) { roles, err := h.svc.List(c.Request.Context()); if err != nil { response.InternalError(c); return }; response.OKJSON(c, roles) }
func (h *RoleHandler) GetByID(c *gin.Context) { id,_:=strconv.ParseUint(c.Param("id"),10,64); role,err:=h.svc.GetByID(c.Request.Context(),uint(id)); if err!=nil { response.NotFound(c); return }; response.OKJSON(c, role) }
func (h *RoleHandler) Create(c *gin.Context) { var r domain.Role; if err:=c.ShouldBindJSON(&r); err!=nil { response.BadRequest(c,err.Error()); return }; h.svc.Create(c.Request.Context(),&r); response.OKJSON(c, r) }
func (h *RoleHandler) Update(c *gin.Context) { id,_:=strconv.ParseUint(c.Param("id"),10,64); var m map[string]any; c.ShouldBindJSON(&m); h.svc.Update(c.Request.Context(),uint(id),m); response.OKJSON(c,nil) }
func (h *RoleHandler) Delete(c *gin.Context) { id,_:=strconv.ParseUint(c.Param("id"),10,64); h.svc.Delete(c.Request.Context(),uint(id)); response.OKJSON(c,nil) }
type AssignPermReq struct { PermissionIDs []uint `json:"permission_ids"` }
func (h *RoleHandler) AssignPerms(c *gin.Context) { id,_:=strconv.ParseUint(c.Param("id"),10,64); var req AssignPermReq; c.ShouldBindJSON(&req); h.svc.AssignPermissions(c.Request.Context(),uint(id),req.PermissionIDs); response.OKJSON(c,nil) }
type AssignMenuReq struct { MenuIDs []uint `json:"menu_ids"` }
func (h *RoleHandler) AssignMenus(c *gin.Context) { id,_:=strconv.ParseUint(c.Param("id"),10,64); var req AssignMenuReq; c.ShouldBindJSON(&req); h.svc.AssignMenus(c.Request.Context(),uint(id),req.MenuIDs); response.OKJSON(c,nil) }
func (h *RoleHandler) GetPermissions(c *gin.Context) { id,_:=strconv.ParseUint(c.Param("id"),10,64); perms,_:=h.svc.GetRolePermissions(c.Request.Context(),uint(id)); response.OKJSON(c, perms) }
func (h *RoleHandler) GetMenus(c *gin.Context) { id,_:=strconv.ParseUint(c.Param("id"),10,64); menus,_:=h.svc.GetRoleMenus(c.Request.Context(),uint(id)); response.OKJSON(c, menus) }
type PermHandler struct{ svc *app.PermService }
func NewPermHandler(svc *app.PermService) *PermHandler { return &PermHandler{svc: svc} }
func (h *PermHandler) List(c *gin.Context) { perms,err:=h.svc.List(c.Request.Context()); if err!=nil { response.InternalError(c); return }; response.OKJSON(c, perms) }
