// Package dbtest contains supporting code for running tests that hit the DB.
package dbtest

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gorm.io/gorm"

	dbUser "github.com/rmsj/service/business/core/user/db"
	"github.com/rmsj/service/business/data/dbschema"
	"github.com/rmsj/service/business/sys/database"
	"github.com/rmsj/service/business/web/auth"
	"github.com/rmsj/service/foundation/docker"
	"github.com/rmsj/service/foundation/keystore"
)

// Success and failure markers.
const (
	Success = "\u2713"
	Failed  = "\u2717"
	dbPort  = 5432
)

// StartDB starts a database instance.
func StartDB() (*docker.Container, error) {
	image := "postgres:14-alpine"
	port := strconv.Itoa(dbPort)
	args := []string{"-e", "POSTGRES_PASSWORD=postgres"}

	container, err := docker.StartContainer(image, port, args...)
	// having trouble connecting to DB not yet ready
	time.Sleep(time.Second * 5)
	return container, err
}

// StopDB stops a running database instance.
func StopDB(c *docker.Container) {
	docker.StopContainer(c.ID)
}

// NewUnit creates a test database inside a Docker container. It creates the
// required table structure but the database is otherwise empty. It returns
// the database to use as well as a function to call at the end of the test.
func NewUnit(t *testing.T, c *docker.Container, dbName string) (*zap.SugaredLogger, *gorm.DB, func()) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// log
	var buf bytes.Buffer
	encoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
	writer := bufio.NewWriter(&buf)
	log := zap.New(
		zapcore.NewCore(encoder, zapcore.AddSync(writer), zapcore.DebugLevel)).
		Sugar()

	// database
	dbM, err := database.Open(database.Config{
		User:       "postgres",
		Password:   "postgres",
		Host:       c.Host,
		Port:       c.Port,
		Name:       "postgres",
		DisableTLS: true,
		Logger:     log,
		Debug:      true,
	})
	if err != nil {
		t.Fatalf("Opening database connection: %v", err)
	}

	t.Log("Waiting for database to be ready ...")

	if err := database.StatusCheck(ctx, dbM); err != nil {
		t.Fatalf("status check database: %v", err)
	}

	t.Log("Database ready")

	if err := database.ExecDDL(context.Background(), dbM, log, "CREATE DATABASE "+dbName, nil); err != nil {
		t.Fatalf("creating database %s: %v", dbName, err)
	}
	database.Close(dbM)

	// =========================================================================

	db, err := database.Open(database.Config{
		User:       "postgres",
		Password:   "postgres",
		Host:       c.Host,
		Port:       c.Port,
		Name:       dbName,
		DisableTLS: true,
		Logger:     log,
		Debug:      true,
	})
	if err != nil {
		t.Fatalf("Opening database connection: %v", err)
	}

	t.Log("Migrate and seed database ...")

	if err := dbschema.Migrate(ctx, db); err != nil {
		docker.DumpContainerLogs(t, c.ID)
		t.Fatalf("Migrating error: %s", err)
	}

	if err := dbschema.Seed(ctx, db); err != nil {
		docker.DumpContainerLogs(t, c.ID)
		t.Fatalf("Seeding error: %s", err)
	}

	t.Log("Ready for testing ...")

	// teardown is the function that should be invoked when the caller is done
	// with the database.
	teardown := func() {
		t.Helper()
		database.Close(db)

		log.Sync()

		writer.Flush()
		fmt.Println("******************** LOGS ********************")
		fmt.Print(buf.String())
		fmt.Println("******************** LOGS ********************")
	}

	return log, db, teardown
}

// Test owns state for running and shutting down tests.
type Test struct {
	DB       *gorm.DB
	Log      *zap.SugaredLogger
	Auth     *auth.Auth
	Teardown func()

	t *testing.T
}

// NewIntegration creates a database, seeds it, constructs an authenticator.
func NewIntegration(t *testing.T, c *docker.Container, dbName string) *Test {
	log, db, teardown := NewUnit(t, c, dbName)

	// Create RSA keys to enable authentication in our service.
	keyID := "4754d86b-7a6d-4df5-9c65-224741361492"
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}

	// Build an authenticator using this private key and id for the key store.
	auth, err := auth.New(keyID, keystore.NewMap(map[string]*rsa.PrivateKey{keyID: privateKey}))
	if err != nil {
		t.Fatal(err)
	}

	test := Test{
		DB:       db,
		Log:      log,
		Auth:     auth,
		t:        t,
		Teardown: teardown,
	}

	return &test
}

// Token generates an authenticated token for a user.
func (test *Test) Token(email, pass string) string {
	test.t.Log("Generating token for test ...")

	dbUsr, err := dbUser.QueryByEmail(context.Background(), test.DB, email)
	if err != nil {
		return ""
	}

	claims := auth.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   dbUsr.UserID,
			Issuer:    "service project",
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		},
		Roles: dbUsr.Roles,
	}

	token, err := test.Auth.GenerateToken(claims)
	if err != nil {
		test.t.Fatal(err)
	}

	return token
}

// StringPointer is a helper to get a *string from a string. It is in the tests
// package because we normally don't want to deal with pointers to basic types
// but it's useful in some tests.
func StringPointer(s string) *string {
	return &s
}

// IntPointer is a helper to get a *int from a int. It is in the tests package
// because we normally don't want to deal with pointers to basic types but it's
// useful in some tests.
func IntPointer(i int) *int {
	return &i
}
