package pg_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	authdomain "github.com/LeHuuHai/server-management/internal/domain/auth"
	apperr "github.com/LeHuuHai/server-management/internal/error"
	pg "github.com/LeHuuHai/server-management/internal/infra/postgres"
	"github.com/LeHuuHai/server-management/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	pgdriver "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// ---------------------------------------------------------------------------
// Suite setup
// ---------------------------------------------------------------------------

type PostgresIntegrationSuite struct {
	suite.Suite
	container   testcontainers.Container
	db          *gorm.DB
	serverRepo  *pg.ServerRepo
	accountRepo *pg.AccountRepo
}

func (s *PostgresIntegrationSuite) SetupSuite() {
	ctx := context.Background()

	container, err := tcpostgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:16-alpine"),
		tcpostgres.WithDatabase("testdb"),
		tcpostgres.WithUsername("test"),
		tcpostgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	require.NoError(s.T(), err)
	s.container = container

	host, err := container.Host(ctx)
	require.NoError(s.T(), err)
	port, err := container.MappedPort(ctx, "5432")
	require.NoError(s.T(), err)

	dsn := fmt.Sprintf("host=%s user=test password=test dbname=testdb port=%s sslmode=disable", host, port.Port())
	db, err := gorm.Open(pgdriver.Open(dsn), &gorm.Config{})
	require.NoError(s.T(), err)

	// migrate
	err = db.AutoMigrate(&model.Server{}, &model.Account{})
	require.NoError(s.T(), err)

	s.db = db
	s.serverRepo = pg.NewServerRepository(db)
	s.accountRepo = pg.NewAccountRepository(db)
}

func (s *PostgresIntegrationSuite) TearDownSuite() {
	s.container.Terminate(context.Background())
}

func (s *PostgresIntegrationSuite) SetupTest() {
	// xóa data trước mỗi test
	s.db.Exec("DELETE FROM servers")
	s.db.Exec("DELETE FROM accounts")
}

func TestPostgresIntegrationSuite(t *testing.T) {
	suite.Run(t, new(PostgresIntegrationSuite))
}

// ---------------------------------------------------------------------------
// ServerRepo tests
// ---------------------------------------------------------------------------

func (s *PostgresIntegrationSuite) TestServerRepo_Create_Success() {
	server := &model.Server{ServerID: "s1", ServerName: "Server1", IPv4: "1.2.3.4"}
	err := s.serverRepo.Create(context.Background(), server)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), model.StatusUnknown, server.Status)
}

func (s *PostgresIntegrationSuite) TestServerRepo_Create_Duplicate_Returns409() {
	server := &model.Server{ServerID: "s1", ServerName: "Server1", IPv4: "1.2.3.4"}
	require.NoError(s.T(), s.serverRepo.Create(context.Background(), server))

	err := s.serverRepo.Create(context.Background(), &model.Server{ServerID: "s1", ServerName: "Server1", IPv4: "1.2.3.4"})
	assert.ErrorIs(s.T(), err, apperr.ErrDuplicateServer)
}

func (s *PostgresIntegrationSuite) TestServerRepo_Update_Success() {
	require.NoError(s.T(), s.serverRepo.Create(context.Background(), &model.Server{
		ServerID: "s1", ServerName: "Server1", IPv4: "1.2.3.4",
	}))

	updated, err := s.serverRepo.Update(context.Background(), "s1", map[string]any{
		"server_name": "Updated",
		"ipv4":        "9.9.9.9",
	})

	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "Updated", updated.ServerName)
	assert.Equal(s.T(), "9.9.9.9", updated.IPv4)
}

func (s *PostgresIntegrationSuite) TestServerRepo_Update_NotFound() {
	_, err := s.serverRepo.Update(context.Background(), "ghost", map[string]any{"server_name": "X"})
	assert.ErrorIs(s.T(), err, apperr.ErrRecordNotFound)
}

func (s *PostgresIntegrationSuite) TestServerRepo_Delete_Success() {
	require.NoError(s.T(), s.serverRepo.Create(context.Background(), &model.Server{
		ServerID: "s1", ServerName: "Server1", IPv4: "1.2.3.4",
	}))

	err := s.serverRepo.Delete(context.Background(), "s1")
	assert.NoError(s.T(), err)

	// verify soft delete
	res, err := s.serverRepo.List(context.Background(), model.ListServerFilter{From: 0, To: 10, SortField: model.SortByName})
	assert.NoError(s.T(), err)
	assert.Empty(s.T(), res.Servers)
}

func (s *PostgresIntegrationSuite) TestServerRepo_Delete_NotFound() {
	err := s.serverRepo.Delete(context.Background(), "ghost")
	assert.ErrorIs(s.T(), err, apperr.ErrRecordNotFound)
}

