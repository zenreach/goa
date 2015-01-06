package v3

// Note how we don't import goa into "." here so that types used in controllers don't clash with types coming from
// goa. We import it in "." in resource and media type definitions to take advantage of the DSL.
import (
	"./db"
	"github.com/raphael/goa"
)

type postController struct{}

/* list */
func (c *postController) list(r *goa.Request) {
	if author, err := db.getAuthor(r.Params("authorId")); err != nil {
		r.RespondEmpty("notFound")
	}
	if context.Param("ids") != nil {
		r := goa.Multipart()
		for id := range context.Param("ids") {
			if post, err := getDraft(context.Param("authorId"), id); err == nil {
				r.AddPart(d.id, postMediaType.render(d, context.Param("view")))
			} else {
				r.AddPart(d.id, goa.InternalError(err.Error()))
			}
		}
		return &r

	} else {
		if post, err := getDraft(context.Param("authorId"), id); err == nil {
			return &postMediaType.render(post, context.Param("view"))
		} else {
			return &goa.InternalError(err.Error())
		}
	}
}

/* get */
func (c *postController) show(context goa.RequestContext) *httpResponse {
	if author, err := getAuthor(context.Param("authorId")); err != nil {
		r.RespondEmpty("notFound")
	}
	if context.Param("ids") != nil {
		r := goa.Multipart()
		for id := range context.Param("ids") {
			if post, err := getDraft(context.Param("authorId"), id); err == nil {
				r.AddPart(d.id, postMediaType.render(d, context.Param("view")))
			} else {
				r.AddPart(d.id, goa.InternalError(err.Error()))
			}
		}
		return &r

	} else {
		if post, err := getDraft(context.Param("authorId"), context.Param("id")); err == nil {
			return &postMediaType.render(post, context.Param("view"))
		} else {
			return &goa.InternalError(err.Error())
		}
	}
}

/* create */
func (c *postController) create(context goa.RequestContext) *http.Response {
	if author, err := getAuthor(context.Param("authorId")); err != nil {
		return &goa.NotFound()
	}

	if Payload().IsMultipart() {
		r := goa.Multipart()
		for k, p := range Payload().GetParts() {
			if post, err := createDraft(author, p); err == nil {
				r.AddPart(k, goa.Created(postResource.href(post)))
			} else {
				r.AddPart(k, goa.InternalError(err.Error()))
			}
		}
		return &r

	} else {
		if post, err := createDraft(author, Payload()); err == nil {
			return &goa.Created(postResource.href(post))
		} else {
			return &goa.InternalError(err.Error())
		}
	}
}

/* delete */
func (c *postController) del(context goa.RequestContext) *httpResponse {
	if author, err := getAuthor(context.Param("authorId")); err != nil {
		return &goa.NotFound()
	}
	if err := deleteDraft(context.Param("authorId"), context.Param("id")); err == nil {
		return &goa.NoContent()
	} else {
		return &goa.InternalError(err.Error())
	}
}

/* bulk delete */
func (c *postController) bulk_del(context goa.RequestContext) *httpResponse {
	if author, err := getAuthor(context.Param("authorId")); err != nil {
		return &goa.NotFound()
	}
	r := goa.MultipartResponse()
	for id := range context.Param("ids") {
		if err := deleteDraft(context.Param("authorId"), id); err == nil {
			r.AddPart(d.id, goa.NoContent())
		} else {
			r.AddPart(d.id, goa.InternalError(err.Error()))
		}
	}
	return &r
}
