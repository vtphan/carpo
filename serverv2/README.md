### Migrations:
UP: migrate -path migrations -database $DATABASE_URL -verbose up
DOWN: migrate -path migrations -database $DATABASE_URL -verbose down
DOWN 1 step: migrate -path migrations -database $DATABASE_URL -verbose down 1
Create New Migration: migrate create -ext sql -dir migrations -seq <give some name>