func (s *PostgresIntegrationSuite) TestServerRepo_List_Success() {
	for i := 1; i <= 3; i++ {
		require.NoError(s.T(), s.serverRepo.Create(context.Background(), &model.Server{
			ServerID:   fmt.Sprintf("s%d", i),
			ServerName: fmt.Sprintf("Server%d", i),
			IPv4:       fmt.Sprintf("1.2.3.%d", i),
		}))
	}

	res, err := s.serverRepo.List(context.Background(), model.ListServerFilter{
		From: 0, To: 10, SortField: model.SortByName,
	})

	assert.NoError(s.T(), err)
	assert.Equal(s.T(), 3, res.Total)
	assert.Len(s.T(), res.Servers, 3)
}

func (s *PostgresIntegrationSuite) TestServerRepo_List_Pagination() {
	for i := 1; i <= 5; i++ {
		require.NoError(s.T(), s.serverRepo.Create(context.Background(), &model.Server{
			ServerID:   fmt.Sprintf("s%d", i),
			ServerName: fmt.Sprintf("Server%d", i),
			IPv4:       fmt.Sprintf("1.2.3.%d", i),
		}))
	}

	res, err := s.serverRepo.List(context.Background(), model.ListServerFilter{
		From: 0, To: 2, SortField: model.SortByName,
	})

	assert.NoError(s.T(), err)
	assert.Equal(s.T(), 5, res.Total) // total vẫn là 5
	assert.Len(s.T(), res.Servers, 2) // chỉ lấy 2
}

func (s *PostgresIntegrationSuite) TestServerRepo_CreateBatch_Success() {
	servers := []model.Server{
		{ServerID: "s1", ServerName: "Server1", IPv4: "1.2.3.4"},
		{ServerID: "s2", ServerName: "Server2", IPv4: "5.6.7.8"},
	}

	res, err := s.serverRepo.CreateBatch(context.Background(), servers)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), 2, res.SuccessCnt)
	assert.Equal(s.T(), 0, res.FailedCnt)
}

func (s *PostgresIntegrationSuite) TestServerRepo_CreateBatch_PartialFail() {
	// tạo s1 trước để gây duplicate
	require.NoError(s.T(), s.serverRepo.Create(context.Background(), &model.Server{
		ServerID: "s1", ServerName: "Server1", IPv4: "1.2.3.4",
	}))

	servers := []model.Server{
		{ServerID: "s1", ServerName: "Server1", IPv4: "1.2.3.4"}, // duplicate → fail
		{ServerID: "s2", ServerName: "Server2", IPv4: "5.6.7.8"}, // success
	}

	res, err := s.serverRepo.CreateBatch(context.Background(), servers)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), 1, res.SuccessCnt)
	assert.Equal(s.T(), 1, res.FailedCnt)
	assert.Contains(s.T(), res.Failed, "s1")
	assert.Contains(s.T(), res.Success, "s2")
}

func (s *PostgresIntegrationSuite) TestServerRepo_BulkUpdateServers_Success() {
	require.NoError(s.T(), s.serverRepo.Create(context.Background(), &model.Server{
		ServerID: "s1", ServerName: "Server1", IPv4: "1.2.3.4",
	}))

	now := time.Now()
	err := s.serverRepo.BulkUpdateServers(context.Background(), []model.Server{
		{ServerID: "s1", Status: model.StatusOnline, LastPingAt: now},
	})

	assert.NoError(s.T(), err)
}

func (s *PostgresIntegrationSuite) TestServerRepo_AllMetadata_Success() {
	require.NoError(s.T(), s.serverRepo.Create(context.Background(), &model.Server{
		ServerID: "s1", ServerName: "Server1", IPv4: "1.2.3.4",
	}))

	metadata, err := s.serverRepo.AllMetadata(context.Background())
	assert.NoError(s.T(), err)
	assert.Len(s.T(), metadata, 1)
	assert.Equal(s.T(), "s1", metadata[0].ServerID)
}

// ---------------------------------------------------------------------------
// AccountRepo tests
// ---------------------------------------------------------------------------

func (s *PostgresIntegrationSuite) TestAccountRepo_FindByUserName_Success() {
	s.db.Create(&model.Account{
		UserID: 1, Username: "hai", Password: "hashed", Role: authdomain.RoleAdmin,
	})

	account, err := s.accountRepo.FindByUserName("hai")
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "hai", account.Username)
}

func (s *PostgresIntegrationSuite) TestAccountRepo_FindByUserName_NotFound() {
	_, err := s.accountRepo.FindByUserName("nobody")
	assert.ErrorIs(s.T(), err, apperr.ErrRecordNotFound)
}

func (s *PostgresIntegrationSuite) TestAccountRepo_FindByUserID_Success() {
	s.db.Create(&model.Account{
		UserID: 99, Username: "hai", Password: "hashed", Role: authdomain.RoleAdmin,
	})

	account, err := s.accountRepo.FindByUserID(99)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), uint(99), account.UserID)
}

func (s *PostgresIntegrationSuite) TestAccountRepo_FindByUserID_NotFound() {
	_, err := s.accountRepo.FindByUserID(999)
	assert.ErrorIs(s.T(), err, apperr.ErrRecordNotFound)
}
