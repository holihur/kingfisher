package transport
import ("github.com/gin-gonic/gin"; "kingfisher/core/response"; "kingfisher/extends/menu/app"; "kingfisher/extends/menu/domain"; "strconv")
type MenuHandler struct{ svc *app.MenuService }
func NewMenuHandler(svc *app.MenuService) *MenuHandler { return &MenuHandler{svc: svc} }
func (h *MenuHandler) GetTree(c *gin.Context) { tree,err:=h.svc.GetTree(c.Request.Context()); if err!=nil { response.InternalError(c); return }; response.OKJSON(c, tree) }
func (h *MenuHandler) GetByID(c *gin.Context) { id,_:=strconv.ParseUint(c.Param("id"),10,64); m,err:=h.svc.GetByID(c.Request.Context(),uint(id)); if err!=nil { response.NotFound(c); return }; response.OKJSON(c, m) }
func (h *MenuHandler) Create(c *gin.Context) { var m domain.Menu; if err:=c.ShouldBindJSON(&m); err!=nil { response.BadRequest(c,err.Error()); return }; h.svc.Create(c.Request.Context(),&m); response.OKJSON(c, m) }
func (h *MenuHandler) Update(c *gin.Context) { id,_:=strconv.ParseUint(c.Param("id"),10,64); var m map[string]any; c.ShouldBindJSON(&m); h.svc.Update(c.Request.Context(),uint(id),m); response.OKJSON(c,nil) }
func (h *MenuHandler) Delete(c *gin.Context) { id,_:=strconv.ParseUint(c.Param("id"),10,64); h.svc.Delete(c.Request.Context(),uint(id)); response.OKJSON(c,nil) }
