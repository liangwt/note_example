package internal

import (
	"database/sql"
	"log"

	"github.com/go-redis/redis"
	_ "github.com/go-sql-driver/mysql"
	"go.uber.org/dig"
)

type PostGateway struct {
	rwDB *sql.DB
}

type CommentGateway struct {
	rwDB *sql.DB
}

type UserGateway struct {
	roDB  *sql.DB
	rwDB  *sql.DB
	cache *redis.Client
}

func (g *UserGateway) GetUserName(id string) (name string, err error) {
	if g.cache != nil {
		name := g.cache.Get(id).Val()
		if name != "" {
			return name, nil
		}
	}

	row := g.roDB.QueryRow("select name from users")
	err = row.Scan(&name)
	if err != nil && g.cache != nil {
		g.cache.Set(id, name, -1)
	}
	return
}

type Option struct {
	driver string
	dsn    string
}

func Init() *dig.Container {
	c := dig.New()

	err := c.Provide(func() *Option {
		return &Option{
			driver: "mysql",
			dsn:    "user:password@/dbname",
		}
	})
	if err != nil {
		// ...
	}

	// 读写库
	// err = c.Provide(func(opt *Option) (*sql.DB, error) {
	// 	return sql.Open(opt.driver, opt.dsn)
	// }, dig.Name("rw"))
	// if err != nil {
	// 	// ...
	// }

	// // 只读库
	// err = c.Provide(func() (*sql.DB, error) {
	// 	return sql.Open("mysql", "user:password@/ro_dbname")
	// }, dig.Name("ro"))
	// if err != nil {
	// 	// ...
	// }

	////////////////以下和上文代码等价////////////////
	type DBResult struct {
		dig.Out
		
		RWDB *sql.DB `name:"rw"`
		RODB *sql.DB `name:"ro"`
	}

	err = c.Provide(func(opt *Option) (DBResult, error) {
		rw, err := sql.Open(opt.driver, opt.dsn)
		if err != nil {
			return DBResult{}, err
		}

		ro, err := sql.Open("mysql", "user:password@/ro_dbname")
		if err != nil {
			return DBResult{}, err
		}

		return DBResult{RWDB: rw, RODB: ro}, nil
	})
	if err != nil {
		// ...
	}
	////////////////////////////

	err = c.Provide(func() (*log.Logger, error) {
		return log.Default(), nil
	})
	if err != nil {
		// ...
	}

	// err = c.Provide(func(Logger *log.Logger, rwDB, roDB *sql.DB) (
	// 	Comments *CommentGateway,
	// 	Posts *PostGateway,
	// 	Users *UserGateway,
	// 	err error,
	// ) {
	// 	return &CommentGateway{rwDB: rwDB},
	// 		&PostGateway{rwDB: rwDB},
	// 		&UserGateway{rwDB: rwDB, roDB: roDB},
	// 		nil
	// })
	// if err != nil {
	// 	// ...
	// }

	type Connection struct {
		dig.In

		Logger *log.Logger
		Cache  *redis.Client `optional:"true"`
		RODB   *sql.DB       `name:"ro"`
		RWDB   *sql.DB       `name:"rw"`
	}

	type Gateways struct {
		dig.Out

		Comments *CommentGateway
		Posts    *PostGateway
		Users    *UserGateway
	}
	err = c.Provide(func(conn Connection) (Gateways, error) {
		return Gateways{
			Comments: &CommentGateway{rwDB: conn.RWDB},
			Posts:    &PostGateway{rwDB: conn.RWDB},
			Users:    &UserGateway{rwDB: conn.RWDB, roDB: conn.RODB, cache: conn.Cache},
		}, nil
	})
	if err != nil {
		// ...
	}

	return c
}
