package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/google/generative-ai-go/genai"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/plugins/ghupdate"
	"github.com/pocketbase/pocketbase/plugins/jsvm"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"
	"google.golang.org/api/option"
)

func main() {
	app := pocketbase.New()

	// ---------------------------------------------------------------
	// Optional plugin flags:
	// ---------------------------------------------------------------

	var hooksDir string
	app.RootCmd.PersistentFlags().StringVar(
		&hooksDir,
		"hooksDir",
		"",
		"the directory with the JS app hooks",
	)

	var hooksWatch bool
	app.RootCmd.PersistentFlags().BoolVar(
		&hooksWatch,
		"hooksWatch",
		true,
		"auto restart the app on pb_hooks file change",
	)

	var hooksPool int
	app.RootCmd.PersistentFlags().IntVar(
		&hooksPool,
		"hooksPool",
		25,
		"the total prewarm goja.Runtime instances for the JS app hooks execution",
	)

	var migrationsDir string
	app.RootCmd.PersistentFlags().StringVar(
		&migrationsDir,
		"migrationsDir",
		"",
		"the directory with the user defined migrations",
	)

	var automigrate bool
	app.RootCmd.PersistentFlags().BoolVar(
		&automigrate,
		"automigrate",
		true,
		"enable/disable auto migrations",
	)

	var publicDir string
	app.RootCmd.PersistentFlags().StringVar(
		&publicDir,
		"publicDir",
		defaultPublicDir(),
		"the directory to serve static files",
	)

	var indexFallback bool
	app.RootCmd.PersistentFlags().BoolVar(
		&indexFallback,
		"indexFallback",
		true,
		"fallback the request to index.html on missing static path (eg. when pretty urls are used with SPA)",
	)

	var queryTimeout int
	app.RootCmd.PersistentFlags().IntVar(
		&queryTimeout,
		"queryTimeout",
		30,
		"the default SELECT queries timeout in seconds",
	)

	app.RootCmd.ParseFlags(os.Args[1:])

	// ---------------------------------------------------------------
	// Plugins and hooks:
	// ---------------------------------------------------------------

	// load jsvm (hooks and migrations)
	jsvm.MustRegister(app, jsvm.Config{
		MigrationsDir: migrationsDir,
		HooksDir:      hooksDir,
		HooksWatch:    hooksWatch,
		HooksPoolSize: hooksPool,
	})

	// migrate command (with js templates)
	migratecmd.MustRegister(app, app.RootCmd, migratecmd.Config{
		TemplateLang: migratecmd.TemplateLangJS,
		Automigrate:  automigrate,
		Dir:          migrationsDir,
	})

	// GitHub selfupdate
	ghupdate.MustRegister(app, app.RootCmd, ghupdate.Config{})

	app.OnAfterBootstrap().PreAdd(func(e *core.BootstrapEvent) error {
		app.Dao().ModelQueryTimeout = time.Duration(queryTimeout) * time.Second
		return nil
	})

	type Wrapper struct {
		Data string
	}
	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {

		// serves static files from the provided public dir (if exists)
		e.Router.GET("/*", apis.StaticDirectoryHandler(os.DirFS(publicDir), indexFallback))

		e.Router.POST("/ask", func(c echo.Context) error {
			jsonBody := make(map[string]interface{})
			err := json.NewDecoder(c.Request().Body).Decode(&jsonBody)
			if err != nil {
				return c.JSON(http.StatusBadRequest, err)
			}
			req := jsonBody["req"].(string)
			if req == "" {
				var val []byte = []byte("{\"data\":" + "Empty Request" + "}")
				return c.JSONBlob(http.StatusBadRequest, val)
			}
			resp := loadGemini(req)
			data, err := json.Marshal(resp.Candidates[0].Content.Parts[0])
			if err != nil {
				var val []byte = []byte("{\"data\":" + err.Error() + "}")
				return c.JSONBlob(http.StatusBadRequest, val)
			}
			// fmt.Println(resp)
			jsonTemp := string(data)
			jsonString := strings.Replace(jsonTemp, "```", "", -1)
			var val []byte = []byte("{\"data\":" + jsonString + "}")
			return c.JSONBlob(http.StatusOK, []byte(val))
		}, apis.ActivityLogger(app))

		e.Router.POST("/gemini", func(c echo.Context) error {
			jsonBody := make(map[string]interface{})
			err := json.NewDecoder(c.Request().Body).Decode(&jsonBody)
			if err != nil {
				return c.JSON(http.StatusBadRequest, err)
			}
			req := jsonBody["req"].(string)
			if req == "" {
				var val []byte = []byte("{\"data\":" + "Empty Request" + "}")
				return c.JSONBlob(http.StatusBadRequest, val)
			}
			preReq := `you are a pocketbase JSON configuration generator, 
			from now on, only output JSON without any backticks ('\n','\t',..) 
			and whitespaces. Create the tables required for this system :  ` + req + ` below 280 len. 
			The structure of the json should be like below, but please replace < > 
			items with the tables or row accordingly: ` + `
		[
				{
						"id" : "<Random 15 char length Id>",
						"name": "<Table Name>",
						"type": "<If need login, use 'auth', if not 'base' >",
						"system": false,
						"schema": [
							{
								"system": false,
								"id": "<Random 8 char length Id>",
								"name": "<Row Name>",
								"type": "<text or number or bool or email or url or editor or date or select or json or file or relation>",
								"required": false,
								"presentable": false,
								"unique": false,
								"options": {
									"min": null,
									"max": null,
									"pattern": ""
								}
							},
							{
								"system": false,
								"id": "<Random 8 char length Id>",
								"name": "<Row Number>",
								"type": "<text or number or bool or email or url or editor or date or select or json or file or relation>",
								"required": false,
								"presentable": false,
								"unique": false,
								"options": {
									"min": null,
									"max": null,
									"noDecimal": false
								}
							},
							{
								"system": false,
								"id": "<Random 8 char length Id>",
								"name": "<Row Boolean>",
								"type": "<text or number or bool or email or url or editor or date or select or json or file or relation>",
								"required": false,
								"presentable": false,
								"unique": false,
								"options": {}
							},
							{
								"system": false,
								"id": "<Random 8 char length Id>",
								"name": "editor",
								"type": "<text or number or bool or email or url or editor or date or select or json or file or relation>",
								"required": false,
								"presentable": false,
								"unique": false,
								"options": {
									"convertUrls": false
								}
							},
							{
								"system": false,
								"id": "<Random 8 char length Id>",
								"name": "email",
								"type": "<text or number or bool or email or url or editor or date or select or json or file or relation>",
								"required": false,
								"presentable": false,
								"unique": false,
								"options": {
									"exceptDomains": null,
									"onlyDomains": null
								}
							},
							{
								"system": false,
								"id": "<Random 8 char length Id>",
								"name": "url",
								"type": "<text or number or bool or email or url or editor or date or select or json or file or relation>",
								"required": false,
								"presentable": false,
								"unique": false,
								"options": {
									"exceptDomains": null,
									"onlyDomains": null
								}
							},
							{
								"system": false,
								"id": "<Random 8 char length Id>",
								"name": "date",
								"type": "<text or number or bool or email or url or editor or date or select or json or file or relation>",
								"required": false,
								"presentable": false,
								"unique": false,
								"options": {
									"min": "",
									"max": ""
								}
							},
							{
								"system": false,
								"id": "<Random 8 char length Id>",
								"name": "gender",
								"type": "<text or number or bool or email or url or editor or date or select or json or file or relation>",
								"required": false,
								"presentable": false,
								"unique": false,
								"options": {
									"maxSelect": 1,
									"values": [
										"male",
										"female"
									]
								}
							},
							{
								"system": false,
								"id": "<Random 8 char length Id>",
								"name": "file",
								"type": "<text or number or bool or email or url or editor or date or select or json or file or relation>",
								"required": false,
								"presentable": false,
								"unique": false,
								"options": <if type json , use '"maxSize": 2000000' else use '{}> 
							},
							{
								"system": false,
								"id": "<Random 8 char length Id>",
								"name": "relation",
								"type": "<text or number or bool or email or url or editor or date or select or json or file or relation>",
								"required": false,
								"presentable": false,
								"unique": false,
								"options": {
									"collectionId": "<Table Id of the related table>",
									"cascadeDelete": false,
									"minSelect": null,
									"maxSelect": 1,
									"displayFields": null
								}
							},
							{
								"system": false,
								"id": ""<Random 8 char length Id>",
								"name": "jsonObj",
								"type": "<text or number or bool or email or url or editor or date or select or json or file or relation>",
								"required": false,
								"presentable": false,
								"unique": false,
								"options": <if type json , use '"maxSize": 2000000' else use '{}>
							}
						],
						"indexes": [],
						"listRule": null,
						"viewRule": null,
						"createRule": null,
						"updateRule": null,
						"deleteRule": null,
						"options": <If table type is 'auth' , use '"allowEmailAuth": true,
									"allowOAuth2Auth": true,
									"allowUsernameAuth": true,
									"exceptEmailDomains": null,
									"manageRule": null,
									"minPasswordLength": 8,
									"onlyEmailDomains": null,
									"onlyVerified": false,
									"requireEmail": false', else use '{}'>
					}
		]
. Please remove json keyword in the beginning. Please use lowercase for table name and row name. Remove all whitespace.

			`
			resp := loadGemini(preReq)
			data, err := json.Marshal(resp.Candidates[0].Content.Parts[0])
			if err != nil {
				return c.JSON(http.StatusBadRequest, err)
			}
			jsonTemp := string(data)
			file, err := os.Create("config.json")
			if err != nil {
				return c.JSON(http.StatusBadRequest, err)
			}
			jsonString := strings.Replace(jsonTemp, "```", "", -1)
			_, er := strconv.Unquote(jsonString)
			if er != nil {
				return c.JSON(http.StatusBadRequest, er)
			}
			//additional wrap json to remove escape slashes
			var val []byte = []byte("{\"data\":" + jsonString + "}")
			var wrapper Wrapper
			err = json.Unmarshal([]byte(val), &wrapper)
			fmt.Fprintln(file, wrapper.Data)
			return c.JSON(http.StatusOK, wrapper.Data)
		}, apis.ActivityLogger(app))

		return nil
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}

}
func loadGemini(req string) *genai.GenerateContentResponse {
	ctx := context.Background()
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()
	model := client.GenerativeModel("gemini-pro")
	resp, err := model.GenerateContent(ctx, genai.Text(req))
	if err != nil {
		log.Fatal(err)
	}
	return resp
}
func printResponse(resp *genai.GenerateContentResponse) {
	for _, cand := range resp.Candidates {
		if cand.Content != nil {
			for _, part := range cand.Content.Parts {
				fmt.Println(part)
			}
		}
	}
}

// the default pb_public dir location is relative to the executable
func defaultPublicDir() string {
	if strings.HasPrefix(os.Args[0], os.TempDir()) {
		// most likely ran with go run
		return "./pb_public"
	}

	return filepath.Join(os.Args[0], "../pb_public")
}
