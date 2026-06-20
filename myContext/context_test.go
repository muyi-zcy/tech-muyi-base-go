package myContext

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestHTTPIngressMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(HTTPIngressMiddleware())
	r.GET("/ping", func(c *gin.Context) {
		if TryGetTraceId(c.Request.Context()) == "" {
			t.Fatal("traceId missing")
		}
		if TryGetSsoId(c.Request.Context()) != "" {
			t.Fatal("ssoId should not be set before auth")
		}
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
}

func TestBindScalarsAndResolveActor(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

	BindScalars(c, ScalarBinding{SsoId: "42", Token: "tok"})
	ctx := c.Request.Context()
	if TryGetSsoId(ctx) != "42" {
		t.Fatalf("ssoId = %q", TryGetSsoId(ctx))
	}
	if TryGetToken(ctx) != "tok" {
		t.Fatalf("token = %q", TryGetToken(ctx))
	}
	if ResolveActor(ctx) != "42" {
		t.Fatalf("ResolveActor = %q", ResolveActor(ctx))
	}
}

func TestResolveActorAnonymous(t *testing.T) {
	if ResolveActor(context.Background()) != "system" {
		t.Fatal("anonymous actor should be system")
	}
}

func TestRequireSsoIdMissing(t *testing.T) {
	if _, err := RequireSsoId(context.Background()); err == nil {
		t.Fatal("expected error for missing ssoId")
	}
}
