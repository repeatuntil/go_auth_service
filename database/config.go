package database

import (
	"auth_service/logger"
)

var (
	maxConn, maxIdle int = 25, 5
)

func ConfigDataBase(url, driver, name, migrDir string) (*dbConnection, error) {
	var conn = new(dbConnection) 

	if err := conn.Init(url, driver, name, maxConn, maxIdle); err != nil {
		logger.Err.Fatalln("Can't establish connection with database -", err)
		return nil, err
	}

	m := new(Migrator)
	if err := m.Init(conn, migrDir); err != nil {
		logger.Err.Fatalln("Can't create migrator -", err)
		return nil, err
	}

	if err := m.Apply(); err != nil {
		return nil, err
	}

	logger.Debug.Println("Successfully establish connection with db")
	return conn, nil
}
