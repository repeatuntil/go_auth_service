package database

import (
	"auth_service/logger"
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

type dbConnection struct {
	db *sql.DB
	isActive bool
	driver string
	baseUrl string
}

var Conn = new(dbConnection)

func (c *dbConnection) Init(url, driver, dbName string, maxConn int, maxIdleTime int) error {
	if c.isActive {
		return fmt.Errorf("this connection has been already initialized")
	}

	c.baseUrl = url
	c.driver = driver

	logger.Debug.Println("dbURL -", c.baseUrl)
	if err := c.createDBIfNotExists(dbName); err != nil {
		logger.Debug.Printf("мудила\n")
		return err
	}

	open, err := sql.Open(driver, fmt.Sprintf(c.baseUrl, dbName))

	if err != nil {
		logger.Err.Fatalln("Connection refused with database")
		return err
	}

	idleCoef := 5

	c.db = open
	c.isActive = true
	c.db.SetMaxOpenConns(maxConn)
	c.db.SetMaxIdleConns(maxConn / idleCoef)
	c.db.SetConnMaxIdleTime(time.Duration(maxIdleTime) * time.Minute)

	return c.db.Ping()
}

func (c *dbConnection) createDBIfNotExists(dbName string) error {
	baseDelay := 1 * time.Second
	maxAttempts := 5
	var genConn *sql.DB
	var err error
	for attempt := 0; attempt < maxAttempts; attempt++ {
        logger.Debug.Printf("Attempt %d\n", attempt+1)
        
        genConn, err = sql.Open(c.driver, fmt.Sprintf(c.baseUrl, ""))
        if err == nil {
            ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
            err = genConn.PingContext(ctx)
            cancel()
            
            if err == nil {
                break 
            }
        }

        logger.Debug.Printf("Database connection attempt %d failed: %v\n", attempt+1, err)
        
        if attempt == maxAttempts - 1 {
            return fmt.Errorf("failed after %d attempts: %v", maxAttempts, err)
        }

        waitTime := baseDelay * time.Duration(1<<uint(attempt))
        time.Sleep(waitTime)
    }
	
	defer genConn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2 * time.Second)
	defer cancel()

	var exists bool
	query := fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = '%s')", dbName)
	
    if err := genConn.QueryRowContext(ctx, query).Scan(&exists); err != nil {
		return err
	}

	if !exists {
		if _, err := genConn.ExecContext(ctx, "CREATE DATABASE " + dbName); err != nil {
			logger.Err.Fatalln("Database creation failed -", err.Error())
			return err
		}
	}

	return nil
}

func (c *dbConnection) Close() {
	c.db.Close()
}
