package postgres

import (
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// функция миграции
func RunMigrations(databaseURL string) error {
	// Указываем путь к папке с миграциями и строку подключения
	m, err := migrate.New(
		"file://db/migrations",
		databaseURL,
	)
	if err != nil {
		return err
	}

	// Выполняем все up миграции
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}

	return nil
}
