//go:build testtools
// +build testtools

package api_test

import (
	"bytes"
	"context"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"entgo.io/ent/dialect"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/brianvoe/gofakeit/v6"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"go.uber.org/zap"

	"go.infratographer.com/x/echojwtx"
	"go.infratographer.com/x/echox"
	"go.infratographer.com/x/gidx"

	"go.infratographer.com/x/testing/containersx"

	"go.infratographer.com/resource-provider-api/internal/api"
	ent "go.infratographer.com/resource-provider-api/internal/ent/generated"
	"go.infratographer.com/resource-provider-api/internal/testclient"
)

const (
	organizationalUnitPrefix = "testtnt"
)

var (
	TestDBURI   = os.Getenv("RESOURCEPROVIDERAPI_TESTDB_URI")
	EntClient   *ent.Client
	DBContainer *containersx.DBContainer
)

func TestMain(m *testing.M) {
	// setup the database if needed
	setupDB()
	// run the tests
	code := m.Run()
	// teardown the database
	teardownDB()
	// return the test response code
	os.Exit(code)
}

func parseDBURI(ctx context.Context) (string, string, *containersx.DBContainer) {
	switch {
	// if you don't pass in a database we default to an in memory sqlite
	case TestDBURI == "":
		return dialect.SQLite, "file:ent?mode=memory&cache=shared&_fk=1", nil
	case strings.HasPrefix(TestDBURI, "sqlite://"):
		return dialect.SQLite, strings.TrimPrefix(TestDBURI, "sqlite://"), nil
	case strings.HasPrefix(TestDBURI, "postgres://"), strings.HasPrefix(TestDBURI, "postgresql://"):
		return dialect.Postgres, TestDBURI, nil
	case strings.HasPrefix(TestDBURI, "docker://"):
		dbImage := strings.TrimPrefix(TestDBURI, "docker://")

		switch {
		case strings.HasPrefix(dbImage, "cockroach"), strings.HasPrefix(dbImage, "cockroachdb"), strings.HasPrefix(dbImage, "crdb"):
			cntr, err := containersx.NewCockroachDB(ctx, dbImage)
			errPanic("error starting db test container", err)

			return dialect.Postgres, cntr.URI, cntr
		case strings.HasPrefix(dbImage, "postgres"):
			cntr, err := containersx.NewPostgresDB(ctx, dbImage,
				postgres.WithInitScripts(filepath.Join("testdata", "postgres_init.sh")),
			)
			errPanic("error starting db test container", err)

			return dialect.Postgres, cntr.URI, cntr
		default:
			panic("invalid testcontainer URI, uri: " + TestDBURI)
		}

	default:
		panic("invalid DB URI, uri: " + TestDBURI)
	}
}

func setupDB() {
	// don't setup the datastore if we already have one
	if EntClient != nil {
		return
	}

	ctx := context.Background()

	dia, uri, cntr := parseDBURI(ctx)

	c, err := ent.Open(dia, uri, ent.Debug())
	if err != nil {
		errPanic("failed terminating test db container after failing to connect to the db", cntr.Container.Terminate(ctx))
		errPanic("failed opening connection to database:", err)
	}

	switch dia {
	case dialect.SQLite:
		// Run automatic migrations for SQLite
		errPanic("failed creating db scema", c.Schema.Create(ctx))
	case dialect.Postgres:
		log.Println("Running database migrations")

		cmd := exec.Command("atlas", "migrate", "apply",
			"--dir", "file://../../db/migrations",
			"--url", uri,
		)

		// write all output to stdout and stderr as it comes through
		var stdBuffer bytes.Buffer
		mw := io.MultiWriter(os.Stdout, &stdBuffer)

		cmd.Stdout = mw
		cmd.Stderr = mw

		// Execute the command
		errPanic("atlas returned an error running database migrations", cmd.Run())
	}

	EntClient = c
}

func teardownDB() {
	ctx := context.Background()

	if EntClient != nil {
		errPanic("teardown failed to close database connection", EntClient.Close())
	}

	if DBContainer != nil {
		errPanic("teardown failed to terminate test db container", DBContainer.Container.Terminate(ctx))
	}
}

func errPanic(msg string, err error) {
	if err != nil {
		log.Panicf("%s err: %s", msg, err.Error())
	}
}

type c struct {
	srvURL     string
	httpClient *http.Client
}

type clientOptions func(*c)

func graphTestClient(options ...clientOptions) testclient.TestClient {
	g := &c{
		srvURL: "graph",
		httpClient: &http.Client{Transport: localRoundTripper{handler: handler.NewDefaultServer(
			api.NewExecutableSchema(
				api.Config{Resolvers: api.NewResolver(EntClient, zap.NewNop().Sugar())},
			))}},
	}

	for _, opt := range options {
		opt(g)
	}

	return testclient.NewClient(g.httpClient, g.srvURL)
}

// localRoundTripper is an http.RoundTripper that executes HTTP transactions
// by using handler directly, instead of going over an HTTP connection.
type localRoundTripper struct {
	handler http.Handler
}

func (l localRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	w := httptest.NewRecorder()
	l.handler.ServeHTTP(w, req)

	return w.Result(), nil
}

func newTestServer(authConfig *echojwtx.AuthConfig) (*httptest.Server, error) { //nolint:unused
	echoCfg := echox.Config{}

	if authConfig != nil {
		auth, err := echojwtx.NewAuth(context.Background(), *authConfig)
		if err != nil {
			return nil, err
		}

		echoCfg = echoCfg.WithMiddleware(auth.Middleware())
	}

	srv, err := echox.NewServer(zap.NewNop(), echoCfg, nil)
	if err != nil {
		return nil, err
	}

	r := api.NewResolver(EntClient, zap.NewNop().Sugar())
	srv.AddHandler(r.Handler(false))

	return httptest.NewServer(srv.Handler()), nil
}

func newString(s string) *string {
	return &s
}

type ResourceProviderBuilder struct {
	Name                 string
	Description          string
	OrganizationalUnitID gidx.PrefixedID
}

func (t *ResourceProviderBuilder) MustNew(ctx context.Context) *ent.ResourceProvider {
	resourceProviderCreate := EntClient.ResourceProvider.Create()

	if t.Name == "" {
		t.Name = gofakeit.Name()
	}

	resourceProviderCreate.SetName(t.Name)

	if t.Description != "" {
		resourceProviderCreate.SetDescription(t.Description)
	}

	if t.OrganizationalUnitID == "" {
		t.OrganizationalUnitID = gidx.MustNewID(organizationalUnitPrefix)
	}

	resourceProviderCreate.SetOrganizationalUnitID(t.OrganizationalUnitID)

	return resourceProviderCreate.SaveX(ctx)
}
