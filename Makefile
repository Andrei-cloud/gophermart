
runaccrual:
	./cmd/accrual/accrual_darwin_amd64 -a "localhost:9090" -d "postgres://postgres:rootpassword@localhost:5432/gophermart"
.PHONY:
	runaccrual
	