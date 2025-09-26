package dbfns


import (
    "context"
    "database/sql"
    "fmt"
    "log"
    "os"
    "testing"
)
import (
    _ "github.com/lib/pq"
    tc "github.com/testcontainers/testcontainers-go"
    "github.com/testcontainers/testcontainers-go/wait"
)


var (
    CONTAINER   tc.Container
    DB          *sql.DB
)


func StartPostgresContainer() (tc.Container, string, error) {
    // ???
    ctx := context.Background()

    // Initalize things that are needed to start docker
    req := tc.ContainerRequest{
        Image:          "postgres:15",
        ExposedPorts:   []string{"5432/tcp"},
        Env: map[string]string{
            "POSTGRES_USER":        "test_user",
            "POSTGRES_PASSWORD":    "test_password_8hst3",
            "POSTGRES_DB":          "test_db",
        },
        WaitingFor: wait.ForListeningPort("5432/tcp"),
    }

    // Start container
    container, err := tc.GenericContainer(ctx, tc.GenericContainerRequest{
        ContainerRequest: req,
        Started:          true,
    })
    if err != nil {
        return nil, "", fmt.Errorf("Could not start container: %w", err)
    }

    // Setting up host and port to container ???
    host, _ := container.Host(ctx)
    port, _ := container.MappedPort(ctx, "5432")
    // Formatting URL for db
    dbURL := fmt.Sprintf("postgres://test_user:test_password_8hst3@%s:%s/test_db?sslmode=disable", host, port.Port())

    log.Printf("Postgres container started, url: %s", dbURL)
    return container, dbURL, nil
}


func ExecSQLFile(db *sql.DB, filePath string) {
    sqlBytes, err := os.ReadFile(filePath)
    if err != nil {
        log.Fatalf("Failed to read %s: %v", filePath, err)
    }
    if _, err := db.Exec(string(sqlBytes)); err != nil {
        log.Fatalf("Failed to execute %s: %v", filePath, err)
    }

    fmt.Printf("Executed SQL file: %s", filePath)
}


func TestMain(m *testing.M) {
    cont, url, err := StartPostgresContainer()
    if err != nil {
        log.Fatalf("Failed to start Postgres container: %v", err)
    }
    CONTAINER = cont

    // Connect to DB
    db, err := sql.Open("postgres", url)
    if err != nil {
        log.Fatalf("Failed to connect to DB: %v", err)
    }
    DB = db

    // Initialize DB
    ExecSQLFile(DB, "init.sql")

    // Run tests
    code := m.Run()
    // Clean up
    DB.Close()
    CONTAINER.Terminate(context.Background())
    os.Exit(code)
}


