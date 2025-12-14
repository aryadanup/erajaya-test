package repository

import (
	"context"
	"database/sql"
	"erajaya-test/internal/interfaces"
	"erajaya-test/internal/models/entity"
	"erajaya-test/internal/models/request"
	"erajaya-test/shared/constant"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type PostgresSuite struct {
	suite.Suite
	mock sqlmock.Sqlmock
	repo interfaces.ProductRepository
	db   *sql.DB
}

func (s *PostgresSuite) SetupTest() {
	var err error
	var gormDB *gorm.DB

	s.db, s.mock, err = sqlmock.New()
	s.Require().NoError(err)

	dialector := postgres.New(postgres.Config{
		Conn:       s.db,
		DriverName: "postgres",
	})
	gormDB, err = gorm.Open(dialector, &gorm.Config{})
	s.Require().NoError(err)

	s.repo = NewProductRepository(gormDB)
}

func (s *PostgresSuite) TearDownTest() {
	s.db.Close()
}

func (s *PostgresSuite) TestCreate() {
	price := int64(5000000)
	product := &entity.Product{Name: "LG TV", Price: &price}

	s.mock.ExpectBegin()
	s.mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "products"`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	s.mock.ExpectCommit()

	err := s.repo.Create(context.Background(), product)
	s.NoError(err)
}

func (s *PostgresSuite) TestGetByID() {

	columns := []string{"id", "name", "price", "description", "quantity", "created_by", "created_at", "updated_by", "updated_at", "deleted_by", "deleted_at"}

	s.Run("Found", func() {
		rows := sqlmock.NewRows(columns).
			AddRow(1, "LG TV", 5000000, "Desc", 10, "arya", time.Now(), nil, nil, nil, nil)

		s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "products" WHERE "products"."id" = $1 ORDER BY "products"."id" LIMIT $2`)).
			WithArgs(1, 1).
			WillReturnRows(rows)

		res, err := s.repo.GetByID(context.Background(), 1)

		if err != nil {
			s.T().Logf("Error Found Case: %v", err)
		}

		s.NoError(err)
		s.NotNil(res)
		s.Equal("LG TV", res.Name)
	})

	s.Run("Not Found", func() {
		s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "products" WHERE "products"."id" = $1 ORDER BY "products"."id" LIMIT $2`)).
			WithArgs(999, 1).
			WillReturnError(gorm.ErrRecordNotFound)

		res, err := s.repo.GetByID(context.Background(), 999)

		s.ErrorIs(err, constant.ErrNotFound)
		s.Nil(res)
	})

	s.Run("DB Error", func() {
		s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "products" WHERE "products"."id" = $1 ORDER BY "products"."id" LIMIT $2`)).
			WithArgs(1, 1).
			WillReturnError(sql.ErrConnDone)

		res, err := s.repo.GetByID(context.Background(), 1)
		s.Error(err)
		s.Nil(res)
	})
}

func (s *PostgresSuite) TestFetch() {
	filter := request.ProductFilter{
		Search: "LG",
		Sort:   "cheapest",
		Page:   1,
		Limit:  10,
	}

	s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "products" WHERE name ILIKE $1 OR description ILIKE $2`)).
		WithArgs("%LG%", "%LG%").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(10))

	rows := sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "LG TV")
	s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "products" WHERE name ILIKE $1 OR description ILIKE $2 ORDER BY price ASC LIMIT $3`)).
		WithArgs("%LG%", "%LG%", 10).
		WillReturnRows(rows)

	res, total, err := s.repo.Fetch(context.Background(), filter)
	s.NoError(err)
	s.Equal(int64(10), total)
	s.Len(res, 1)
}

func (s *PostgresSuite) TestFetchSort() {

	columns := []string{"id", "name", "price", "description", "quantity", "created_at"}

	tests := []struct {
		name            string
		sortParam       string
		expectedOrderBy string
	}{
		{
			name:            "Sort Newest",
			sortParam:       "newest",
			expectedOrderBy: "created_at DESC",
		},
		{
			name:            "Sort Cheapest",
			sortParam:       "cheapest",
			expectedOrderBy: "price ASC",
		},
		{
			name:            "Sort Expensive",
			sortParam:       "expensive",
			expectedOrderBy: "price DESC",
		},
		{
			name:            "Sort Name ASC",
			sortParam:       "name asc",
			expectedOrderBy: "name ASC",
		},
		{
			name:            "Sort Name DESC",
			sortParam:       "name desc",
			expectedOrderBy: "name DESC",
		},
		{
			name:            "Sort Default (Empty)",
			sortParam:       "",
			expectedOrderBy: "created_at DESC",
		},
		{
			name:            "Sort Default (Invalid)",
			sortParam:       "random_string",
			expectedOrderBy: "created_at DESC",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			filter := request.ProductFilter{
				Sort:  tc.sortParam,
				Page:  1,
				Limit: 10,
			}

			s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "products"`)).
				WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))

			expectedQueryRegex := `SELECT \* FROM "products" .*ORDER BY ` + regexp.QuoteMeta(tc.expectedOrderBy) + ` LIMIT \$1`

			s.mock.ExpectQuery(expectedQueryRegex).
				WithArgs(10).
				WillReturnRows(sqlmock.NewRows(columns).AddRow(1, "Test Item", 10000, "Desc", 5, time.Now()))

			products, _, err := s.repo.Fetch(context.Background(), filter)

			s.NoError(err)
			s.Len(products, 1)

			s.NoError(s.mock.ExpectationsWereMet())
		})
	}
}

func (s *PostgresSuite) TestFetchErrors() {
	filter := request.ProductFilter{
		Search: "LG",
		Page:   1,
		Limit:  10,
	}

	s.Run("Count Error", func() {
		s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "products" WHERE name ILIKE $1 OR description ILIKE $2`)).
			WithArgs("%LG%", "%LG%").
			WillReturnError(sql.ErrConnDone)

		res, total, err := s.repo.Fetch(context.Background(), filter)
		s.Error(err)
		s.Equal(int64(0), total)
		s.Nil(res)
	})

	s.Run("Find Error", func() {
		s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "products"`)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(10))

		s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "products" WHERE name ILIKE $1 OR description ILIKE $2 ORDER BY created_at DESC LIMIT $3`)).
			WithArgs("%LG%", "%LG%", 10).
			WillReturnError(sql.ErrConnDone)

		res, total, err := s.repo.Fetch(context.Background(), filter)
		s.Error(err)
		s.Equal(int64(0), total)
		s.Nil(res)
	})
}

func TestPostgresSuite(t *testing.T) {
	suite.Run(t, new(PostgresSuite))
}
