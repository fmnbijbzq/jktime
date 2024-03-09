package web

import (
	"example/wb/internal/domain"
	"example/wb/internal/service"
	"example/wb/internal/web/jwt"
	"example/wb/pkg/logger"
	"net/http"

	"github.com/ecodeclub/ekit/slice"
	"github.com/gin-gonic/gin"
)

type ArticleHandler struct {
	svc service.ArticleService
	l   logger.Logger
}

func NewArticleHandler(svc service.ArticleService, l logger.Logger) *ArticleHandler {
	return &ArticleHandler{
		svc: svc,
		l:   l,
	}
}

func (h *ArticleHandler) RegisterRoutes(g *gin.Engine) {
	ag := g.Group("/article")
	ag.POST("/publish", h.Publish)
	ag.POST("/edit", h.Edit)
	ag.POST("/withdraw", h.Withdraw)

	// 创作者接口
	g.POST("/detail/:id", h.Detail)
	g.POST("/list", h.List)

}

func (h *ArticleHandler) Detail(ctx *gin.Context) {

}

func (h *ArticleHandler) List(ctx *gin.Context) {
	var page Page
	if err := ctx.Bind(&page); err != nil {
		return
	}

	uc := ctx.MustGet("user").(jwt.UserClaims)
	arts, err := h.svc.GetByAuthor(ctx, uc.Id, page.Limit, page.Offset)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		h.l.Error("查找文章列表失败",
			logger.Error(err),
			logger.Int64("uid", uc.Id),
			logger.Int("limit", page.Limit),
			logger.Int("offset", page.Offset),
		)
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Data: slice.Map[domain.Article, ArticleVo](arts, func(idx int, src domain.Article) ArticleVo {
			return ArticleVo{
				Id:       src.Id,
				Title:    src.Title,
				Content:  src.Content,
				AuthorId: src.Author.Id,
				// AuthorName: src.Author.Name, // 正常来说列表不需要作者名称
				Status: uint8(src.Status),
				Ctime:  src.Ctime,
				Utime:  src.Utime,
			}
		}),
	})

}

func (h *ArticleHandler) Withdraw(ctx *gin.Context) {
	type Req struct {
		Id int64 `json:"id"`
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		return
	}
	uc := ctx.MustGet("user").(jwt.UserClaims)
	err := h.svc.Withdraw(ctx, uc.Id, req.Id)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		h.l.Error("撤回文章失败",
			logger.Int64("uid", uc.Id),
			logger.Int64("aid", req.Id),
		)
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Msg: "OK",
	})
}

func (h *ArticleHandler) Edit(ctx *gin.Context) {
	type Req struct {
		Id      int64
		Title   string `json:"title"`
		Content string `json:"content"`
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		return
	}

	uc, ok := ctx.MustGet("user").(jwt.UserClaims)
	if !ok {
		// 正常来说用户不会到这里，可以在这里进行监控
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		h.l.Error("未发现用户的 session 信息")
		return
	}
	// 跳过检测输入数据
	// 调用svc的代码
	id, err := h.svc.Save(ctx, domain.Article{
		Id:      req.Id,
		Title:   req.Title,
		Content: req.Content,
		Author: domain.Author{
			Id: uc.Id,
		},
	})
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		h.l.Error("保存帖子失败",
			logger.Int64("uid", uc.Id),
			logger.Error(err),
		)
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Data: id,
	})
}

func (h *ArticleHandler) Publish(ctx *gin.Context) {
	type Req struct {
		Id      int64
		Title   string `json:"title"`
		Content string `json:"content"`
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		return
	}
	uc := ctx.MustGet("user").(jwt.UserClaims)
	id, err := h.svc.Publish(ctx, domain.Article{
		Id:      req.Id,
		Title:   req.Title,
		Content: req.Content,
		Author: domain.Author{
			Id: uc.Id,
		},
	})
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		h.l.Error("发表文章失败",
			logger.Int64("uid", uc.Id),
			logger.Error(err),
		)
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Data: id,
	})

}
