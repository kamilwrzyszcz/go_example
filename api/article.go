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

func (server *Server) createArticle(ctx *gin.Context) {
	var req createArticleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

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
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		} else {
			ctx.JSON(http.StatusInternalServerError, errorResponse(err))
			return
		}
	}

	ctx.JSON(http.StatusOK, articles)
}

type deleteArticleRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

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
