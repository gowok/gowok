package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"runtime"
	"time"

	"github.com/gowok/gowok"
	gowok_json "github.com/gowok/gowok/json"
	"github.com/gowok/gowok/web"
	"github.com/gowok/gowok/web/request"
	"github.com/gowok/gowok/web/response"

	// _ "github.com/mattn/go-sqlite3"
	"github.com/ngamux/ngamux"
	"github.com/spf13/cobra"
)

type broker struct {
	clients []chan struct{}
}

var b broker

func (b broker) Send() {
	for _, c := range b.clients {
		c <- struct{}{}
	}
}

var cmdRoot = &cobra.Command{
	Run: func(cmd *cobra.Command, args []string) {
	},
}

type User struct {
	ID       string
	Email    string
	Password string
}

func ConfigureDB() {
	db := gowok.SQL.Conn().OrPanic()

	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			email TEXT NOT NULL,
			password TEXT NOT NULL
		)`)
	if err != nil {
		panic(err)
	}
}

func ConfigureRoute() {
	db := gowok.SQL.Conn().OrPanic()
	r := gowok.Web
	r.Get("/users", func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.QueryContext(r.Context(), "SELECT * FROM users")
		if err != nil {
			response.New(w).InternalServerError(err)
			return
		}
		defer rows.Close()

		users := make([]User, 0)
		for rows.Next() {
			user := User{}
			err := rows.Scan(&user.ID, &user.Email, &user.Password)
			if err != nil {
				continue
			}
			users = append(users, user)
		}

		response.New(w).Ok(users)
	})

	r.Post("/users", func(w http.ResponseWriter, r *http.Request) {
		input := User{}
		err := request.New(r).JSON(&input)
		if err != nil {
			response.New(w).BadRequest(err)
			return
		}

		_, err = db.ExecContext(r.Context(), `INSERT INTO users VALUES(?, ?, ?)`,
			input.ID,
			input.Email,
			input.Password,
		)
		if err != nil {
			response.New(w).InternalServerError(err)
			return
		}

		response.New(w).Created()
	})
	r.Put("/users/{id}", func(w http.ResponseWriter, r *http.Request) {
	})

	r.Delete("/users/{id}", func(w http.ResponseWriter, r *http.Request) {
		_, err := db.ExecContext(r.Context(),
			`DELETE FROM users WHERE id = ?`,
			r.PathValue("id"),
		)
		if err != nil {
			response.New(w).InternalServerError(err)
			return
		}

		response.New(w).Created()
	})
}

var cmdRun = &cobra.Command{
	Use: "run",
	Run: func(cmd *cobra.Command, args []string) {
		b = broker{}
		// gowok.Router().Get("/events", gowok.HandlerSse(sseHandler))
		// gowok.Router().Get("/fire", func(w http.ResponseWriter, r *http.Request) {
		// 	b.Send()
		// 	gowok.HttpOk(w, "mantap")
		//
		// 	p := gowok.PaginationFromReq[broker](r)
		// 	gowok.PaginationFromPagination[int](p)
		// })
		//
		// gowok.Router().Get("/ping", gowok.Handler(func(ctx *web.Ctx) error {
		// 	return errors.New("mantap")
		// }))

		v2 := gowok.Web.Group("/v2")
		v2.GroupFunc("/users", func(mux *ngamux.HttpServeMux) {
			mux.Get("", func(w http.ResponseWriter, r *http.Request) {
				rows, err := gowok.SQL.Conn().OrPanic().Query("SELECT 1")
				if err != nil {
					response.New(w).BadRequest(err)
					return
				}
				defer rows.Close()

				var res int
				for rows.Next() {
					rows.Scan(&res)
				}

				password := "123123"
				passwordHashed := gowok.Hash.Password(password)

				gowok.Web.Response.New(w).Ok(ngamux.Map{
					"sql":            res,
					"passwordHashed": passwordHashed,
					"config":         gowok.Config,
					"configMap":      gowok.Config.Map(),
				})
			})
		})
		v2.GroupFunc("/chats", func(mux *ngamux.HttpServeMux) {
			mux.Get("", func(w http.ResponseWriter, r *http.Request) {
				pagination := web.PaginationFromReq[int](r)
				response.New(w).JSON(pagination)
			})
		})
		v2.Get("/gc", func(w http.ResponseWriter, r *http.Request) {
			runtime.GC()
		})

		gowok.Web.Resource.New("/products", products{}, gowok.Web.Resource.WithMiddleware())

		gowok.Net.HandleFunc(func(conn net.Conn) {
			defer conn.Close()

			reader := bufio.NewReader(conn)
			for {
				msg, err := reader.ReadString('\n')
				if err != nil {
					fmt.Println(err)
					break
				}

				fmt.Println(msg)
			}
		})

		gowok.Configures(
			gowok_json.Configure(json.Marshal, json.Unmarshal),
		// ConfigureDB,
		// ConfigureRoute,
		).Run()
	},
}

type products struct{}

func (p products) Index(w http.ResponseWriter, r *http.Request) {
	gowok.Shutdown()
	go gowok.Run()
}
func (p products) Show(w http.ResponseWriter, r *http.Request)    {}
func (p products) Store(w http.ResponseWriter, r *http.Request)   {}
func (p products) Update(w http.ResponseWriter, r *http.Request)  {}
func (p products) Destroy(w http.ResponseWriter, r *http.Request) {}

var cmdMigrate = &cobra.Command{
	Use: "migrate",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.DebugFlags()
	},
}

// func main() {
// 	gowok.CMD.AddCommand(gowok.CMD.Wrap(cmdRoot))
// 	gowok.CMD.AddCommand(
// 		gowok.CMD.Wrap(cmdRun),
// 		cmdMigrate,
// 	)
// 	gowok.CMD.Execute()
// }

func main() {
	cmdRun.Run(cmdRun, []string{})
}

func sseHandler(ctx *web.CtxSse) {
	ctx.Publish([]byte("mantap " + time.Now().String()))
	client := make(chan struct{}, 1)
	b.clients = append(b.clients, client)

	pingTicker := time.NewTicker(3 * time.Second)
	for {
		select {
		case <-ctx.Done():
			fmt.Println("Client disconnected")
			return
		case <-client:
			ctx.Emit("asoy", []byte("mantap "+time.Now().String()))
		case <-pingTicker.C:
			fmt.Println(":keep-alive")
			ctx.PublishRaw(":keep-alive\n\n")
		}
	}
}
