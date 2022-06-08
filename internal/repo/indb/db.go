package indb

import (
	"context"
	"database/sql"
	"os"
	"time"

	"github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4/stdlib"

	"github.com/andrei-cloud/gophermart/internal/repo"
	"github.com/rs/zerolog/log"
)

var _ repo.Repository = &dbRepo{}

type dbRepo struct {
	db *sql.DB
}

func NewDB(dsn string) *dbRepo {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatal().AnErr("NewDB", err)
		os.Exit(1)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		log.Fatal().AnErr("NewDB", err)
		os.Exit(1)
	}
	log.Debug().Msg("create schema if not already exists")
	if err := createTables(ctx, db); err != nil {
		log.Fatal().AnErr("NewDB", err)
		os.Exit(1)
	}
	return &dbRepo{db: db}
}

func createTables(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx,
		`CREATE TABLE IF NOT EXISTS "users" (
		"id" BIGSERIAL PRIMARY KEY,
		"username" varchar,
		"password" varchar,
		"balance" float,
		"withdrawn" float,
		"created_at" timestamp
	  );`)
	if err != nil {
		return err
	}

	_, err = db.ExecContext(ctx,
		`CREATE UNIQUE INDEX CONCURRENTLY IF NOT EXISTS unique_username 
		ON users ("username");`)
	if err != nil {
		return err
	}

	_, err = db.ExecContext(ctx,
		`ALTER TABLE users 
		ADD CONSTRAINT unique_username 
		UNIQUE USING INDEX unique_username;`)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); !ok {
			return err
		} else if pgErr.Code != "55000" {
			return err
		}

	}

	_, err = db.ExecContext(ctx,
		`CREATE TABLE IF NOT EXISTS "orders" (
			"id" BIGSERIAL PRIMARY KEY,
			"number" varchar,
			"type" varchar,
			"user_id" bigint,
			"value" float,
			"status" varchar,
			"uploaded_at" timestamp
		  );
		  `)
	if err != nil {
		return err
	}

	_, err = db.ExecContext(ctx,
		`ALTER TABLE "orders" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id");`)
	if err != nil {
		return err
	}

	_, err = db.ExecContext(ctx,
		`CREATE UNIQUE INDEX CONCURRENTLY IF NOT EXISTS unique_order 
		ON orders("number");`)
	if err != nil {
		return err
	}

	_, err = db.ExecContext(ctx,
		`ALTER TABLE orders
		ADD CONSTRAINT unique_order 
		UNIQUE USING INDEX unique_order;`)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); !ok {
			return err
		} else if pgErr.Code != "55000" {
			return err
		}

	}

	return nil
}

func checkUserExists(db *sql.DB, u *repo.User) bool {
	var count int
	err := db.QueryRow("SELECT count(1) FROM users WHERE username=$1", u.Username).Scan(&count)
	if err != nil || count > 0 {
		return true
	}
	return false
}

func (r *dbRepo) UserCreate(u *repo.User) (int64, error) {
	if checkUserExists(r.db, u) {
		return 0, repo.ErrAlreadyExists
	}

	var id int64
	err := r.db.QueryRow(`
	INSERT INTO users(username, password, balance, withdrawn, created_at) 
	VALUES ($1, $2, $3, $4, $5) 
	RETURNING id`,
		u.Username, u.Password, 0, 0, time.Now()).
		Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}
func (r *dbRepo) UserGet(username string) (*repo.User, error) {
	user := repo.User{}
	err := r.db.QueryRow(`
	SELECT id, username, password, balance, withdrawn FROM users
	WHERE username=$1`,
		username).
		Scan(&user.ID, &user.Username, &user.Password, &user.Balance, &user.Withdrawal)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *dbRepo) UserUpdate(u *repo.User) error {
	_, err := r.db.Exec(`
		UPDATE users SET balance = $2, withdrawn = $3
		WHERE id=$1`,
		u.ID, u.Balance, u.Withdrawal)
	if err != nil {
		return err
	}
	return nil
}

func (r *dbRepo) UserGetByID(id int64) (*repo.User, error) {
	user := repo.User{}
	err := r.db.QueryRow(`
	SELECT id, username, password, balance, withdrawn FROM users
	WHERE id=$1`,
		id).
		Scan(&user.ID, &user.Username, &user.Password, &user.Balance, &user.Withdrawal)
	if err != nil {
		return nil, err
	}
	return &user, nil
}
func (r *dbRepo) UserDelete(string) error { return nil }

func (r *dbRepo) OrderCreate(o *repo.Order) (int64, error) {
	var id int64
	err := r.db.QueryRow(`
	INSERT INTO orders(number, type, user_id, value, status, uploaded_at) 
	VALUES ($1, $2, $3, $4, $5, $6) 
	RETURNING id`,
		o.Order, string(o.Type), o.UserID, o.Value, string(o.Status), o.UploadedAt).
		Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (r *dbRepo) OrderGet(number string) (*repo.Order, error) {
	order := repo.Order{}
	err := r.db.QueryRow(`
		SELECT id, number, type, user_id, value, status, uploaded_at
		FROM orders
		WHERE number=$1`,
		number).
		Scan(&order.ID, &order.Order, &order.Type,
			&order.UserID, &order.Value, &order.Status,
			&order.UploadedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, repo.ErrNotExists
		}
		return nil, err
	}
	return &order, nil
}
func (r *dbRepo) OrderGetList(uid int64, t repo.OrderType) ([]repo.Order, error) {
	orders := make([]repo.Order, 0)
	rows, err := r.db.Query(`
		SELECT id, number, type, value, status, uploaded_at
		FROM orders
		WHERE user_id=$1 and type= $2`,
		uid, t)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		order := repo.Order{}
		err := rows.Scan(&order.ID, &order.Order, &order.Type, &order.Value, &order.Status, &order.UploadedAt)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return orders, nil
}
func (r *dbRepo) OrderDelete(string) error { return nil }

func (r *dbRepo) OrderToProcess() ([]string, error) {
	orders := make([]string, 0)
	rows, err := r.db.Query(`
		SELECT number 
		FROM orders o 
		WHERE o.status NOT IN ('PROCESSED', 'INVALID', '');`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var number string
		err := rows.Scan(&number)
		if err != nil {
			return nil, err
		}
		orders = append(orders, number)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return orders, nil
}

func (r *dbRepo) OrderUpdate(number string, status repo.OrderStatus, accrual float64) error {
	order, err := r.OrderGet(number)
	if err != nil {
		return err
	}
	order.Value = accrual
	order.Status = status
	_, err = r.db.Exec(`
		UPDATE orders SET value = $2, status = $3
		WHERE id=$1`,
		order.ID, order.Value, order.Status)
	if err != nil {
		return err
	}

	user, err := r.UserGetByID(order.UserID)
	if err != nil {
		return err
	}

	user.Balance += accrual

	err = r.UserUpdate(user)
	if err != nil {
		return err
	}

	return nil
}
