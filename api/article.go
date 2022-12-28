package api

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	db "github.com/kamilwrzyszcz/go_example/db/sqlc"
	"github.com/kamilwrzyszcz/go_example/token"
	"gopkg.in/guregu/null.v3"
)

type createArticleRequest struct {
	Headline string `json:"headline" binding:"required"`
	Content  string `json:"content" binding:"required"`
}

// CreateArticle godoc
// @Summary      Create an article
// @Description  Create a new article
// @Tags         articles
// @Accept       json
// @Produce      json
// @Param   payload   body    api.createArticleRequest    true  "Article payload"
// @Success      201  {object}  db.Article
// @Failure      400  {object} object{error=string}
// @Failure      500  {object} object{error=string}
// @Security BearerAuth
// @Router       /articles [post]
func (server *Server) createArticle(ctx *gin.Context) {
	var req createArticleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// Info about author is taken from token part
	// Could be handled differently. Wanted to try out getting data that way
	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	arg := db.CreateArticleParams{
		Author:   authPayload.Username,
		Headline: req.Headline,
		Content:  req.Content,
	}

	article, err := server.store.CreateArticle(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusCreated, article)
}

type getArticleRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

// GetArticle godoc
// @Summary      Get an article
// @Description  Get a specific article by ID
// @Tags         articles
// @Accept       json
// @Produce      json
// @Param   id   path    int64   true  "Article ID path param"
// @Success      200  {object}  db.Article
// @Failure      400  {object} object{error=string}
// @Failure      404  {object} object{error=string}
// @Failure      500  {object} object{error=string}
// @Security BearerAuth
// @Router       /articles/{id} [get]
func (server *Server) getArticle(ctx *gin.Context) {
	var req getArticleRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	article, err := server.store.GetArticle(ctx, req.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		} else {
			ctx.JSON(http.StatusInternalServerError, errorResponse(err))
			return
		}
	}

	ctx.JSON(http.StatusOK, article)
}

type listArticlesRequest struct {
	PageID   int32 `form:"page_id" binding:"required,min=1"`
	PageSize int32 `form:"page_size" binding:"required,min=5,max=10"`
}

// ListArticles godoc
// @Summary      Get the list of articles
// @Description  Get the list of articles accoring to specified params
// @Tags         articles
// @Accept       json
// @Produce      json
// @Param   page_id   query    int32   true  "Article PageID query param"
// @Param   page_size  query    int32   true  "Article PageSize query param"
// @Success      200  {object}  []db.Article
// @Failure      400  {object} object{error=string}
// @Failure      500  {object} object{error=string}
// @Security BearerAuth
// @Router       /articles [get]
func (server *Server) listArticles(ctx *gin.Context) {
	var req listArticlesRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	arg := db.ListArticlesParams{
		Author: authPayload.Username,
		Limit:  req.PageSize,
		Offset: (req.PageID - 1) * req.PageSize,
	}

	articles, err := server.store.ListArticles(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, articles)
}

type deleteArticleRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

// DeleteArticle godoc
// @Summary      Delete an article
// @Description  Delete an article as an article owner
// @Tags         articles
// @Accept       json
// @Produce      json
// @Param   id   path    int64   true  "Article ID path param"
// @Success      200  {object} object{}
// @Failure      400  {object} object{error=string}
// @Failure      401  {object} object{error=string}
// @Failure      404  {object} object{error=string}
// @Failure      500  {object} object{error=string}
// @Security BearerAuth
// @Router       /articles/{id} [delete]
func (server *Server) deleteArticle(ctx *gin.Context) {
	var req deleteArticleRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	article, err := server.store.GetArticle(ctx, req.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		} else {
			ctx.JSON(http.StatusInternalServerError, errorResponse(err))
			return
		}
	}

	if authPayload.Username != article.Author {
		err = errors.New("article doesn't belong to the authenticated user")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	err = server.store.DeleteArticle(ctx, req.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{})
}

type updateArticleRequest struct {
	ID   int64 `uri:"id" binding:"required"`
	Data struct {
		Headline null.String `json:"headline"`
		Content  null.String `json:"content"`
	}
}

// UpdateArticle godoc
// @Summary      Update an article
// @Description  Update an article as a article owner
// @Tags         articles
// @Accept       json
// @Produce      json
// @Param   id   path    int64   true  "Article ID path param"
// @Param   payload   body    object{headline=string,content=string}   true  "Article update payload"
// @Success      200  {object}  db.Article
// @Failure      400  {object} object{error=string}
// @Failure      401  {object} object{error=string}
// @Failure      404  {object} object{error=string}
// @Failure      500  {object} object{error=string}
// @Security BearerAuth
// @Router       /articles/{id} [patch]
func (server *Server) updateArticle(ctx *gin.Context) {
	var req updateArticleRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	if err := ctx.ShouldBindJSON(&req.Data); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	article, err := server.store.GetArticle(ctx, req.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		} else {
			ctx.JSON(http.StatusInternalServerError, errorResponse(err))
			return
		}
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	if article.Author != authPayload.Username {
		err := errors.New("article doesn't belong to the authenticated user")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	arg := db.UpdateArticleParams{
		ID:       req.ID,
		Headline: req.Data.Headline.NullString,
		Content:  req.Data.Content.NullString,
	}

	updatedArticle, err := server.store.UpdateArticle(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, updatedArticle)
}
