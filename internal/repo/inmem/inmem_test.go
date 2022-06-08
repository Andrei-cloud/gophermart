package inmem

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/andrei-cloud/gophermart/internal/repo"
)

func TestInMemRepoUser(t *testing.T) {
	createTests := []struct {
		name    string
		user    repo.User
		want    int64
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "Create: Unique user",
			user: repo.User{
				Username: "test",
				Password: "test",
			},
			want:    1,
			wantErr: false,
		},
		{
			name: "Create: Another unique user",
			user: repo.User{
				Username: "test1",
				Password: "test1",
			},
			want:    2,
			wantErr: false,
		},
		{
			name: "Create: user Existing name",
			user: repo.User{
				Username: "test1",
				Password: "test1",
			},
			want:    0,
			wantErr: true,
		},
	}

	r := NewInMemRepo()

	for _, tt := range createTests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := r.UserCreate(&tt.user)
			if tt.wantErr {
				require.Equal(t, repo.ErrAlreadyExists, err)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}

	getTests := []struct {
		name     string
		username string
		want     *repo.User
		wantErr  bool
	}{
		// TODO: Add test cases.
		{
			name:     "Get: Exiting user",
			username: "test",
			want: &repo.User{
				ID:         1,
				Username:   "test",
				Password:   "test",
				Balance:    0,
				Withdrawal: 0,
				CreatedAt:  time.Time{},
			},
			wantErr: false,
		},
		{
			name:     "Get: non exiting user",
			username: "test2",
			want:     nil,
			wantErr:  true,
		},
	}

	for _, tt := range getTests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := r.UserGet(tt.username)
			if tt.wantErr {
				require.Equal(t, repo.ErrNotExists, err)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}

	deleteTests := []struct {
		name     string
		username string
		wantErr  bool
	}{
		// TODO: Add test cases.
		{
			name:     "Delete: Exiting user",
			username: "test",
			wantErr:  false,
		},
		{
			name:     "Delete: non exiting user",
			username: "test2",
			wantErr:  true,
		},
		{
			name:     "Delete: deleted user",
			username: "test",
			wantErr:  true,
		},
	}

	for _, tt := range deleteTests {
		t.Run(tt.name, func(t *testing.T) {
			err := r.UserDelete(tt.username)
			if tt.wantErr {
				require.Equal(t, repo.ErrNotExists, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestInMemRepoOrder(t *testing.T) {
	createTests := []struct {
		name    string
		order   repo.Order
		want    int64
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "Create: Unique Order",
			order: repo.Order{
				Order:  "test",
				Type:   repo.CREDIT,
				UserID: 1,
				Value:  100,
				Status: repo.NEW,
			},
			want:    1,
			wantErr: false,
		},
		{
			name: "Create: Another unitq Order",
			order: repo.Order{
				Order:  "test1",
				Type:   repo.CREDIT,
				UserID: 1,
				Value:  100,
				Status: repo.NEW,
			},
			want:    2,
			wantErr: false,
		},
		{
			name: "Create: Order Existing numebr",
			order: repo.Order{
				Order:  "test1",
				Type:   repo.CREDIT,
				UserID: 1,
				Value:  100,
				Status: repo.NEW,
			},
			want:    0,
			wantErr: true,
		},
	}

	r := NewInMemRepo()

	for _, tt := range createTests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := r.OrderCreate(&tt.order)
			if tt.wantErr {
				require.Equal(t, repo.ErrAlreadyExists, err)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}

	getTests := []struct {
		name    string
		order   string
		want    *repo.Order
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name:  "Get: Exiting order",
			order: "test",
			want: &repo.Order{
				ID:     1,
				Order:  "test",
				Type:   repo.CREDIT,
				UserID: 1,
				Value:  100,
				Status: repo.NEW,
			},
			wantErr: false,
		},
		{
			name:    "Get: non exiting order",
			order:   "test2",
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range getTests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := r.OrderGet(tt.order)
			if tt.wantErr {
				require.Equal(t, repo.ErrNotExists, err)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}

	deleteTests := []struct {
		name    string
		order   string
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name:    "Delete: Exiting order",
			order:   "test",
			wantErr: false,
		},
		{
			name:    "Delete: non exiting order",
			order:   "test2",
			wantErr: true,
		},
		{
			name:    "Delete: deleted order",
			order:   "test",
			wantErr: true,
		},
	}

	for _, tt := range deleteTests {
		t.Run(tt.name, func(t *testing.T) {
			err := r.OrderDelete(tt.order)
			if tt.wantErr {
				require.Equal(t, repo.ErrNotExists, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
